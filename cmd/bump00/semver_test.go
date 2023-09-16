package main

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
)

func TestSortTags(t *testing.T) {
	tests := []struct {
		name string
		tags []*semver.Version
		want []*semver.Version
	}{
		{
			name: "only release",
			tags: []*semver.Version{
				semver.MustParse("0.0.1"),
				semver.MustParse("0.0.2"),
				semver.MustParse("0.0.3"),
			},
			want: []*semver.Version{
				semver.MustParse("0.0.3"),
				semver.MustParse("0.0.2"),
				semver.MustParse("0.0.1"),
			},
		},
		{
			name: "only rc",
			tags: []*semver.Version{
				semver.MustParse("0.0.1-RC1"),
				semver.MustParse("0.0.1-RC2"),
				semver.MustParse("0.0.1-RC9"),
				semver.MustParse("0.0.1-RC11"),
			},
			want: []*semver.Version{
				semver.MustParse("0.0.1-RC11"),
				semver.MustParse("0.0.1-RC9"),
				semver.MustParse("0.0.1-RC2"),
				semver.MustParse("0.0.1-RC1"),
			},
		},
		{
			name: "mixed",
			tags: []*semver.Version{
				semver.MustParse("0.0.1-RC1"),
				semver.MustParse("0.0.1-RC2"),
				semver.MustParse("0.0.1-RC9"),
				semver.MustParse("0.0.1-RC11"),
				semver.MustParse("0.0.2"),
				semver.MustParse("0.0.3-RC1"),
				semver.MustParse("0.0.3-RC2"),
				semver.MustParse("0.0.3-RC20"),
				semver.MustParse("0.0.3-RC100"),
				semver.MustParse("0.0.3"),
			},
			want: []*semver.Version{
				semver.MustParse("0.0.3"),
				semver.MustParse("0.0.3-RC100"),
				semver.MustParse("0.0.3-RC20"),
				semver.MustParse("0.0.3-RC2"),
				semver.MustParse("0.0.3-RC1"),
				semver.MustParse("0.0.2"),
				semver.MustParse("0.0.1-RC11"),
				semver.MustParse("0.0.1-RC9"),
				semver.MustParse("0.0.1-RC2"),
				semver.MustParse("0.0.1-RC1"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sortTags(tc.tags)
			assert.Equal(t, tc.want, tc.tags)
		})
	}
}

func TestGenNewTag(t *testing.T) {
	tests := []struct {
		name       string
		rawTags    []string
		isRelease  bool
		wantOldTag string
		wantNewTag string
	}{
		{
			name:       "release empty",
			rawTags:    []string{},
			isRelease:  true,
			wantOldTag: "",
			wantNewTag: "v0.0.1",
		},
		{
			name: "release with previous release",
			rawTags: []string{
				"v0.0.1",
			},
			isRelease:  true,
			wantOldTag: "v0.0.1",
			wantNewTag: "v0.0.2",
		},
		{
			name: "release with previous release",
			rawTags: []string{
				"v0.0.1",
				"v0.0.1-RC1",
			},
			isRelease:  true,
			wantOldTag: "v0.0.1",
			wantNewTag: "v0.0.2",
		},
		{
			name: "release with previous rc",
			rawTags: []string{
				"v0.0.1",
				"v0.0.1-RC1",
				"v0.0.2-RC1",
			},
			isRelease:  true,
			wantOldTag: "v0.0.2-RC1",
			wantNewTag: "v0.0.2",
		},
		{
			name:       "rc empty",
			rawTags:    []string{},
			isRelease:  false,
			wantOldTag: "",
			wantNewTag: "v0.0.1-RC1",
		},
		{
			name: "rc with previous release",
			rawTags: []string{
				"v0.0.1",
			},
			isRelease:  false,
			wantOldTag: "v0.0.1",
			wantNewTag: "v0.0.2-RC1",
		},
		{
			name: "rc with previous release",
			rawTags: []string{
				"v0.0.1",
				"v0.0.1-RC1",
			},
			isRelease:  false,
			wantOldTag: "v0.0.1",
			wantNewTag: "v0.0.2-RC1",
		},
		{
			name: "rc with previous rc",
			rawTags: []string{
				"v0.0.1",
				"v0.0.1-RC1",
				"v0.0.2-RC1",
			},
			isRelease:  false,
			wantOldTag: "v0.0.2-RC1",
			wantNewTag: "v0.0.2-RC2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOldTag, gotNewTag := genNewTag(tc.rawTags, tc.isRelease)
			assert.Equal(t, tc.wantOldTag, gotOldTag)
			assert.Equal(t, tc.wantNewTag, gotNewTag)
		})
	}
}
