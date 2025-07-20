// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package framework

import (
	"sync"

	"github.com/huhouhua/go-nuget/internal/consts"
	"github.com/huhouhua/go-nuget/version"
)

var (
	frameworkNameProvider FrameworkNameProvider
	once                  sync.Once
)

func GetProviderInstance() FrameworkNameProvider {
	once.Do(func() {
		mappings := []FrameworkMappings{&DefaultFrameworkMappings{}}
		portableMappings := []PortableFrameworkMappings{&DefaultPortableFrameworkMappings{}}
		frameworkNameProvider = *NewFrameworkNameProvider(mappings, portableMappings)
	})
	return frameworkNameProvider
}

type KeyValuePair[K comparable, V any] struct {
	Key   K
	Value V
}

// FrameworkSpecificMapping A key value pair specific to a framework identifier
type FrameworkSpecificMapping struct {
	FrameworkIdentifier string

	Mapping *KeyValuePair[string, string]
}

type PortableFrameworkMappings interface {
	// GetProfileFrameworkMap  Ex: 5 -> net4, win8
	GetProfileFrameworkMap() []*KeyValuePair[int, []*Framework]

	// GetProfileOptionalFrameworkMap  Additional optional frameworks supported in a portable profile.
	// Ex: 5 -> MonoAndroid1+MonoTouch1
	GetProfileOptionalFrameworkMap() []*KeyValuePair[int, []*Framework]
}

type FrameworkMappings interface {
	// GetIdentifierSynonymsMap Synonym &#8210;&gt; Identifier
	// Ex: NET Framework .NET Framework
	GetIdentifierSynonymsMap() []*KeyValuePair[string, string]

	// GetIdentifierShortNameMap Ex: .NET Framework  net
	GetIdentifierShortNameMap() []*KeyValuePair[string, string]

	// GetProfileShortNamesMap Ex: WindowsPhone  wp
	GetProfileShortNamesMap() []*FrameworkSpecificMapping

	// GetEquivalentFrameworkMap Equal frameworks. Used for legacy conversions.
	// ex: Framework: Win8  Framework: NetCore45 Platform: Win8
	GetEquivalentFrameworkMap() []*KeyValuePair[*Framework, *Framework]

	// GetShortNameReplacementMap Rewrite folder short names to the given value. Ex: dotnet50 ‒> dotnet
	GetShortNameReplacementMap() []*KeyValuePair[*Framework, *Framework]

	// GetFullNameReplacementMap Rewrite full framework names to the given value. Ex: .NETPlatform,Version=v0.0 ‒>
	// .NETPlatform,Version=v5.0
	GetFullNameReplacementMap() []*KeyValuePair[*Framework, *Framework]
}

type DefaultFrameworkMappings struct {
}

func (d *DefaultFrameworkMappings) GetIdentifierSynonymsMap() []*KeyValuePair[string, string] {
	return []*KeyValuePair[string, string]{

		// .NET
		{
			Key:   "NETFramework",
			Value: consts.Net,
		},
		{
			Key:   ".NET",
			Value: consts.Net,
		},
		// .NET Core
		{
			Key:   "NETCore",
			Value: consts.NetCore,
		},
		// Portable
		{
			Key:   "NETPortable",
			Value: consts.Portable,
		},
		// ASP
		{
			Key:   "asp.net",
			Value: consts.AspNet,
		},
		{
			Key:   "asp.netcore",
			Value: consts.AspNetCore,
		},
		// Mono/Xamarin
		{
			Key:   "Xamarin.PlayStationThree",
			Value: consts.XamarinPlayStation3,
		},
		{
			Key:   "XamarinPlayStationThree",
			Value: consts.XamarinPlayStation3,
		},
		{
			Key:   "Xamarin.PlayStationFour",
			Value: consts.XamarinPlayStation4,
		},
		{
			Key:   "XamarinPlayStationFour",
			Value: consts.XamarinPlayStation4,
		},
		{
			Key:   "XamarinPlayStationVita",
			Value: consts.XamarinPlayStationVita,
		},
	}
}

func (d *DefaultFrameworkMappings) GetIdentifierShortNameMap() []*KeyValuePair[string, string] {
	return []*KeyValuePair[string, string]{
		{
			Key:   consts.NetCoreApp,
			Value: "netcoreapp",
		},
		{
			Key:   consts.NetStandardApp,
			Value: "netstandardapp",
		},
		{
			Key:   consts.NetStandard,
			Value: "netstandard",
		},
		{
			Key:   consts.NetPlatform,
			Value: "dotnet",
		},
		{
			Key:   consts.Net,
			Value: "net",
		},
		{
			Key:   consts.NetMicro,
			Value: "netmf",
		},
		{
			Key:   consts.Silverlight,
			Value: "sl",
		},
		{
			Key:   consts.Portable,
			Value: "portable",
		},
		{
			Key:   consts.WindowsPhone,
			Value: "wp",
		},
		{
			Key:   consts.WindowsPhoneApp,
			Value: "wpa",
		},
		{
			Key:   consts.Windows,
			Value: "win",
		},
		{
			Key:   consts.AspNet,
			Value: "aspnet",
		},
		{
			Key:   consts.AspNetCore,
			Value: "aspnetcore",
		},
		{
			Key:   consts.FrameworkNative,
			Value: "native",
		},
		{
			Key:   consts.MonoAndroid,
			Value: "monoandroid",
		},
		{
			Key:   consts.MonoTouch,
			Value: "monotouch",
		},
		{
			Key:   consts.MonoMac,
			Value: "monomac",
		},
		{
			Key:   consts.XamarinIOs,
			Value: "xamarinios",
		},
		{
			Key:   consts.XamarinMac,
			Value: "xamarinmac",
		},
		{
			Key:   consts.XamarinPlayStation3,
			Value: "xamarinpsthree",
		},
		{
			Key:   consts.XamarinPlayStation4,
			Value: "xamarinpsfour",
		},
		{
			Key:   consts.XamarinPlayStationVita,
			Value: "xamarinpsvita",
		},
		{
			Key:   consts.XamarinWatchOS,
			Value: "xamarinwatchos",
		},
		{
			Key:   consts.XamarinTVOS,
			Value: "xamarintvos",
		},
		{
			Key:   consts.XamarinXbox360,
			Value: "xamarinxboxthreesixty",
		},
		{
			Key:   consts.XamarinXboxOne,
			Value: "xamarinxboxone",
		},
		{
			Key:   consts.Dnx,
			Value: "dnx",
		},
		{
			Key:   consts.DnxCore,
			Value: "dnxcore",
		},
		{
			Key:   consts.NetCore,
			Value: "netcore",
		},
		{
			Key:   consts.WinRT,
			Value: "winrt",
		},
		{
			Key:   consts.UAP,
			Value: "uap",
		},
		{
			Key:   consts.Tizen,
			Value: "tizen",
		},
		{
			Key:   consts.NanoFramework,
			Value: "netnano",
		},
	}
}

func (d *DefaultFrameworkMappings) GetProfileShortNamesMap() []*FrameworkSpecificMapping {
	return []*FrameworkSpecificMapping{
		{
			FrameworkIdentifier: consts.Net,
			Mapping: &KeyValuePair[string, string]{
				Key:   "Client",
				Value: "Client",
			},
		},
		{
			FrameworkIdentifier: consts.Net,
			Mapping: &KeyValuePair[string, string]{
				Key:   "CF",
				Value: "CompactFramework",
			},
		},
		{
			FrameworkIdentifier: consts.Net,
			Mapping: &KeyValuePair[string, string]{
				Key:   "Full",
				Value: "",
			},
		},
		{
			FrameworkIdentifier: consts.Silverlight,
			Mapping: &KeyValuePair[string, string]{
				Key:   "WP",
				Value: "WindowsPhone",
			},
		},
		{
			FrameworkIdentifier: consts.Silverlight,
			Mapping: &KeyValuePair[string, string]{
				Key:   "WP71",
				Value: "WindowsPhone71",
			},
		},
	}
}

func (d *DefaultFrameworkMappings) GetEquivalentFrameworkMap() []*KeyValuePair[*Framework, *Framework] {
	return []*KeyValuePair[*Framework, *Framework]{
		{
			// UAP 0.0 <-> UAP 10.0
			Key:   NewFrameworkWithVersion(consts.UAP, consts.EmptyVersion),
			Value: UAP10,
		},

		{
			// win <-> win8
			Key:   NewFrameworkWithVersion(consts.Windows, consts.EmptyVersion),
			Value: Win8,
		},

		{
			// win8 <-> netcore45
			Key:   NewFrameworkWithVersion(consts.NetCore, version.NewVersionFrom(0, 5, 0, "", "")),
			Value: Win8,
		},

		{
			// netcore45 <-> winrt45
			Key:   NewFrameworkWithVersion(consts.NetCore, version.NewVersionFrom(4, 5, 0, "", "")),
			Value: NewFrameworkWithVersion(consts.WinRT, version.NewVersionFrom(4, 5, 0, "", "")),
		},

		{
			// netcore <-> netcore45
			Key:   NewFrameworkWithVersion(consts.NetCore, consts.EmptyVersion),
			Value: NewFrameworkWithVersion(consts.NetCore, version.NewVersionFrom(4, 5, 0, "", "")),
		},

		{
			// winrt <-> winrt45
			Key:   NewFrameworkWithVersion(consts.WinRT, consts.EmptyVersion),
			Value: NewFrameworkWithVersion(consts.WinRT, version.NewVersionFrom(4, 5, 0, "", "")),
		},

		{
			// win81 <-> netcore451
			Key:   Win81,
			Value: NewFrameworkWithVersion(consts.NetCore, version.NewVersionFrom(4, 5, 1, "", "")),
		},

		{
			// wp <-> wp7
			Key:   NewFrameworkWithVersion(consts.WindowsPhone, consts.EmptyVersion),
			Value: WP7,
		},

		{
			// wp7 <-> f:sl3-wp
			Key:   WP7,
			Value: NewFrameworkWithProfile(consts.Silverlight, version.NewVersionFrom(3, 0, 0, "", ""), "WindowsPhone"),
		},
		{
			// wp71 <-> f:sl4-wp71
			Key:   NewFrameworkWithVersion(consts.WindowsPhone, version.NewVersionFrom(7, 1, 0, "", "")),
			Value: NewFrameworkWithProfile(consts.Silverlight, version.NewVersionFrom(4, 0, 0, "", ""), "WindowsPhone"),
		},
		{
			// wp8 <-> f:sl8-wp
			Key:   WP8,
			Value: NewFrameworkWithProfile(consts.Silverlight, version.NewVersionFrom(8, 0, 0, "", ""), "WindowsPhone"),
		},
		{
			// wp81 <-> f:sl81-wp
			Key:   WP81,
			Value: NewFrameworkWithProfile(consts.Silverlight, version.NewVersionFrom(8, 1, 0, "", ""), "WindowsPhone"),
		},
		{
			// wpa <-> wpa81
			Key:   NewFrameworkWithVersion(consts.WindowsPhoneApp, consts.EmptyVersion),
			Value: WPA81,
		},
		{
			// tizen <-> tizen3
			Key:   NewFrameworkWithVersion(consts.Tizen, consts.EmptyVersion),
			Value: Tizen3,
		},
		{
			// dnx <-> dnx45
			Key:   Dnx,
			Value: Dnx45,
		},
		{
			// dnxcore <-> dnxcore50
			Key:   DnxCore,
			Value: DnxCore50,
		},
		{
			// dotnet <-> dotnet50
			Key:   DotNet,
			Value: DotNet50,
		},
		{
			// TODO: remove these rules post-RC
			// aspnet <-> aspnet50
			Key:   AspNet,
			Value: AspNet50,
		},
		{
			// aspnetcore <-> aspnetcore50
			Key:   AspNetCore,
			Value: AspNetCore50,
		},
		{
			// dnx451 <-> aspnet50
			Key:   Dnx45,
			Value: AspNet50,
		},
		{
			// dnxcore50 <-> aspnetcore50
			Key:   DnxCore50,
			Value: AspNetCore50,
		},
	}
}

func (d *DefaultFrameworkMappings) GetShortNameReplacementMap() []*KeyValuePair[*Framework, *Framework] {
	return []*KeyValuePair[*Framework, *Framework]{
		{
			Key:   DotNet50,
			Value: DotNet,
		},
	}
}

func (d *DefaultFrameworkMappings) GetFullNameReplacementMap() []*KeyValuePair[*Framework, *Framework] {
	return []*KeyValuePair[*Framework, *Framework]{
		{
			Key:   DotNet,
			Value: DotNet50,
		},
	}
}

// DefaultPortableFrameworkMappings Contains the standard portable framework mappings
type DefaultPortableFrameworkMappings struct {
}

// profiles that also support monotouch1+monoandroid1
var profilesWithOptionalFrameworks = []int{
	5,
	6,
	7,
	14,
	19,
	24,
	37,
	42,
	44,
	47,
	49,
	78,
	92,
	102,
	111,
	136,
	147,
	151,
	158,
	225,
	255,
	259,
	328,
	336,
	344,
}

func (d *DefaultPortableFrameworkMappings) GetProfileFrameworkMap() []*KeyValuePair[int, []*Framework] {
	return []*KeyValuePair[int, []*Framework]{
		// v4.6
		createProfileFrameworks(31, Win81, WP81),
		createProfileFrameworks(32, Win81, WPA81),
		createProfileFrameworks(44, Net451, Win81),
		createProfileFrameworks(84, WP81, WPA81),
		createProfileFrameworks(151, Net451, Win81, WPA81),
		createProfileFrameworks(157, Win81, WP81, WPA81),

		// v4.5
		createProfileFrameworks(7, Net45, Win8),
		createProfileFrameworks(49, Net45, WP8),
		createProfileFrameworks(78, Net45, Win8, WP8),
		createProfileFrameworks(111, Net45, Win8, WPA81),
		createProfileFrameworks(259, Net45, Win8, WPA81, WP8),

		// v4.0
		createProfileFrameworks(2, Net4, Win8, SL4, WP7),
		createProfileFrameworks(3, Net4, SL4),
		createProfileFrameworks(4, Net45, SL4, Win8, WP7),
		createProfileFrameworks(5, Net4, Win8),
		createProfileFrameworks(6, Net403, Win8),
		createProfileFrameworks(14, Net4, SL5),
		createProfileFrameworks(18, Net403, SL4),
		createProfileFrameworks(19, Net403, SL5),
		createProfileFrameworks(23, Net45, SL4),
		createProfileFrameworks(24, Net45, SL5),
		createProfileFrameworks(36, Net4, SL4, Win8, WP8),
		createProfileFrameworks(37, Net4, SL5, Win8),
		createProfileFrameworks(41, Net403, SL4, Win8),
		createProfileFrameworks(42, Net403, SL5, Win8),
		createProfileFrameworks(46, Net45, SL4, Win8),
		createProfileFrameworks(47, Net45, SL5, Win8),
		createProfileFrameworks(88, Net4, SL4, Win8, WP75),
		createProfileFrameworks(92, Net4, Win8, WPA81),
		createProfileFrameworks(95, Net403, SL4, Win8, WP7),
		createProfileFrameworks(96, Net403, SL4, Win8, WP75),
		createProfileFrameworks(102, Net403, Win8, WPA81),
		createProfileFrameworks(104, Net45, SL4, Win8, WP75),
		createProfileFrameworks(136, Net4, SL5, Win8, WP8),
		createProfileFrameworks(143, Net403, SL4, Win8, WP8),
		createProfileFrameworks(147, Net403, SL5, Win8, WP8),
		createProfileFrameworks(154, Net45, SL4, Win8, WP8),
		createProfileFrameworks(158, Net45, SL5, Win8, WP8),
		createProfileFrameworks(225, Net4, SL5, Win8, WPA81),
		createProfileFrameworks(240, Net403, SL5, Win8, WPA81),
		createProfileFrameworks(255, Net45, SL5, Win8, WPA81),
		createProfileFrameworks(328, Net4, SL5, Win8, WPA81, WP8),
		createProfileFrameworks(336, Net403, SL5, Win8, WPA81, WP8),
		createProfileFrameworks(344, Net45, SL5, Win8, WPA81, WP8),
	}
}

func createProfileFrameworks(profile int, framework ...*Framework) *KeyValuePair[int, []*Framework] {
	frameworkPtrs := make([]*Framework, len(framework))
	frameworkPtrs = append(frameworkPtrs, framework...)
	return &KeyValuePair[int, []*Framework]{
		Key:   profile,
		Value: frameworkPtrs,
	}
}

func (d *DefaultPortableFrameworkMappings) GetProfileOptionalFrameworkMap() []*KeyValuePair[int, []*Framework] {
	monoandroid := NewFrameworkWithVersion(consts.MonoAndroid, version.NewVersionFrom(0, 0, 0, "", ""))
	monotouch := NewFrameworkWithVersion(consts.MonoTouch, version.NewVersionFrom(0, 0, 0, "", ""))
	xamarinIOs := NewFrameworkWithVersion(consts.XamarinIOs, version.NewVersionFrom(0, 0, 0, "", ""))
	xamarinMac := NewFrameworkWithVersion(consts.XamarinMac, version.NewVersionFrom(0, 0, 0, "", ""))
	xamarinTVOS := NewFrameworkWithVersion(consts.XamarinTVOS, version.NewVersionFrom(0, 0, 0, "", ""))
	xamarinWatchOS := NewFrameworkWithVersion(consts.XamarinWatchOS, version.NewVersionFrom(0, 0, 0, "", ""))
	monoFrameworks := []*Framework{monoandroid, monotouch, xamarinIOs, xamarinMac, xamarinTVOS, xamarinWatchOS}

	profileOptionalFrameworks := make([]*KeyValuePair[int, []*Framework], 0)

	for _, profile := range profilesWithOptionalFrameworks {
		profileOptionalFrameworks = append(profileOptionalFrameworks, &KeyValuePair[int, []*Framework]{
			Key:   profile,
			Value: monoFrameworks,
		})
	}
	return profileOptionalFrameworks
}
