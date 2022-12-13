package main

import (
	"bytes"
	"context"
	"io/fs"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
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

func (asset Asset) Mode() fs.FileMode {
	switch asset.Type {
	case ExecutableType:
		return 0o777
	default:
		return 0o666
	}
}

func (asset Asset) ExtractBuildID(ctx context.Context, releaseDir string) (string, bool) {
	assetPath := filepath.Join(releaseDir, asset.Name)
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

func (asset Asset) CompareTo(other Asset) CompareResult {
	return compareReduce(
		compareByte(asset.OS, other.OS),
		compareByte(asset.Arch, other.Arch),
		compareByte(asset.Type, other.Type),
		compareString(asset.Name, other.Name),
		compareInt64(asset.ID, other.ID),
		compareString(asset.URL, other.URL),
		compareString(asset.Base, other.Base),
	)
}

type sortable[T any] interface {
	CompareTo(T) CompareResult
}

type sortableList[T sortable[T]] []T

func (list sortableList[T]) Len() int {
	return len(list)
}

func (list sortableList[T]) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list sortableList[T]) Less(i, j int) bool {
	return list[i].CompareTo(list[j]) < 0
}

func sortList[T sortable[T]](list []T) {
	sort.Sort(sortableList[T](list))
}

func compareByte[T ~byte](a T, b T) CompareResult {
	switch {
	case a == b:
		return EQ
	case a < b:
		return LT
	default:
		return GT
	}
}

func compareInt64[T ~int64](a T, b T) CompareResult {
	switch {
	case a == b:
		return EQ
	case a < b:
		return LT
	default:
		return GT
	}
}

func compareString[T ~string](a T, b T) CompareResult {
	switch {
	case a == b:
		return EQ
	case a < b:
		return LT
	default:
		return GT
	}
}

func compareReduce(list ...CompareResult) CompareResult {
	for _, item := range list {
		if item != EQ {
			return item
		}
	}
	return EQ
}
