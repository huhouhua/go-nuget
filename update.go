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

// Delete deletes a package from the server.
// please note that this package can only be soft deleted
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
	req, err := p.client.NewRequest(http.MethodDelete, u, baseURL, nil, options)
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

// PushWithStream pushes a package stream to the server.
func (p *PackageUpdateResource) PushWithStream(
	packageStream io.Reader,
	opt *PushPackageOptions,
	options ...RequestOptionFunc,
) (*http.Response, error) {
	tempDir := os.TempDir()
	extension := PackageExtension
	if opt.IsSnupkg {
		extension = SnupkgExtension
	}
	millis := time.Now().UnixNano() / int64(time.Millisecond)
	fileName := fmt.Sprintf("%s%s", "package", extension)
	tempFilePath := filepath.Join(tempDir, "_nuget", strconv.FormatInt(millis, 10), fileName)

	if err := os.MkdirAll(filepath.Dir(tempFilePath), 0755); err != nil {
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
	return p.Push(tempFilePath, opt, options...)
}

// Push pushes a package to the server. It supports pushing multiple packages.
// please note that it takes a while to see the successfully pushed packages.
// the pushed packages can only be soft-deleted.
func (p *PackageUpdateResource) Push(
	packagePath string,
	opt *PushPackageOptions,
	options ...RequestOptionFunc,
) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opt.TimeoutInDuration)
	defer cancel()
	packageUrl, err := p.getResourceUrl(PackagePublish)
	if err != nil {
		return nil, err
	}
	symbolUrl := &url.URL{}
	if opt.SymbolSource != "" {
		if symbolUrl, err = createSourceUri(opt.SymbolSource); err != nil {
			return nil, err
		}
	}
	resultChan := make(chan *resultContext)
	go func() {
		defer close(resultChan)
		var resp *http.Response
		var err error
		if !strings.HasSuffix(packagePath, SnupkgExtension) {
			resp, err = p.pushPackagePath(opt, packagePath, packageUrl, symbolUrl, options...)
		} else if strings.TrimSpace(opt.SymbolSource) != "" {
			// symbolSource is only set when:
			// - The user specified it on the command line
			// - The endpoint for main package supports pushing snupkgs
			resp, err = p.pushWithSymbol(opt, packagePath, symbolUrl, options...)
		}
		resultChan <- &resultContext{
			Resp:  resp,
			Error: err,
		}
	}()

	for {
		select {
		// context deadline exceeded
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-resultChan:
			return result.Resp, result.Error
		}
	}
}

// getResourceUrl returns the resource URL for the given service type.
func (p *PackageUpdateResource) getResourceUrl(value ServiceType) (*url.URL, error) {
	baseURL := p.client.getResourceUrl(value)
	sourceUri, err := createSourceUri(baseURL.String())
	if err != nil {
		return nil, err
	}
	return sourceUri, nil
}

// pushPackagePath Push nupkgs, and if successful, push any corresponding symbols.
func (p *PackageUpdateResource) pushPackagePath(
	opt *PushPackageOptions,
	path string,
	sourceUri, symbolUrl *url.URL,
	options ...RequestOptionFunc,
) (*http.Response, error) {
	paths, err := resolvePackageFromPath(path, false)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no packages found in %s", path)
	}

	if p.client.apiKey == "" && sourceUri.Scheme != "file" {
		return nil, fmt.Errorf("api key is required")
	}
	for _, nupkgToPush := range paths {
		if resp, err := p.pushPackageCore(nupkgToPush, sourceUri, opt, options...); err != nil {
			return resp, err
		}
		// If the package was pushed successfully, push the symbol package.
		if strings.TrimSpace(opt.SymbolSource) == "" {
			continue
		}
		symbolPackagePath := GetSymbolsPath(nupkgToPush, opt.IsSnupkg)
		if _, err = os.Stat(symbolPackagePath); os.IsNotExist(err) {
			continue
		}
		if resp, err := p.pushPackageCore(symbolPackagePath, symbolUrl, opt, options...); err != nil {
			return resp, err
		}
	}
	return nil, nil
}

func (p *PackageUpdateResource) pushPackageCore(
	packageToPush string,
	sourceUri *url.URL,
	opt *PushPackageOptions,
	options ...RequestOptionFunc,
) (*http.Response, error) {
	log.Printf("push package %s to %s", filepath.Base(packageToPush), sourceUri.Path)

	// TODO: file system push
	if sourceUri.Scheme == "file" {
		return nil, fmt.Errorf("no support file system push")
	}
	return p.push(packageToPush, sourceUri, options...)
}

// pushWithSymbol handle push to https://nuget.smbsrc.net/
func (p *PackageUpdateResource) pushWithSymbol(
	opt *PushPackageOptions,
	path string,
	symbolUrl *url.URL,
	options ...RequestOptionFunc,
) (*http.Response, error) {
	// Get the symbol package for this package
	symbolPackagePath := GetSymbolsPath(path, opt.IsSnupkg)

	paths, err := resolvePackageFromPath(symbolPackagePath, opt.IsSnupkg)
	if err != nil {
		return nil, err
	}
	// No files were resolved.
	if len(paths) == 0 {
		return nil, fmt.Errorf("unable to find file %s", path)
	}
	// See if the api key exists
	if p.client.apiKey == "" && symbolUrl.Scheme != "file" {
		log.Printf("warning symbol server not configured %s", filepath.Base(symbolPackagePath))
	}
	for _, packageToPush := range paths {
		if resp, err := p.pushPackageCore(packageToPush, symbolUrl, opt, options...); err != nil {
			return resp, err
		}
	}
	return nil, nil
}

// createVerificationApiKey Get a temp API key from nuget.org for pushing to https://nuget.smbsrc.net/
func (p *PackageUpdateResource) createVerificationApiKey(
	pathToPackage string,
	options ...RequestOptionFunc,
) (string, error) {
	packageFile, err := os.Open(pathToPackage)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = packageFile.Close()
	}()
	reader, err := NewPackageArchiveReader(packageFile)
	if err != nil {
		return "", err
	}
	nuspec, err := reader.Nuspec()
	if err != nil {
		return "", err
	}
	u := fmt.Sprintf(TempApiKeyServiceEndpoint, nuspec.Metadata.ID, nuspec.Metadata.Version)
	sourceUri, err := getServiceEndpointUrl(DefaultGalleryServerUrl, u, false)
	if err != nil {
		return "", err
	}
	req, err := p.client.NewRequest(http.MethodPost, sourceUri.Path, sourceUri, nil, options)
	if err != nil {
		return "", err
	}
	// Execute request
	var keyMap map[string]string
	if _, err = p.client.Do(req, &keyMap, DecoderTypeJSON); err != nil {
		return "", err
	}
	return keyMap["Key"], nil
}

// push pushes a package to the server.
func (p *PackageUpdateResource) push(
	pathToPackage string,
	sourceUrl *url.URL,
	options ...RequestOptionFunc,
) (*http.Response, error) {
	file, err := os.Open(pathToPackage)
	if err != nil {
		return nil, err
	}
	endpointUrl, err := getServiceEndpointUrl(sourceUrl.String(), "", false)
	if err != nil {
		return nil, err
	}
	req, err := p.client.UploadRequest(
		http.MethodPut,
		endpointUrl.Path,
		endpointUrl,
		file,
		"package",
		"package.nupkg",
		nil,
		options,
	)
	if err != nil {
		return nil, err
	}
	if isSourceNuGetSymbolServer(sourceUrl) {
		if key, err := p.createVerificationApiKey(pathToPackage, options...); err != nil {
			return nil, err
		} else {
			req.Header.Add("X-NuGet-ApiKey", key)
		}
	}
	return p.client.Do(req, nil, DecoderEmpty)
}
