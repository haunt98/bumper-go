package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/make-go-great/color-go"
)

const NameApp = "bump00"

var (
	flagDebug          bool
	flagDryRun         bool
	flagRelease        bool
	flagReleaseMinor   bool
	flagReleaseRemote  bool
	flagReleaseMessage string
)

const (
	netrcPath = "~/.netrc"

	// Git related
	gitRemoteOrigin = "origin"
)

func init() {
	flag.BoolVar(&flagDebug, "debug", false, "Debug mode, print eveyrthing")
	flag.BoolVar(&flagDryRun, "dry-run", false, "Will not do anything dangerous")
	flag.BoolVar(&flagRelease, "release", false, "Bump patch version, not RC anymore")
	flag.BoolVar(&flagReleaseMinor, "release-minor", false, "Bump minor not patch when release")
	flag.BoolVar(&flagReleaseRemote, "release-remote", false, "Release to the wild, support GitLab for now")
	flag.StringVar(&flagReleaseMessage, "release-message", "", "What you release my friend?")
}

func main() {
	flag.Parse()

	// Init slog
	slogLevel := slog.LevelInfo
	if flagDebug {
		slogLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slogLevel,
	})))

	ctx := context.Background()

	remoteURL, err := gitRemote(ctx)
	if err != nil {
		slog.Error("git remote", "error", err)
		return
	}
	color.PrintAppOK(NameApp, fmt.Sprintf("Hacking %s", remoteURL))

	rawTags, err := gitGetRawTags(ctx)
	if err != nil {
		slog.Error("git get raw tags", "error", err)
		return
	}

	oldTag, newTag := genNewTag(rawTags, flagRelease, flagReleaseMinor)
	if oldTag != "" {
		color.PrintAppOK(NameApp, fmt.Sprintf("Tag: %s -> %s", oldTag, newTag))
	} else {
		color.PrintAppOK(NameApp, fmt.Sprintf("New tag: %s", newTag))
	}

	if flagDryRun {
		return
	}

	if err := gitTag(ctx, newTag); err != nil {
		slog.Error("git tag", "error", err)
		return
	}

	if err := gitPush(ctx, newTag); err != nil {
		slog.Error("git push", "error", err)
		return
	}

	if flagReleaseRemote {
		if err := gitRelease(ctx, newTag, remoteURL, flagReleaseMessage); err != nil {
			slog.Error("git release", "error", err)
			return
		}
	}
}
