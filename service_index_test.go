// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"testing"
)

func TestServiceResource_GetIndex(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/v3/index.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/index.json")
	})

	data, err := os.ReadFile("testdata/index.json")
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
