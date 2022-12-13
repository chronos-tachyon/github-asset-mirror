package indexfile

import (
	"bytes"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"sync"
	"unicode"
)

var reTag = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-((?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?$`)

type Version struct {
	Major      uint   `json:"major"`
	Minor      uint   `json:"minor"`
	Patch      uint   `json:"patch"`
	Prerelease string `json:"prerelease,omitempty"`
	BuildID    string `json:"buildID,omitempty"`
}

func (v Version) GoAppendTo(out []byte) []byte {
	out = append(out, "Version{Major:"...)
	out = strconv.AppendUint(out, uint64(v.Major), 10)
	out = append(out, ", Minor:"...)
	out = strconv.AppendUint(out, uint64(v.Minor), 10)
	out = append(out, ", Patch:"...)
	out = strconv.AppendUint(out, uint64(v.Patch), 10)
	out = append(out, ", Prerelease:"...)
	out = strconv.AppendQuote(out, v.Prerelease)
	out = append(out, ", BuildID:"...)
	out = strconv.AppendQuote(out, v.BuildID)
	out = append(out, "}"...)
	return out
}

func (v Version) AppendTo(out []byte) []byte {
	out = strconv.AppendUint(out, uint64(v.Major), 10)
	out = append(out, '.')
	out = strconv.AppendUint(out, uint64(v.Minor), 10)
	out = append(out, '.')
	out = strconv.AppendUint(out, uint64(v.Patch), 10)
	if v.Prerelease != "" {
		out = append(out, '-')
		out = append(out, v.Prerelease...)
	}
	if v.BuildID != "" {
		out = append(out, '+')
		out = append(out, v.BuildID...)
	}
	return out
}

func (v Version) GoString() string {
	var tmp [64]byte
	return string(v.GoAppendTo(tmp[:0]))
}

func (v Version) String() string {
	var tmp [64]byte
	return string(v.AppendTo(tmp[:0]))
}

func (v *Version) Parse(str string) bool {
	match := reTag.FindStringSubmatch(str)
	if match == nil {
		return false
	}

	major, err := strconv.ParseUint(match[1], 10, 0)
	if err != nil {
		return false
	}

	minor, err := strconv.ParseUint(match[2], 10, 0)
	if err != nil {
		return false
	}

	patch, err := strconv.ParseUint(match[3], 10, 0)
	if err != nil {
		return false
	}

	*v = Version{
		Major:      uint(major),
		Minor:      uint(minor),
		Patch:      uint(patch),
		Prerelease: match[4],
	}
	return true
}

func (v Version) CompareTo(other Version) CompareResult {
	cmp := CompareUint(v.Major, other.Major)
	if cmp == EQ {
		cmp = CompareUint(v.Minor, other.Minor)
	}
	if cmp == EQ {
		cmp = CompareUint(v.Patch, other.Patch)
	}
	if cmp == EQ {
		aVSL := ParseVersionString(v.Prerelease)
		bVSL := ParseVersionString(other.Prerelease)
		cmp = aVSL.CompareTo(bVSL)
	}
	if cmp == EQ {
		cmp = CompareString(v.BuildID, other.BuildID)
	}
	return cmp
}

var (
	_ fmt.GoStringer        = Version{}
	_ fmt.Stringer          = Version{}
	_ ComparableTo[Version] = Version{}
)

type VersionElement struct {
	Type     VersionElementType
	IntValue *big.Int
	StrValue string
}

func (elem VersionElement) CompareTo(other VersionElement) CompareResult {
	cmp := elem.Type.CompareTo(other.Type)
	if cmp == EQ {
		switch elem.Type {
		case DigitsVET:
			cmp = CompareBigInt(elem.IntValue, other.IntValue)
		default:
			cmp = CompareString(elem.StrValue, other.StrValue)
		}
	}
	return cmp
}

var (
	_ ComparableTo[VersionElement] = VersionElement{}
)

type VersionElementList []VersionElement

func (list VersionElementList) CompareTo(other VersionElementList) CompareResult {
	aList := list
	bList := other

	aLen := uint(len(aList))
	bLen := uint(len(bList))

	minLen := aLen
	if minLen > bLen {
		minLen = bLen
	}

	for index := uint(0); index < minLen; index++ {
		aElem := aList[index]
		bElem := bList[index]
		cmp := aElem.CompareTo(bElem)
		if cmp != EQ {
			return cmp
		}
	}
	return CompareUint(aLen, bLen)
}

var _ ComparableTo[VersionElementList] = VersionElementList(nil)

var (
	gCacheMutex sync.Mutex
	gCacheMap   map[string]VersionElementList
)

func ParseVersionString(str string) VersionElementList {
	gCacheMutex.Lock()
	defer gCacheMutex.Unlock()

	out, found := gCacheMap[str]
	if !found {
		out = ParseVersionStringRaw(str)
		gCacheMap[str] = out
	}
	return out
}

func ParseVersionStrings(inList []string) []VersionElementList {
	inLen := uint(len(inList))
	if inLen == 0 {
		return nil
	}

	gCacheMutex.Lock()
	defer gCacheMutex.Unlock()

	outList := make([]VersionElementList, inLen)
	for index := uint(0); index < inLen; index++ {
		str := inList[index]
		out, found := gCacheMap[str]
		if !found {
			out = ParseVersionStringRaw(str)
			gCacheMap[str] = out
		}
		outList[index] = out
	}
	return outList
}

func ParseVersionStringRaw(str string) VersionElementList {
	type state byte
	const (
		initState state = iota
		digitState
		letterState
		symbolState
	)

	var out VersionElementList
	var buf bytes.Buffer
	buf.Grow(len(str))
	s := initState

	flush := func() {
		tmp := buf.String()
		buf.Reset()

		switch s {
		case digitState:
			bi, ok := new(big.Int).SetString(tmp, 10)
			if !ok {
				panic(fmt.Errorf("failed to parse %q as *big.Int", tmp))
			}
			out = append(out, VersionElement{Type: DigitsVET, IntValue: bi})
		case letterState:
			out = append(out, VersionElement{Type: LettersVET, StrValue: tmp})
		case symbolState:
			out = append(out, VersionElement{Type: SymbolsVET, StrValue: tmp})
		}

		s = initState
	}

	for _, ch := range str {
		var goal state
		switch {
		case unicode.IsDigit(ch):
			goal = digitState
		case unicode.IsLetter(ch):
			goal = letterState
		default:
			goal = symbolState
		}

		if s != goal {
			flush()
			s = goal
		}

		buf.WriteRune(ch)
	}
	flush()
	return out
}
