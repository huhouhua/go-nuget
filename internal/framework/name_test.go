// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package framework

import (
	"errors"
	"fmt"
	"testing"

	"github.com/huhouhua/go-nuget/version"

	"github.com/stretchr/testify/require"
)

func TestNewFrameworkName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *FrameworkName
		error error
	}{
		{
			name:  "valid framework name with version",
			input: ".NET Framework, Version=v4.5",
			want: &FrameworkName{
				identifier: ".NET Framework",
				version:    *version.NewVersionFrom(4, 5, 0, "", ""),
				profile:    "",
			},
		},
		{
			name:  "valid framework name with version and profile",
			input: ".NET Framework, Version=v4.5, Profile=Client",
			want: &FrameworkName{
				identifier: ".NET Framework",
				version:    *version.NewVersionFrom(4, 5, 0, "", ""),
				profile:    "Client",
			},
		},
		{
			name:  "valid framework name with uppercase V in version",
			input: ".NET Framework, Version=V4.5",
			want: &FrameworkName{
				identifier: ".NET Framework",
				version:    *version.NewVersionFrom(4, 5, 0, "", ""),
				profile:    "",
			},
		},
		{
			name:  "empty framework name",
			input: "",
			error: fmt.Errorf("frameworkName cannot be empty"),
		},
		{
			name:  "invalid component count",
			input: ".NET Framework, Version=v4.5, Profile=Client, Extra=Value",
			error: fmt.Errorf("frameworkName must have 2 or 3 components"),
		},
		{
			name:  "missing version",
			input: ".NET Framework, Profile=Client",
			error: fmt.Errorf("frameworkName must contain a version"),
		},
		{
			name:  "invalid version format",
			input: ".NET Framework, Version=invalid",
			error: fmt.Errorf("invalid version: %w", errors.New("invalid semantic version")),
		},
		{
			name:  "invalid key in component",
			input: ".NET Framework, InvalidKey=Value",
			error: fmt.Errorf("invalid key: \"InvalidKey\""),
		},
		{
			name:  "invalid component format",
			input: ".NET Framework, Version",
			error: fmt.Errorf("invalid component: \" Version\""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewFrameworkName(tc.input)
			require.Equal(t, tc.error, err)
			if err != nil {
				return
			}
			require.Equal(t, tc.want.GetIdentifier(), got.GetIdentifier())
			gotVersion := got.GetVersion()
			wantVersion := tc.want.GetVersion()
			require.True(t, wantVersion.Semver.Equal(gotVersion.Semver))
			require.Equal(t, tc.want.GetProfile(), got.GetProfile())
		})
	}
}
