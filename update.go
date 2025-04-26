// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PackageUpdateResource Contains logics to push or delete packages in Http server or file system
type PackageUpdateResource struct {
	client *Client
}

type PushOptions struct {
	PackagePaths []string `json:"packagePaths,omitempty"`

	SymbolSource string `json:"symbolSource,omitempty"`

	TimeoutInDuration time.Duration `json:"TimeoutInDuration"`

	DisableBuffering bool `json:"disableBuffering,omitempty"`

	NoServiceEndpoint bool `json:"noServiceEndpoint"`

	SkipDuplicate bool `json:"skipDuplicate"`

	IsSnupkg bool `json:"isSnupkg"`
}

func (p *PackageUpdateResource) Push(opt *PushOptions, options ...RequestOptionFunc) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opt.TimeoutInDuration)
	defer cancel()
	baseURL := p.client.getResourceUrl(PackagePublish)
	sourceUri, err := createSourceUri(baseURL.RawPath)
	if err != nil {
		return nil, err
	}
	for _, path := range opt.PackagePaths {
		if !strings.HasSuffix(path, SnupkgExtension) {
			resp, err := p.pushPackagePath(opt, path, sourceUri, options...)
			if err != nil {
				return resp, err
			}
		} else {
			// TODO: explicit snupkg push
			// symbolSource is only set when:
			// - The user specified it on the command line
			// - The endpoint for main package supports pushing snupkgs
		}

	}
	<-ctx.Done()
	return nil, nil
}

// pushPackagePath Push nupkgs, and if successful, push any corresponding symbols.
func (p *PackageUpdateResource) pushPackagePath(opt *PushOptions, path string, sourceUri *url.URL, options ...RequestOptionFunc) (*http.Response, error) {
	paths, err := resolvePackageFromPath(path, false)
	if err != nil {
		return nil, err
	}
	if paths == nil || len(paths) == 0 {
		return nil, fmt.Errorf("no packages found in %s", path)
	}

	if p.client.apiKey == "" && sourceUri.Scheme != "file" {
		return nil, fmt.Errorf("api key is required")
	}
	//var alreadyWarnedSymbolServerNotConfigured, warnForHttpSources = false, true
	for _, nupkgToPush := range paths {
		resp, err := p.pushPackageCore(nupkgToPush, sourceUri, opt, options...)
		if err != nil {
			return resp, err
		}
	}
	return nil, nil
}

func (p *PackageUpdateResource) pushPackageCore(packageToPush string, sourceUri *url.URL, opt *PushOptions, options ...RequestOptionFunc) (*http.Response, error) {
	sourceUri, err := createSourceUri(sourceUri.RawPath)
	if err != nil {
		return nil, err
	}
	log.Printf("push package %s to %s", filepath.Base(packageToPush), sourceUri.RawPath)

	if sourceUri.Scheme == "file" {
		// TODO: file system push
		return nil, nil
	}
	return p.pushPackageToServer(sourceUri, packageToPush, options...)
}

func (p *PackageUpdateResource) pushPackageToServer(sourceUri *url.URL, packageToPush string, options ...RequestOptionFunc) (*http.Response, error) {
	if isSourceNuGetSymbolServer(sourceUri) {
		// TODO: push to symbol server
		return nil, nil
	}
	return p.push(packageToPush, sourceUri.Path, options...)
}

// https://nuget.smbsrc.net/
func (p *PackageUpdateResource) pushWithSymbol() {

}
func (p *PackageUpdateResource) push(sourcePath, pathToPackage string, options ...RequestOptionFunc) (*http.Response, error) {
	file, err := os.Open(pathToPackage)
	if err != nil {
		return nil, err
	}
	req, err := p.client.UploadRequest(http.MethodPut, sourcePath, file, "package.nupkg", nil, options)
	if err != nil {
		return nil, err
	}
	return p.client.Do(req, nil, DecoderEmpty)
}
