// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"errors"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/require"
)

func TestNewPackageDependencyGroup(t *testing.T) {
	tests := []struct {
		name            string
		packages        []*Dependency
		targetFramework string
		wantError       error
	}{
		{
			name: "valid dependencies return success",
			packages: []*Dependency{
				{
					Id:         "Package1",
					VersionRaw: "1.0.0",
				},
				{
					Id:         "Package2",
					VersionRaw: "2.0.0",
				},
			},
			targetFramework: "net5.0",
			wantError:       nil,
		},
		{
			name:            "empty packages return success",
			targetFramework: "net5.0",
			packages:        nil,
			wantError:       nil,
		},
		{
			name:            "invalid dependency version return error",
			targetFramework: "net5.0",
			packages: []*Dependency{
				{
					Id:         "Invalid",
					VersionRaw: "invalid_version",
				},
			},
			wantError: errors.New("invalid version: Invalid Semantic Version"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := NewPackageDependencyGroup(tt.targetFramework, tt.packages...)
			require.Equal(t, tt.wantError, err)
			if err == nil {
				require.NotNil(t, group)
			}
		})
	}
}

func TestNewPackageIdentity(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		version   string
		wantError error
	}{
		{
			name:    "valid version return success",
			id:      "TestPackage",
			version: "1.2.3",
		},
		{
			name:      "invalid version return error",
			id:        "TestPackage",
			version:   "invalid_version",
			wantError: errors.New("Invalid Semantic Version"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity, err := NewPackageIdentity(tt.id, tt.version)
			require.Equal(t, tt.wantError, err)
			if err == nil {
				require.NotNil(t, identity)
				require.Equal(t, tt.id, identity.Id)
			}
		})
	}
}

func TestFrameworkSpecificGroup(t *testing.T) {
	tests := []struct {
		name            string
		input           []string
		targetFramework string
		wantGroup       *FrameworkSpecificGroup
		wantError       error
	}{
		{
			name:  "valid items return success",
			input: []string{"file1.dll", "file2.dll"},
			wantGroup: &FrameworkSpecificGroup{
				TargetFramework: "net5.0",
				Items:           []string{"file1.dll", "file2.dll"},
			},
			targetFramework: "net5.0",
		},
		{
			name:            "empty items return error",
			targetFramework: "net5.0",
			input:           nil,
			wantError:       errors.New("items cannot be nil"),
		},
		{
			name: "has empty folder return success",
			input: []string{
				"path/to/package/_._",
			},
			wantGroup: &FrameworkSpecificGroup{
				TargetFramework: "net5.0",
				Items:           make([]string, 0, 1),
				HasEmptyFolder:  true,
			},
			targetFramework: "net5.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := NewFrameworkSpecificGroup(tt.targetFramework, tt.input...)
			require.Equal(t, tt.wantError, err)
			if err == nil {
				require.NotNil(t, group)
				require.Equal(t, tt.wantGroup, group)
			}
		})
	}
}

func TestConfigurePackageDependency(t *testing.T) {
	tests := []struct {
		name         string
		optionsFunc  func() []PackageDependencyInfoFunc
		wantDataFunc func(t *testing.T) *PackageDependencyInfo
		wantError    error
	}{
		{
			name: "no options",
			optionsFunc: func() []PackageDependencyInfoFunc {
				return []PackageDependencyInfoFunc{}
			},
			wantDataFunc: func(t *testing.T) *PackageDependencyInfo {
				return &PackageDependencyInfo{}
			},
		},
		{
			name: "with identity return success",
			optionsFunc: func() []PackageDependencyInfoFunc {
				meta := &Metadata{
					PackageInfo: PackageInfo{
						ID:      "TestPackage",
						Version: "1.0.0",
					},
				}
				return []PackageDependencyInfoFunc{
					WithIdentity(meta),
				}
			},
			wantDataFunc: func(t *testing.T) *PackageDependencyInfo {
				identity, err := NewPackageIdentity("TestPackage", "1.0.0")
				require.NoError(t, err)
				require.True(t, identity.HasVersion())
				return &PackageDependencyInfo{
					PackageIdentity: identity,
				}
			},
		},
		{
			name: "with identity parse version return error",
			optionsFunc: func() []PackageDependencyInfoFunc {
				meta := &Metadata{
					PackageInfo: PackageInfo{
						ID:      "TestPackage",
						Version: "^0.0.1",
					},
				}
				return []PackageDependencyInfoFunc{
					WithIdentity(meta),
				}
			},
			wantError: errors.New("Invalid Semantic Version"),
		},
		{
			name: "with dependencyGroups in groups return success",
			optionsFunc: func() []PackageDependencyInfoFunc {
				dependencies := &Dependencies{
					Groups: []*DependenciesGroup{
						{
							TargetFramework: ".NETFramework4.8",
							Dependencies: []*Dependency{
								{
									Id:              "Newtonsoft.Json",
									VersionRaw:      "12.0.3",
									ExcludeRaw:      "Build,Analyzers",
									IncludeRaw:      "",
									VersionRangeRaw: "",
								},
								{
									Id:              "Microsoft.Extensions.Logging",
									VersionRaw:      "5.0.0",
									ExcludeRaw:      "",
									IncludeRaw:      "",
									VersionRangeRaw: "",
								},
							},
						},
						{
							TargetFramework: ".NETStandard2.0",
							Dependencies: []*Dependency{
								{
									Id:              "Newtonsoft.Json",
									VersionRaw:      "12.0.3",
									ExcludeRaw:      "Build,Analyzers",
									IncludeRaw:      "",
									VersionRangeRaw: "",
								},
							},
						},
					},
				}
				return []PackageDependencyInfoFunc{
					WithDependencyGroups(dependencies),
				}
			},
			wantDataFunc: func(t *testing.T) *PackageDependencyInfo {
				versionRange1203, err := ParseVersionRange("12.0.3")
				require.NoError(t, err)

				versionRange500, err := ParseVersionRange("5.0.0")
				require.NoError(t, err)

				return &PackageDependencyInfo{
					DependencyGroups: []*PackageDependencyGroup{
						{
							TargetFramework: ".NETFramework4.8",
							Packages: []*Dependency{
								{
									Id:           "Newtonsoft.Json",
									VersionRaw:   "12.0.3",
									ExcludeRaw:   "Build,Analyzers",
									VersionRange: versionRange1203,
									//Version:    &NuGetVersion{semver.New(12, 0, 3, "", "")},
									Exclude: []string{"Build", "Analyzers"},
								},
								{
									Id:           "Microsoft.Extensions.Logging",
									VersionRaw:   "5.0.0",
									VersionRange: versionRange500,
									//Version:    &NuGetVersion{semver.New(5, 0, 0, "", "")},
								},
							},
						},
						{
							TargetFramework: ".NETStandard2.0",
							Packages: []*Dependency{
								{
									Id:           "Newtonsoft.Json",
									VersionRaw:   "12.0.3",
									ExcludeRaw:   "Build,Analyzers",
									VersionRange: versionRange1203,
									//Version:    &NuGetVersion{semver.New(12, 0, 3, "", "")},
									Exclude: []string{"Build", "Analyzers"},
								},
							},
						},
					},
				}
			},
		},
		{
			name: "with dependencyGroups in dependency return success",
			optionsFunc: func() []PackageDependencyInfoFunc {
				dependencies := &Dependencies{
					Dependency: []*Dependency{
						{
							Id:              "Newtonsoft.Json",
							VersionRaw:      "12.0.3",
							ExcludeRaw:      "Build,Analyzers",
							IncludeRaw:      "",
							VersionRangeRaw: "",
						},
						{
							Id:              "Microsoft.Extensions.Logging",
							VersionRaw:      "5.0.0",
							ExcludeRaw:      "",
							IncludeRaw:      "",
							VersionRangeRaw: "",
						},
					},
				}
				return []PackageDependencyInfoFunc{
					WithDependencyGroups(dependencies),
				}
			},
			wantDataFunc: func(t *testing.T) *PackageDependencyInfo {
				versionRange1203, err := ParseVersionRange("12.0.3")
				require.NoError(t, err)

				versionRange500, err := ParseVersionRange("5.0.0")
				require.NoError(t, err)

				return &PackageDependencyInfo{
					DependencyGroups: []*PackageDependencyGroup{
						{
							TargetFramework: "Any",
							Packages: []*Dependency{
								{
									Id:           "Newtonsoft.Json",
									VersionRaw:   "12.0.3",
									ExcludeRaw:   "Build,Analyzers",
									VersionRange: versionRange1203,
									//Version:    &NuGetVersion{semver.New(12, 0, 3, "", "")},
									Exclude: []string{"Build", "Analyzers"},
								},
							},
						},
						{
							TargetFramework: "Any",
							Packages: []*Dependency{
								{
									Id:           "Microsoft.Extensions.Logging",
									VersionRaw:   "5.0.0",
									VersionRange: versionRange500,
									//Version:    &NuGetVersion{semver.New(5, 0, 0, "", "")},
								},
							},
						},
					},
				}
			},
		},
		{
			name: "with dependencyGroups return nil",
			optionsFunc: func() []PackageDependencyInfoFunc {
				return []PackageDependencyInfoFunc{
					WithDependencyGroups(nil),
				}
			},
			wantDataFunc: func(t *testing.T) *PackageDependencyInfo {
				return &PackageDependencyInfo{}
			},
		},
		{
			name: "with dependencyGroups parse version in groups return error",
			optionsFunc: func() []PackageDependencyInfoFunc {
				dependencies := &Dependencies{
					Groups: []*DependenciesGroup{
						{
							Dependencies: []*Dependency{
								{
									VersionRaw: "[1.0.0]",
								},
							},
						},
					},
				}
				return []PackageDependencyInfoFunc{
					WithDependencyGroups(dependencies),
				}
			},
			wantError: errors.New("invalid range format: [1.0.0]"),
		},
		{
			name: "with dependencyGroups parse version in dependency return error",
			optionsFunc: func() []PackageDependencyInfoFunc {
				dependencies := &Dependencies{
					Dependency: []*Dependency{
						{
							VersionRaw: "[1.0.0]",
						},
					},
				}
				return []PackageDependencyInfoFunc{
					WithDependencyGroups(dependencies),
				}
			},
			wantError: errors.New("invalid range format: [1.0.0]"),
		},
		{
			name: "with frameworkReferenceGroups return success",
			optionsFunc: func() []PackageDependencyInfoFunc {
				frameworkAssemblies := &FrameworkAssemblies{
					FrameworkAssembly: []*FrameworkAssembly{
						{
							AssemblyName:    []string{"System.Net.Http"},
							TargetFramework: ".NETFramework4.8",
						},
					},
				}
				return []PackageDependencyInfoFunc{
					WithFrameworkReferenceGroups(frameworkAssemblies),
				}
			},
			wantDataFunc: func(t *testing.T) *PackageDependencyInfo {
				return &PackageDependencyInfo{
					FrameworkReferenceGroups: []*FrameworkSpecificGroup{
						{
							Items:           []string{"System.Net.Http"},
							TargetFramework: ".NETFramework4.8",
						},
					},
				}
			},
		},
		{
			name: "with frameworkReferenceGroups return nil",
			optionsFunc: func() []PackageDependencyInfoFunc {
				return []PackageDependencyInfoFunc{
					WithFrameworkReferenceGroups(nil),
				}
			},
			wantDataFunc: func(t *testing.T) *PackageDependencyInfo {
				return &PackageDependencyInfo{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &PackageDependencyInfo{}
			err := ApplyPackageDependency(input, tt.optionsFunc()...)
			require.Equal(t, tt.wantError, err)
			if err == nil {
				require.Equal(t, tt.wantDataFunc(t), input)
			}
		})
	}
}

func TestConfigureDependencyInfo(t *testing.T) {
	nuspec := Nuspec{
		Metadata: &Metadata{
			PackageInfo: PackageInfo{
				ID:      "TestPackage",
				Version: "1.0.0",
			},
			Dependencies: &Dependencies{
				Dependency: nil,
				Groups: []*DependenciesGroup{
					{
						TargetFramework: ".NETFramework4.8",
						Dependencies: []*Dependency{
							{
								Id:              "Newtonsoft.Json",
								VersionRaw:      "12.0.3",
								ExcludeRaw:      "Build,Analyzers",
								IncludeRaw:      "",
								VersionRangeRaw: "",
							},
							{
								Id:              "Microsoft.Extensions.Logging",
								VersionRaw:      "5.0.0",
								ExcludeRaw:      "",
								IncludeRaw:      "",
								VersionRangeRaw: "",
							},
						},
					},
					{
						TargetFramework: ".NETStandard2.0",
						Dependencies: []*Dependency{
							{
								Id:              "Newtonsoft.Json",
								VersionRaw:      "12.0.3",
								ExcludeRaw:      "Build,Analyzers",
								IncludeRaw:      "",
								VersionRangeRaw: "",
							},
						},
					},
				},
			},
			FrameworkAssemblies: &FrameworkAssemblies{
				FrameworkAssembly: []*FrameworkAssembly{
					{
						AssemblyName:    []string{"System.Net.Http"},
						TargetFramework: ".NETFramework4.8",
					},
				},
			},
			References: nil,
		},
	}

	versionRange1203, err := ParseVersionRange("12.0.3")
	require.NoError(t, err)

	versionRange500, err := ParseVersionRange("5.0.0")
	require.NoError(t, err)

	want := &PackageDependencyInfo{
		PackageIdentity: &PackageIdentity{
			Id:      "TestPackage",
			Version: &NuGetVersion{semver.New(1, 0, 0, "", "")},
		},
		DependencyGroups: []*PackageDependencyGroup{
			{
				TargetFramework: ".NETFramework4.8",
				Packages: []*Dependency{
					{
						Id:           "Newtonsoft.Json",
						VersionRaw:   "12.0.3",
						ExcludeRaw:   "Build,Analyzers",
						VersionRange: versionRange1203,
						//Version:    &NuGetVersion{semver.New(12, 0, 3, "", "")},
						Exclude: []string{"Build", "Analyzers"},
					},
					{
						Id:           "Microsoft.Extensions.Logging",
						VersionRaw:   "5.0.0",
						VersionRange: versionRange500,
						//Version:    &NuGetVersion{semver.New(5, 0, 0, "", "")},
					},
				},
			},
			{
				TargetFramework: ".NETStandard2.0",
				Packages: []*Dependency{
					{
						Id:           "Newtonsoft.Json",
						VersionRaw:   "12.0.3",
						ExcludeRaw:   "Build,Analyzers",
						VersionRange: versionRange1203,
						//Version:    &NuGetVersion{semver.New(12, 0, 3, "", "")},
						Exclude: []string{"Build", "Analyzers"},
					},
				},
			},
		},
		FrameworkReferenceGroups: []*FrameworkSpecificGroup{
			{
				Items:           []string{"System.Net.Http"},
				HasEmptyFolder:  false,
				TargetFramework: ".NETFramework4.8",
			},
		},
	}
	input := &PackageDependencyInfo{}
	err = ConfigureDependencyInfo(input, nuspec)
	require.NoError(t, err)
	require.Equal(t, want, input)
}
