// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestEnsurePackageExtension(t *testing.T) {
	tests := []struct {
		packagePath string
		isSnupkg    bool
		expected    string
	}{
		{
			packagePath: "example-package",
			isSnupkg:    false,
			expected:    "example-package*.nupkg",
		},
		{
			packagePath: "example-package",
			isSnupkg:    true,
			expected:    "example-package*.snupkg",
		},
		{
			packagePath: "example-package.nupkg",
			isSnupkg:    false,
			expected:    "example-package.nupkg",
		},
		{
			packagePath: "example-package.snupkg",
			isSnupkg:    true,
			expected:    "example-package.snupkg",
		},
		{
			packagePath: "example-package*",
			isSnupkg:    false,
			expected:    "example-package*.nupkg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.packagePath, func(t *testing.T) {
			result := EnsurePackageExtension(tt.packagePath, tt.isSnupkg)
			require.Equalf(t, tt.expected, result, "EnsurePackageExtension(%q, %v) = %q; want %q", tt.packagePath, tt.isSnupkg, result, tt.expected)
		})
	}
}

func TestWildcardToRegex(t *testing.T) {
	tests := []struct {
		wildcard string
		match    []string
		nomatch  []string
	}{
		{
			wildcard: "*.txt",
			match:    []string{"notes.txt", "README.TXT"},
			nomatch:  []string{"image.png", "notes.txt.bak"},
		},
		{
			wildcard: "data/*.csv",
			match:    []string{"data/file.csv", "data/test.CSV"},
			nomatch:  []string{"data/file.csvx", "databack/file.csv"},
		},
		{
			wildcard: "**/*.go",
			match:    []string{"main.go", "src/util/main.go", "lib/test/hello.go"},
			nomatch:  []string{"main.go.old", "main.go.bak"},
		},
		{
			wildcard: "config?.json",
			match:    []string{"config1.json", "configA.json"},
			nomatch:  []string{"config10.json", "conf.json"},
		},
		{
			wildcard: "**/test?.*",
			match:    []string{"test1.py", "src/test2.go", "lib/testA.java"},
			nomatch:  []string{"test10.py", "test.py", "lib/test.py"},
		},
	}

	for _, tc := range tests {
		re := wildcardToRegex(tc.wildcard)

		for _, input := range tc.match {
			if !re.MatchString(input) {
				t.Errorf("Expected match: pattern=%q input=%q", tc.wildcard, input)
			}
		}

		for _, input := range tc.nomatch {
			if re.MatchString(input) {
				t.Errorf("Expected no match: pattern=%q input=%q", tc.wildcard, input)
			}
		}
	}
}

func TestGetPathToEnumerateFrom(t *testing.T) {
	tests := []struct {
		name           string
		basePath       string
		searchPath     string
		expectedResult string
	}{
		{
			name:           "No wildcard",
			basePath:       "/home/user",
			searchPath:     "src/utils/helper.go",
			expectedResult: filepath.Join("/home/user", "src/utils"),
		},
		{
			name:           "Wildcard with directory",
			basePath:       "/project",
			searchPath:     "src/**/*.cs",
			expectedResult: filepath.Join("/project", "src"),
		},
		{
			name:           "Wildcard in root",
			basePath:       "/project",
			searchPath:     "*.*",
			expectedResult: "/project",
		},
		{
			name:           "No directory",
			basePath:       "/base",
			searchPath:     "file.cs",
			expectedResult: "/base",
		},
		{
			name:           "Wildcard with deep nested path",
			basePath:       "/project",
			searchPath:     "a/b/**/c/*.go",
			expectedResult: filepath.Join("/project", "a/b"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := getPathToEnumerateFrom(tt.basePath, tt.searchPath)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			expected := filepath.Clean(tt.expectedResult)
			if actual != expected {
				t.Errorf("Expected %s, but got %s", expected, actual)
			}
		})
	}
}

func TestNormalizeBasePath(t *testing.T) {
	tests := []struct {
		name           string
		basePath       string
		searchPath     string
		expectedBase   string
		expectedSearch string
	}{
		{
			name:           "Empty base path with no parent refs",
			basePath:       "",
			searchPath:     "src/**/*.go",
			expectedBase:   mustAbs("."),
			expectedSearch: "src/**/*.go",
		},
		{
			name:           "No change to search path",
			basePath:       "/home/user",
			searchPath:     "src/**/*.go",
			expectedBase:   mustAbs("/home/user"),
			expectedSearch: "src/**/*.go",
		},
		{
			name:           "search testdata/test.1.0.0.nupkg",
			basePath:       "testdata",
			searchPath:     "test.1.0.0.nupkg",
			expectedBase:   mustAbs("testdata"),
			expectedSearch: "test.1.0.0.nupkg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			search := tt.searchPath
			base := normalizeBasePath(tt.basePath, &search)

			if base != tt.expectedBase {
				t.Errorf("Expected base path: %s, got: %s", tt.expectedBase, base)
			}
			if search != tt.expectedSearch {
				t.Errorf("Expected search path: %s, got: %s", tt.expectedSearch, search)
			}
		})
	}
}

// Helper to make absolute path in test cases
func mustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return abs
}
