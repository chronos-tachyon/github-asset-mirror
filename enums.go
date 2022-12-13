package main

import (
	"bytes"
	"encoding"
	"fmt"
	"strings"
)

type EnumData struct {
	GoName  string
	Name    string
	Aliases []string
}

type CompareResult int

const (
	LT CompareResult = -1
	EQ CompareResult = 0
	GT CompareResult = 1
)

var compareResultDataMap = map[CompareResult]EnumData{
	LT: {"LT", "less-than", []string{"lessthan", "lt"}},
	EQ: {"EQ", "equal-to", []string{"equalto", "equals", "equal", "eq"}},
	GT: {"GT", "greater-than", []string{"greaterthan", "gt"}},
}

func (compareResult CompareResult) Data() EnumData {
	if data, found := compareResultDataMap[compareResult]; found {
		return data
	}
	switch {
	case compareResult < 0:
		return compareResultDataMap[LT]
	case compareResult > 0:
		return compareResultDataMap[GT]
	default:
		return compareResultDataMap[EQ]
	}
}

func (compareResult CompareResult) GoString() string {
	return compareResult.Data().GoName
}

func (compareResult CompareResult) String() string {
	return compareResult.Data().Name
}

func (compareResult CompareResult) MarshalText() ([]byte, error) {
	str := compareResult.String()
	return []byte(str), nil
}

func (compareResult *CompareResult) UnmarshalText(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	str := string(raw)
	for enum, data := range compareResultDataMap {
		if str == data.GoName || strings.EqualFold(str, data.Name) {
			*compareResult = enum
			return nil
		}
		for _, alias := range data.Aliases {
			if strings.EqualFold(str, alias) {
				*compareResult = enum
				return nil
			}
		}
	}
	*compareResult = 0
	return fmt.Errorf("failed to parse %q as CompareResult", str)
}

var (
	_ fmt.GoStringer           = CompareResult(0)
	_ fmt.Stringer             = CompareResult(0)
	_ encoding.TextMarshaler   = CompareResult(0)
	_ encoding.TextUnmarshaler = (*CompareResult)(nil)
)

type AssetOS byte

const (
	UnknownAssetOS AssetOS = iota
	AnyOS
	LinuxOS
	NumAssetOSes
)

var assetOSDataArray = [NumAssetOSes]EnumData{
	{"UnknownAssetOS", "unknown", []string{""}},
	{"AnyOS", "any", nil},
	{"LinuxOS", "linux", nil},
}

func (assetOS AssetOS) Data() EnumData {
	if assetOS < NumAssetOSes {
		return assetOSDataArray[assetOS]
	}
	goName := fmt.Sprintf("AssetOS(0x%02x)", uint(assetOS))
	name := fmt.Sprintf("asset-os-%02x", uint(assetOS))
	return EnumData{goName, name, nil}
}

func (assetOS AssetOS) GoString() string {
	return assetOS.Data().GoName
}

func (assetOS AssetOS) String() string {
	return assetOS.Data().Name
}

func (assetOS AssetOS) MarshalText() ([]byte, error) {
	str := assetOS.String()
	return []byte(str), nil
}

func (assetOS *AssetOS) UnmarshalText(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	str := string(raw)
	for enum := AssetOS(0); enum < NumAssetOSes; enum++ {
		data := assetOSDataArray[enum]
		if str == data.GoName || strings.EqualFold(str, data.Name) {
			*assetOS = enum
			return nil
		}
		for _, alias := range data.Aliases {
			if strings.EqualFold(str, alias) {
				*assetOS = enum
				return nil
			}
		}
	}
	*assetOS = 0
	return fmt.Errorf("failed to parse %q as AssetOS", str)
}

var (
	_ fmt.GoStringer           = AssetOS(0)
	_ fmt.Stringer             = AssetOS(0)
	_ encoding.TextMarshaler   = AssetOS(0)
	_ encoding.TextUnmarshaler = (*AssetOS)(nil)
)

type AssetArch byte

const (
	UnknownAssetArch AssetArch = iota
	AnyArch
	AMD64Arch
	ARM64Arch
	NumAssetArches
)

var assetArchDataArray = [NumAssetArches]EnumData{
	{"UnknownAssetArch", "unknown", []string{""}},
	{"AnyArch", "any", nil},
	{"AMD64Arch", "amd64", []string{"x86-64", "x64"}},
	{"ARM64Arch", "arm64", []string{"aarch64"}},
}

func (assetArch AssetArch) Data() EnumData {
	if assetArch < NumAssetArches {
		return assetArchDataArray[assetArch]
	}
	goName := fmt.Sprintf("AssetArch(0x%02x)", uint(assetArch))
	name := fmt.Sprintf("asset-arch-%02x", uint(assetArch))
	return EnumData{goName, name, nil}
}

func (assetArch AssetArch) GoString() string {
	return assetArch.Data().GoName
}

func (assetArch AssetArch) String() string {
	return assetArch.Data().Name
}

func (assetArch AssetArch) MarshalText() ([]byte, error) {
	str := assetArch.String()
	return []byte(str), nil
}

func (assetArch *AssetArch) UnmarshalText(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	str := string(raw)
	for enum := AssetArch(0); enum < NumAssetArches; enum++ {
		data := assetArchDataArray[enum]
		if str == data.GoName || strings.EqualFold(str, data.Name) {
			*assetArch = enum
			return nil
		}
		for _, alias := range data.Aliases {
			if strings.EqualFold(str, alias) {
				*assetArch = enum
				return nil
			}
		}
	}
	*assetArch = 0
	return fmt.Errorf("failed to parse %q as AssetArch", str)
}

var (
	_ fmt.GoStringer           = AssetArch(0)
	_ fmt.Stringer             = AssetArch(0)
	_ encoding.TextMarshaler   = AssetArch(0)
	_ encoding.TextUnmarshaler = (*AssetArch)(nil)
)

type AssetType byte

const (
	UnknownAssetType AssetType = iota
	SourceTarType
	SourceZipType
	ExecutableType
	ProvenanceType
	NumAssetTypes
)

var assetTypeDataArray = [NumAssetTypes]EnumData{
	{"UnknownAssetType", "unknown", []string{""}},
	{"SourceTarType", "source-tar", []string{"sourcetar"}},
	{"SourceZipType", "source-zip", []string{"sourcezip"}},
	{"ExecutableType", "executable", []string{"binary"}},
	{"ProvenanceType", "provenance", nil},
}

func (assetType AssetType) Data() EnumData {
	if assetType < NumAssetTypes {
		return assetTypeDataArray[assetType]
	}
	goName := fmt.Sprintf("AssetType(0x%02x)", uint(assetType))
	name := fmt.Sprintf("asset-type-%02x", uint(assetType))
	return EnumData{goName, name, nil}
}

func (assetType AssetType) GoString() string {
	return assetType.Data().GoName
}

func (assetType AssetType) String() string {
	return assetType.Data().Name
}

func (assetType AssetType) MarshalText() ([]byte, error) {
	str := assetType.String()
	return []byte(str), nil
}

func (assetType *AssetType) UnmarshalText(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	str := string(raw)
	for enum := AssetType(0); enum < NumAssetTypes; enum++ {
		data := assetTypeDataArray[enum]
		if str == data.GoName || strings.EqualFold(str, data.Name) {
			*assetType = enum
			return nil
		}
		for _, alias := range data.Aliases {
			if strings.EqualFold(str, alias) {
				*assetType = enum
				return nil
			}
		}
	}
	*assetType = 0
	return fmt.Errorf("failed to parse %q as AssetType", str)
}

var (
	_ fmt.GoStringer           = AssetType(0)
	_ fmt.Stringer             = AssetType(0)
	_ encoding.TextMarshaler   = AssetType(0)
	_ encoding.TextUnmarshaler = (*AssetType)(nil)
)
