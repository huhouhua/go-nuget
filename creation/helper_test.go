// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"testing"
)

func TestGetPathWithDirectorySeparator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      rune
		expected string
	}{
		{"UnixStyle", "lib/net6.0/foo.dll", '/', "lib/net6.0/foo.dll"},
		{
			"WindowsStyle",
			"lib\\net6.0\\foo.dll",
			'\\',
			normalizedPath("lib/net6.0/foo.dll", "lib\\net6.0\\foo.dll", '\\'),
		},
		{
			"MixedSeparators",
			"lib/net6.0\\foo.dll",
			'\\',
			normalizedPath("lib/net6.0/foo.dll", "lib\\net6.0\\foo.dll", '\\'),
		},
		{
			"MixedSeparatorsUnix",
			"lib/net6.0\\foo.dll",
			'/',
			normalizedPath("lib/net6.0/foo.dll", "lib\\net6.0\\foo.dll", '/'),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPathWithDirectorySeparator(tt.input, tt.sep)
			if result != tt.expected {
				t.Errorf("getPathWithDirectorySeparator(%q, %q) = %q; want %q", tt.input, tt.sep, result, tt.expected)
			}
		})
	}
}
func normalizedPath(unixPath, windowsPath string, sep rune) string {
	if sep == '/' {
		return unixPath
	}
	return windowsPath
}
