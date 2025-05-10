// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/stretchr/testify/require"
)

func TestPackageUpdateResource_PushWithStream(t *testing.T) {
	mux, client := setup(t, index_V3)
	baseURL := client.getResourceUrl(PackagePublish)
	mux.HandleFunc(baseURL.Path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data;") {
			t.Fatalf(
				"PackageUpdateResource.PushWithStream request content-type %+v want multipart/form-data;",
				r.Header.Get("Content-Type"),
			)
		}
		if !strings.Contains(r.Header.Get("X-NuGet-Client-Version"), "4.1.0") {
			t.Fatalf(
				"PackageUpdateResource.PushWithStream request x-nuget-client-version %+v want 4.1.0",
				r.Header.Get("X-NuGet-Client-Version"),
			)
		}
		if r.ContentLength == -1 {
			t.Fatalf("PackageUpdateResource.PushWithStream request content-length is -1")
		}
		_, err := fmt.Fprint(w, `{}`)
		require.NoError(t, err)
	})

	var opt = &PushPackageOptions{
		TimeoutInDuration: time.Second * 5,
		SymbolSource:      "",
	}
	packageBuf := new(bytes.Buffer)
	_, err := client.UpdateResource.PushWithStream(packageBuf, opt)
	if err != nil {
		t.Fatalf("PackageUpdateResource.PushWithStream returns an error: %v", err)
	}
}

func TestPackageUpdateResource_Delete(t *testing.T) {
	mux, client := setup(t, index_V3)
	baseURL := client.getResourceUrl(PackagePublish)
	u := fmt.Sprintf("%s/%s/%s", baseURL.Path, PathEscape("newtonsoft.json"), PathEscape("1.0.0"))
	mux.HandleFunc(u, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})
	_, err := client.UpdateResource.Delete("newtonsoft.json", "1.0.0")
	if err != nil {
		t.Errorf("UpdateResource.Delete returned error: %v", err)
	}
}

func TestPackageUpdateResource_Push(t *testing.T) {
	tmpDir := t.TempDir()
	emptyPath := filepath.Join(tmpDir, "empty.nupkg")
	nupkgPath := filepath.Join(tmpDir, "mynuget.nupkg")
	createFile(t, emptyPath, "")
	createFile(t, nupkgPath, "TestPackageUpdateResource_Push allows apikey when pushing")

	defaultTimeOut := time.Second * 10
	tests := []struct {
		name        string
		opt         *PushPackageOptions
		packagePath string
		configFunc  func(client *Client, mux *http.ServeMux)
		error       error
	}{
		{
			name: "valid resource url",
			opt: &PushPackageOptions{
				TimeoutInDuration: defaultTimeOut,
			},
			configFunc: func(client *Client, mux *http.ServeMux) {
				u := createUrl(t, "http://abc")
				u.Scheme = ":"
				client.serviceUrls[PackagePublish] = u
			},
			error: &url.Error{
				Op:  "parse",
				URL: ":://abc",
				Err: errors.New("missing protocol scheme"),
			},
		}, {
			name: "valid symbolSource",
			opt: &PushPackageOptions{
				TimeoutInDuration: defaultTimeOut,
				SymbolSource:      "://abc",
			},
			error: &url.Error{
				Op:  "parse",
				URL: "://abc",
				Err: errors.New("missing protocol scheme"),
			},
		}, {
			name: "push request timeout for 5 millisecond",
			opt: &PushPackageOptions{
				TimeoutInDuration: time.Millisecond * 5,
			},
			packagePath: emptyPath,
			configFunc: func(client *Client, mux *http.ServeMux) {
				u := client.getResourceUrl(PackagePublish)
				require.NotNil(t, u)
				q := u.Query()
				q.Add("timeout_millisecond", strconv.FormatInt(int64(time.Millisecond*5), 10))
				u.RawQuery = q.Encode()
				client.serviceUrls[PackagePublish] = u
			},
			error: context.DeadlineExceeded,
		},
		{
			name: "push package empty",
			opt: &PushPackageOptions{
				TimeoutInDuration: defaultTimeOut,
			},
			packagePath: emptyPath,
			error:       errors.New("{error: package content size is 0}"),
		},
		{
			name: "allows apiKey when pushing",
			opt: &PushPackageOptions{
				TimeoutInDuration: defaultTimeOut,
			},
			packagePath: nupkgPath,
			error:       nil,
		},
		{
			name: "push with symbol package",
			opt: &PushPackageOptions{
				TimeoutInDuration: defaultTimeOut,
				SymbolSource:      "https://nuget.smbsrc.net/",
				IsSnupkg:          true,
			},
			configFunc: func(client *Client, mux *http.ServeMux) {
				baseURL := client.getResourceUrl(PackagePublish)
				wantKey := "0309f180-c810-45dd-bcae-9f0a94557abc"
				apiKeyEndpoint := fmt.Sprintf(TempApiKeyServiceEndpoint, "go.nuget.test", "1.0.0")

				path := fmt.Sprintf("%s/%s", baseURL.Path, apiKeyEndpoint)
				addTestVerificationApiKeyHandler(t, path, client.apiKey, wantKey, mux)
			},
			packagePath: "testdata/go.nuget.test.1.0.0.snupkg",
			error:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, client := setup(t, index_V3)
			require.NotNil(t, client)
			baseURL := client.getResourceUrl(PackagePublish)
			addTestUploadHandler(t, baseURL.Path, mux)
			if tt.configFunc != nil {
				tt.configFunc(client, mux)
			}
			_, err := client.UpdateResource.Push(tt.packagePath, tt.opt, func(request *retryablehttp.Request) error {
				request.URL.Scheme = "http"
				request.URL.Host = client.baseURL.Host
				request.Host = client.baseURL.Host
				return nil
			})
			var errResp *ErrorResponse
			if errors.As(err, &errResp) {
				require.Equal(t, tt.error.Error(), errResp.Message, "PackageUpdateResource.Push returns an error")
			} else {
				require.Equal(t, tt.error, err, "PackageUpdateResource.Push returns an error")
			}
		})
	}
}

func TestPushPackagePath(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)
	tests := []struct {
		name        string
		opt         *PushPackageOptions
		configFunc  func(client *Client)
		packagePath string
		error       error
	}{
		{
			name:        "directory does not exist",
			packagePath: "notfind/test",
			error: &fs.PathError{
				Op:   "lstat",
				Path: fmt.Sprintf("%s/notfind", dir),
				Err:  syscall.Errno(2),
			},
		},
		{
			name:        "url empty",
			packagePath: "",
			error:       errors.New("unable to find file "),
		},
		{
			name:        "api key empty",
			packagePath: "testdata/go.nuget.test.1.0.0.nupkg",
			configFunc: func(client *Client) {
				client.apiKey = ""
			},
			error: errors.New("api key is required"),
		},
		{
			name:        "not fund suffix .symbols.nupkg package",
			packagePath: "testdata/go.nuget.test.1.0.0.nupkg",
			opt: &PushPackageOptions{
				SymbolSource: "https://www.myget.org/F/nuget/api/v2/symbolpackage/",
				IsSnupkg:     false,
			},
			error: nil,
		},
		{
			name:        "push package to file system",
			packagePath: "testdata/go.nuget.test.1.0.0.snupkg",
			opt: &PushPackageOptions{
				SymbolSource: "file:///F/nuget/api/v2/symbolpackage/",
				IsSnupkg:     true,
			},
			error: errors.New("no support file system push"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, client := setup(t, index_V3)
			require.NotNil(t, client)

			baseURL := client.getResourceUrl(PackagePublish)
			addTestUploadHandler(t, baseURL.Path, mux)

			if tt.configFunc != nil {
				tt.configFunc(client)
			}

			packageUrl, err := client.UpdateResource.getResourceUrl(PackagePublish)
			require.NoError(t, err)
			symbolUrl := &url.URL{}
			if tt.opt != nil && tt.opt.SymbolSource != "" {
				symbolUrl, err = createSourceUri(tt.opt.SymbolSource)
				require.NoError(t, err)
				require.NotNil(t, symbolUrl)
			}
			_, err = client.UpdateResource.pushPackagePath(tt.opt, tt.packagePath, packageUrl, symbolUrl)
			require.Equal(t, tt.error, err, "PackageUpdateResource.pushPackagePath returns an error")
		})
	}
}

func TestPushWithSymbol(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)
	tests := []struct {
		name        string
		opt         *PushPackageOptions
		packagePath string
		error       error
	}{
		{
			name:        "directory does not exist",
			packagePath: "notfind/test",
			opt: &PushPackageOptions{
				SymbolSource: "https://www.myget.org/F/nuget/api/v2/symbolpackage/",
				IsSnupkg:     false,
			},
			error: &fs.PathError{
				Op:   "lstat",
				Path: fmt.Sprintf("%s/notfind", dir),
				Err:  syscall.Errno(2),
			},
		},
		{
			name:        "url empty",
			packagePath: "",
			opt: &PushPackageOptions{
				SymbolSource: "https://www.myget.org/F/nuget/api/v2/symbolpackage/",
			},
			error: errors.New("unable to find file "),
		},
		{
			name:        "api key empty",
			packagePath: "testdata/go.nuget.test.1.0.0.snupkg",
			opt: &PushPackageOptions{
				SymbolSource: "https://www.myget.org/F/nuget/api/v2/symbolpackage/",
				IsSnupkg:     true,
			},
			error: errors.New("{msg: api key is required}"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, client := setup(t, index_V3)
			require.NotNil(t, client)
			// test empty api key fail
			client.apiKey = ""

			symbolUrl := &url.URL{}
			if tt.opt != nil && tt.opt.SymbolSource != "" {
				symbolUrl, err = createSourceUri(tt.opt.SymbolSource)
				require.NoError(t, err)
				require.NotNil(t, symbolUrl)
			}

			addTestUploadHandler(t, strings.TrimRight(symbolUrl.Path, "/"), mux)

			_, err = client.UpdateResource.pushWithSymbol(
				tt.opt,
				tt.packagePath,
				symbolUrl,
				func(request *retryablehttp.Request) error {
					request.URL.Scheme = "http"
					request.URL.Host = client.baseURL.Host
					request.Host = client.baseURL.Host
					return nil
				},
			)
			var errResp *ErrorResponse
			if errors.As(err, &errResp) {
				require.Equal(t, tt.error.Error(), errResp.Message, "PackageUpdateResource.Push returns an error")
			} else {
				require.Equal(t, tt.error, err, "PackageUpdateResource.Push returns an error")
			}
		})
	}
}

func TestPushPackage(t *testing.T) {
	invalidSchemeUrlTemplate, err := url.Parse("https://www.myget.org/F/nuget/api/v2/symbolpackage/")
	require.NoError(t, err)
	invalidSchemeUrlTemplate.Scheme = "://abc"

	invalidUrlTemplate, err := url.Parse("https://www.myget.org/F/nuget/api/v2/symbolpackage/")
	require.NoError(t, err)
	invalidUrlTemplate.Path = invalidUrlTemplate.Path + "%eth"

	smbsrcUrl, err := url.Parse("https://nuget.smbsrc.net/")
	require.NoError(t, err)

	_, client := setup(t, index_V3)
	require.NotNil(t, client)

	tests := []struct {
		name        string
		sourceUrl   *url.URL
		packagePath string
		error       error
	}{
		{
			name:        "open not find file",
			packagePath: "notfind/file",
			error: &fs.PathError{
				Op:   "open",
				Path: "notfind/file",
				Err:  syscall.Errno(2),
			},
		},
		{
			name:        "endpoint url fail",
			sourceUrl:   invalidSchemeUrlTemplate,
			packagePath: "testdata/go.nuget.test.1.0.0.snupkg",
			error: &url.Error{
				Op:  "parse",
				URL: invalidSchemeUrlTemplate.String(),
				Err: errors.New("missing protocol scheme"),
			},
		},
		{
			name:        "uploadRequest file",
			sourceUrl:   invalidUrlTemplate,
			packagePath: "testdata/go.nuget.test.1.0.0.snupkg",
			error:       url.EscapeError("%et"),
		},
		{
			name:        "api interface does not exist",
			sourceUrl:   smbsrcUrl,
			packagePath: "testdata/go.nuget.test.1.0.0.snupkg",
			error:       errors.New("404 Not Found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			packageUrl := tt.sourceUrl
			if tt.sourceUrl == nil {
				packageUrl, err = client.UpdateResource.getResourceUrl(PackagePublish)
				require.NoError(t, err)
			}
			_, err = client.UpdateResource.push(
				tt.packagePath,
				packageUrl,
				func(request *retryablehttp.Request) error {
					request.URL.Scheme = "http"
					request.URL.Host = client.baseURL.Host
					request.Host = client.baseURL.Host
					return nil
				},
			)
			require.Equal(t, tt.error, err, "PackageUpdateResource.push returns an error")
		})
	}
}

func TestCreateVerificationApiKey(t *testing.T) {
	tmpDir := t.TempDir()
	emptyPackage := filepath.Join(tmpDir, "empty.nupkg")
	createFile(t, emptyPackage, "")
	tests := []struct {
		name             string
		packagePath      string
		handleConfigFunc func(client *Client, mux *http.ServeMux)
		wantApiKey       string
		error            error
	}{
		{
			name:        "open not find file",
			packagePath: "notfind/file",
			error: &fs.PathError{
				Op:   "open",
				Path: "notfind/file",
				Err:  syscall.Errno(2),
			},
		},
		{
			name:        "empty package",
			packagePath: emptyPackage,
			error:       errors.New("package is empty"),
		},
		{
			name:        "api interface does not exist",
			packagePath: "testdata/go.nuget.test.1.0.0.snupkg",
			error:       errors.New("404 Not Found"),
		},
		{
			name:        "create a apikey",
			packagePath: "testdata/go.nuget.test.1.0.0.snupkg",
			handleConfigFunc: func(client *Client, mux *http.ServeMux) {
				baseURL := client.getResourceUrl(PackagePublish)
				apiKeyEndpoint := fmt.Sprintf(TempApiKeyServiceEndpoint, "go.nuget.test", "1.0.0")
				path := fmt.Sprintf("%s/%s", baseURL.Path, apiKeyEndpoint)
				addTestVerificationApiKeyHandler(t, path, client.apiKey, "0309f180-c810-45dd-bcae-9f0a94557abc", mux)
			},
			wantApiKey: "0309f180-c810-45dd-bcae-9f0a94557abc",
			error:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, client := setup(t, index_V3)
			require.NotNil(t, client)
			if tt.handleConfigFunc != nil {
				tt.handleConfigFunc(client, mux)
			}
			key, err := client.UpdateResource.createVerificationApiKey(tt.packagePath, func(request *retryablehttp.Request) error {
				request.URL.Scheme = "http"
				request.URL.Host = client.baseURL.Host
				request.Host = client.baseURL.Host
				return nil
			})
			require.Equal(t, tt.error, err)
			require.Equal(t, tt.wantApiKey, key)
		})
	}
}

func addTestUploadHandler(t *testing.T, path string, mux *http.ServeMux) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		apiKey := r.Header.Get("X-NuGet-ApiKey")
		if apiKey == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, err := fmt.Fprint(w, `{"msg":"api key is required"}`)
			require.NoError(t, err)
			return
		}
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data;") {
			t.Fatalf(
				"PackageUpdateResource.Push request content-type %+v want multipart/form-data;",
				r.Header.Get("Content-Type"),
			)
		}
		if !strings.Contains(r.Header.Get("X-NuGet-Client-Version"), "4.1.0") {
			t.Fatalf(
				"PackageUpdateResource.Push request x-nuget-client-version %+v want 4.1.0",
				r.Header.Get("X-NuGet-Client-Version"),
			)
		}
		timeout := strings.TrimRight(r.URL.Query().Get("timeout_millisecond"), "/")
		if millisecond, err := strconv.Atoi(timeout); err == nil {
			time.Sleep(time.Duration(millisecond + int(time.Millisecond*5)))
		}
		if r.ContentLength == 248 {
			w.WriteHeader(http.StatusBadRequest)
			_, err := fmt.Fprint(w, `{ "error": "package content size is 0" }`)
			require.NoError(t, err)
			return
		}
		if r.ContentLength == -1 {
			t.Fatalf("PackageUpdateResource.Push request content-length is -1")
		}
		_, err := fmt.Fprint(w, `{}`)
		require.NoError(t, err)
	})
}
func addTestVerificationApiKeyHandler(t *testing.T, path, apiKey, wantKey string, mux *http.ServeMux) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		if !strings.Contains(r.Header.Get("X-NuGet-ApiKey"), apiKey) {
			t.Fatalf(
				"PackageUpdateResource.createVerificationApiKey request x-nuget-apikey %+v want %s",
				r.Header.Get("X-NuGet-ApiKey"),
				apiKey,
			)
		}

		if !strings.Contains(r.Header.Get("X-NuGet-Client-Version"), "4.1.0") {
			t.Fatalf(
				"PackageUpdateResource.createVerificationApiKey request x-nuget-client-version %+v want 4.1.0",
				r.Header.Get("X-NuGet-Client-Version"),
			)
		}
		if r.ContentLength == -1 {
			t.Fatalf("PackageUpdateResource.createVerificationApiKey request content-length is -1")
		}
		data := fmt.Sprintf(`{"Key":"%s","Expires":"2025-05-08T18:35:17.2531692Z"}`, wantKey)
		_, err := fmt.Fprint(w, data)
		require.NoError(t, err)
	})
}
