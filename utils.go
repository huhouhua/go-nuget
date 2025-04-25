// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	PackageExtension = ".nupkg"
	SnupkgExtension  = ".snupkg"
)

// SearchPathResult stores the result of a search, including the file path and whether it is a file
type SearchPathResult struct {
	Path   string // Path to the file or directory
	IsFile bool   // True if it's a file, false if it's a directory
}

// wildcardToRegex converts a wildcard string to a regular expression.
func wildcardToRegex(wildcard string) *regexp.Regexp {
	// 转义所有正则特殊字符
	escaped := regexp.QuoteMeta(wildcard)

	var pattern string
	if os.PathSeparator != '/' {
		pattern = strings.ReplaceAll(escaped, "/", "\\\\")
		pattern = strings.ReplaceAll(pattern, "\\.\\*\\*", "\\.[^\\\\.]*")
		pattern = strings.ReplaceAll(pattern, "\\*\\*\\\\", `(\\\\)?([^\\\\]+\\\\)*?`)
		pattern = strings.ReplaceAll(pattern, "\\*\\*", ".*")
		pattern = strings.ReplaceAll(pattern, "\\*", `[^\\\\]*(\\\\)?`)
		pattern = strings.ReplaceAll(pattern, "\\?", ".")
	} else {
		pattern = strings.ReplaceAll(escaped, "\\.\\*\\*", "\\.[^/.]*")
		pattern = strings.ReplaceAll(pattern, "\\*\\*/", "/?([^/]+/)*?")
		pattern = strings.ReplaceAll(pattern, "\\*\\*", ".*")
		pattern = strings.ReplaceAll(pattern, "\\*", `[^/]*(/)?`)
		pattern = strings.ReplaceAll(pattern, "\\?", ".")
	}

	finalPattern := "^" + pattern + "$"
	// 编译正则表达式，使用 `(?i)` 前缀实现不区分大小写（等效于 RegexOptions.IgnoreCase）
	re := regexp.MustCompile(`(?i)` + finalPattern)
	return re
}

// PerformWildcardSearch searches for files and optionally empty directories
// matching the wildcard pattern.
//func PerformWildcardSearch(basePath, searchPath string, includeEmptyDirectories bool) ([]SearchPathResult, string, error) {
//	flag1 := false
//	if isDirectoryPath(searchPath) {
//		searchPath = filepath.Join(searchPath, "**", "*")
//		flag1 = true
//	}
//
//	basePath, searchPath = normalizeBasePath(basePath, searchPath)
//	normalizedBasePath := GetPathToEnumerateFrom(basePath, searchPath)
//	searchRegex := WildcardToRegex(filepath.Join(basePath, searchPath))
//
//	searchOption := filepath.Walk
//	if !strings.Contains(searchPath, "**") && !strings.Contains(filepath.Dir(searchPath), "*") {
//		// Only top directory
//		searchOption = func(root string, walkFn filepath.WalkFunc) error {
//			entries, err := os.ReadDir(root)
//			if err != nil {
//				return err
//			}
//			for _, entry := range entries {
//				info, err := entry.Info()
//				if err != nil {
//					return err
//				}
//				err = walkFn(filepath.Join(root, entry.Name()), info, nil)
//				if err != nil {
//					return err
//				}
//			}
//			return nil
//		}
//	}
//
//	var results []SearchPathResult
//	err := searchOption(normalizedBasePath, func(path string, info os.FileInfo, err error) error {
//		if err != nil {
//			return err
//		}
//		relPath, err := filepath.Rel(basePath, path)
//		if err != nil {
//			return err
//		}
//		if searchRegex.MatchString(filepath.Join(basePath, relPath)) {
//			if info.IsDir() {
//				if includeEmptyDirectories && IsEmptyDirectory(path) {
//					results = append(results, SearchPathResult{Path: path, IsFile: false})
//				}
//			} else {
//				results = append(results, SearchPathResult{Path: path, IsFile: true})
//			}
//		}
//		return nil
//	})
//	if err != nil {
//		return nil, "", err
//	}
//
//	if flag1 && isDirectoryPath(normalizedBasePath) {
//		results = append(results, SearchPathResult{Path: normalizedBasePath, IsFile: false})
//	}
//
//	return results, normalizedBasePath, nil
//}

// getPathToEnumerateFrom determines the path to enumerate from based on the base path and search path.
func getPathToEnumerateFrom(basePath, searchPath string) (string, error) {
	// Find the index of the first '*' character, which indicates the wildcard
	startIndex := strings.Index(searchPath, "*")
	var pathToEnumerateFrom string

	// If no wildcard is found, the directory is part of the base path
	if startIndex == -1 {
		dirName := filepath.Dir(searchPath)
		if dirName == "" {
			return "", errors.New("filepath.Dir(searchPath) returned null")
		}
		pathToEnumerateFrom = filepath.Join(basePath, dirName)
	} else {
		// Find the last directory separator before the wildcard
		lastIndex := strings.LastIndex(searchPath[:startIndex], string(filepath.Separator))
		if lastIndex == -1 {
			// If no directory separator is found, the search is at the base level
			pathToEnumerateFrom = basePath
		} else {
			// Get the part of the path before the wildcard
			pathPart := searchPath[:lastIndex]
			pathToEnumerateFrom = filepath.Join(basePath, pathPart)
		}
	}
	return pathToEnumerateFrom, nil
}

// normalizeBasePath normalizes the base path by handling relative paths, including parent directory references ("..").
func normalizeBasePath(basePath string, searchPath *string) string {
	path2 := ".."
	str := "."
	if strings.TrimSpace(basePath) == "" {
		basePath = str
	}
	for strings.HasPrefix(*searchPath, path2) {
		basePath = filepath.Join(basePath, path2)
		*searchPath = (*searchPath)[len(path2):]
	}
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		// In production code, you might want to log this error or handle it differently
		return filepath.Clean(basePath)
	}
	return absBasePath

}

// EnsurePackageExtension Ensure any wildcards in packagePath end with *.nupkg or *.snupkg.
func EnsurePackageExtension(packagePath string, isSnupkg bool) string {
	// If packagePath doesn't contain '*' and already ends with .nupkg or .snupkg, return the path as is
	if !strings.Contains(packagePath, "*") &&
		(strings.HasSuffix(packagePath, PackageExtension) || strings.HasSuffix(packagePath, SnupkgExtension)) {
		return packagePath
	}

	// If packagePath ends with "**", we modify it by adding "*"
	if strings.HasSuffix(packagePath, "**") {

		// Add directory separator and wildcard
		packagePath = packagePath + string(filepath.Separator) + "*"

	} else if !strings.HasSuffix(packagePath, "*") {
		// If it doesn't end with "*", append "*" at the end
		packagePath += "*"
	}

	// Add the appropriate extension based on isSnupkg
	if isSnupkg {
		packagePath += SnupkgExtension
	} else {
		packagePath += PackageExtension
	}

	return packagePath
}

func isDirectoryPath(path string) bool {
	if len(path) <= 1 {
		return path == "/" || path == "\\"
	}
	lastChar := path[len(path)-1]
	return lastChar == '/' || lastChar == '\\'
}

// IsEmptyDirectory checks if the given directory is empty.
func isEmptyDirectory(directory string) (bool, error) {
	// Open the directory
	dir, err := os.Open(directory)
	if err != nil {
		return false, err
	}
	defer dir.Close()

	// Read the directory entries
	entries, err := dir.Readdirnames(0) // 0 means read all entries
	if err != nil {
		return false, err
	}

	// If the length of entries is 0, then the directory is empty
	return len(entries) == 0, nil
}
