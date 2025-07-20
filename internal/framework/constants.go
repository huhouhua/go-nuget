// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package framework

import (
	"github.com/huhouhua/go-nuget/internal/consts"
	"github.com/huhouhua/go-nuget/version"
)

const (
	Any         = "Any"
	Agnostic    = "Agnostic"
	Unsupported = "Unsupported"
)

var (
	// Net2 is net20 (.NETFramework,Version=v2.0)
	Net2 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(2, 0, 0, "", ""))
	// Net35 is net35 (.NETFramework,Version=v3.5)
	Net35 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(3, 5, 0, "", ""))
	// Net4 is net40 (.NETFramework,Version=v4.0)
	Net4 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 0, 0, "", ""))
	// Net403 is net403 (.NETFramework,Version=v4.0.3)
	Net403 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 0, 3, "", ""))
	// Net45 is net45 (.NETFramework,Version=v4.5)
	Net45 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 5, 0, "", ""))
	// Net451 is net451 (.NETFramework,Version=v4.5.1)
	Net451 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 5, 1, "", ""))
	// Net452 is net452 (.NETFramework,Version=v4.5.2)
	Net452 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 5, 2, "", ""))
	// Net46 is net46 (.NETFramework,Version=v4.6)
	Net46 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 6, 0, "", ""))
	// Net461 is net461 (.NETFramework,Version=v4.6.1)
	Net461 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 6, 1, "", ""))
	// Net462 is net462 (.NETFramework,Version=v4.6.2)
	Net462 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 6, 2, "", ""))
	// Net463 is net463 (.NETFramework,Version=v4.6.3)
	Net463 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 6, 3, "", ""))
	// Net47 is net47 (.NETFramework,Version=v4.7)
	Net47 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 7, 0, "", ""))
	// Net471 is net471 (.NETFramework,Version=v4.7.1)
	Net471 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 7, 1, "", ""))
	// Net472 is net472 (.NETFramework,Version=v4.7.2)
	Net472 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 7, 2, "", ""))
	// Net48 is net48 (.NETFramework,Version=v4.8)
	Net48 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 8, 0, "", ""))
	// Net481 is net481 (.NETFramework,Version=v4.8.1)
	Net481 = NewFrameworkWithVersion(consts.Net, version.NewVersionFrom(4, 8, 1, "", ""))

	// NetCore45 is netcore45 (.NETCore,Version=v4.5)
	// This is not .NET Core. You are probably looking for netcoreapp.
	NetCore45 = NewFrameworkWithVersion(consts.NetCore, version.NewVersionFrom(4, 5, 0, "", ""))
	// NetCore451 is netcore451 (.NETCore,Version=v4.5.1)
	// This is not .NET Core. You are probably looking for netcoreapp.
	NetCore451 = NewFrameworkWithVersion(consts.NetCore, version.NewVersionFrom(4, 5, 1, "", ""))
	// NetCore50 is netcore50 (.NETCore,Version=v5.0)
	// This is not .NET 5. You are probably looking for net50 (.NETCoreApp,Version=v5.0)
	NetCore50 = NewFrameworkWithVersion(consts.NetCore, version.NewVersionFrom(5, 0, 0, "", ""))

	Win8  = NewFrameworkWithVersion(consts.Windows, version.NewVersionFrom(8, 0, 0, "", ""))
	Win81 = NewFrameworkWithVersion(consts.Windows, version.NewVersionFrom(8, 1, 0, "", ""))
	Win10 = NewFrameworkWithVersion(consts.Windows, version.NewVersionFrom(10, 0, 0, "", ""))

	SL4 = NewFrameworkWithVersion(consts.Silverlight, version.NewVersionFrom(4, 0, 0, "", ""))
	SL5 = NewFrameworkWithVersion(consts.Silverlight, version.NewVersionFrom(5, 0, 0, "", ""))

	WP7   = NewFrameworkWithVersion(consts.WindowsPhone, version.NewVersionFrom(7, 0, 0, "", ""))
	WP75  = NewFrameworkWithVersion(consts.WindowsPhone, version.NewVersionFrom(7, 5, 0, "", ""))
	WP8   = NewFrameworkWithVersion(consts.WindowsPhone, version.NewVersionFrom(8, 0, 0, "", ""))
	WP81  = NewFrameworkWithVersion(consts.WindowsPhone, version.NewVersionFrom(8, 1, 0, "", ""))
	WPA81 = NewFrameworkWithVersion(consts.WindowsPhoneApp, version.NewVersionFrom(8, 1, 0, "", ""))

	Tizen3 = NewFrameworkWithVersion(consts.Tizen, version.NewVersionFrom(3, 0, 0, "", ""))
	Tizen4 = NewFrameworkWithVersion(consts.Tizen, version.NewVersionFrom(4, 0, 0, "", ""))
	Tizen6 = NewFrameworkWithVersion(consts.Tizen, version.NewVersionFrom(6, 0, 0, "", ""))

	AspNet       = NewFrameworkWithVersion(consts.AspNet, consts.EmptyVersion)
	AspNetCore   = NewFrameworkWithVersion(consts.AspNetCore, consts.EmptyVersion)
	AspNet50     = NewFrameworkWithVersion(consts.AspNet, consts.Version5)
	AspNetCore50 = NewFrameworkWithVersion(consts.AspNetCore, consts.Version5)

	Dnx       = NewFrameworkWithVersion(consts.Dnx, consts.EmptyVersion)
	Dnx45     = NewFrameworkWithVersion(consts.Dnx, version.NewVersionFrom(4, 5, 0, "", ""))
	Dnx452    = NewFrameworkWithVersion(consts.Dnx, version.NewVersionFrom(4, 5, 2, "", ""))
	DnxCore   = NewFrameworkWithVersion(consts.DnxCore, consts.EmptyVersion)
	DnxCore50 = NewFrameworkWithVersion(consts.DnxCore, consts.Version5)

	DotNet   = NewFrameworkWithVersion(consts.NetPlatform, consts.EmptyVersion)
	DotNet50 = NewFrameworkWithVersion(consts.NetPlatform, consts.Version5)

	DotNet51 = NewFrameworkWithVersion(consts.NetPlatform, version.NewVersionFrom(5, 1, 0, "", ""))
	DotNet52 = NewFrameworkWithVersion(consts.NetPlatform, version.NewVersionFrom(5, 2, 0, "", ""))
	DotNet53 = NewFrameworkWithVersion(consts.NetPlatform, version.NewVersionFrom(5, 3, 0, "", ""))
	DotNet54 = NewFrameworkWithVersion(consts.NetPlatform, version.NewVersionFrom(5, 4, 0, "", ""))
	DotNet55 = NewFrameworkWithVersion(consts.NetPlatform, version.NewVersionFrom(5, 5, 0, "", ""))
	DotNet56 = NewFrameworkWithVersion(consts.NetPlatform, version.NewVersionFrom(5, 6, 0, "", ""))

	NetStandard = NewFrameworkWithVersion(consts.NetStandard, consts.EmptyVersion)
	// NetStandard10 is netstandard1.0 (.NETStandard,Version=v1.0)
	NetStandard10 = NewFrameworkWithVersion(consts.NetStandard, version.NewVersionFrom(1, 0, 0, "", ""))
	// NetStandard11 is netstandard1.1 (.NETStandard,Version=v1.1)
	NetStandard11 = NewFrameworkWithVersion(consts.NetStandard, version.NewVersionFrom(1, 1, 0, "", ""))
	// NetStandard12 is netstandard1.2 (.NETStandard,Version=v1.2)
	NetStandard12 = NewFrameworkWithVersion(consts.NetStandard, version.NewVersionFrom(1, 2, 0, "", ""))
	// NetStandard13 is netstandard1.3 (.NETStandard,Version=v1.3)
	NetStandard13 = NewFrameworkWithVersion(consts.NetStandard, version.NewVersionFrom(1, 3, 0, "", ""))
	// NetStandard14 is netstandard1.4 (.NETStandard,Version=v1.4)
	NetStandard14 = NewFrameworkWithVersion(consts.NetStandard, version.NewVersionFrom(1, 4, 0, "", ""))
	// NetStandard15 is netstandard1.5 (.NETStandard,Version=v1.5)
	NetStandard15 = NewFrameworkWithVersion(consts.NetStandard, version.NewVersionFrom(1, 5, 0, "", ""))
	// NetStandard16 is netstandard1.6 (.NETStandard,Version=v1.6)
	NetStandard16 = NewFrameworkWithVersion(consts.NetStandard, version.NewVersionFrom(1, 6, 0, "", ""))
	// NetStandard17 is netstandard1.7 (.NETStandard,Version=v1.7
	NetStandard17 = NewFrameworkWithVersion(consts.NetStandard, version.NewVersionFrom(1, 7, 0, "", ""))
	// NetStandard20 is netstandard2.0 (.NETStandard,Version=v2.0)
	NetStandard20 = NewFrameworkWithVersion(consts.NetStandard, version.NewVersionFrom(2, 0, 0, "", ""))
	// NetStandard21 is netstandard2.1 (.NETStandard,Version=v2.1)
	NetStandard21 = NewFrameworkWithVersion(consts.NetStandard, version.NewVersionFrom(2, 1, 0, "", ""))

	NetStandardApp15 = NewFrameworkWithVersion(consts.NetStandardApp, version.NewVersionFrom(1, 5, 0, "", ""))

	UAP10 = NewFrameworkWithVersion(consts.UAP, consts.Version10)

	// NetCoreApp10 is netcoreapp1.0 (.NETCoreApp,Version=v1.0)
	NetCoreApp10 = NewFrameworkWithVersion(consts.NetCoreApp, version.NewVersionFrom(1, 0, 0, "", ""))
	// NetCoreApp11 is netcoreapp1.1 (.NETCoreApp,Version=v1.1)
	NetCoreApp11 = NewFrameworkWithVersion(consts.NetCoreApp, version.NewVersionFrom(1, 1, 0, "", ""))
	// NetCoreApp20 is netcoreapp2.0 (.NETCoreApp,Version=v2.0)
	NetCoreApp20 = NewFrameworkWithVersion(consts.NetCoreApp, version.NewVersionFrom(2, 0, 0, "", ""))
	// NetCoreApp21 is netcoreapp2.1 (.NETCoreApp,Version=v2.1)
	NetCoreApp21 = NewFrameworkWithVersion(consts.NetCoreApp, version.NewVersionFrom(2, 1, 0, "", ""))
	// NetCoreApp22 is netcoreapp2.2 (.NETCoreApp,Version=v2.2)
	NetCoreApp22 = NewFrameworkWithVersion(consts.NetCoreApp, version.NewVersionFrom(2, 2, 0, "", ""))
	// NetCoreApp30 is netcoreapp3.0 (.NETCoreApp,Version=v3.0)
	NetCoreApp30 = NewFrameworkWithVersion(consts.NetCoreApp, version.NewVersionFrom(3, 0, 0, "", ""))
	// NetCoreApp31 is netcoreapp3.1 (.NETCoreApp,Version=v3.1)
	NetCoreApp31 = NewFrameworkWithVersion(consts.NetCoreApp, version.NewVersionFrom(3, 1, 0, "", ""))

	// Net50 .NET 5.0 and later has NetCoreApp identifier
	// Net50 is net5.0 (.NETCoreApp,Version=v5.0)
	Net50 = NewFrameworkWithVersion(consts.NetCoreApp, consts.Version5)
	// Net60 is net6.0 (.NETCoreApp,Version=v6.0)
	Net60 = NewFrameworkWithVersion(consts.NetCoreApp, consts.Version6)
	// Net70 is net7.0 (.NETCoreApp,Version=v7.0)
	Net70 = NewFrameworkWithVersion(consts.NetCoreApp, consts.Version7)
	// Net80 is net8.0 (.NETCoreApp,Version=v8.0)
	Net80 = NewFrameworkWithVersion(consts.NetCoreApp, consts.Version8)
	// Net90  is net9.0 (.NETCoreApp,Version=v9.0)
	Net90 = NewFrameworkWithVersion(consts.NetCoreApp, consts.Version9)
	// Net10_0  is net10.0 (.NETCoreApp,Version=v10.0)
	Net10_0 = NewFrameworkWithVersion(consts.NetCoreApp, consts.Version10)
	Native  = NewFrameworkWithVersion(consts.FrameworkNative, consts.EmptyVersion)
)
