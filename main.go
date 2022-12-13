package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-github/v48/github"
	"github.com/pborman/getopt/v2"
)

var (
	AppVersion string = "devel"
)

const UserAgentFormat = "github-asset-mirror/%s (+https://github.com/chronos-tachyon/github-asset-mirror)"

const ReleasesPerPage = 10
const AssetsPerPage = 10

type MyRoundTripper struct {
	Next  http.RoundTripper
	Token string
}

func UserAgent() string {
	return fmt.Sprintf(UserAgentFormat, AppVersion)
}

func (rt *MyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	header := make(http.Header, 4+len(req.Header))
	header.Set("user-agent", UserAgent())
	header.Set("authorization", "Bearer "+rt.Token)
	for key, values := range req.Header {
		header[key] = values
	}

	req = req.WithContext(req.Context())
	req.Header = header
	return rt.Next.RoundTrip(req)
}

func main() {
	var tokenFile string
	var ghOwner string
	var ghRepo string
	var outputDir string

	getopt.FlagLong(&tokenFile, "token-file", 'T', "path to file containing your GitHub token")
	getopt.FlagLong(&ghOwner, "github-owner", 'O', "name of GitHub repository's owner user or owner organization")
	getopt.FlagLong(&ghRepo, "github-repo", 'R', "name of GitHub repository")
	getopt.FlagLong(&outputDir, "output-dir", 'd', "path to the output directory")
	getopt.Parse()

	if tokenFile == "" {
		panic(fmt.Errorf("missing required flag -T / --token-file"))
	}
	if ghOwner == "" {
		panic(fmt.Errorf("missing required flag -O / --github-owner"))
	}
	if ghRepo == "" {
		panic(fmt.Errorf("missing required flag -R / --github-repo"))
	}
	if outputDir == "" {
		panic(fmt.Errorf("missing required flag -d / --output-dir"))
	}

	raw, err := os.ReadFile(tokenFile)
	if err != nil {
		panic(fmt.Errorf("failed to read token file: %q: %w", tokenFile, err))
	}
	raw = bytes.TrimSpace(raw)
	accessToken := string(raw)

	var rt http.RoundTripper = http.DefaultClient.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	rt = &MyRoundTripper{Next: rt, Token: accessToken}
	http.DefaultClient.Transport = rt

	var releaseTags []string
	var releasesByTag map[string]Release

	releaseDataPath := filepath.Join(outputDir, "data.json")
	raw, err = os.ReadFile(releaseDataPath)
	switch {
	case err == nil:
		d := json.NewDecoder(bytes.NewReader(raw))
		d.UseNumber()
		d.DisallowUnknownFields()
		err = d.Decode(&releasesByTag)
		if err != nil {
			panic(fmt.Errorf("failed to decode contents of JSON data file: %q: %w", releaseDataPath, err))
		}

	case errors.Is(err, fs.ErrNotExist):
		releasesByTag = make(map[string]Release, 128)

	default:
		panic(fmt.Errorf("failed to read JSON data file: %q: %w", releaseDataPath, err))
	}

	n := uint(len(releasesByTag))
	n += 64
	if n < 128 {
		n = 128
	}
	releaseTags = make([]string, 0, n)

	ctx := context.Background()
	client := github.NewClient(http.DefaultClient)

	err = Iterate(
		ReleasesPerPage,
		func(options *github.ListOptions) (list []*github.RepositoryRelease, resp *github.Response, err error) {
			list, resp, err = client.Repositories.ListReleases(ctx, ghOwner, ghRepo, options)
			if err != nil {
				list = nil
				resp = nil
				err = fmt.Errorf("failed to list releases for %s/%s: %w", ghOwner, ghRepo, err)
			}
			return
		},
		func(ghr *github.RepositoryRelease) error {
			if *ghr.Draft {
				return nil
			}

			id := *ghr.ID
			tag := *ghr.TagName
			release, found := releasesByTag[tag]

			release.ID = id
			release.Name = *ghr.Name
			release.Body = *ghr.Body

			if !found {
				release.Tag = tag
				err := release.Version.Parse(tag)
				if err != nil {
					return err
				}
			}

			release.Assets = make([]Asset, 2, 16)
			release.Assets[0] = Asset{
				URL:  *ghr.TarballURL,
				Name: "source.tar.gz",
				Type: SourceTarType,
				OS:   AnyOS,
				Arch: AnyArch,
			}
			release.Assets[1] = Asset{
				URL:  *ghr.ZipballURL,
				Name: "source.zip",
				Type: SourceZipType,
				OS:   AnyOS,
				Arch: AnyArch,
			}

			err := Iterate(
				AssetsPerPage,
				func(options *github.ListOptions) (list []*github.ReleaseAsset, resp *github.Response, err error) {
					list, resp, err = client.Repositories.ListReleaseAssets(ctx, ghOwner, ghRepo, id, options)
					if err != nil {
						err = fmt.Errorf("failed to list assets for %s/%s release %d: %w", ghOwner, ghRepo, id, err)
					}
					return
				},
				func(gha *github.ReleaseAsset) error {
					assetID := gha.GetID()
					assetURL := gha.GetBrowserDownloadURL()
					assetName := gha.GetName()
					release.Assets = append(release.Assets, MakeAsset(assetID, assetURL, assetName))
					return nil
				},
			)
			if err != nil {
				return err
			}

			sortList(release.Assets)
			releaseTags = append(releaseTags, tag)
			releasesByTag[tag] = release
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	for _, tag := range releaseTags {
		release := releasesByTag[tag]
		fmt.Printf("%q\n", tag)
		for _, asset := range release.Assets {
			assetPath := filepath.Join(outputDir, tag, asset.Name)

			_, err := os.Stat(assetPath)
			if err == nil {
				continue
			}
			if !errors.Is(err, fs.ErrNotExist) {
				panic(fmt.Errorf("os.Stat: %q: %w", assetPath, err))
			}

			fmt.Printf("\t%q ← %q\n", assetPath, asset.URL)

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, http.NoBody)
			if err != nil {
				panic(err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}

			if resp.StatusCode != http.StatusOK {
				_ = resp.Body.Close()
				panic(fmt.Errorf("unexpected HTTP status code %03d", resp.StatusCode))
			}

			_, err = writeFileToDisk(assetPath, resp.Body, asset.Mode())
			if err2 := resp.Body.Close(); err == nil {
				err = err2
			}
			if err != nil {
				panic(err)
			}
		}
	}

	for _, tag := range releaseTags {
		release := releasesByTag[tag]
		releaseDir := filepath.Join(outputDir, tag)
		for _, asset := range release.Assets {
			if asset.Type == ExecutableType && release.Version.BuildID == "" {
				if buildID, ok := asset.ExtractBuildID(ctx, releaseDir); ok {
					release.Version.BuildID = buildID
					releasesByTag[tag] = release
				}
			}
		}
	}

	var buf bytes.Buffer
	e := json.NewEncoder(&buf)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")
	err = e.Encode(releasesByTag)
	if err != nil {
		panic(fmt.Errorf("failed to encode index data to JSON: %w", err))
	}
	raw = buf.Bytes()

	_, err = writeFileToDisk(releaseDataPath, bytes.NewReader(raw), 0o666)
	if err != nil {
		panic(fmt.Errorf("failed to write index data to file: %w", err))
	}
}

type CallFunc[T any] func(*github.ListOptions) ([]*T, *github.Response, error)

type ProcessFunc[T any] func(*T) error

func Iterate[T any](pageSize int, callFn CallFunc[T], processFn ProcessFunc[T]) error {
	var options github.ListOptions
	options.Page = 0
	options.PerPage = pageSize
	for {
		list, resp, err := callFn(&options)
		if err != nil {
			return err
		}
		for _, item := range list {
			err = processFn(item)
			if err != nil {
				return err
			}
		}
		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}
	return nil
}

func writeFileToDisk(fileName string, r io.Reader, mode fs.FileMode) (n int64, err error) {
	fileDir := filepath.Dir(fileName)
	fileBase := filepath.Base(fileName)
	fileTemp := filepath.Join(fileDir, ".tmp."+fileBase+"~")

	err = os.MkdirAll(fileDir, 0o777)
	if err != nil {
		return n, fmt.Errorf("os.MkdirAll: %q: %w", fileDir, err)
	}

	dir, err := os.OpenFile(fileDir, os.O_RDONLY, 0)
	if err != nil {
		return n, fmt.Errorf("os.OpenFile: %q, O_RDONLY: %w", fileDir, err)
	}

	needDirClose := true
	defer func() {
		if needDirClose {
			_ = dir.Close()
		}
	}()

	_ = os.Remove(fileTemp)

	file, err := os.OpenFile(fileTemp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return n, fmt.Errorf("os.OpenFile: %q, O_WRONLY|O_CREATE|O_EXCL: %w", fileTemp, err)
	}

	needFileClose := true
	needFileRemove := true
	defer func() {
		if needFileClose {
			_ = file.Close()
		}
		if needFileRemove {
			_ = os.Remove(fileTemp)
		}
	}()

	n, err = io.Copy(file, r)
	if err != nil {
		return n, fmt.Errorf("io.Copy: %q: %w", fileTemp, err)
	}

	err = file.Sync()
	if err != nil {
		return n, fmt.Errorf("os.File.Sync: %q: %w", fileTemp, err)
	}

	needFileClose = false
	err = file.Close()
	if err != nil {
		return n, fmt.Errorf("os.File.Close: %q: %w", fileTemp, err)
	}

	err = os.Rename(fileTemp, fileName)
	if err != nil {
		return n, fmt.Errorf("os.Rename: %q, %q: %w", fileTemp, fileName, err)
	}

	needFileRemove = false
	err = dir.Sync()
	if err != nil {
		return n, fmt.Errorf("os.File.Sync: %q: %w", fileDir, err)
	}

	needDirClose = false
	err = dir.Close()
	if err != nil {
		return n, fmt.Errorf("os.File.Close: %q: %w", fileDir, err)
	}

	return n, nil
}
