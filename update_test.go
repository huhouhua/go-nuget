// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPackageUpdateResource_AllowsApiKeyWhenPushing(t *testing.T) {
	mux, client := setup(t, "testdata/index.json")
	baseURL := client.getResourceUrl(PackagePublish)
	mux.HandleFunc(baseURL.Path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data;") {
			t.Fatalf("PackageUpdateResource.Push request content-type %+v want multipart/form-data;", r.Header.Get("Content-Type"))
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
	mux, client := setup(t, "testdata/index.json")
	baseURL := client.getResourceUrl(PackagePublish)
	mux.HandleFunc(baseURL.Path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data;") {
			t.Fatalf("PackageUpdateResource.Push request content-type %+v want multipart/form-data;", r.Header.Get("Content-Type"))
		}
		if r.ContentLength == -1 {
			t.Fatalf("PackageUpdateResource.Push request content-length is -1")
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
	mux, client := setup(t, "testdata/index.json")
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
