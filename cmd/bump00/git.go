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

func gitRemote(ctx context.Context) (*url.URL, error) {
	gitOuput, err := exec.CommandContext(ctx, "git", "remote", "-v").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git: failed to remote: %w", err)
	}
	slog.Debug("git remote", "output", gitOuput)

	// Example gitOuput
	// origin  https://github.com/haunt98/haha.git (fetch)
	// origin  https://github.com/haunt98/haha.git (push)
	var rawURL string
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
			rawURL = lines2[1]
			break
		}
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("git: failed to parse %s: %w", rawURL, err)
	}

	return parsedURL, nil
}

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

func gitTag(ctx context.Context, tag string) error {
	gitOuput, err := exec.CommandContext(ctx, "git", "tag", tag).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git: failed to tag: %w", err)
	}
	slog.Debug("git tag", "output", gitOuput)

	return nil
}

func gitPush(ctx context.Context, newTag string) error {
	gitOuput, err := exec.CommandContext(ctx, "git", "push", gitRemoteOrigin, newTag).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git: failed to push: %w", err)
	}
	slog.Debug("git push", "output", gitOuput)

	return nil
}

func gitRelease(ctx context.Context, tag string, remoteURL *url.URL, releaseMsg string) error {
	netrcData, err := netrc.ParseFile(netrcPath)
	if err != nil {
		return fmt.Errorf("netrc: failed to parse file: %w", err)
	}

	var netrcPassword string
	for _, machine := range netrcData.Machines {
		if strings.EqualFold(
			strings.TrimSpace(machine.Name),
			strings.TrimSpace(remoteURL.Hostname()),
		) {
			netrcPassword = machine.Password
		}
	}

	if netrcPassword == "" {
		return ErrNetrcMissing
	}

	if strings.Contains(strings.ToLower(remoteURL.Hostname()), "gitlab") {
		return gitReleaseGitLab(ctx, netrcPassword, remoteURL, tag, releaseMsg)
	}

	return ErrNotSupportHost
}

func gitReleaseGitLab(ctx context.Context, token string, remoteURL *url.URL, tag, releaseMsg string) error {
	g, err := gitlab.NewClient(token, gitlab.WithBaseURL("https://"+remoteURL.Hostname()+"/api/v4"))
	if err != nil {
		return fmt.Errorf("gitlab: failed to new client: %w", err)
	}

	pid := strings.Trim(remoteURL.Path, "/")
	pid = strings.TrimSuffix(pid, ".git")

	if releaseMsg == "" {
		color.PrintAppOK(NameApp, "Release message:")
		releaseMsg = ioe.ReadInput()
	}

	if _, _, err := g.Releases.CreateRelease(pid, &gitlab.CreateReleaseOptions{
		Name:        &tag,
		TagName:     &tag,
		Description: &releaseMsg,
	}); err != nil {
		return fmt.Errorf("gitlab: failed to create release: %w", err)
	}

	return nil
}
