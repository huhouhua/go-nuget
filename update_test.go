// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPackageUpdateResource_AllowsApiKeyWhenPushing(t *testing.T) {
	mux, client := setup(t, index_V3)
	baseURL := client.getResourceUrl(PackagePublish)
	mux.HandleFunc(baseURL.Path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

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
		if r.ContentLength == -1 {
			t.Fatalf("PackageUpdateResource.Push request content-length is -1")
		}
		_, err := fmt.Fprint(w, `{}`)
		require.NoError(t, err)
	})
	tmpDir := t.TempDir()
	nupkgPath := filepath.Join(tmpDir, "mynuget.nupkg")
	createFile(t, nupkgPath, "TestPackageUpdateResource_AllowsApiKeyWhenPushing")

	var opt = &PushPackageOptions{
		TimeoutInDuration: time.Second * 5,
		SymbolSource:      "",
	}
	_, err := client.UpdateResource.Push([]string{nupkgPath}, opt)
	if err != nil {
		t.Fatalf("PackageUpdateResource.Push returns an error: %v", err)
	}
}

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
	createFile(t, emptyPath, "")

	defaultTimeOut := time.Second * 10
	tests := []struct {
		name         string
		opt          *PushPackageOptions
		packagePaths []string
		clientFunc   func(client *Client)
		error        error
	}{
		{
			name: "valid resource url",
			opt: &PushPackageOptions{
				TimeoutInDuration: defaultTimeOut,
			},
			clientFunc: func(client *Client) {
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
			packagePaths: []string{
				emptyPath,
			},
			clientFunc: func(client *Client) {
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
			packagePaths: []string{
				emptyPath,
			},
			error: errors.New("{error: package content size is 0}"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, client := setup(t, index_V3)
			require.NotNil(t, client)
			baseURL := client.getResourceUrl(PackagePublish)
			addTestUploadHandler(t, baseURL.Path, mux)
			if tt.clientFunc != nil {
				tt.clientFunc(client)
			}

			_, err := client.UpdateResource.Push(tt.packagePaths, tt.opt)
			var errResp *ErrorResponse
			if errors.As(err, &errResp) {
				require.Equal(t, tt.error.Error(), errResp.Message)
			} else {
				require.Equal(t, tt.error, err)
			}
		})
	}
}

func addTestUploadHandler(t *testing.T, path string, mux *http.ServeMux) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

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
		timeout := r.URL.Query().Get("timeout_millisecond")
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
