package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cast"

	"github.com/make-go-great/color-go"
)

const NameApp = "bump00"

var (
	flagDebug   bool
	flagDryRun  bool
	flagRelease bool
)

func init() {
	flag.BoolVar(&flagDebug, "debug", false, "Debug mode, print eveyrthing")
	flag.BoolVar(&flagDryRun, "dry-run", false, "Will not do anything dangerous")
	flag.BoolVar(&flagRelease, "release", false, "Bump minor version, not RC anymore")
}

func main() {
	flag.Parse()

	ctx := context.Background()

	// Make sure we have latest tags from default remote
	gitOuput, err := exec.CommandContext(ctx, "git", "fetch", "--tags").CombinedOutput()
	if err != nil {
		color.PrintAppError(NameApp, fmt.Sprintf("Failed to git fetch tags: %s", err.Error()))
	} else if flagDebug {
		log.Printf("git fetch tags:\n%s\n", string(gitOuput))
	}

	// List tags with reversed sort for semver
	gitOuput, err = exec.CommandContext(ctx, "git", "tag").CombinedOutput()
	if err != nil {
		color.PrintAppError(NameApp, fmt.Sprintf("Failed to git list tags: %s", err.Error()))
	} else if flagDebug {
		log.Printf("git list tags:\n%s\n", string(gitOuput))
	}

	rawTags := make([]string, 0, 100)
	for _, rawTag := range strings.Split(string(gitOuput), "\n") {
		rawTag = strings.TrimSpace(rawTag)
		if rawTag == "" {
			continue
		}

		rawTags = append(rawTags, rawTag)
	}

	newTagStr := genNewTag(rawTags, flagDebug, flagRelease)

	if !flagDryRun {
		// Tag
		// TODO: Handle if tag need comment
		gitOuput, err = exec.CommandContext(ctx, "git", "tag", newTagStr).CombinedOutput()
		if err != nil {
			color.PrintAppError(NameApp, fmt.Sprintf("Failed to git tag: %s", err.Error()))
		} else if flagDebug {
			log.Printf("git tag:\n%s\n", string(gitOuput))
		}

		// Push tag
		// TODO: Handle different remote
		gitOuput, err = exec.CommandContext(ctx, "git", "push", "origin", newTagStr).CombinedOutput()
		if err != nil {
			color.PrintAppError(NameApp, fmt.Sprintf("Failed to git push: %s", err.Error()))
		} else if flagDebug {
			log.Printf("git push:\n%s\n", string(gitOuput))
		}
	} else {
		log.Println("Will tag: ", newTagStr)
		log.Println("Will push tag: ", newTagStr)
	}
}

func genNewTag(rawTags []string, isDebug, isRelease bool) string {
	tags := make([]*semver.Version, 0, 100)
	for _, rawTag := range rawTags {
		tag, err := semver.NewVersion(rawTag)
		if err != nil {
			continue
		}
		tags = append(tags, tag)
	}

	sortTags(tags)
	if isDebug {
		log.Printf("tags: %+v\n", tags)
	}

	var newTagStr string
	if isRelease {
		// Default tag for release
		newTagStr = "v0.0.1"
		if len(tags) > 0 {
			latestTag := tags[0]
			color.PrintAppOK(NameApp, fmt.Sprintf("Latest tag: %s", latestTag.String()))

			if latestTag.Prerelease() == "" {
				// Latest tag is release
				// Only bump patch
				// v0.2.3 -> v0.2.4
				newTagStr = fmt.Sprintf("v%d.%d.%d",
					latestTag.Major(),
					latestTag.Minor(),
					latestTag.Patch()+1,
				)
			} else {
				// Latest tag is RC
				// Release tag is missing
				// Only remove RC
				// v0.2.3-RC1 -> v0.2.3
				newTagStr = fmt.Sprintf("v%d.%d.%d",
					latestTag.Major(),
					latestTag.Minor(),
					latestTag.Patch(),
				)
			}
		}
	} else {
		// Default tag for RC
		newTagStr = "v0.0.1-RC1"
		if len(tags) > 0 {
			latestTag := tags[0]
			color.PrintAppOK(NameApp, fmt.Sprintf("Latest tag: %s", latestTag.String()))

			// If latest tag don't have RC
			// Bump patch and add RC1
			// v0.2.0, v0.1.0-RC2, v0.1.0-RC1 -> v0.2.1-RC1
			// Otherwise latest tag already have RC
			// Only bump RC
			// v0.2.0-RC1, v0.2.0, v0.1.0-RC2, v0.1.0-RC1 -> v0.2.0-RC2
			latestPrerelease := latestTag.Prerelease()
			if latestPrerelease == "" {
				// Latest tag is already release
				// Bump patch with RC1
				// v0.2.3 -> v0.2.4-RC1
				newTagStr = fmt.Sprintf("v%d.%d.%d-RC1",
					latestTag.Major(),
					latestTag.Minor(),
					latestTag.Patch()+1,
				)
			} else {
				// Latest tag is RC
				// Release tag is missing
				// Only bump RC
				// v0.2.3-RC1 -> v0.2.3-RC2
				latestPrereleaseNum := cast.ToInt(strings.TrimLeft(latestPrerelease, "RC"))
				newTagStr = fmt.Sprintf("v%d.%d.%d-RC%d",
					latestTag.Major(),
					latestTag.Minor(),
					latestTag.Patch(),
					latestPrereleaseNum+1,
				)
			}
		}
	}

	color.PrintAppOK(NameApp, fmt.Sprintf("New tag: %s", newTagStr))

	return newTagStr
}

// Copy and modified from semver
type Collection00 []*semver.Version

func (c Collection00) Len() int {
	return len(c)
}

func (c Collection00) Less(i, j int) bool {
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

func (c Collection00) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Sort from latest to earliest
func sortTags(tags []*semver.Version) {
	sort.Sort(Collection00(tags))
}
