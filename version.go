package main

import (
	"fmt"
	"regexp"
	"strconv"
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

func (v *Version) Parse(str string) error {
	match := reTag.FindStringSubmatch(str)
	if match == nil {
		return fmt.Errorf("failed to parse %q as semantic version", str)
	}

	major, err := strconv.ParseUint(match[1], 10, 0)
	if err != nil {
		return fmt.Errorf("failed to parse major version number %q as uint: %w", match[1], err)
	}

	minor, err := strconv.ParseUint(match[2], 10, 0)
	if err != nil {
		return fmt.Errorf("failed to parse minor version number %q as uint: %w", match[2], err)
	}

	patch, err := strconv.ParseUint(match[3], 10, 0)
	if err != nil {
		return fmt.Errorf("failed to parse patch version number %q as uint: %w", match[3], err)
	}

	*v = Version{
		Major:      uint(major),
		Minor:      uint(minor),
		Patch:      uint(patch),
		Prerelease: match[4],
	}
	return nil
}
