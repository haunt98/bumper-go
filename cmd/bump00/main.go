package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cast"
)

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
		log.Fatalln("Failed to git fetch tags: ", err)
	} else if flagDebug {
		log.Printf("git fetch tags:\n%s\n", string(gitOuput))
	}

	// List tags with reversed sort for semver
	gitOuput, err = exec.CommandContext(ctx, "git", "tag", "--sort=-version:refname").CombinedOutput()
	if err != nil {
		log.Fatalln("Failed to git list tags: ", err)
	} else if flagDebug {
		log.Printf("git list tags:\n%s\n", string(gitOuput))
	}

	tags := make([]*semver.Version, 0, 100)
	for _, rawTag := range strings.Split(string(gitOuput), "\n") {
		tag, err := semver.NewVersion(rawTag)
		if err != nil {
			continue
		}
		tags = append(tags, tag)
	}
	if flagDebug {
		log.Printf("tags: %+v\n", tags)
	}

	var newTagStr string
	if flagRelease {
		// Default tag for release
		newTagStr = "v0.0.1"
		if len(tags) > 0 {
			latestTag := tags[0]
			if flagDebug {
				log.Printf("Latest tag: %+v\n", latestTag)
			}

			// Ignore prerelease, always bump patch
			// 0.2.3 -> 0.2.4
			// v0.2.3-RC2 -> v0.2.4
			newTagStr = fmt.Sprintf("v%d.%d.%d",
				latestTag.Major(),
				latestTag.Minor(),
				latestTag.Patch()+1,
			)
		}
	} else {
		// Default tag for RC
		newTagStr = "v0.0.1-RC1"
		if len(tags) > 0 {
			latestTag := tags[0]
			if flagDebug {
				log.Printf("Latest tag: %+v\n", latestTag)
			}

			// If latest tag don't have RC
			// Bump patch and add RC1
			// v0.2.0, v0.1.0-RC2, v0.1.0-RC1 -> v0.2.1-RC1
			// Otherwise latest tag already have RC
			// Only bump RC
			// v0.2.0-RC1, v0.2.0, v0.1.0-RC2, v0.1.0-RC1 -> v0.2.0-RC2
			latestPrerelease := latestTag.Prerelease()
			if latestPrerelease != "" && strings.HasPrefix(latestPrerelease, "RC") {
				latestPrereleaseNum := cast.ToInt(strings.TrimLeft(latestPrerelease, "RC"))
				newTagStr = fmt.Sprintf("v%d.%d.%d-RC%d",
					latestTag.Major(),
					latestTag.Minor(),
					latestTag.Patch(),
					latestPrereleaseNum+1,
				)
			} else {
				newTagStr = fmt.Sprintf("v%d.%d.%d-RC1",
					latestTag.Major(),
					latestTag.Minor(),
					latestTag.Patch()+1,
				)
			}

		}
	}
	if flagDebug {
		log.Printf("New tag: %+v\n", newTagStr)
	}

	if !flagDryRun {
		// Tag
		// TODO: Handle if tag need comment
		gitOuput, err = exec.CommandContext(ctx, "git", "tag", newTagStr).CombinedOutput()
		if err != nil {
			log.Fatalln("Failed to git tag: ", err)
		} else if flagDebug {
			log.Printf("git tag:\n%s\n", string(gitOuput))
		}

		// Push tag
		// TODO: Handle different remote
		gitOuput, err = exec.CommandContext(ctx, "git", "push", "origin", newTagStr).CombinedOutput()
		if err != nil {
			log.Fatalln("Failed to git push: ", err)
		} else if flagDebug {
			log.Printf("git push:\n%s\n", string(gitOuput))
		}
	} else {
		log.Println("Will tag: ", newTagStr)
		log.Println("Will push tag: ", newTagStr)
	}
}
