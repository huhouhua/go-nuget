// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

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
	Nano                   = ".NETnanoFramework"
)

//var (
//	Net2   = NewFrameworkWithVersion(Net, &nuget.NuGetVersion{Version: semver.New(2, 0, 0, "", "")})
//	Net35  = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(3, 5, 0, "", "")})
//	Net4   = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 0, 0, "", "")})
//	Net403 = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 0, 3, "", "")})
//	Net45  = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 5, 0, "", "")})
//	Net451 = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 5, 1, "", "")})
//	Net452 = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 5, 2, "", "")})
//	Net46  = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 6, 0, "", "")})
//	Net461 = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 6, 1, "", "")})
//	Net462 = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 6, 2, "", "")})
//	Net463 = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 6, 3, "", "")})
//	Net47  = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 7, 0, "", "")})
//	Net471 = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 7, 1, "", "")})
//	Net472 = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 7, 2, "", "")})
//	Net48  = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 8, 0, "", "")})
//	Net481 = NewFrameworkWithVersion(nuget.Net, &nuget.NuGetVersion{Version: semver.New(4, 8, 1, "", "")})
//
//	NetCore45  = NewFrameworkWithVersion(nuget.NetCore, &nuget.NuGetVersion{Version: semver.New(4, 5, 0, "", "")})
//	NetCore451 = NewFrameworkWithVersion(nuget.NetCore, &nuget.NuGetVersion{Version: semver.New(4, 5, 1, "", "")})
//	NetCore50  = NewFrameworkWithVersion(nuget.NetCore, &nuget.NuGetVersion{Version: semver.New(5, 0, 0, "", "")})
//
//	Win8  = NewFrameworkWithVersion(nuget.Windows, &nuget.NuGetVersion{Version: semver.New(8, 0, 0, "", "")})
//	Win81 = NewFrameworkWithVersion(nuget.Windows, &nuget.NuGetVersion{Version: semver.New(8, 1, 0, "", "")})
//	Win10 = NewFrameworkWithVersion(nuget.Windows, &nuget.NuGetVersion{Version: semver.New(10, 0, 0, "", "")})
//
//	SL4 = NewFrameworkWithVersion(nuget.Silverlight, &nuget.NuGetVersion{Version: semver.New(4, 0, 0, "", "")})
//	SL5 = NewFrameworkWithVersion(nuget.Silverlight, &nuget.NuGetVersion{Version: semver.New(5, 0, 0, "", "")})
//
//	WP7   = NewFrameworkWithVersion(nuget.WindowsPhone, &nuget.NuGetVersion{Version: semver.New(7, 0, 0, "", "")})
//	WP75  = NewFrameworkWithVersion(nuget.WindowsPhone, &nuget.NuGetVersion{Version: semver.New(7, 5, 0, "", "")})
//	WP8   = NewFrameworkWithVersion(nuget.WindowsPhone, &nuget.NuGetVersion{Version: semver.New(8, 0, 0, "", "")})
//	WP81  = NewFrameworkWithVersion(nuget.WindowsPhone, &nuget.NuGetVersion{Version: semver.New(8, 1, 0, "", "")})
//	WPA81 = NewFrameworkWithVersion(nuget.WindowsPhoneApp, &nuget.NuGetVersion{Version: semver.New(8, 1, 0, "", "")})
//
//	Tizen3 = NewFrameworkWithVersion(nuget.Tizen, &nuget.NuGetVersion{Version: semver.New(3, 0, 0, "", "")})
//	Tizen4 = NewFrameworkWithVersion(nuget.Tizen, &nuget.NuGetVersion{Version: semver.New(4, 0, 0, "", "")})
//	Tizen6 = NewFrameworkWithVersion(nuget.Tizen, &nuget.NuGetVersion{Version: semver.New(6, 0, 0, "", "")})
//
//	AspNet       = NewFrameworkWithVersion(nuget.AspNet, &nuget.EmptyVersion)
//	AspNetCore   = NewFrameworkWithVersion(nuget.AspNetCore, &nuget.EmptyVersion)
//	AspNet50     = NewFrameworkWithVersion(nuget.AspNet, &nuget.Version5)
//	AspNetCore50 = NewFrameworkWithVersion(nuget.AspNetCore, &nuget.Version5)
//
//	Dnx       = NewFrameworkWithVersion(nuget.Dnx, &nuget.EmptyVersion)
//	Dnx45     = NewFrameworkWithVersion(nuget.Dnx, &nuget.NuGetVersion{Version: semver.New(4, 5, 0, "", "")})
//	Dnx452    = NewFrameworkWithVersion(nuget.Dnx, &nuget.NuGetVersion{Version: semver.New(4, 5, 2, "", "")})
//	DnxCore   = NewFrameworkWithVersion(nuget.DnxCore, &nuget.EmptyVersion)
//	DnxCore50 = NewFrameworkWithVersion(nuget.DnxCore, &nuget.Version5)
//
//	DotNet   = NewFrameworkWithVersion(nuget.NetPlatform, &nuget.EmptyVersion)
//	DotNet50 = NewFrameworkWithVersion(nuget.NetPlatform, &nuget.Version5)
//
//	DotNet51 = NewFrameworkWithVersion(nuget.NetPlatform, &nuget.NuGetVersion{Version: semver.New(5, 1, 0, "", "")})
//	DotNet52 = NewFrameworkWithVersion(nuget.NetPlatform, &nuget.NuGetVersion{Version: semver.New(5, 2, 0, "", "")})
//	DotNet53 = NewFrameworkWithVersion(nuget.NetPlatform, &nuget.NuGetVersion{Version: semver.New(5, 3, 0, "", "")})
//	DotNet54 = NewFrameworkWithVersion(nuget.NetPlatform, &nuget.NuGetVersion{Version: semver.New(5, 4, 0, "", "")})
//	DotNet55 = NewFrameworkWithVersion(nuget.NetPlatform, &nuget.NuGetVersion{Version: semver.New(5, 5, 0, "", "")})
//	DotNet56 = NewFrameworkWithVersion(nuget.NetPlatform, &nuget.NuGetVersion{Version: semver.New(5, 6, 0, "", "")})
//
//	NetStandard   = NewFrameworkWithVersion(nuget.NetStandard, &nuget.EmptyVersion)
//	NetStandard10 = NewFrameworkWithVersion(nuget.NetStandard, &nuget.NuGetVersion{Version: semver.New(1, 0, 0, "", "")})
//	NetStandard11 = NewFrameworkWithVersion(nuget.NetStandard, &nuget.NuGetVersion{Version: semver.New(1, 1, 0, "", "")})
//	NetStandard12 = NewFrameworkWithVersion(nuget.NetStandard, &nuget.NuGetVersion{Version: semver.New(1, 2, 0, "", "")})
//	NetStandard13 = NewFrameworkWithVersion(nuget.NetStandard, &nuget.NuGetVersion{Version: semver.New(1, 3, 0, "", "")})
//	NetStandard14 = NewFrameworkWithVersion(nuget.NetStandard, &nuget.NuGetVersion{Version: semver.New(1, 4, 0, "", "")})
//	NetStandard15 = NewFrameworkWithVersion(nuget.NetStandard, &nuget.NuGetVersion{Version: semver.New(1, 5, 0, "", "")})
//	NetStandard16 = NewFrameworkWithVersion(nuget.NetStandard, &nuget.NuGetVersion{Version: semver.New(1, 6, 0, "", "")})
//	NetStandard17 = NewFrameworkWithVersion(nuget.NetStandard, &nuget.NuGetVersion{Version: semver.New(1, 7, 0, "", "")})
//	NetStandard20 = NewFrameworkWithVersion(nuget.NetStandard, &nuget.NuGetVersion{Version: semver.New(2, 0, 0, "", "")})
//	NetStandard21 = NewFrameworkWithVersion(nuget.NetStandard, &nuget.NuGetVersion{Version: semver.New(2, 1, 0, "", "")})
//
//	NetStandardApp15 = NewFrameworkWithVersion(nuget.NetStandardApp, &nuget.NuGetVersion{Version: semver.New(1, 5, 0, "", "")})
//
//	UAP10 = NewFrameworkWithVersion(nuget.UAP, &nuget.Version10)
//
//	NetCoreApp10 = NewFrameworkWithVersion(nuget.NetCoreApp, &nuget.NuGetVersion{Version: semver.New(1, 0, 0, "", "")})
//	NetCoreApp11 = NewFrameworkWithVersion(nuget.NetCoreApp, &nuget.NuGetVersion{Version: semver.New(1, 1, 0, "", "")})
//	NetCoreApp20 = NewFrameworkWithVersion(nuget.NetCoreApp, &nuget.NuGetVersion{Version: semver.New(2, 0, 0, "", "")})
//	NetCoreApp21 = NewFrameworkWithVersion(nuget.NetCoreApp, &nuget.NuGetVersion{Version: semver.New(2, 1, 0, "", "")})
//	NetCoreApp22 = NewFrameworkWithVersion(nuget.NetCoreApp, &nuget.NuGetVersion{Version: semver.New(2, 2, 0, "", "")})
//	NetCoreApp30 = NewFrameworkWithVersion(nuget.NetCoreApp, &nuget.NuGetVersion{Version: semver.New(3, 0, 0, "", "")})
//	NetCoreApp31 = NewFrameworkWithVersion(nuget.NetCoreApp, &nuget.NuGetVersion{Version: semver.New(3, 1, 0, "", "")})
//
//	Net50   = NewFrameworkWithVersion(nuget.NetCoreApp, &nuget.Version5)
//	Net60   = NewFrameworkWithVersion(nuget.NetCoreApp, &nuget.Version6)
//	Net70   = NewFrameworkWithVersion(nuget.NetCoreApp, &nuget.Version7)
//	Net80   = NewFrameworkWithVersion(nuget.NetCoreApp, &nuget.Version8)
//	Net90   = NewFrameworkWithVersion(nuget.NetCoreApp, &nuget.Version9)
//	Net10_0 = NewFrameworkWithVersion(nuget.NetCoreApp, &nuget.Version10)
//	Native  = NewFrameworkWithVersion(nuget.FrameworkNative, &nuget.EmptyVersion)
//)
