// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"archive/zip"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/huhouhua/go-nuget"
)

const (
	DefaultVersion                                             = 1
	XdtTransformationVersion                                   = 6
	TargetFrameworkSupportForDependencyContentsAndToolsVersion = 4
)

func xmlEscape(s string) string {
	replacer := strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&quot;",
		`'`, "&apos;",
	)
	return replacer.Replace(s)
}
func getPathWithDirectorySeparator(path string, sep rune) string {
	if sep == '/' {
		return getPathWithForwardSlashes(path)
	} else {
		return strings.ReplaceAll(path, "/", "\\")
	}
}

func generateRelationshipId(path string) string {
	hash := sha512.Sum512([]byte(path))
	hexStr := hex.EncodeToString(hash[:])
	return "R" + hexStr[:16]
}
func calcPsmdcpName(files []PackageFile, deterministic bool) (string, error) {
	if deterministic {
		hash := sha512.New()
		for _, file := range files {
			if stream, err := file.GetStream(); err != nil {
				return "", err
			} else {
				if data, err := io.ReadAll(stream); err != nil {
					return "", err
				} else {
					hash.Write(data)
				}
			}
		}
		sum := hash.Sum(nil)
		return hex.EncodeToString(sum)[:32], nil
	} else {
		return randomString(32), nil
	}
}

const letterBytes = "abcdef0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
func createPart(zipWriter *zip.Writer, filePath string,
	sourceStream io.Reader, lastWriteTime time.Time, warningMessage *strings.Builder) error {
	if strings.HasSuffix(strings.ToLower(filePath), nuget.NuspecExtension) {
		return nil
	}
	// Split on '/', '\\', and OS-specific separator (assuming Unix-like system here)
	separators := []string{"/", "\\"}
	for _, sep := range separators {
		filePath = strings.ReplaceAll(filePath, sep, "/")
	}

	// Escape each segment
	segments := strings.Split(filePath, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}

	escapedPath := strings.Join(segments, "/")

	//Create an absolute URI to get the refinement on the relative path
	partURL, err := defaultURL.Parse(escapedPath)
	if err != nil {
		return err
	}
	cleanPath := path.Clean(partURL.Path)

	entryName, err := url.PathUnescape(cleanPath)
	if err != nil {
		return err
	}
	entry, err := createPackageFileEntry(zipWriter, entryName, lastWriteTime, warningMessage)
	if err != nil {
		return err
	}
	_, err = io.Copy(entry, sourceStream)
	return err
}

func createEntry(zipWriter *zip.Writer, entryName string, deterministic bool) (io.Writer, error) {
	header := &zip.FileHeader{
		Name:   entryName,
		Method: zip.Deflate,
	}
	if deterministic {
		header.Modified = zipFormatMinDate
	}
	return zipWriter.CreateHeader(header)
}

func createPackageFileEntry(
	zipWriter *zip.Writer,
	entryName string,
	timeOffset time.Time,
	warningMessage *strings.Builder,
) (io.Writer, error) {
	header := &zip.FileHeader{
		Name:     entryName,
		Method:   zip.Deflate,
		Modified: timeOffset,
	}
	if timeOffset.Before(zipFormatMinDate) {
		warningMessage.WriteString(fmt.Sprintf("Timestamp for '%s' (%s) is before minimum. Adjusted to %s.\n",
			entryName, timeOffset.Format("2006-01-02"), zipFormatMinDate.Format("2006-01-02")))
		header.Modified = zipFormatMinDate
	} else if timeOffset.After(zipFormatMaxDate) {
		warningMessage.WriteString(fmt.Sprintf("Timestamp for '%s' (%s) is after maximum. Adjusted to %s.\n",
			entryName, timeOffset.Format("2006-01-02"), zipFormatMaxDate.Format("2006-01-02")))
		header.Modified = zipFormatMaxDate
	}
	return zipWriter.CreateHeader(header)
}

func resolveSearchPattern(
	basePath, searchPath, targetPath string,
	includeEmptyDirectories bool,
) ([]*PhysicalPackageFile, error) {
	searchResults, normalizedBasePath, err := nuget.PerformWildcardSearch(basePath, searchPath, includeEmptyDirectories)
	if err != nil {
		return nil, err
	}
	files := make([]*PhysicalPackageFile, 0)
	for _, result := range searchResults {
		file := &PhysicalPackageFile{
			sourcePath: result.Path,
			targetPath: resolvePackagePath(normalizedBasePath, searchPath, result.Path, targetPath),
		}
		if !result.IsFile {
			file.targetPath = path.Join(file.targetPath, nuget.PackageEmptyFileName)
		}
		files = append(files, file)
	}
	return files, nil
}

// resolvePackagePath the path of the file inside a package. For recursive wildcard paths, we preserve the
// path portion beginning
// with the wildcard. For non-recursive wildcard paths, we use the file name from the actual file path on disk.
func resolvePackagePath(searchDirectory, searchPattern, fullPath, targetPath string) string {
	var packagePath string
	isWildcardSearch := strings.Contains(searchPattern, "*")
	isRecursiveWildcardSearch := isWildcardSearch && strings.Contains(searchPattern, "**")
	if (isRecursiveWildcardSearch || isWildcardSearch) && strings.HasPrefix(fullPath, searchDirectory) {
		// The search pattern is recursive. Preserve the non-wildcard portion of the path.
		// e.g. Search: X:\foo\**\*.cs results in SearchDirectory: X:\foo and a file path of X:\foo\bar\biz\boz.cs
		// Truncating X:\foo\ would result in the package path.
		relPath := fullPath[len(searchDirectory):]
		packagePath = strings.TrimLeft(relPath, string(filepath.Separator))
	} else if !isWildcardSearch && strings.EqualFold(path.Ext(searchPattern), path.Ext(targetPath)) {
		// If the search does not contain wild cards, and the target path shares the same extension, copy it
		// e.g. <file src="ie\css\style.css" target="Content\css\ie.css" /> --> Content\css\ie.css
		return targetPath
	} else {
		packagePath = path.Base(fullPath)
	}
	return path.Join(targetPath, packagePath)
}

// isKnownFolder Returns true if the path uses a known folder root.
func isKnownFolder(targetPath string) bool {
	if strings.TrimSpace(targetPath) == "" {
		return false
	}
	parts := nuget.SplitWithFilter(targetPath, []rune{'\\', '/'})
	if len(parts) > 1 {
		topLevelDirectory := parts[0]
		return nuget.Some(nuget.Known, func(folder nuget.Folder) bool {
			return strings.EqualFold(string(folder), topLevelDirectory)
		})
	}
	return false
}
func contains(slice []string, item string) bool {
	return nuget.Some(slice, func(s string) bool {
		return strings.Contains(s, item)
	})
}
func hasContentFilesV2(files []PackageFile) bool {
	return nuget.Some(files, func(file PackageFile) bool {
		prefix := fmt.Sprintf("contentFiles%v", os.PathSeparator)
		return strings.HasPrefix(file.GetPath(), prefix)
	})
}
func hasIncludeExclude(dependencyGroups []*PackageDependencyGroup) bool {
	return nuget.Some(dependencyGroups, func(group *PackageDependencyGroup) bool {
		return nuget.Some(group.Packages, func(dependency *nuget.Dependency) bool {
			return dependency.Include != nil || dependency.Exclude != nil
		})
	})
}
func hasXdtTransformFile(files []PackageFile) bool {
	return nuget.Some(files, func(file PackageFile) bool {
		prefix := fmt.Sprintf("content%v", os.PathSeparator)
		return strings.HasPrefix(file.GetPath(), prefix) &&
			(strings.HasSuffix(file.GetPath(), ".install.xdt") ||
				strings.HasSuffix(file.GetPath(), ".uninstall.xdt"))
	})
}
func requiresV4TargetFrameworkSchema(files []PackageFile) bool {
	// check if any file under Content or Tools has TargetFramework defined
	hasContentOrTool := nuget.Some(files, func(file PackageFile) bool {
		framework := file.GetNuGetFramework()
		contentPrefix := fmt.Sprintf("content%v", os.PathSeparator)
		toolsPrefix := fmt.Sprintf("tools%v", os.PathSeparator)
		return framework != nil && framework.IsUnsupported && (strings.HasPrefix(file.GetPath(), contentPrefix) ||
			strings.HasPrefix(file.GetPath(), toolsPrefix))
	})
	if hasContentOrTool {
		return true
	}
	// now check if the Lib folder has any empty framework folder
	return nuget.Some(files, func(file PackageFile) bool {
		framework := file.GetNuGetFramework()
		libPrefix := fmt.Sprintf("lib%v", os.PathSeparator)
		return framework != nil && strings.HasPrefix(file.GetPath(), libPrefix) &&
			file.GetEffectivePath() == nuget.PackageEmptyFileName
	})
}

func determineMinimumSchemaVersion(files []PackageFile, dependencyGroups []*PackageDependencyGroup) int {
	if hasContentFilesV2(files) || hasIncludeExclude(dependencyGroups) || hasXdtTransformFile(files) {
		// version 5
		return XdtTransformationVersion
	}
	if requiresV4TargetFrameworkSchema(files) {
		// version 4
		return TargetFrameworkSupportForDependencyContentsAndToolsVersion
	}
	return DefaultVersion
}

// getPathWithForwardSlashes Replace all back slashes with forward slashes.
// If the path does not contain a back slash
// the original string is returned.
func getPathWithForwardSlashes(path string) string {
	if strings.Contains(path, "\\") {
		return strings.ReplaceAll(path, "\\", "/")
	}
	return path
}

func stripLeadingDirectorySeparators(fileName string) string {
	filename := getPathWithForwardSlashes(fileName)
	currentDirectoryPath := "./"
	if strings.HasPrefix(filename, currentDirectoryPath) {
		filename = filename[len(currentDirectoryPath):]
	}
	return strings.TrimLeft(filename, "/")
}
