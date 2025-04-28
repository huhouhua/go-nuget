// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// PackageUpdateResource Contains logics to push or delete packages in Http server or file system
type PackageUpdateResource struct {
	client *Client
}

type resultContext struct {
	Resp  *http.Response
	Error error
}

func (p *PackageUpdateResource) Delete(id, version string, options ...RequestOptionFunc) (*http.Response, error) {
	baseURL, err := p.getResourceUrl(PackagePublish)
	if err != nil {
		return nil, err
	}
	sourceUri, err := getServiceEndpointUrl(baseURL.String(), "", false)
	if err != nil {
		return nil, err
	}
	if sourceUri.Scheme == "file" {
		return nil, fmt.Errorf("no support file system delete")
	}
	u := fmt.Sprintf("%s/%s/%s", baseURL.Path, PathEscape(id), PathEscape(version))
	req, err := p.client.NewRequest(http.MethodDelete, u, nil, nil, options)
	if err != nil {
		return nil, err
	}
	return p.client.Do(req, nil, DecoderEmpty)
}

type PushPackageOptions struct {
	SymbolSource string `json:"symbolSource,omitempty"`

	TimeoutInDuration time.Duration `json:"TimeoutInDuration"`

	IsSnupkg bool `json:"isSnupkg"`
}

func (p *PackageUpdateResource) PushWithStream(packageStream io.Reader, opt *PushPackageOptions, options ...RequestOptionFunc) (*http.Response, error) {
	tempDir := os.TempDir()
	extension := PackageExtension
	if opt.IsSnupkg {
		extension = SnupkgExtension
	}
	millis := time.Now().UnixNano() / int64(time.Millisecond)
	fileName := fmt.Sprintf("%s%s", "package", extension)
	tempFilePath := filepath.Join(tempDir, "_nuget", strconv.FormatInt(millis, 10), fileName)

	err := os.MkdirAll(filepath.Dir(tempFilePath), 0755)
	if err != nil {
		return nil, err
	}
	fileInfo, err := os.Create(tempFilePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = fileInfo.Close()
		_ = os.RemoveAll(filepath.Dir(tempFilePath))
	}()
	if _, err = io.Copy(fileInfo, packageStream); err != nil {
		return nil, err
	}
	return p.Push([]string{tempFilePath}, opt, options...)
}

func (p *PackageUpdateResource) PushSingle(packagePath string, opt *PushPackageOptions, options ...RequestOptionFunc) (*http.Response, error) {
	return p.Push([]string{packagePath}, opt, options...)
}

func (p *PackageUpdateResource) Push(packagePaths []string, opt *PushPackageOptions, options ...RequestOptionFunc) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opt.TimeoutInDuration)
	defer cancel()
	resultChan := make(chan *resultContext)

	packageUrl, err := p.getResourceUrl(PackagePublish)
	if err != nil {
		return nil, err
	}
	symbolUrl := &url.URL{}
	if opt.SymbolSource != "" {
		symbolUrl, err = createSourceUri(opt.SymbolSource)
		if err != nil {
			return nil, err
		}
	}
	go func() {
		for _, path := range packagePaths {
			if !strings.HasSuffix(path, SnupkgExtension) {
				resp, err := p.pushPackagePath(opt, path, packageUrl, symbolUrl, options...)
				if err != nil {
					resultChan <- &resultContext{
						Resp:  resp,
						Error: err,
					}
				}
			} else {
				// symbolSource is only set when:
				// - The user specified it on the command line
				// - The endpoint for main package supports pushing snupkgs
				if strings.TrimSpace(opt.SymbolSource) != "" {
					resp, err := p.pushWithSymbol(opt, path, symbolUrl, options...)
					if err != nil {
						resultChan <- &resultContext{
							Resp:  resp,
							Error: err,
						}
					}
				}
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
func (p *PackageUpdateResource) pushPackagePath(opt *PushPackageOptions, path string, sourceUri, symbolUrl *url.URL, options ...RequestOptionFunc) (*http.Response, error) {
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
	for _, nupkgToPush := range paths {
		resp, err := p.pushPackageCore(nupkgToPush, sourceUri, opt, options...)
		if err != nil {
			return resp, err
		}
		// If the package was pushed successfully, push the symbol package.
		if strings.TrimSpace(opt.SymbolSource) == "" {
			continue
		}
		symbolPackagePath := GetSymbolsPath(nupkgToPush, opt.IsSnupkg)
		if _, err = os.Stat(symbolPackagePath); !os.IsNotExist(err) {
			continue
		}
		resp, err = p.pushPackageCore(symbolPackagePath, symbolUrl, opt, options...)
		if err != nil {
			return resp, err
		}
	}
	return nil, nil
}

func (p *PackageUpdateResource) pushPackageCore(packageToPush string, sourceUri *url.URL, opt *PushPackageOptions, options ...RequestOptionFunc) (*http.Response, error) {
	log.Printf("push package %s to %s", filepath.Base(packageToPush), sourceUri.Path)

	// TODO: file system push
	if sourceUri.Scheme == "file" {
		return nil, fmt.Errorf("no support file system push")
	}
	return p.push(packageToPush, sourceUri, options...)
}

// https://nuget.smbsrc.net/
func (p *PackageUpdateResource) pushWithSymbol(opt *PushPackageOptions, path string, symbolUrl *url.URL, options ...RequestOptionFunc) (*http.Response, error) {

	// Get the symbol package for this package
	symbolPackagePath := GetSymbolsPath(path, opt.IsSnupkg)

	paths, err := resolvePackageFromPath(symbolPackagePath, opt.IsSnupkg)
	if err != nil {
		return nil, err
	}
	// No files were resolved.
	if paths == nil || len(paths) == 0 {
		return nil, fmt.Errorf("unable to find file %s", path)
	}
	// See if the api key exists
	if p.client.apiKey == "" && symbolUrl.Scheme != "file" {
		log.Printf("warning symbol server not configured %s", filepath.Base(symbolPackagePath))
	}
	for _, packageToPush := range paths {
		resp, err := p.pushPackageCore(packageToPush, symbolUrl, opt, options...)
		if err != nil {
			return resp, err
		}
	}
	return nil, nil
}

func (p *PackageUpdateResource) push(pathToPackage string, sourceUrl *url.URL, options ...RequestOptionFunc) (*http.Response, error) {
	file, err := os.Open(pathToPackage)
	if err != nil {
		return nil, err
	}
	req, err := p.client.UploadRequest(http.MethodPut, sourceUrl.Path, sourceUrl, file, "package", "package.nupkg", nil, options)
	if err != nil {
		return nil, err
	}
	return p.client.Do(req, nil, DecoderEmpty)
}
