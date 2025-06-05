package creation

import (
	"github.com/Masterminds/semver/v3"
	"github.com/huhouhua/go-nuget"
	"sync"
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
	// Ex: NET Framework &#8210;&gt; .NET Framework
	GetIdentifierSynonymsMap() []*KeyValuePair[string, string]

	// GetIdentifierShortNameMap Ex: .NET Framework &#8210;&gt; net
	GetIdentifierShortNameMap() []*KeyValuePair[string, string]

	// GetProfileShortNamesMap Ex: WindowsPhone &#8210;&gt; wp
	GetProfileShortNamesMap() []*FrameworkSpecificMapping

	// GetEquivalentFrameworkMap Equal frameworks. Used for legacy conversions.
	// ex: Framework: Win8 &lt;&#8210;&gt; Framework: NetCore45 Platform: Win8
	GetEquivalentFrameworkMap() []*KeyValuePair[*Framework, *Framework]

	// GetFullNameReplacementMap Rewrite full framework names to the given value.
	// Ex: .NETPlatform,Version=v0.0 &#8210;&gt; .NETPlatform,Version=v5.0
	GetFullNameReplacementMap() []*KeyValuePair[*Framework, *Framework]
}

type DefaultFrameworkMappings struct {
}

func (d *DefaultFrameworkMappings) GetIdentifierSynonymsMap() []*KeyValuePair[string, string] {
	return []*KeyValuePair[string, string]{

		// .NET
		{
			Key:   "NETFramework",
			Value: nuget.Net,
		},
		{
			Key:   ".NET",
			Value: nuget.Net,
		},
		// .NET Core
		{
			Key:   "NETCore",
			Value: nuget.NetCore,
		},
		// Portable
		{
			Key:   "NETPortable",
			Value: nuget.Portable,
		},
		// ASP
		{
			Key:   "asp.net",
			Value: nuget.AspNet,
		},
		{
			Key:   "asp.netcore",
			Value: nuget.AspNetCore,
		},
		// Mono/Xamarin
		{
			Key:   "Xamarin.PlayStationThree",
			Value: nuget.XamarinPlayStation3,
		},
		{
			Key:   "XamarinPlayStationThree",
			Value: nuget.XamarinPlayStation3,
		},
		{
			Key:   "Xamarin.PlayStationFour",
			Value: nuget.XamarinPlayStation4,
		},
		{
			Key:   "XamarinPlayStationFour",
			Value: nuget.XamarinPlayStation4,
		},
		{
			Key:   "XamarinPlayStationVita",
			Value: nuget.XamarinPlayStationVita,
		},
	}
}

func (d *DefaultFrameworkMappings) GetIdentifierShortNameMap() []*KeyValuePair[string, string] {
	return []*KeyValuePair[string, string]{
		{
			Key:   nuget.NetCoreApp,
			Value: "netcoreapp",
		},
		{
			Key:   nuget.NetStandardApp,
			Value: "netstandardapp",
		},
		{
			Key:   nuget.NetStandard,
			Value: "netstandard",
		},
		{
			Key:   nuget.NetPlatform,
			Value: "dotnet",
		},
		{
			Key:   nuget.Net,
			Value: "net",
		},
		{
			Key:   nuget.NetMicro,
			Value: "netmf",
		},
		{
			Key:   nuget.Silverlight,
			Value: "sl",
		},
		{
			Key:   nuget.Portable,
			Value: "portable",
		},
		{
			Key:   nuget.WindowsPhone,
			Value: "wp",
		},
		{
			Key:   nuget.WindowsPhoneApp,
			Value: "wpa",
		},
		{
			Key:   nuget.Windows,
			Value: "win",
		},
		{
			Key:   nuget.AspNet,
			Value: "aspnet",
		},
		{
			Key:   nuget.AspNetCore,
			Value: "aspnetcore",
		},
		{
			Key:   nuget.FrameworkNative,
			Value: "native",
		},
		{
			Key:   nuget.MonoAndroid,
			Value: "monoandroid",
		},
		{
			Key:   nuget.MonoTouch,
			Value: "monotouch",
		},
		{
			Key:   nuget.MonoMac,
			Value: "monomac",
		},
		{
			Key:   nuget.XamarinIOs,
			Value: "xamarinios",
		},
		{
			Key:   nuget.XamarinMac,
			Value: "xamarinmac",
		},
		{
			Key:   nuget.XamarinPlayStation3,
			Value: "xamarinpsthree",
		},
		{
			Key:   nuget.XamarinPlayStation4,
			Value: "xamarinpsfour",
		},
		{
			Key:   nuget.XamarinPlayStationVita,
			Value: "xamarinpsvita",
		},
		{
			Key:   nuget.XamarinWatchOS,
			Value: "xamarinwatchos",
		},
		{
			Key:   nuget.XamarinTVOS,
			Value: "xamarintvos",
		},
		{
			Key:   nuget.XamarinXbox360,
			Value: "xamarinxboxthreesixty",
		},
		{
			Key:   nuget.XamarinXboxOne,
			Value: "xamarinxboxone",
		},
		{
			Key:   nuget.Dnx,
			Value: "dnx",
		},
		{
			Key:   nuget.DnxCore,
			Value: "dnxcore",
		},
		{
			Key:   nuget.NetCore,
			Value: "netcore",
		},
		{
			Key:   nuget.WinRT,
			Value: "winrt",
		},
		{
			Key:   nuget.UAP,
			Value: "uap",
		},
		{
			Key:   nuget.Tizen,
			Value: "tizen",
		},
		{
			Key:   nuget.NanoFramework,
			Value: "netnano",
		},
	}
}

func (d *DefaultFrameworkMappings) GetProfileShortNamesMap() []*FrameworkSpecificMapping {
	return []*FrameworkSpecificMapping{
		{
			FrameworkIdentifier: nuget.Net,
			Mapping: &KeyValuePair[string, string]{
				Key:   "Client",
				Value: "Client",
			},
		},
		{
			FrameworkIdentifier: nuget.Net,
			Mapping: &KeyValuePair[string, string]{
				Key:   "CF",
				Value: "CompactFramework",
			},
		},
		{
			FrameworkIdentifier: nuget.Net,
			Mapping: &KeyValuePair[string, string]{
				Key:   "Full",
				Value: "",
			},
		},
		{
			FrameworkIdentifier: nuget.Silverlight,
			Mapping: &KeyValuePair[string, string]{
				Key:   "WP",
				Value: "WindowsPhone",
			},
		},
		{
			FrameworkIdentifier: nuget.Silverlight,
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
			Key:   NewFrameworkWithVersion(nuget.UAP, nuget.EmptyVersion),
			Value: UAP10,
		},

		{
			// win <-> win8
			Key:   NewFrameworkWithVersion(nuget.Windows, nuget.EmptyVersion),
			Value: Win8,
		},

		{
			// win8 <-> netcore45
			Key:   NewFrameworkWithVersion(nuget.NetCore, semver.New(0, 5, 0, "", "")),
			Value: Win8,
		},

		{
			// netcore45 <-> winrt45
			Key:   NewFrameworkWithVersion(nuget.NetCore, semver.New(4, 5, 0, "", "")),
			Value: NewFrameworkWithVersion(nuget.WinRT, semver.New(4, 5, 0, "", "")),
		},

		{
			// netcore <-> netcore45
			Key:   NewFrameworkWithVersion(nuget.NetCore, nuget.EmptyVersion),
			Value: NewFrameworkWithVersion(nuget.NetCore, semver.New(4, 5, 0, "", "")),
		},

		{
			// winrt <-> winrt45
			Key:   NewFrameworkWithVersion(nuget.WinRT, nuget.EmptyVersion),
			Value: NewFrameworkWithVersion(nuget.WinRT, semver.New(4, 5, 0, "", "")),
		},

		{
			// win81 <-> netcore451
			Key:   Win81,
			Value: NewFrameworkWithVersion(nuget.NetCore, semver.New(4, 5, 1, "", "")),
		},

		{
			// wp <-> wp7
			Key:   NewFrameworkWithVersion(nuget.WindowsPhone, nuget.EmptyVersion),
			Value: WP7,
		},

		{
			// wp7 <-> f:sl3-wp
			Key:   WP7,
			Value: NewFrameworkWithProfile(nuget.Silverlight, semver.New(3, 0, 0, "", ""), "WindowsPhone"),
		},
		{
			// wp71 <-> f:sl4-wp71
			Key:   NewFrameworkWithVersion(nuget.WindowsPhone, semver.New(7, 1, 0, "", "")),
			Value: NewFrameworkWithProfile(nuget.Silverlight, semver.New(4, 0, 0, "", ""), "WindowsPhone"),
		},
		{
			// wp8 <-> f:sl8-wp
			Key:   WP8,
			Value: NewFrameworkWithProfile(nuget.Silverlight, semver.New(8, 0, 0, "", ""), "WindowsPhone"),
		},
		{
			// wp81 <-> f:sl81-wp
			Key:   WP81,
			Value: NewFrameworkWithProfile(nuget.Silverlight, semver.New(8, 1, 0, "", ""), "WindowsPhone"),
		},
		{
			// wpa <-> wpa81
			Key:   NewFrameworkWithVersion(nuget.WindowsPhoneApp, nuget.EmptyVersion),
			Value: WPA81,
		},
		{
			// tizen <-> tizen3
			Key:   NewFrameworkWithVersion(nuget.Tizen, nuget.EmptyVersion),
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
var profilesWithOptionalFrameworks = []int{5, 6, 7, 14, 19, 24, 37, 42, 44, 47, 49, 78, 92, 102, 111, 136, 147, 151, 158, 225, 255, 259, 328, 336, 344}

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
	for i := range framework {
		frameworkPtrs[i] = framework[i]
	}
	return &KeyValuePair[int, []*Framework]{
		Key:   profile,
		Value: frameworkPtrs,
	}
}

func (d *DefaultPortableFrameworkMappings) GetProfileOptionalFrameworkMap() []*KeyValuePair[int, []*Framework] {
	monoandroid := NewFrameworkWithVersion(nuget.MonoAndroid, semver.New(0, 0, 0, "", ""))
	monotouch := NewFrameworkWithVersion(nuget.MonoTouch, semver.New(0, 0, 0, "", ""))
	xamarinIOs := NewFrameworkWithVersion(nuget.XamarinIOs, semver.New(0, 0, 0, "", ""))
	xamarinMac := NewFrameworkWithVersion(nuget.XamarinMac, semver.New(0, 0, 0, "", ""))
	xamarinTVOS := NewFrameworkWithVersion(nuget.XamarinTVOS, semver.New(0, 0, 0, "", ""))
	xamarinWatchOS := NewFrameworkWithVersion(nuget.XamarinWatchOS, semver.New(0, 0, 0, "", ""))
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
