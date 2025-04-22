package nuget

import (
	"errors"
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
func wildcardToRegex(wildcard string) (*regexp.Regexp, error) {
	// Escape all special characters in the wildcard
	str := regexp.QuoteMeta(wildcard)

	// Replace the wildcard patterns with appropriate regex equivalents
	if string(filepath.Separator) != "/" {
		// For non-unix-like file systems (Windows)
		str = "^" + strings.ReplaceAll(str, "/", "\\\\")
		str = strings.Replace(str, "\\.\\*\\*", "\\.[^\\\\.]*", -1)
		str = strings.Replace(str, "\\*\\*\\\\", "(\\\\\\\\)?([^\\\\]+\\\\)*?", -1)
		str = strings.Replace(str, "\\*\\*", ".*", -1)
		str = strings.Replace(str, "\\*", "[^\\\\]*(\\\\)?", -1)
		str = strings.Replace(str, "\\?", ".", -1)
	} else {
		// For unix-like file systems
		str = "^" + strings.ReplaceAll(str, "/", "/?([^/]+/)*?")
		str = strings.Replace(str, "\\.\\*\\*", "\\.[^/.]*", -1)
		str = strings.Replace(str, "\\*\\*", ".*", -1)
		str = strings.Replace(str, "\\*", "[^/]*(/)?", -1)
		str = strings.Replace(str, "\\?", ".", -1) + "$"
	}

	// Compile the regular expression with appropriate options
	return regexp.MustCompile(str), nil
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
