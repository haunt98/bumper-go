package main

import "testing"

func TestGenNewTag(t *testing.T) {
	tests := []struct {
		name      string
		rawTags   []string
		isRelease bool
		want      string
	}{
		{
			name:      "release empty",
			rawTags:   []string{},
			isRelease: true,
			want:      "v0.0.1",
		},
		{
			name: "release with previous release",
			rawTags: []string{
				"v0.0.1",
			},
			isRelease: true,
			want:      "v0.0.2",
		},
		{
			name: "release with previous release",
			rawTags: []string{
				"v0.0.1",
				"v0.0.1-RC1",
			},
			isRelease: true,
			want:      "v0.0.2",
		},
		{
			name: "release with previous rc",
			rawTags: []string{
				"v0.0.1",
				"v0.0.1-RC1",
				"v0.0.2-RC1",
			},
			isRelease: true,
			want:      "v0.0.2",
		},
		{
			name:      "rc empty",
			rawTags:   []string{},
			isRelease: false,
			want:      "v0.0.1-RC1",
		},
		{
			name: "rc with previous release",
			rawTags: []string{
				"v0.0.1",
			},
			isRelease: false,
			want:      "v0.0.2-RC1",
		},
		{
			name: "rc with previous release",
			rawTags: []string{
				"v0.0.1",
				"v0.0.1-RC1",
			},
			isRelease: false,
			want:      "v0.0.2-RC1",
		},
		{
			name: "rc with previous rc",
			rawTags: []string{
				"v0.0.1",
				"v0.0.1-RC1",
				"v0.0.2-RC1",
			},
			isRelease: false,
			want:      "v0.0.2-RC2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := genNewTag(tc.rawTags, false, tc.isRelease)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
