// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import "github.com/huhouhua/go-nuget"

type Framework struct {
	// Framework Target framework
	Framework string

	// Version Target framework version
	Version nuget.NuGetVersion

	// Platform Framework Platform (net5.0+)
	Platform string

	// PlatformVersion Framework Platform Version (net5.0+)
	PlatformVersion nuget.NuGetVersion

	// TODO ShortFolderName the shortened version of the framework using the default mappings.
	ShortFolderName string

	// TODO IsUnsupported True if this framework was invalid or unknown. This framework is only compatible with Any and
	// Agnostic.
	IsUnsupported bool

	// TODO IsSpecificFramework True if this framework is real and not one of the special identifiers.
	IsSpecificFramework bool
}

// GetFrameworkString TODO
func (f *Framework) GetFrameworkString() string {
	return ""
}

type FrameworkAssemblyReference struct {
	AssemblyName        string
	SupportedFrameworks []*Framework
}
