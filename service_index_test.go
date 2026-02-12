// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

func TestServiceResource_GetIndex(t *testing.T) {
	_, client := setup(t, index_V3)

	data, err := os.ReadFile(index_V3)
	require.NoError(t, err)
	require.NotNil(t, data)

	var want ServiceIndex
	err = json.Unmarshal(data, &want)
	require.NoError(t, err)

	index, resp, err := client.IndexResource.GetIndex()
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, &want, index)
}

func TestServiceResource_GetIndexRequestOptions(t *testing.T) {
	_, client := setup(t, index_V3)
	_, _, err := client.IndexResource.GetIndex(func(request *retryablehttp.Request) error {
		return fmt.Errorf("test requestOptions error")
	})
	require.NotNil(t, err)
	require.Equal(t, err.Error(), "test requestOptions error")
}

func TestServiceResource_GetIndexResponseFatal(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/v3/index.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{ "error": "test request error" }`))
		require.NoError(t, err)
	})

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	wantError := `{error: test request error}`

	sourceURL := fmt.Sprintf("%s/v3/index.json", server.URL)
	_, err := NewClient(WithSourceURL(sourceURL))
	require.NotNil(t, err)

	var errResp *ErrorResponse
	require.True(t, errors.As(err, &errResp))
	require.Equal(t, wantError, errResp.Message)
}
