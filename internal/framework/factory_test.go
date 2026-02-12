// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package framework

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDeprecatedFramework(t *testing.T) {
	tests := []struct {
		input string
		want  *Framework
	}{
		{
			input: "45",
			want:  Net45,
		},
		{
			input: "4.5",
			want:  Net45,
		},
		{
			input: "40",
			want:  Net4,
		},
		{
			input: "4.0",
			want:  Net4,
		},
		{
			input: "4",
			want:  Net4,
		},
		{
			input: "35",
			want:  Net35,
		},
		{
			input: "3.5",
			want:  Net35,
		},
		{
			input: "20",
			want:  Net2,
		},
		{
			input: "2.0",
			want:  Net2,
		},
		{
			input: "2",
			want:  Net2,
		},
		{
			input: "",
			want:  nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			actual := parseDeprecatedFramework(tc.input)
			require.Equal(t, tc.want, actual)
		})
	}
}

func TestParseCommonFramework(t *testing.T) {
	tests := []struct {
		input string
		want  *Framework
	}{
		{input: "dotnet", want: DotNet50},
		{input: "dotnet50", want: DotNet50},
		{input: "dotnet5.0", want: DotNet50},
		{input: "net40", want: Net4},
		{input: "net4", want: Net4},
		{input: "net403", want: Net403},
		{input: "net45", want: Net45},
		{input: "net451", want: Net451},
		{input: "net452", want: Net452},
		{input: "net46", want: Net46},
		{input: "net461", want: Net461},
		{input: "net462", want: Net462},
		{input: "net463", want: Net463},
		{input: "net47", want: Net47},
		{input: "net471", want: Net471},
		{input: "net472", want: Net472},
		{input: "net48", want: Net48},
		{input: "net481", want: Net481},
		{input: "win8", want: Win8},
		{input: "win81", want: Win81},
		{input: "netstandard", want: NetStandard},
		{input: "netstandard1.0", want: NetStandard10},
		{input: "netstandard10", want: NetStandard10},
		{input: "netstandard1.1", want: NetStandard11},
		{input: "netstandard11", want: NetStandard11},
		{input: "netstandard1.2", want: NetStandard12},
		{input: "netstandard12", want: NetStandard12},
		{input: "netstandard1.3", want: NetStandard13},
		{input: "netstandard13", want: NetStandard13},
		{input: "netstandard1.4", want: NetStandard14},
		{input: "netstandard14", want: NetStandard14},
		{input: "netstandard1.5", want: NetStandard15},
		{input: "netstandard15", want: NetStandard15},
		{input: "netstandard1.6", want: NetStandard16},
		{input: "netstandard16", want: NetStandard16},
		{input: "netstandard1.7", want: NetStandard17},
		{input: "netstandard17", want: NetStandard17},
		{input: "netstandard2.0", want: NetStandard20},
		{input: "netstandard20", want: NetStandard20},
		{input: "netstandard2.1", want: NetStandard21},
		{input: "netstandard21", want: NetStandard21},
		{input: "netcoreapp1.0", want: NetCoreApp10},
		{input: "netcoreapp1.1", want: NetCoreApp11},
		{input: "netcoreapp2.0", want: NetCoreApp20},
		{input: "netcoreapp2.1", want: NetCoreApp21},
		{input: "netcoreapp21", want: NetCoreApp21},
		{input: "netcoreapp2.2", want: NetCoreApp22},
		{input: "netcoreapp3.0", want: NetCoreApp30},
		{input: "netcoreapp30", want: NetCoreApp30},
		{input: "netcoreapp3.1", want: NetCoreApp31},
		{input: "netcoreapp31", want: NetCoreApp31},
		{input: "netcoreapp5.0", want: Net50},
		{input: "netcoreapp50", want: Net50},
		{input: "net5.0", want: Net50},
		{input: "net50", want: Net50},
		{input: "netcoreapp6.0", want: Net60},
		{input: "netcoreapp60", want: Net60},
		{input: "net6.0", want: Net60},
		{input: "net60", want: Net60},
		{input: "netcoreapp7.0", want: Net70},
		{input: "netcoreapp70", want: Net70},
		{input: "net7.0", want: Net70},
		{input: "net70", want: Net70},
		{input: "netcoreapp8.0", want: Net80},
		{input: "netcoreapp80", want: Net80},
		{input: "net8.0", want: Net80},
		{input: "net80", want: Net80},
		{input: "net9.0", want: Net90},
		{input: "net10.0", want: Net10_0},
		{input: "", want: nil},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			actual := parseCommonFramework(tc.input)
			require.Equal(t, tc.want, actual)
		})
	}
}
