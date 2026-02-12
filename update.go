// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
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

	"github.com/huhouhua/go-nuget/internal/consts"
	"github.com/huhouhua/go-nuget/internal/util"
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
	baseURL, err := p.getResourceURL(PackagePublish)
	if err != nil {
		return nil, err
	}
	sourceURL, err := util.GetServiceEndpointUrl(baseURL.String(), "", false)
	if err != nil {
		return nil, err
	}
	if sourceURL.Scheme == "file" {
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
	extension := consts.PackageExtension
	if opt.IsSnupkg {
		extension = consts.SnupkgExtension
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

// Push push the package to the server.
// please note that if the push is successful, it will take some time to see the successfully pushed package.
// pushed packages can only be soft deleted.
func (p *PackageUpdateResource) Push(
	packagePath string,
	opt *PushPackageOptions,
	options ...RequestOptionFunc,
) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opt.TimeoutInDuration)
	defer cancel()
	packageURL, err := p.getResourceURL(PackagePublish)
	if err != nil {
		return nil, err
	}
	symbolURL := &url.URL{}
	if opt.SymbolSource != "" {
		if symbolURL, err = util.CreateSourceURL(opt.SymbolSource); err != nil {
			return nil, err
		}
	}
	resultChan := make(chan *resultContext)
	go func() {
		defer close(resultChan)
		var resp *http.Response
		var err error
		if !strings.HasSuffix(packagePath, consts.SnupkgExtension) {
			resp, err = p.pushPackagePath(opt, packagePath, packageURL, symbolURL, options...)
		} else if strings.TrimSpace(opt.SymbolSource) != "" {
			// symbolSource is only set when:
			// - The user specified it on the command line
			// - The endpoint for main package supports pushing snupkgs
			resp, err = p.pushWithSymbol(opt, packagePath, symbolURL, options...)
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

// getResourceURL returns the resource URL for the given service type.
func (p *PackageUpdateResource) getResourceURL(value ServiceType) (*url.URL, error) {
	baseURL := p.client.getResourceURL(value)
	sourceURL, err := util.CreateSourceURL(baseURL.String())
	if err != nil {
		return nil, err
	}
	return sourceURL, nil
}

// pushPackagePath Push nupkgs, and if successful, push any corresponding symbols.
func (p *PackageUpdateResource) pushPackagePath(
	opt *PushPackageOptions,
	path string,
	sourceURL, symbolURL *url.URL,
	options ...RequestOptionFunc,
) (*http.Response, error) {
	paths, err := util.ResolvePackageFromPath(path, false)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("unable to find file %s", path)
	}

	if p.client.apiKey == "" && sourceURL.Scheme != "file" {
		return nil, fmt.Errorf("api key is required")
	}
	for _, nupkgToPush := range paths {
		if resp, err := p.pushPackageCore(nupkgToPush, sourceURL, options...); err != nil {
			return resp, err
		}
		// If the package was pushed successfully, push the symbol package.
		if strings.TrimSpace(opt.SymbolSource) == "" {
			continue
		}
		symbolPackagePath := util.GetSymbolsPath(nupkgToPush, opt.IsSnupkg)
		if _, err = os.Stat(symbolPackagePath); os.IsNotExist(err) {
			continue
		}
		if resp, err := p.pushPackageCore(symbolPackagePath, symbolURL, options...); err != nil {
			return resp, err
		}
	}
	return nil, nil
}

func (p *PackageUpdateResource) pushPackageCore(
	packageToPush string,
	sourceURL *url.URL,
	options ...RequestOptionFunc,
) (*http.Response, error) {
	log.Printf("push package %s to %s", filepath.Base(packageToPush), sourceURL.String())

	// TODO: file system push
	if sourceURL.Scheme == "file" {
		return nil, fmt.Errorf("no support file system push")
	}
	return p.push(packageToPush, sourceURL, options...)
}

// pushWithSymbol handle push to https://nuget.smbsrc.net/
func (p *PackageUpdateResource) pushWithSymbol(
	opt *PushPackageOptions,
	path string,
	symbolURL *url.URL,
	options ...RequestOptionFunc,
) (*http.Response, error) {
	// Get the symbol package for this package
	symbolPackagePath := util.GetSymbolsPath(path, opt.IsSnupkg)

	paths, err := util.ResolvePackageFromPath(symbolPackagePath, opt.IsSnupkg)
	if err != nil {
		return nil, err
	}
	// No files were resolved.
	if len(paths) == 0 {
		return nil, fmt.Errorf("unable to find file %s", path)
	}
	// See if the api key exists
	if p.client.apiKey == "" && symbolURL.Scheme != "file" {
		log.Printf("warning symbol server not configured %s", filepath.Base(symbolPackagePath))
	}
	for _, packageToPush := range paths {
		if resp, err := p.pushPackageCore(packageToPush, symbolURL, options...); err != nil {
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
	endpointURL, err := util.GetServiceEndpointUrl(DefaultGalleryServerURL, u, false)
	if err != nil {
		return "", err
	}
	req, err := p.client.NewRequest(http.MethodPost, endpointURL.Path, endpointURL, nil, options)
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
	sourceURL *url.URL,
	options ...RequestOptionFunc,
) (*http.Response, error) {
	file, err := os.Open(pathToPackage)
	if err != nil {
		return nil, err
	}
	endpointURL, err := util.GetServiceEndpointUrl(sourceURL.String(), "", false)
	if err != nil {
		return nil, err
	}
	req, err := p.client.UploadRequest(
		http.MethodPut,
		endpointURL.Path,
		endpointURL,
		file,
		"package",
		"package.nupkg",
		nil,
		options,
	)
	if err != nil {
		return nil, err
	}
	if util.IsSourceNuGetSymbolServer(sourceURL) {
		if key, err := p.createVerificationApiKey(pathToPackage, options...); err != nil {
			return nil, err
		} else {
			req.Header.Add("X-NuGet-ApiKey", key)
		}
	}
	return p.client.Do(req, nil, DecoderEmpty)
}
