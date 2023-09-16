package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
)

const NameApp = "bump00"

var (
	flagDebug         bool
	flagDryRun        bool
	flagRelease       bool
	flagReleaseRemote bool
)

const (
	netrcPath = "~/.netrc"

	// Git related
	gitRemoteOrigin = "origin"
)

func init() {
	flag.BoolVar(&flagDebug, "debug", false, "Debug mode, print eveyrthing")
	flag.BoolVar(&flagDryRun, "dry-run", false, "Will not do anything dangerous")
	flag.BoolVar(&flagRelease, "release", false, "Bump minor version, not RC anymore")
	flag.BoolVar(&flagReleaseRemote, "release-remote", false, "Release to the wild")
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

	rawTags, err := gitGetRawTags(ctx)
	if err != nil {
		slog.Error("git get raw tags", "error", err)
		return
	}

	newTag := genNewTag(rawTags, flagRelease)

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
		if err := gitRelease(ctx, newTag); err != nil {
			slog.Error("git release", "error", err)
			return
		}
	}
}
