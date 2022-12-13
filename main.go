package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-github/v48/github"
	"github.com/pborman/getopt/v2"
	"github.com/rs/zerolog"

	"github.com/chronos-tachyon/github-asset-mirror/indexfile"
	"github.com/chronos-tachyon/github-asset-mirror/indexutil"
	"github.com/chronos-tachyon/github-asset-mirror/logging"
)

var (
	AppVersion string = "devel"
)

const (
	UserAgentFormat = "github-asset-mirror/%s (+https://github.com/chronos-tachyon/github-asset-mirror)"
	IndexFileName   = "index.json"
	ReleasesPerPage = 10
	AssetsPerPage   = 10
)

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
	logging.Init()
	defer logging.Done()

	ctx := context.Background()
	logger := zerolog.Ctx(ctx)

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
		logger.Fatal().Msg("missing required flag -T / --token-file")
	}
	if ghOwner == "" {
		logger.Fatal().Msg("missing required flag -O / --github-owner")
	}
	if ghRepo == "" {
		logger.Fatal().Msg("missing required flag -R / --github-repo")
	}
	if outputDir == "" {
		logger.Fatal().Msg("missing required flag -d / --output-dir")
	}

	raw, err := os.ReadFile(tokenFile)
	if err != nil {
		logger.Fatal().
			Str("tokenFile", tokenFile).
			Err(err).
			Msg("failed to read GitHub access token from file")
		panic(nil)
	}
	raw = bytes.TrimSpace(raw)
	accessToken := string(raw)

	releases := make([]indexfile.Release, 0, 256)

	releaseDataPath := filepath.Join(outputDir, IndexFileName)
	indexLogger := logger.With().
		Str("path", releaseDataPath).
		Logger()

	raw, err = os.ReadFile(releaseDataPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		indexLogger.Fatal().
			Str("path", releaseDataPath).
			Err(err).
			Msg("failed to read contents of JSON index file")
		panic(nil)
	}
	if err == nil {
		ctx2 := indexLogger.WithContext(ctx)
		indexutil.FromJSON(ctx2, &releases, raw)
	}

	releaseIndexByTag := make(map[string]uint, 256)
	for index, release := range releases {
		releaseIndexByTag[release.Tag] = uint(index)
	}

	var rt http.RoundTripper = http.DefaultClient.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	rt = &MyRoundTripper{Next: rt, Token: accessToken}
	http.DefaultClient.Transport = rt
	client := github.NewClient(http.DefaultClient)

	ghLogger := logger.With().
		Str("githubOwner", ghOwner).
		Str("githubRepo", ghRepo).
		Logger()

	Iterate(
		ReleasesPerPage,
		func(options *github.ListOptions) ([]*github.RepositoryRelease, *github.Response) {
			list, resp, err := client.Repositories.ListReleases(ctx, ghOwner, ghRepo, options)
			if err == nil {
				return list, resp
			}
			ghLogger.Fatal().
				Int("pageNumber", options.Page).
				Err(err).
				Msg("failed to list GitHub releases")
			panic(nil)
		},
		func(ghr *github.RepositoryRelease) {
			if ghr.GetDraft() {
				return
			}

			id := ghr.GetID()
			tag := ghr.GetTagName()

			ghrLogger := ghLogger.With().
				Int64("releaseID", id).
				Str("releaseTag", tag).
				Logger()

			var release indexfile.Release
			releaseIndex, found := releaseIndexByTag[tag]
			switch {
			case found:
				release = releases[releaseIndex]
			default:
				release.Tag = tag
				if !release.Version.Parse(tag) {
					ghrLogger.Error().
						Msg("failed to parse GitHub release tag as a semantic version")
					return
				}
			}

			release.ID = id
			release.Name = ghr.GetName()
			release.Body = ghr.GetBody()
			release.Assets = make([]indexfile.Asset, 2, 16)
			release.Assets[0] = indexfile.MakeSourceTarballAsset(ghr.GetTarballURL())
			release.Assets[1] = indexfile.MakeSourceZipballAsset(ghr.GetZipballURL())

			Iterate(
				AssetsPerPage,
				func(options *github.ListOptions) ([]*github.ReleaseAsset, *github.Response) {
					list, resp, err := client.Repositories.ListReleaseAssets(ctx, ghOwner, ghRepo, id, options)
					if err == nil {
						return list, resp
					}
					ghrLogger.Fatal().
						Int("pageNumber", options.Page).
						Err(err).
						Msg("failed to list assets for GitHub release")
					panic(nil)
				},
				func(gha *github.ReleaseAsset) {
					release.Assets = append(
						release.Assets,
						indexfile.MakeAsset(
							gha.GetID(),
							gha.GetBrowserDownloadURL(),
							gha.GetName(),
						),
					)
				},
			)

			type AssetList = indexfile.SortableList[indexfile.Asset]
			AssetList(release.Assets).Sort()

			switch {
			case found:
				releases[releaseIndex] = release
			default:
				releaseIndex = uint(len(releases))
				releases = append(releases, release)
				releaseIndexByTag[tag] = releaseIndex
			}
		},
	)

	type ReleaseList = indexfile.SortableList[indexfile.Release]
	ReleaseList(releases).Sort()

	for _, release := range releases {
		for _, asset := range release.Assets {
			assetPath := filepath.Join(outputDir, release.Tag, asset.Name)

			assetLogger := logger.With().
				Int64("releaseID", release.ID).
				Str("releaseTag", release.Tag).
				Int64("assetID", asset.ID).
				Str("assetURL", asset.URL).
				Str("assetPath", assetPath).
				Logger()

			_, err := os.Stat(assetPath)
			if err == nil {
				continue
			}
			if !errors.Is(err, fs.ErrNotExist) {
				assetLogger.Fatal().
					Err(err).
					Msg("failed to stat file containing downloaded asset")
				panic(nil)
			}

			assetLogger.Info().
				Msg("downloading asset to local file")

			reqURL := asset.URL
			reqMethod := http.MethodGet

			req, err := http.NewRequestWithContext(ctx, reqMethod, reqURL, http.NoBody)
			if err != nil {
				assetLogger.Fatal().
					Err(err).
					Msg("failed to create HTTP request object")
				panic(nil)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				assetLogger.Fatal().
					Err(err).
					Msg("HTTP request failed")
				panic(nil)
			}

			assetLogger = assetLogger.With().
				Int("statusCode", resp.StatusCode).
				Logger()

			if resp.StatusCode != http.StatusOK {
				_ = resp.Body.Close()
				assetLogger.Fatal().
					Msg("unexpected HTTP status code")
				panic(nil)
			}

			raw, err := io.ReadAll(resp.Body)
			if err2 := resp.Body.Close(); err == nil {
				err = err2
			}
			if err != nil {
				assetLogger.Fatal().
					Err(err).
					Msg("I/O error while reading HTTP response body")
				panic(nil)
			}

			ctx2 := assetLogger.WithContext(ctx)
			indexutil.WriteFile(ctx2, assetPath, raw, asset.Mode())
		}
	}

	for _, release := range releases {
		releaseDir := filepath.Join(outputDir, release.Tag)
		for _, asset := range release.Assets {
			if asset.Type == indexfile.ExecutableType && release.Version.BuildID == "" {
				if buildID, ok := asset.ExtractBuildID(ctx, releaseDir); ok {
					release.Version.BuildID = buildID
				}
			}
		}
	}

	ctx2 := indexLogger.WithContext(ctx)
	raw = indexutil.ToJSON(ctx2, releases)
	indexutil.WriteFile(ctx2, releaseDataPath, raw, 0o666)
	if err != nil {
		indexLogger.Fatal().
			Err(err).
			Msgf("failed to write contents of new JSON index file")
		panic(nil)
	}
}
