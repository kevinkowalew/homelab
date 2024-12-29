package poller

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Version struct {
	major, minor, patch   int
	prNumber, buildNumber int
}

func NewVersion(s string) (*Version, error) {
	if !strings.HasPrefix(s, "v") {
		return nil, fmt.Errorf("unexpected Version tag formatting: %s", s)
	}

	s = s[1:]

	p := strings.Split(s, ".")
	if len(p) != 3 {
		return nil, fmt.Errorf("unexpected Version tag formatting: %s", s)
	}

	major, err := strconv.Atoi(p[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse major Version tag: %w", err)
	}

	minor, err := strconv.Atoi(p[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse minor Version tag: %w", err)
	}

	pp := strings.Split(p[2], "-")
	patch, err := strconv.Atoi(pp[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse patch Version tag: %w", err)
	}

	if len(pp) != 1 && len(pp) != 5 {
		return nil, fmt.Errorf("unexpected Version tag formatting: %s", s)
	}

	if len(pp) == 1 {
		return &Version{major, minor, patch, 0, 0}, nil
	}

	prNumber, err := strconv.Atoi(pp[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse PR number: %s", s)
	}

	buildNumber, err := strconv.Atoi(pp[4])
	if err != nil {
		return nil, fmt.Errorf("failed to parse build number: %s", s)
	}

	return &Version{major, minor, patch, prNumber, buildNumber}, nil
}

func (v Version) Increment(major, minor, patch int) Version {
	return Version{
		major: v.major + major,
		minor: v.minor + minor,
		patch: v.patch + patch,
	}
}

func (v Version) String() string {
	if v.buildNumber != 0 && v.prNumber != 0 {
		return fmt.Sprintf("v%d.%d.%d-pr-%d-build-%d", v.major, v.minor, v.patch, v.prNumber, v.buildNumber)
	}
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

func (v Version) Less(o Version) bool {
	if v.major < o.major {
		return true
	} else if v.minor < o.minor {
		return true
	} else if v.patch < o.patch {
		return true
	}

	return false
}

func (v Version) Diff(o Version) (int, int, int) {
	return int(math.Abs(float64(v.major) - float64(o.major))),
		int(math.Abs(float64(v.minor) - float64(o.minor))),
		int(math.Abs(float64(v.patch) - float64(o.patch)))
}
