// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"testing"
)

func TestGetVersionFromNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		want      int
	}{
		{
			name:      "V1 Schema",
			namespace: SchemaVersionV1,
			want:      1,
		},
		{
			name:      "V2 Schema",
			namespace: SchemaVersionV2,
			want:      2,
		},
		{
			name:      "V3 Schema",
			namespace: SchemaVersionV3,
			want:      3,
		},
		{
			name:      "V4 Schema",
			namespace: SchemaVersionV4,
			want:      4,
		},
		{
			name:      "V5 Schema",
			namespace: SchemaVersionV5,
			want:      5,
		},
		{
			name:      "V6 Schema",
			namespace: SchemaVersionV6,
			want:      6,
		},
		{
			name:      "Unknown Schema",
			namespace: "http://unknown.schema",
			want:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VersionToSchemaMaps.GetVersionFromNamespace(tt.namespace)
			if got != tt.want {
				t.Errorf("GetVersionFromNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSchemaNamespace(t *testing.T) {
	tests := []struct {
		name    string
		version int
		want    string
		wantErr bool
	}{
		{
			name:    "V1 Schema",
			version: 1,
			want:    SchemaVersionV1,
			wantErr: false,
		},
		{
			name:    "V2 Schema",
			version: 2,
			want:    SchemaVersionV2,
			wantErr: false,
		},
		{
			name:    "V3 Schema",
			version: 3,
			want:    SchemaVersionV3,
			wantErr: false,
		},
		{
			name:    "V4 Schema",
			version: 4,
			want:    SchemaVersionV4,
			wantErr: false,
		},
		{
			name:    "V5 Schema",
			version: 5,
			want:    SchemaVersionV5,
			wantErr: false,
		},
		{
			name:    "V6 Schema",
			version: 6,
			want:    SchemaVersionV6,
			wantErr: false,
		},
		{
			name:    "Invalid Version - Zero",
			version: 0,
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid Version - Negative",
			version: -1,
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid Version - Too High",
			version: 7,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VersionToSchemaMaps.GetSchemaNamespace(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSchemaNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetSchemaNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsKnownSchema(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		want      bool
	}{
		{
			name:      "V1 Schema",
			namespace: SchemaVersionV1,
			want:      true,
		},
		{
			name:      "V2 Schema",
			namespace: SchemaVersionV2,
			want:      true,
		},
		{
			name:      "V3 Schema",
			namespace: SchemaVersionV3,
			want:      true,
		},
		{
			name:      "V4 Schema",
			namespace: SchemaVersionV4,
			want:      true,
		},
		{
			name:      "V5 Schema",
			namespace: SchemaVersionV5,
			want:      true,
		},
		{
			name:      "V6 Schema",
			namespace: SchemaVersionV6,
			want:      true,
		},
		{
			name:      "Unknown Schema",
			namespace: "http://unknown.schema",
			want:      false,
		},
		{
			name:      "Empty Schema",
			namespace: "",
			want:      false,
		},
		{
			name:      "Case Insensitive V1",
			namespace: "HTTP://SCHEMAS.MICROSOFT.COM/PACKAGING/2010/07/NUSPEC.XSD",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VersionToSchemaMaps.IsKnownSchema(tt.namespace)
			if got != tt.want {
				t.Errorf("IsKnownSchema() = %v, want %v", got, tt.want)
			}
		})
	}
}
