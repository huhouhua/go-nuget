// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package meta

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromFile(t *testing.T) {
	_, err := FromFile("non_existent_file.nuspec")
	require.Error(t, err, "expected error when file does not exist")

	nuspec, err := FromFile("../../testdata/myTestLibrary.nuspec")
	require.NoError(t, err)
	require.NotNil(t, nuspec)
}
func TestFromReader(t *testing.T) {
	_, err := FromReader(&errorReader{})
	if err == nil || !strings.Contains(err.Error(), "read error") {
		t.Fatal("expected read error")
	}
	nuspecFile, err := os.Open("../../testdata/myTestLibrary.nuspec")
	require.NoError(t, err)

	nuspec, err := FromReader(nuspecFile)
	require.NoError(t, err)
	require.NotNil(t, nuspec)
}

func TestFromBytes(t *testing.T) {
	t.Run("invalid xml", func(t *testing.T) {
		invalidXML := []byte("<invalid><xml>")
		_, err := FromBytes(invalidXML)
		if err == nil {
			t.Fatal("expected error for invalid XML")
		}
	})
	t.Run("empty input", func(t *testing.T) {
		_, err := FromBytes([]byte{})
		if err == nil {
			t.Fatal("expected error for empty input")
		}
	})
	t.Run("valid xml", func(t *testing.T) {
		validXML := []byte(`
		<package>
			<metadata>
				<id>TestPackage</id>
			</metadata>
		</package>`)

		nuspec, err := FromBytes(validXML)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if nuspec.Metadata.ID != "TestPackage" {
			t.Errorf("expected ID to be 'TestPackage', got '%s'", nuspec.Metadata.ID)
		}
	})
	t.Run("read return success", func(t *testing.T) {
		nuspecFile, err := os.Open("../../testdata/myTestLibrary.nuspec")
		require.NoError(t, err)

		nuspecBytes, err := io.ReadAll(nuspecFile)
		require.NoError(t, err)

		nuspec, err := FromBytes(nuspecBytes)
		require.NoError(t, err)
		require.NotNil(t, nuspec)
	})
}

func TestToBytes(t *testing.T) {
	nuspec, err := FromFile("../../testdata/myTestLibrary.nuspec")
	require.NoError(t, err)
	require.NotNil(t, nuspec)

	nuspecBytes, err := nuspec.ToBytes()
	require.NoError(t, err)
	require.NotNil(t, nuspecBytes)
}

type errorReader struct{}

func (e *errorReader) Close() error {
	return nil
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}
