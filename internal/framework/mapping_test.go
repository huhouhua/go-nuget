// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package framework

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFactoryMapping(t *testing.T) {
	instance := GetProviderInstance()
	require.NotNil(t, instance)
	require.Equal(t, instance, GetProviderInstance())
}
