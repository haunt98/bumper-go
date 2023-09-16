package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os/exec"
	"strings"

	"github.com/xanzy/go-gitlab"

	"github.com/make-go-great/color-go"
	"github.com/make-go-great/ioe-go"
	"github.com/make-go-great/netrc-go"
)

var (
	ErrNetrcMissing   = errors.New("netrc missing")
	ErrNotSupportHost = errors.New("not support host")
)

func gitGetRawTags(ctx context.Context) ([]string, error) {
	// Make sure we have latest tags from default remote
	gitOutput, err := exec.CommandContext(ctx, "git", "fetch", "--tags").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git: failed to fetch: %w", err)
	}
	slog.Debug("git fetch", "output", gitOutput)

	// List tags with reversed sort for semver
	gitOutput, err = exec.CommandContext(ctx, "git", "tag").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git: failed to tag: %w", err)
	}
	slog.Debug("git tag", "output", gitOutput)

	// Extract raw tags from git output
	rawTags := make([]string, 0, 100)
	for _, rawTag := range strings.Split(string(gitOutput), "\n") {
		rawTag := strings.TrimSpace(rawTag)
		if rawTag == "" {
			continue
		}

		rawTags = append(rawTags, rawTag)
	}
	slog.Debug("raw tags", "tags", rawTags)

	return rawTags, nil
}

// Tag
func gitTag(ctx context.Context, tag string) error {
	gitOuput, err := exec.CommandContext(ctx, "git", "tag", tag).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git: failed to tag: %w", err)
	}
	slog.Debug("git tag", "output", gitOuput)

	return nil
}

// Push tag
func gitPush(ctx context.Context, newTag string) error {
	gitOuput, err := exec.CommandContext(ctx, "git", "push", "origin", newTag).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git: failed to push: %w", err)
	}
	slog.Debug("git push", "output", gitOuput)

	return nil
}

func gitRelease(ctx context.Context, tag string) error {
	netrcData, err := netrc.ParseFile(netrcPath)
	if err != nil {
		return fmt.Errorf("netrc: failed to parse file: %w", err)
	}

	gitOuput, err := exec.CommandContext(ctx, "git", "remote", "-v").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git: failed to remote: %w", err)
	}
	slog.Debug("git remote", "output", gitOuput)

	// Example gitOuput
	// origin  https://github.com/haunt98/haha.git (fetch)
	// origin  https://github.com/haunt98/haha.git (push)
	var gitRemoteURLRaw string
	for _, line := range strings.Split(string(gitOuput), "\n") {
		line := strings.TrimSpace(line)
		if line == "" {
			continue
		}

		lines2 := strings.Fields(line)
		if len(lines2) != 3 {
			continue
		}

		// Expect origin
		if strings.EqualFold(lines2[0], gitRemoteOrigin) {
			// Only get first
			gitRemoteURLRaw = lines2[1]
			break
		}
	}

	gitRemoteURL, err := url.Parse(gitRemoteURLRaw)
	if err != nil {
		return fmt.Errorf("git: failed to parse %s: %w", gitRemoteURLRaw, err)
	}
	slog.Debug("git remote", "url", gitRemoteURL)

	var netrcPassword string
	for _, machine := range netrcData.Machines {
		if strings.EqualFold(
			strings.TrimSpace(machine.Name),
			strings.TrimSpace(gitRemoteURL.Hostname()),
		) {
			netrcPassword = machine.Password
		}
	}

	if netrcPassword == "" {
		return ErrNetrcMissing
	}

	if strings.Contains(strings.ToLower(gitRemoteURL.Hostname()), "gitlab") {
		return gitReleaseGitLab(ctx, netrcPassword, gitRemoteURL, tag)
	}

	return ErrNotSupportHost
}

func gitReleaseGitLab(ctx context.Context, token string, remoteURL *url.URL, tag string) error {
	g, err := gitlab.NewClient(token, gitlab.WithBaseURL("https://"+remoteURL.Hostname()+"/api/v4"))
	if err != nil {
		return fmt.Errorf("gitlab: failed to new client: %w", err)
	}

	pid := strings.Trim(remoteURL.Path, "/")
	pid = strings.TrimSuffix(pid, ".git")

	// Like a demo
	// If OK should we process to create release later
	if _, _, err := g.Releases.ListReleases(pid, nil); err != nil {
		return fmt.Errorf("gitlab: failed to list releases: %w", err)
	}

	color.PrintAppOK(NameApp, "Release description:")
	description := ioe.ReadInput()

	if _, _, err := g.Releases.CreateRelease(pid, &gitlab.CreateReleaseOptions{
		Name:        &tag,
		TagName:     &tag,
		Description: &description,
	}); err != nil {
		return fmt.Errorf("gitlab: failed to create release: %w", err)
	}

	return nil
}
