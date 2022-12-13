package indexfile

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

func (value CompareResult) Data() EnumData {
	if data, found := compareResultDataMap[value]; found {
		return data
	}
	switch {
	case value < 0:
		return compareResultDataMap[LT]
	case value > 0:
		return compareResultDataMap[GT]
	default:
		return compareResultDataMap[EQ]
	}
}

func (value CompareResult) GoString() string {
	return value.Data().GoName
}

func (value CompareResult) String() string {
	return value.Data().Name
}

func (value CompareResult) MarshalText() ([]byte, error) {
	str := value.String()
	return []byte(str), nil
}

func (value *CompareResult) UnmarshalText(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	str := string(raw)
	for enum, data := range compareResultDataMap {
		if str == data.GoName || strings.EqualFold(str, data.Name) {
			*value = enum
			return nil
		}
		for _, alias := range data.Aliases {
			if strings.EqualFold(str, alias) {
				*value = enum
				return nil
			}
		}
	}
	*value = 0
	return fmt.Errorf("failed to parse %q as CompareResult", str)
}

func (value CompareResult) CompareTo(other CompareResult) CompareResult {
	return CompareInt(value, other)
}

var (
	_ fmt.GoStringer              = CompareResult(0)
	_ fmt.Stringer                = CompareResult(0)
	_ encoding.TextMarshaler      = CompareResult(0)
	_ encoding.TextUnmarshaler    = (*CompareResult)(nil)
	_ ComparableTo[CompareResult] = CompareResult(0)
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

func (value AssetOS) Data() EnumData {
	if value < NumAssetOSes {
		return assetOSDataArray[value]
	}
	goName := fmt.Sprintf("AssetOS(0x%02x)", uint(value))
	name := fmt.Sprintf("asset-os-%02x", uint(value))
	return EnumData{goName, name, nil}
}

func (value AssetOS) GoString() string {
	return value.Data().GoName
}

func (value AssetOS) String() string {
	return value.Data().Name
}

func (value AssetOS) MarshalText() ([]byte, error) {
	str := value.String()
	return []byte(str), nil
}

func (value *AssetOS) UnmarshalText(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	str := string(raw)
	for enum := AssetOS(0); enum < NumAssetOSes; enum++ {
		data := assetOSDataArray[enum]
		if str == data.GoName || strings.EqualFold(str, data.Name) {
			*value = enum
			return nil
		}
		for _, alias := range data.Aliases {
			if strings.EqualFold(str, alias) {
				*value = enum
				return nil
			}
		}
	}
	*value = 0
	return fmt.Errorf("failed to parse %q as AssetOS", str)
}

func (value AssetOS) CompareTo(other AssetOS) CompareResult {
	return CompareByte(value, other)
}

var (
	_ fmt.GoStringer           = AssetOS(0)
	_ fmt.Stringer             = AssetOS(0)
	_ encoding.TextMarshaler   = AssetOS(0)
	_ encoding.TextUnmarshaler = (*AssetOS)(nil)
	_ ComparableTo[AssetOS]    = AssetOS(0)
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

func (value AssetArch) Data() EnumData {
	if value < NumAssetArches {
		return assetArchDataArray[value]
	}
	goName := fmt.Sprintf("AssetArch(0x%02x)", uint(value))
	name := fmt.Sprintf("asset-arch-%02x", uint(value))
	return EnumData{goName, name, nil}
}

func (value AssetArch) GoString() string {
	return value.Data().GoName
}

func (value AssetArch) String() string {
	return value.Data().Name
}

func (value AssetArch) MarshalText() ([]byte, error) {
	str := value.String()
	return []byte(str), nil
}

func (value *AssetArch) UnmarshalText(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	str := string(raw)
	for enum := AssetArch(0); enum < NumAssetArches; enum++ {
		data := assetArchDataArray[enum]
		if str == data.GoName || strings.EqualFold(str, data.Name) {
			*value = enum
			return nil
		}
		for _, alias := range data.Aliases {
			if strings.EqualFold(str, alias) {
				*value = enum
				return nil
			}
		}
	}
	*value = 0
	return fmt.Errorf("failed to parse %q as AssetArch", str)
}

func (value AssetArch) CompareTo(other AssetArch) CompareResult {
	return CompareByte(value, other)
}

var (
	_ fmt.GoStringer           = AssetArch(0)
	_ fmt.Stringer             = AssetArch(0)
	_ encoding.TextMarshaler   = AssetArch(0)
	_ encoding.TextUnmarshaler = (*AssetArch)(nil)
	_ ComparableTo[AssetArch]  = AssetArch(0)
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

func (value AssetType) Data() EnumData {
	if value < NumAssetTypes {
		return assetTypeDataArray[value]
	}
	goName := fmt.Sprintf("AssetType(0x%02x)", uint(value))
	name := fmt.Sprintf("asset-type-%02x", uint(value))
	return EnumData{goName, name, nil}
}

func (value AssetType) GoString() string {
	return value.Data().GoName
}

func (value AssetType) String() string {
	return value.Data().Name
}

func (value AssetType) MarshalText() ([]byte, error) {
	str := value.String()
	return []byte(str), nil
}

func (value *AssetType) UnmarshalText(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	str := string(raw)
	for enum := AssetType(0); enum < NumAssetTypes; enum++ {
		data := assetTypeDataArray[enum]
		if str == data.GoName || strings.EqualFold(str, data.Name) {
			*value = enum
			return nil
		}
		for _, alias := range data.Aliases {
			if strings.EqualFold(str, alias) {
				*value = enum
				return nil
			}
		}
	}
	*value = 0
	return fmt.Errorf("failed to parse %q as AssetType", str)
}

func (value AssetType) CompareTo(other AssetType) CompareResult {
	return CompareByte(value, other)
}

var (
	_ fmt.GoStringer           = AssetType(0)
	_ fmt.Stringer             = AssetType(0)
	_ encoding.TextMarshaler   = AssetType(0)
	_ encoding.TextUnmarshaler = (*AssetType)(nil)
	_ ComparableTo[AssetType]  = AssetType(0)
)

type VersionElementType byte

const (
	UnknownVET VersionElementType = iota
	SymbolsVET
	DigitsVET
	LettersVET
	NumVersionElementTypes
)

var versionElementTypeDataArray = [NumVersionElementTypes]EnumData{
	{"UnknownVET", "unknown", []string{""}},
	{"SymbolsVET", "symbols", nil},
	{"DigitsVET", "digits", nil},
	{"LettersVET", "letters", nil},
}

func (value VersionElementType) Data() EnumData {
	if value < NumVersionElementTypes {
		return versionElementTypeDataArray[value]
	}
	goName := fmt.Sprintf("VersionElementType(0x%02x)", uint(value))
	name := fmt.Sprintf("version-element-type-%02x", uint(value))
	return EnumData{goName, name, nil}
}

func (value VersionElementType) GoString() string {
	return value.Data().GoName
}

func (value VersionElementType) String() string {
	return value.Data().Name
}

func (value VersionElementType) MarshalText() ([]byte, error) {
	str := value.String()
	return []byte(str), nil
}

func (value *VersionElementType) UnmarshalText(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	str := string(raw)
	for enum := VersionElementType(0); enum < NumVersionElementTypes; enum++ {
		data := versionElementTypeDataArray[enum]
		if str == data.GoName || strings.EqualFold(str, data.Name) {
			*value = enum
			return nil
		}
		for _, alias := range data.Aliases {
			if strings.EqualFold(str, alias) {
				*value = enum
				return nil
			}
		}
	}
	*value = 0
	return fmt.Errorf("failed to parse %q as VersionElementType", str)
}

func (value VersionElementType) CompareTo(other VersionElementType) CompareResult {
	return CompareByte(value, other)
}

var (
	_ fmt.GoStringer                   = VersionElementType(0)
	_ fmt.Stringer                     = VersionElementType(0)
	_ encoding.TextMarshaler           = VersionElementType(0)
	_ encoding.TextUnmarshaler         = (*VersionElementType)(nil)
	_ ComparableTo[VersionElementType] = VersionElementType(0)
)
