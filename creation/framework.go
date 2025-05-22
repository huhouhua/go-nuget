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

	// the shortened version of the framework using the default mappings.
	ShortFolderName string
}

type FrameworkAssemblyReference struct {
	AssemblyName        string
	SupportedFrameworks []*Framework
}
