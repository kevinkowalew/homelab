package poller

import (
	"fmt"
	"strconv"
	"strings"
)

type version struct {
	major, minor, patch int
}

func newVersion(s string) (*version, error) {
	if !strings.HasPrefix(s, "v") {
		return nil, fmt.Errorf("unexpected version tag formatting: %s", s)
	}

	s = s[1:]

	p := strings.Split(s, ".")
	if len(p) != 3 {
		return nil, fmt.Errorf("unexpected version tag formatting: %s", s)
	}

	major, err := strconv.Atoi(p[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse major version tag: %w", err)
	}

	minor, err := strconv.Atoi(p[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse minor version tag: %w", err)
	}

	pp := strings.Split(p[2], "-")
	patch, err := strconv.Atoi(pp[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse patch version tag: %w", err)
	}

	if len(pp) != 1 && len(pp) != 3 {
		return nil, fmt.Errorf("unexpected version tag formatting: %s", s)
	}

	if len(pp) == 1 {
		return &version{major, minor, patch}, nil
	}

	return &version{major, minor, patch}, nil
}

func (v version) Increment(major, minor, patch int) version {
	return version{
		major: v.major + major,
		minor: v.minor + minor,
		patch: v.patch + patch,
	}
}

func (v version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

func (v version) Less(o version) bool {
	if v.major < o.major {
		return true
	} else if v.minor < o.minor {
		return true
	} else if v.patch < o.patch {
		return true
	}

	return false
}
