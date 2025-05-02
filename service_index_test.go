// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"encoding/json"
	"os"
	"testing"

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
