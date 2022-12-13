package indexfile

import (
	"bytes"
	"context"
	"io/fs"
	"os/exec"
	"path/filepath"
	"regexp"
)

var (
	reAssetExecutable = regexp.MustCompile(`^([0-9A-Za-z]+(?:[_-][0-9A-Za-z]+)*)-(linux)-(amd64|arm64)(?:\.exe)?$`)
	reAssetProvenance = regexp.MustCompile(`^([0-9A-Za-z]+(?:[_-][0-9A-Za-z]+)*)-(linux)-(amd64|arm64)\.intoto\.jsonl?$`)
	reBuildID         = regexp.MustCompile(`^\tbuild\tvcs\.revision=([0-9a-f]{40})\s*$`)
)

var osMap = map[string]AssetOS{
	"any":   AnyOS,
	"linux": LinuxOS,
}

var archMap = map[string]AssetArch{
	"any":   AnyArch,
	"amd64": AMD64Arch,
	"arm64": ARM64Arch,
}

type Asset struct {
	ID   int64     `json:"id,omitempty"`
	URL  string    `json:"url"`
	Name string    `json:"name"`
	Base string    `json:"base,omitempty"`
	OS   AssetOS   `json:"os"`
	Arch AssetArch `json:"arch"`
	Type AssetType `json:"type"`
}

func MakeSourceTarballAsset(assetURL string) Asset {
	return Asset{
		URL:  assetURL,
		Name: "source.tar.gz",
		OS:   AnyOS,
		Arch: AnyArch,
		Type: SourceTarType,
	}
}

func MakeSourceZipballAsset(assetURL string) Asset {
	return Asset{
		URL:  assetURL,
		Name: "source.zip",
		OS:   AnyOS,
		Arch: AnyArch,
		Type: SourceZipType,
	}
}

func MakeAsset(assetID int64, assetURL string, assetName string) Asset {
	assetBase := ""
	assetOS := UnknownAssetOS
	assetArch := UnknownAssetArch
	assetType := UnknownAssetType

	match := reAssetExecutable.FindStringSubmatch(assetName)
	if match != nil {
		assetBase = match[1]
		assetOS = osMap[match[2]]
		assetArch = archMap[match[3]]
		assetType = ExecutableType
	}

	match = reAssetProvenance.FindStringSubmatch(assetName)
	if match != nil {
		assetBase = match[1]
		assetOS = osMap[match[2]]
		assetArch = archMap[match[3]]
		assetType = ProvenanceType
	}

	return Asset{
		ID:   assetID,
		URL:  assetURL,
		Name: assetName,
		Base: assetBase,
		OS:   assetOS,
		Arch: assetArch,
		Type: assetType,
	}
}

func (a Asset) Mode() fs.FileMode {
	switch a.Type {
	case ExecutableType:
		return 0o777
	default:
		return 0o666
	}
}

func (a Asset) ExtractBuildID(ctx context.Context, releaseDir string) (string, bool) {
	assetPath := filepath.Join(releaseDir, a.Name)
	cmd := exec.CommandContext(ctx, "go", "version", "-m", assetPath)
	raw, err := cmd.Output()
	if err == nil {
		for _, line := range bytes.Split(raw, []byte{'\n'}) {
			match := reBuildID.FindStringSubmatch(string(line))
			if match != nil {
				return match[1], true
			}
		}
	}
	return "", false
}

func (a Asset) CompareTo(other Asset) CompareResult {
	cmp := a.OS.CompareTo(other.OS)
	if cmp == EQ {
		cmp = a.Arch.CompareTo(other.Arch)
	}
	if cmp == EQ {
		cmp = a.Type.CompareTo(other.Type)
	}
	if cmp == EQ {
		cmp = CompareString(a.Name, other.Name)
	}
	if cmp == EQ {
		cmp = CompareInt64(a.ID, other.ID)
	}
	if cmp == EQ {
		cmp = CompareString(a.URL, other.URL)
	}
	if cmp == EQ {
		cmp = CompareString(a.Base, other.Base)
	}
	return cmp
}

var (
	_ ComparableTo[Asset] = Asset{}
)
