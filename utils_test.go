// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPerformWildcardSearch(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup structure
	createFile(t, filepath.Join(tmpDir, "file1.txt"), "file1.txt")
	createFile(t, filepath.Join(tmpDir, "file2.log"), "file2.log")
	createFile(t, filepath.Join(tmpDir, "data.json"), "data.json")
	createFile(t, filepath.Join(tmpDir, "a1.txt"), "a1.txt")
	createFile(t, filepath.Join(tmpDir, "sub1", "b.txt"), "sub1")
	createFile(t, filepath.Join(tmpDir, "sub2", "c.md"), "sub2")
	createEmptyDir(t, filepath.Join(tmpDir, "empty"))
	createEmptyDir(t, filepath.Join(tmpDir, "sub2", "innerempty"))

	tests := []struct {
		name              string
		basePath          string
		searchPath        string
		includeEmptyDirs  bool
		expectedFiles     []string
		expectedEmptyDirs []string
	}{
		{
			name:             "match *.txt at top-level only",
			basePath:         tmpDir,
			searchPath:       "*.txt",
			includeEmptyDirs: false,
			expectedFiles:    []string{"file1.txt", "a1.txt"},
		},
		{
			name:             "match **/*.txt recursively",
			basePath:         tmpDir,
			searchPath:       "**/*.txt",
			includeEmptyDirs: false,
			expectedFiles:    []string{"file1.txt", "a1.txt", filepath.Join("sub1", "b.txt")},
		},
		{
			name:             "match *.md in any folder",
			basePath:         tmpDir,
			searchPath:       "**/*.md",
			includeEmptyDirs: false,
			expectedFiles:    []string{filepath.Join("sub2", "c.md")},
		},
		{
			name:             "no matching files",
			basePath:         tmpDir,
			searchPath:       "*.go",
			includeEmptyDirs: false,
			expectedFiles:    []string{},
		},
		{
			name:             "match empty directories",
			basePath:         tmpDir,
			searchPath:       "**",
			includeEmptyDirs: true,
			expectedFiles: []string{
				"file1.txt",
				"file2.log",
				"data.json",
				"a1.txt",
				filepath.Join("sub1", "b.txt"),
				filepath.Join("sub2", "c.md"),
			},
			expectedEmptyDirs: []string{"empty", filepath.Join("sub2", "innerempty")},
		},
		{
			name:             "match ? wildcard in filename",
			basePath:         tmpDir,
			searchPath:       "a?.txt",
			includeEmptyDirs: false,
			expectedFiles:    []string{"a1.txt"},
		},
		{
			name:             "directory wildcard like sub*",
			basePath:         tmpDir,
			searchPath:       "sub*/**/*.txt",
			includeEmptyDirs: false,
			expectedFiles:    []string{filepath.Join("sub1", "b.txt")},
		},
		{
			name:             "search root directory",
			basePath:         tmpDir,
			searchPath:       "/",
			includeEmptyDirs: false,
			expectedFiles:    []string{filepath.Join("sub1", "b.txt")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, normBase, err := PerformWildcardSearch(tt.basePath, tt.searchPath, tt.includeEmptyDirs)
			require.NoErrorf(t, err, "error in PerformWildcardSearch: %v", err)

			foundFiles := map[string]bool{}
			foundDirs := map[string]bool{}

			for _, res := range results {
				relPath, _ := filepath.Rel(normBase, res.Path)
				if res.IsFile {
					foundFiles[filepath.ToSlash(relPath)] = true
				} else {
					foundDirs[filepath.ToSlash(relPath)] = true
				}
			}

			for _, f := range tt.expectedFiles {
				f = filepath.ToSlash(f)
				if !foundFiles[f] {
					t.Errorf("expected file %q not found", f)
				}
			}

			for _, d := range tt.expectedEmptyDirs {
				d = filepath.ToSlash(d)
				if !foundDirs[d] {
					t.Errorf("expected empty dir %q not found", d)
				}
			}
		})
	}
}

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
		{
			packagePath: "**",
			isSnupkg:    false,
			expected:    "**/*.nupkg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.packagePath, func(t *testing.T) {
			result := EnsurePackageExtension(tt.packagePath, tt.isSnupkg)
			require.Equalf(
				t,
				tt.expected,
				result,
				"EnsurePackageExtension(%q, %v) = %q; want %q",
				tt.packagePath,
				tt.isSnupkg,
				result,
				tt.expected,
			)
		})
	}
}

func TestWildcardToRegex(t *testing.T) {
	tests := []struct {
		name          string
		wildcard      string
		pathSeparator rune
		match         []string
		nomatch       []string
	}{
		{
			name:          "Match .txt files",
			wildcard:      "*.txt",
			pathSeparator: os.PathSeparator,
			match:         []string{"notes.txt", "README.TXT"},
			nomatch:       []string{"image.png", "notes.txt.bak"},
		},
		{
			name:          "Match CSV files in data/ directory",
			wildcard:      "data/*.csv",
			pathSeparator: os.PathSeparator,
			match:         []string{"data/file.csv", "data/test.CSV"},
			nomatch:       []string{"data/file.csvx", "databack/file.csv"},
		},
		{
			name:          "Match all .go files recursively",
			wildcard:      "**/*.go",
			pathSeparator: os.PathSeparator,
			match:         []string{"main.go", "src/util/main.go", "lib/test/hello.go"},
			nomatch:       []string{"main.go.old", "main.go.bak"},
		},
		{
			name:          "Match config files with a single character suffix",
			wildcard:      "config?.json",
			pathSeparator: os.PathSeparator,
			match:         []string{"config1.json", "configA.json"},
			nomatch:       []string{"config10.json", "conf.json"},
		},
		{
			name:          "Match test files with a single character suffix in any directory",
			wildcard:      "**/test?.*",
			pathSeparator: os.PathSeparator,
			match:         []string{"test1.py", "src/test2.go", "lib/testA.java"},
			nomatch:       []string{"test10.py", "test.py", "lib/test.py"},
		},
		{
			name:          "windows",
			wildcard:      "**\\test?.*",
			pathSeparator: '\\',
			match:         []string{"test1.py"},
			nomatch:       []string{"test10.py", "test.py"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			re := WildcardToRegex(tc.wildcard, tc.pathSeparator)

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
		})
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
				t.Fatalf("Unexpected error: %s", err.Error())
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
		{
			name:           "search ../test.1.0.0.nupkg",
			basePath:       "testdata",
			searchPath:     "../test.1.0.0.nupkg",
			expectedBase:   mustAbs(""),
			expectedSearch: "/test.1.0.0.nupkg",
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
func TestNormalizeWildcardForExcludedFiles(t *testing.T) {
	expected := NormalizeWildcardForExcludedFiles("", "**/abc")
	require.Equal(t, "**/abc", expected)
	expected = NormalizeWildcardForExcludedFiles("testdata", "../test.1.0.0.nupkg")
	require.Equal(t, "/test.1.0.0.nupkg", expected)
}

func TestIsDirectoryPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Unix-style directory",
			path:     "/usr/local/",
			expected: true,
		},
		{
			name:     "Unix-style file",
			path:     "/usr/local",
			expected: false,
		},
		{
			name:     "Windows-style directory",
			path:     "C:\\Program Files\\",
			expected: true,
		},
		{
			name:     "Windows-style file",
			path:     "C:\\Program Files",
			expected: false,
		},
		{
			name:     "empty string",
			path:     "",
			expected: false,
		},
		{
			name:     "root directory",
			path:     "/",
			expected: true,
		},
		{
			name:     "mixed Windows path with forward slashes",
			path:     "C:/Windows/System32/",
			expected: true,
		},
		{
			name:     "no trailing slash",
			path:     "C:/Windows/System32",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDirectoryPath(tt.path)
			if result != tt.expected {
				t.Errorf("isDirectoryPath(%q) = %v; expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsEmptyDirectory(t *testing.T) {
	// Test 1: Empty directory
	emptyDir := createTestDirectory(t, "emptyDir", []string{})
	t.Cleanup(func() {
		_ = os.RemoveAll(emptyDir)
	})

	t.Run("Empty directory", func(t *testing.T) {
		if empty := IsEmptyDirectory(emptyDir); !empty {
			t.Errorf("Expected directory to be empty, but it was not")
		}
	})

	// Test 2: Directory with files
	dirWithFiles := createTestDirectory(t, "dirWithFiles", []string{"file1.txt", "file2.txt"})
	t.Cleanup(func() {
		_ = os.RemoveAll(dirWithFiles)
	})

	t.Run("Directory with files", func(t *testing.T) {
		if empty := IsEmptyDirectory(dirWithFiles); empty {
			t.Errorf("Expected directory to have files, but it was empty")
		}
	})

	// Test 3: Non-existent directory
	t.Run("Non-existent directory", func(t *testing.T) {
		nonExistentDir := "/path/to/nonexistent/directory"
		if empty := IsEmptyDirectory(nonExistentDir); empty {
			t.Errorf("Expected error, but directory was considered empty")
		}
	})
}

func TestFixSourceURL(t *testing.T) {
	actual := fixSourceURL("/abc")
	expected := "file:///abc"
	require.Equal(t, expected, actual)
}

func TestPathCombine(t *testing.T) {
	actual := pathCombine()
	expected := ""
	require.Equal(t, expected, actual)
}
func TestGetFileNameWithoutExtension(t *testing.T) {
	require.Equal(t, "abc", getFileNameWithoutExtension(":\\abc"))
	require.Equal(t, "abc", getFileNameWithoutExtension(":/abc"))
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		filter   func(int) bool
		expected []int
	}{
		{
			name:     "filter even numbers",
			input:    []int{1, 2, 3, 4, 5, 6},
			filter:   func(n int) bool { return n%2 == 0 },
			expected: []int{2, 4, 6},
		},
		{
			name:     "filter numbers greater than 3",
			input:    []int{1, 2, 3, 4, 5, 6},
			filter:   func(n int) bool { return n > 3 },
			expected: []int{4, 5, 6},
		},
		{
			name:     "empty slice",
			input:    []int{},
			filter:   func(n int) bool { return n > 0 },
			expected: []int{},
		},
		{
			name:     "no matches",
			input:    []int{1, 2, 3},
			filter:   func(n int) bool { return n > 10 },
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filter(tt.input, tt.filter)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Filter() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSome(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expected  bool
	}{
		{
			name:      "has even number",
			input:     []int{1, 3, 5, 6, 7},
			predicate: func(n int) bool { return n%2 == 0 },
			expected:  true,
		},
		{
			name:      "no even numbers",
			input:     []int{1, 3, 5, 7},
			predicate: func(n int) bool { return n%2 == 0 },
			expected:  false,
		},
		{
			name:      "empty slice",
			input:     []int{},
			predicate: func(n int) bool { return n > 0 },
			expected:  false,
		},
		{
			name:      "has number greater than 5",
			input:     []int{1, 2, 3, 6, 4},
			predicate: func(n int) bool { return n > 5 },
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Some(tt.input, tt.predicate)
			if result != tt.expected {
				t.Errorf("Some() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetMatches(t *testing.T) {
	type testFile struct {
		path string
	}

	tests := []struct {
		name      string
		input     []testFile
		getPath   func(testFile) string
		wildcards []string
		expected  []testFile
	}{
		{
			name: "match txt files",
			input: []testFile{
				{path: "file1.txt"},
				{path: "file2.log"},
				{path: "file3.txt"},
			},
			getPath:   func(f testFile) string { return f.path },
			wildcards: []string{"*.txt"},
			expected: []testFile{
				{path: "file1.txt"},
				{path: "file3.txt"},
			},
		},
		{
			name: "match multiple patterns",
			input: []testFile{
				{path: "file1.txt"},
				{path: "file2.log"},
				{path: "file3.md"},
			},
			getPath:   func(f testFile) string { return f.path },
			wildcards: []string{"*.txt", "*.md"},
			expected: []testFile{
				{path: "file1.txt"},
				{path: "file3.md"},
			},
		},
		{
			name: "no matches",
			input: []testFile{
				{path: "file1.txt"},
				{path: "file2.log"},
			},
			getPath:   func(f testFile) string { return f.path },
			wildcards: []string{"*.md"},
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMatches(tt.input, tt.getPath, tt.wildcards)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetMatches() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetFilteredPackageFiles(t *testing.T) {
	type testFile struct {
		path string
	}

	tests := []struct {
		name      string
		input     []testFile
		getPath   func(testFile) string
		wildcards []string
		expected  []testFile
		remaining []testFile
	}{
		{
			name: "filter txt files",
			input: []testFile{
				{path: "file1.txt"},
				{path: "file2.log"},
				{path: "file3.txt"},
			},
			getPath:   func(f testFile) string { return f.path },
			wildcards: []string{"*.txt"},
			expected: []testFile{
				{path: "file1.txt"},
				{path: "file3.txt"},
			},
			remaining: []testFile{
				{path: "file2.log"},
			},
		},
		{
			name: "filter multiple patterns",
			input: []testFile{
				{path: "file1.txt"},
				{path: "file2.log"},
				{path: "file3.md"},
			},
			getPath:   func(f testFile) string { return f.path },
			wildcards: []string{"*.txt", "*.md"},
			expected: []testFile{
				{path: "file1.txt"},
				{path: "file3.md"},
			},
			remaining: []testFile{
				{path: "file2.log"},
			},
		},
		{
			name: "no matches",
			input: []testFile{
				{path: "file1.txt"},
				{path: "file2.log"},
			},
			getPath:   func(f testFile) string { return f.path },
			wildcards: []string{"*.md"},
			expected:  nil,
			remaining: []testFile{
				{path: "file1.txt"},
				{path: "file2.log"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFilteredPackageFiles(&tt.input, tt.getPath, tt.wildcards)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetFilteredPackageFiles() = %v, want %v", result, tt.expected)
			}
			if !reflect.DeepEqual(tt.input, tt.remaining) {
				t.Errorf("GetFilteredPackageFiles() remaining = %v, want %v", tt.input, tt.remaining)
			}
		})
	}
}

func TestSplitWithFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      []rune
		expected []string
	}{
		{
			name:     "split by comma and semicolon",
			input:    "a,b;c,d;e",
			sep:      []rune{',', ';'},
			expected: []string{"a", "b", "c", "d", "e"},
		},
		{
			name:     "split by space and tab",
			input:    "hello world\tgoodbye",
			sep:      []rune{' ', '\t'},
			expected: []string{"hello", "world", "goodbye"},
		},
		{
			name:     "empty string",
			input:    "",
			sep:      []rune{',', ';'},
			expected: nil,
		},
		{
			name:     "no separators",
			input:    "hello",
			sep:      []rune{',', ';'},
			expected: []string{"hello"},
		},
		{
			name:     "multiple consecutive separators",
			input:    "a,,b;;c",
			sep:      []rune{',', ';'},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "trim spaces",
			input:    " a , b ; c ",
			sep:      []rune{',', ';'},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitWithFilter(tt.input, tt.sep)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SplitWithFilter() = %v, want %v", result, tt.expected)
			}
		})
	}
}
func TestGetPathWithDirectorySeparator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"UnixStyle", "lib/net6.0/foo.dll", "lib/net6.0/foo.dll"},
		{"WindowsStyle", "lib\\net6.0\\foo.dll", normalizedPath("lib/net6.0/foo.dll", "lib\\net6.0\\foo.dll")},
		{"MixedSeparators", "lib/net6.0\\foo.dll", normalizedPath("lib/net6.0/foo.dll", "lib\\net6.0\\foo.dll")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPathWithDirectorySeparator(tt.input)
			if result != tt.expected {
				t.Errorf("GetPathWithDirectorySeparator(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// based on current OS, return expected separator
func normalizedPath(unixPath, windowsPath string) string {
	if os.PathSeparator == '/' {
		return unixPath
	}
	return windowsPath
}
