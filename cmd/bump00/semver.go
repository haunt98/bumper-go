package main

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cast"
)

// Copy and modified from semver
type Collection []*semver.Version

func (c Collection) Len() int {
	return len(c)
}

func (c Collection) Less(i, j int) bool {
	// Compare RC
	if c[i].Major() == c[j].Major() &&
		c[i].Minor() == c[j].Minor() &&
		c[i].Patch() == c[j].Patch() &&
		strings.HasPrefix(c[i].Prerelease(), "RC") &&
		strings.HasPrefix(c[j].Prerelease(), "RC") {
		rcI := cast.ToInt(strings.TrimPrefix(c[i].Prerelease(), "RC"))
		rcJ := cast.ToInt(strings.TrimPrefix(c[j].Prerelease(), "RC"))
		return rcI > rcJ
	}

	return c[i].GreaterThan(c[j])
}

func (c Collection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Sort from latest to earliest
func sortTags(tags []*semver.Version) {
	sort.Sort(Collection(tags))
}

func genNewTag(rawTags []string, isRelease bool) (string, string) {
	tags := make([]*semver.Version, 0, 100)
	for _, rawTag := range rawTags {
		tag, err := semver.NewVersion(rawTag)
		if err != nil {
			continue
		}
		tags = append(tags, tag)
	}

	sortTags(tags)
	slog.Debug("sort tags", "tags", tags)

	var oldTagStr string
	var newTagStr string
	if isRelease {
		// Default tag for release
		newTagStr = "v0.0.1"
		if len(tags) > 0 {
			oldTag := tags[0]
			oldTagStr = oldTag.String()

			if oldTag.Prerelease() == "" {
				// Old tag is release
				// Only bump patch
				// v0.2.3 -> v0.2.4
				newTagStr = fmt.Sprintf("v%d.%d.%d",
					oldTag.Major(),
					oldTag.Minor(),
					oldTag.Patch()+1,
				)
			} else {
				// Old tag is RC
				// Release tag is missing
				// Only remove RC
				// v0.2.3-RC1 -> v0.2.3
				newTagStr = fmt.Sprintf("v%d.%d.%d",
					oldTag.Major(),
					oldTag.Minor(),
					oldTag.Patch(),
				)
			}
		}
	} else {
		// Default tag for RC
		newTagStr = "v0.0.1-RC1"
		if len(tags) > 0 {
			oldTag := tags[0]
			oldTagStr = oldTag.String()

			// If old tag don't have RC
			// Bump patch and add RC1
			// v0.2.0, v0.1.0-RC2, v0.1.0-RC1 -> v0.2.1-RC1
			// Otherwise old tag already have RC
			// Only bump RC
			// v0.2.0-RC1, v0.2.0, v0.1.0-RC2, v0.1.0-RC1 -> v0.2.0-RC2
			oldPrerelease := oldTag.Prerelease()
			if oldPrerelease == "" {
				// Old tag is already release
				// Bump patch with RC1
				// v0.2.3 -> v0.2.4-RC1
				newTagStr = fmt.Sprintf("v%d.%d.%d-RC1",
					oldTag.Major(),
					oldTag.Minor(),
					oldTag.Patch()+1,
				)
			} else {
				// Old tag is RC
				// Release tag is missing
				// Only bump RC
				// v0.2.3-RC1 -> v0.2.3-RC2
				oldPrereleaseNum := cast.ToInt(strings.TrimLeft(oldPrerelease, "RC"))
				newTagStr = fmt.Sprintf("v%d.%d.%d-RC%d",
					oldTag.Major(),
					oldTag.Minor(),
					oldTag.Patch(),
					oldPrereleaseNum+1,
				)
			}
		}
	}

	if oldTagStr != "" && !strings.HasPrefix(oldTagStr, "v") {
		oldTagStr = "v" + oldTagStr
	}

	return oldTagStr, newTagStr
}
