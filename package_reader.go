// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/huhouhua/go-nuget/internal/consts"
	"github.com/huhouhua/go-nuget/internal/meta"
)

type PackageArchiveReader struct {
	nuspec     *meta.Nuspec
	buf        *bytes.Buffer
	archive    *zip.Reader
	nuspecFile io.ReadCloser
	once       sync.Once
}

func NewPackageArchiveReader(reader io.Reader) (*PackageArchiveReader, error) {
	p := &PackageArchiveReader{
		buf: &bytes.Buffer{},
	}
	if _, err := p.buf.ReadFrom(reader); err != nil {
		return nil, err
	}
	if err := p.parse(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *PackageArchiveReader) parse() error {
	if p.buf == nil || p.buf.Len() == 0 {
		return fmt.Errorf("package is empty")
	}
	// Create a zip reader from the buffer
	r := p.buf.Bytes()
	archive := bytes.NewReader(r)
	var err error
	if p.archive, err = zip.NewReader(archive, int64(len(r))); err != nil {
		return err
	}
	// Extract the nuspec file
	if p.nuspecFile, err = p.findNuspecFile(); err != nil {
		return err
	}
	return nil
}

func (p *PackageArchiveReader) Nuspec() (*meta.Nuspec, error) {
	if p.nuspec != nil {
		return p.nuspec, nil
	}
	var err error
	p.once.Do(func() {
		defer func() {
			_ = p.nuspecFile.Close()
		}()
		// Reader the XML content into the Nuspec struct
		p.nuspec, err = meta.FromReader(p.nuspecFile)
	})

	return p.nuspec, err
}

func (p *PackageArchiveReader) GetFiles() []*zip.File {
	return p.archive.File
}

func (p *PackageArchiveReader) GetFilesFromDir(folder string) []*zip.File {
	files := make([]*zip.File, 0)
	prefix := strings.ToLower(folder + "/")
	for _, file := range p.GetFiles() {
		if strings.HasPrefix(strings.ToLower(file.Name), prefix) {
			files = append(files, file)
		}
	}
	return files
}

func (p *PackageArchiveReader) findNuspecFile() (io.ReadCloser, error) {
	for _, file := range p.archive.File {
		if strings.HasSuffix(file.Name, consts.NuspecExtension) {
			if nuspecFile, err := file.Open(); err != nil {
				return nil, err
			} else {
				return nuspecFile, nil
			}
		}
	}
	return nil, fmt.Errorf("no .nuspec file found in the .nupkg archive")
}
