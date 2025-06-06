// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"os"
	"strings"
	"time"
)

type PackageFile interface {
	// GetPath Gets the full path of the file inside the package.
	GetPath() string

	// GetEffectivePath Gets the path that excludes the root folder (content/ lib/ tools) and framework folder (if
	// present).
	// Example: If a package has the Path as 'content\[net40]\scripts\jQuery. js', the EffectivePath
	// will be 'scripts\jQuery. js'. If it is 'tools\init. ps1', the EffectivePath will be 'init. ps1'.
	GetEffectivePath() string

	// GetNuGetFramework object representing this package file's target framework. Use this instead of
	// TargetFramework.
	GetNuGetFramework() *Framework

	GetLastWriteTime() time.Time

	GetStream() (*os.File, error)
}

type PhysicalPackageFile struct {
	PackageFile
	nugetFramework *Framework
	streamFactory  func() *os.File

	effectivePath string

	lastWriteTime time.Time
	// Path on disk
	sourcePath string

	// Path in package
	targetPath string
}

func NewPhysicalPackageFile(sourcePath, targetPath string, streamFactory func() *os.File) PackageFile {
	return &PhysicalPackageFile{
		sourcePath:    sourcePath,
		targetPath:    targetPath,
		streamFactory: streamFactory,
	}
}

func (p *PhysicalPackageFile) GetPath() string {
	return p.targetPath
}
func (p *PhysicalPackageFile) SetPath(targetPath string) {
	if strings.Compare(p.targetPath, targetPath) == 0 {
		return
	}
	p.targetPath = targetPath
	p.nugetFramework = ParseNuGetFrameworkFromFilePath(p.targetPath, &p.effectivePath)
}

func (p *PhysicalPackageFile) GetEffectivePath() string {
	return p.effectivePath
}
func (p *PhysicalPackageFile) GetNuGetFramework() *Framework {
	return p.nugetFramework
}
func (p *PhysicalPackageFile) GetStream() (*os.File, error) {
	if p.streamFactory != nil {
		p.lastWriteTime = time.Now().UTC()
		return p.streamFactory(), nil
	}
	if strings.TrimSpace(p.sourcePath) != "" {
		info, err := os.Stat(p.sourcePath)
		if err != nil {
			return nil, err
		}
		p.lastWriteTime = info.ModTime().UTC()
		return os.Open(p.sourcePath)
	}
	return nil, nil
}

func (p *PhysicalPackageFile) GetLastWriteTime() time.Time {
	return p.lastWriteTime
}
