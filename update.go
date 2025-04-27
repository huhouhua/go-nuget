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

type resultContext struct {
	Resp  *http.Response
	Error error
}

func (p *PackageUpdateResource) Push(opt *PushOptions, options ...RequestOptionFunc) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opt.TimeoutInDuration)
	defer cancel()
	resultChan := make(chan *resultContext)

	packageUrl, err := p.getResourceUrl(PackagePublish)
	if err != nil {
		return nil, err
	}
	symbolUrl, err := p.getResourceUrl(SymbolPackagePublish)
	if err != nil {
		return nil, err
	}
	go func() {
		for _, path := range opt.PackagePaths {
			if !strings.HasSuffix(path, SnupkgExtension) {
				resp, err := p.pushPackagePath(opt, path, packageUrl, options...)
				if err != nil {
					resultChan <- &resultContext{
						Resp:  resp,
						Error: err,
					}
				}
			} else {
				// TODO: explicit snupkg push
				// symbolSource is only set when:
				// - The user specified it on the command line
				// - The endpoint for main package supports pushing snupkgs

				fmt.Println(symbolUrl.Host)
			}
		}
		// execution completed
		resultChan <- &resultContext{
			Resp:  nil,
			Error: nil,
		}
	}()

	select {
	case <-ctx.Done():
		// context deadline exceeded
		return nil, ctx.Err()
	case result := <-resultChan:
		return result.Resp, result.Error
	}
}

func (p *PackageUpdateResource) getResourceUrl(value ServiceType) (*url.URL, error) {
	baseURL := p.client.getResourceUrl(value)
	sourceUri, err := createSourceUri(baseURL.String())
	if err != nil {
		return nil, err
	}
	return sourceUri, nil
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
	//sourceUri, err := createSourceUri(sourceUri.String())
	//if err != nil {
	//	return nil, err
	//}

	log.Printf("push package %s to %s", filepath.Base(packageToPush), sourceUri.Path)

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
	return p.push(sourceUri.Path, packageToPush, options...)
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
