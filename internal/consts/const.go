// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package consts

import nugetVersion "github.com/huhouhua/go-nuget/version"

var (
	EmptyVersion = nugetVersion.NewVersionFrom(0, 0, 0, "", "")
	Version5     = nugetVersion.NewVersionFrom(5, 0, 0, "", "")
	Version6     = nugetVersion.NewVersionFrom(6, 0, 0, "", "")
	Version7     = nugetVersion.NewVersionFrom(7, 0, 0, "", "")
	Version8     = nugetVersion.NewVersionFrom(8, 0, 0, "", "")
	Version9     = nugetVersion.NewVersionFrom(9, 0, 0, "", "")
	Version10    = nugetVersion.NewVersionFrom(10, 0, 0, "", "")
)

type Folder string

const (
	Content             Folder = "content"
	Build               Folder = "build"
	BuildCrossTargeting Folder = "buildCrossTargeting"
	BuildTransitive     Folder = "buildTransitive"
	Tools               Folder = "tools"
	ContentFiles        Folder = "contentFiles"
	Lib                 Folder = "lib"
	Native              Folder = "native"
	Runtimes            Folder = "runtimes"
	Ref                 Folder = "ref"
	Analyzers           Folder = "analyzers"
	Source              Folder = "src"
)

var Known = []Folder{
	Content,
	Build,
	BuildCrossTargeting,
	BuildTransitive,
	Tools,
	ContentFiles,
	Lib,
	Native,
	Runtimes,
	Ref,
	Analyzers,
	Source,
}

const (
	NetCoreApp             = ".NETCoreApp"
	NetStandardApp         = ".NETStandardApp"
	NetStandard            = ".NETStandard"
	NetPlatform            = ".NETPlatform"
	Net                    = ".NETFramework"
	NetCore                = ".NETCore"
	WinRT                  = "WinRT"
	NetMicro               = ".NETMicroFramework"
	Portable               = ".NETPortable"
	WindowsPhone           = "WindowsPhone"
	Windows                = "Windows"
	WindowsPhoneApp        = "WindowsPhoneApp"
	Dnx                    = "DNX"
	DnxCore                = "DNXCore"
	AspNet                 = "ASP.NET"
	AspNetCore             = "ASP.NETCore"
	Silverlight            = "Silverlight"
	FrameworkNative        = "native"
	MonoAndroid            = "MonoAndroid"
	MonoTouch              = "MonoTouch"
	MonoMac                = "MonoMac"
	XamarinIOs             = "Xamarin.iOS"
	XamarinMac             = "Xamarin.Mac"
	XamarinPlayStation3    = "Xamarin.PlayStation3"
	XamarinPlayStation4    = "Xamarin.PlayStation4"
	XamarinPlayStationVita = "Xamarin.PlayStationVita"
	XamarinWatchOS         = "Xamarin.WatchOS"
	XamarinTVOS            = "Xamarin.TVOS"
	XamarinXbox360         = "Xamarin.Xbox360"
	XamarinXboxOne         = "Xamarin.XboxOne"
	UAP                    = "UAP"
	Tizen                  = "Tizen"
	NanoFramework          = ".NETnanoFramework"
)

const (
	PackageExtension    = ".nupkg"
	SnupkgExtension     = ".snupkg"
	NuspecExtension     = ".nuspec"
	SymbolsExtension    = ".symbols" + PackageExtension
	ReadmeExtension     = ".md"
	NuGetSymbolHostName = "nuget.smbsrc.net"
	ServiceEndpoint     = "/api/v2/package"
)
