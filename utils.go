// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// SearchPathResult stores the result of a search, including the file path and whether it is a file
type SearchPathResult struct {
	// Path to the file or directory
	Path string
	// IsFile True if it's a file, false if it's a directory
	IsFile bool
}

// Filter returns a new slice containing all elements from the input slice that satisfy the predicate function.
// The function is generic and can work with any type T.
// It preserves the order of elements in the original slice.
func Filter[T any](slice []T, f func(T) bool) []T {
	for i, value := range slice {
		if !f(value) {
			result := slices.Clone(slice[:i])
			for i++; i < len(slice); i++ {
				value = slice[i]
				if f(value) {
					result = append(result, value)
				}
			}
			return result
		}
	}
	return slice
}

// Some checks if at least one element in the slice satisfies the predicate function.
// The function is generic and can work with any type T.
// Returns true if any element satisfies the predicate, false otherwise.
func Some[T any](slice []T, predicate func(T) bool) bool {
	for _, value := range slice {
		if predicate(value) {
			return true
		}
	}
	return false
}

// getMatches returns items from the source that match any of the wildcard patterns.
// The function is generic and can work with any type T.
// It uses the provided getPath function to extract the path from each item for pattern matching.
// Returns a slice of items whose paths match any of the provided wildcard patterns.
func getMatches[T any](source []T, getPath func(T) string, wildcards []string) []T {
	var matches []T
	var regexps []*regexp.Regexp

	for _, pattern := range wildcards {
		regexps = append(regexps, WildcardToRegex(pattern, os.PathSeparator))
	}

	for _, item := range source {
		path := getPath(item)
		for _, re := range regexps {
			if re.MatchString(path) {
				matches = append(matches, item)
				break
			}
		}
	}

	return matches
}

// GetFilteredPackageFiles filters the source slice based on wildcard patterns and returns two results:
// 1. A slice of items that match any of the wildcard patterns
// 2. Updates the source slice to contain only the non-matching items
// The function is generic and can work with any comparable type T.
// It uses the provided getPath function to extract the path from each item for pattern matching.
func GetFilteredPackageFiles[T comparable](source *[]T, getPath func(T) string, wildcards []string) []T {
	matched := make(map[T]bool)
	for _, item := range getMatches(*source, getPath, wildcards) {
		matched[item] = true
	}

	var filtered []T
	var remaining []T

	for _, item := range *source {
		if _, ok := matched[item]; ok {
			filtered = append(filtered, item)
		} else {
			remaining = append(remaining, item)
		}
	}

	*source = remaining
	return filtered
}

func NormalizeWildcardForExcludedFiles(basePath, wildcard string) string {
	if strings.HasPrefix(wildcard, "**") {
		return wildcard
	}
	basePath = normalizeBasePath(basePath, &wildcard)
	return pathCombine(basePath, wildcard)
}

// WildcardToRegex converts a wildcard string to a regular expression.
func WildcardToRegex(wildcard string, sep rune) *regexp.Regexp {
	// Escape all regular special characters
	escaped := regexp.QuoteMeta(wildcard)

	var pattern string
	if sep != '/' {
		// regex wildcard adjustments for Windows-style file systems
		pattern = strings.ReplaceAll(
			escaped,
			"/",
			"\\\\",
		) // On Windows, / is treated the same as \.
		pattern = strings.ReplaceAll(
			pattern,
			"\\.\\*\\*",
			"\\.[^\\\\.]*",
		) // .** should not match on ../file or ./file but will match .file
		pattern = strings.ReplaceAll(
			pattern,
			"\\*\\*\\\\",
			`(\\\\)?([^\\\\]+\\\\)*?`,
		) //For recursive wildcards \**\, include the current directory.
		pattern = strings.ReplaceAll(
			pattern,
			"\\*\\*",
			".*",
		) // For recursive wildcards that don't end in a slash e.g. **.txt would be treated as a .txt file at any depth
		pattern = strings.ReplaceAll(
			pattern,
			"\\*",
			`[^\\\\]*(\\\\)?`,
		) // For non recursive searches, limit it any character that is not a directory separator
		pattern = strings.ReplaceAll(
			pattern,
			"\\?",
			".",
		) // ? translates to a single any character
	} else {
		// regex wildcard adjustments for *nix-style file systems
		pattern = strings.ReplaceAll(escaped, "\\.\\*\\*", "\\.[^/.]*")  // .** should not match on ../file or ./file but will match .file
		pattern = strings.ReplaceAll(pattern, "\\*\\*/", "/?([^/]+/)*?") //For recursive wildcards /**/, include the current directory.
		pattern = strings.ReplaceAll(pattern, "\\*\\*", ".*")            // For recursive wildcards that don't end in a slash e.g. **.txt would be treated as a .txt file at any depth
		pattern = strings.ReplaceAll(pattern, "\\*", `[^/]*(/)?`)        // For non recursive searches, limit it any character that is not a directory separator
		pattern = strings.ReplaceAll(pattern, "\\?", ".")                // ? translates to a single any character
	}
	finalPattern := fmt.Sprintf("^%s$", pattern)
	// Compile regular expressions to be case-insensitive using the `(?i)` prefix (equivalent to
	// RegexOptions.IgnoreCase)
	re := regexp.MustCompile(`(?i)` + finalPattern)
	return re
}

// PerformWildcardSearch searches for files or directories based on a wildcard search pattern.
func PerformWildcardSearch(basePath, searchPath string, includeEmptyDirs bool) ([]*SearchPathResult, string, error) {
	// Flag to check if the search pattern should include directories recursively
	flag1 := false

	// Check if the search path is a directory, modify it to include '**/*'
	if IsDirectoryPath(searchPath) {
		searchPath = pathCombine(searchPath, "**", "*")
		flag1 = true
	}

	// Normalize the base path and search path
	basePath = normalizeBasePath(basePath, &searchPath)
	normalizedBasePath, err := getPathToEnumerateFrom(basePath, searchPath)
	if err != nil {
		return nil, "", err
	}
	searchPattern := pathCombine(basePath, searchPath)

	// Convert wildcard search pattern to regex
	searchRegex := WildcardToRegex(searchPattern, os.PathSeparator)

	searchRecursively := strings.Contains(searchPath, "**") || strings.Contains(filepath.Dir(searchPath), "*")

	var results []*SearchPathResult

	// Search for files matching the search pattern
	err = filepath.WalkDir(normalizedBasePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories if not searching recursively
		if !searchRecursively && path != normalizedBasePath && filepath.Dir(path) != normalizedBasePath {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Match file or directory path with regex
		if searchRegex.MatchString(path) {
			if d.IsDir() {
				// If it's a directory, check if we should include empty directories
				if ok := IsDirectoryPath(path); ok && includeEmptyDirs {
					results = append(results, &SearchPathResult{Path: path, IsFile: false})
				}
			} else {
				// If it's a file, include it in the results
				results = append(results, &SearchPathResult{Path: path, IsFile: true})
			}
		}

		return nil
	})

	// Handle error during WalkDir
	if err != nil {
		return nil, "", err
	}

	// If flag1 is true and the normalized base path is empty, include the base path as a result
	if ok := IsDirectoryPath(normalizedBasePath); ok && flag1 {
		results = append(results, &SearchPathResult{Path: normalizedBasePath, IsFile: false})
	}
	return results, normalizedBasePath, nil
}

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

		pathToEnumerateFrom = pathCombine(basePath, dirName)
	} else {
		// Find the last directory separator before the wildcard
		lastIndex := strings.LastIndex(searchPath[:startIndex], string(filepath.Separator))
		if lastIndex == -1 {
			// If no directory separator is found, the search is at the base level
			pathToEnumerateFrom = basePath
		} else {
			// Get the part of the path before the wildcard
			pathPart := searchPath[:lastIndex]
			pathToEnumerateFrom = pathCombine(basePath, pathPart)
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
		basePath = pathCombine(basePath, path2)
		*searchPath = (*searchPath)[len(path2):]
	}
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		// In production code, you might want to log this error or handle it differently
		return filepath.Clean(basePath)
	}
	return absBasePath
}

// resolvePackageFromPath Resolves a package path into a list of paths.
// If the path contains wildcards then the path is expanded to all matching entries.
func resolvePackageFromPath(packagePath string, isSnupkg bool) ([]string, error) {
	packagePath = EnsurePackageExtension(packagePath, isSnupkg)
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	results, _, err := PerformWildcardSearch(dir, packagePath, false)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(results))
	for _, item := range results {
		paths = append(paths, item.Path)
	}
	return paths, nil
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
		packagePath = fmt.Sprintf("%s%s%s", packagePath, string(filepath.Separator), "*")
	} else if !strings.HasSuffix(packagePath, "*") {
		// If it doesn't end with "*", append "*" at the end
		packagePath = fmt.Sprintf("%s%s", packagePath, "*")
	}
	// Add the appropriate extension based on isSnupkg
	if isSnupkg {
		packagePath = fmt.Sprintf("%s%s", packagePath, SnupkgExtension)
	} else {
		packagePath = fmt.Sprintf("%s%s", packagePath, PackageExtension)
	}

	return packagePath
}

func IsDirectoryPath(path string) bool {
	if len(path) <= 1 {
		return path == "/" || path == "\\"
	}
	lastChar := path[len(path)-1]
	return lastChar == '/' || lastChar == '\\'
}

// IsEmptyDirectory checks if the given directory is empty.
func IsEmptyDirectory(directory string) bool {
	// Open the directory
	dir, err := os.Open(directory)
	if err != nil {
		log.Printf("open this %s fatal ", directory)
		return false
	}
	defer dir.Close()

	// Read the directory entries
	entries, err := dir.Readdirnames(0) // 0 means read all entries
	if err != nil {
		log.Printf("Read the %s directory entries fatal ", directory)
		return false
	}

	// If the length of entries is 0, then the directory is empty
	return len(entries) == 0
}

// createSourceURL Same as "new Uri" except that it can handle UNIX style paths that start with '/'
func createSourceURL(source string) (*url.URL, error) {
	source = fixSourceURL(source)
	return url.Parse(source)
}

func fixSourceURL(source string) string {
	if filepath.Separator == '/' && source != "" && strings.HasPrefix(source, "/") {
		source = "file://" + source
	}
	return source
}

func isSourceNuGetSymbolServer(source *url.URL) bool {
	return strings.EqualFold(source.Host, NuGetSymbolHostName)
}

// pathCombine combines an array of strings into a path.
func pathCombine(paths ...string) string {
	if len(paths) == 0 {
		return ""
	}

	var combinedPath string
	for _, path := range paths {
		path = strings.TrimSpace(path) // Trim any leading or trailing spaces
		if path == "" {
			continue // Skip empty paths
		}
		if filepath.IsAbs(path) {
			// Reset to the absolute path
			combinedPath = path
		} else {
			// Join with the current combined path
			combinedPath = filepath.Join(combinedPath, path)
		}
	}

	// Normalize the final combined path
	return filepath.Clean(combinedPath)
}

// getFileNameWithoutExtension returns the file name of the specified path string without the extension.
func getFileNameWithoutExtension(path string) string {
	newPath := path
	if strings.Contains(newPath, ":\\") {
		p := strings.Split(newPath, "\\")
		newPath = p[len(p)-1]
	} else if strings.Contains(newPath, ":/") {
		p := strings.Split(newPath, "/")
		newPath = p[len(p)-1]
	}
	// Get the file name
	fileName := filepath.Base(newPath)
	// Remove the extension name
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

// GetSymbolsPath Get the symbols package from the original package. Removes the .nupkg and adds .snupkg or .symbols.
// nupkg.
func GetSymbolsPath(packagePath string, isSnupkg bool) string {
	symbolPath := getFileNameWithoutExtension(packagePath)
	if isSnupkg {
		symbolPath = fmt.Sprintf("%s%s", symbolPath, SnupkgExtension)
	} else {
		symbolPath = fmt.Sprintf("%s%s", symbolPath, SymbolsExtension)
	}
	packageDir := filepath.Dir(packagePath)

	return pathCombine(packageDir, symbolPath)
}

// getServiceEndpointUrl Calculates the URL to the package to.
func getServiceEndpointUrl(source, path string, noServiceEndpoint bool) (*url.URL, error) {
	baseUri, err := ensureTrailingSlash(source)
	if err != nil {
		return nil, err
	}
	requestUri := ""
	if strings.TrimSpace(strings.TrimPrefix(baseUri.Path, "/")) == "" && !noServiceEndpoint {
		requestUri = fmt.Sprintf("%s%s/%s", strings.TrimRight(baseUri.String(), "/"), ServiceEndpoint, path)
	} else {
		requestUri = fmt.Sprintf("%s%s", baseUri.String(), path)
	}
	return url.Parse(requestUri)
}

// ensureTrailingSlash Ensure a trailing slash at the end
func ensureTrailingSlash(value string) (*url.URL, error) {
	if !strings.HasSuffix(value, "/") {
		value = fmt.Sprintf("%s%s", value, "/")
	}
	return createSourceURL(value)
}

// SplitWithFilter splits the input string by any of the specified separator runes and returns a slice of non-empty
// strings.
// The function trims whitespace from each resulting part and filters out empty strings.
// Multiple consecutive separators are treated as a single separator.
func SplitWithFilter(input string, sep []rune) []string {
	var result []string
	var sb strings.Builder

	filter := func(r rune) bool {
		for _, d := range sep {
			if r == d {
				return true
			}
		}
		return false
	}

	for _, r := range input {
		if filter(r) {
			if sb.Len() > 0 {
				part := strings.TrimSpace(sb.String())
				if part != "" {
					result = append(result, part)
				}
				sb.Reset()
			}
		} else {
			sb.WriteRune(r)
		}
	}

	// Add the last part if any
	if sb.Len() > 0 {
		part := strings.TrimSpace(sb.String())
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}
