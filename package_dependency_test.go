// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"errors"
	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewPackageDependencyGroup(t *testing.T) {
	tests := []struct {
		name            string
		packages        []*Dependency
		targetFramework string
		wantError       error
	}{
		{
			name: "Valid dependencies",
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
			name:            "Empty packages",
			targetFramework: "net5.0",
			packages:        nil,
			wantError:       nil,
		},
		{
			name:            "Invalid dependency version",
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
			require.Equal(t, err, tt.wantError)
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
			name:      "Valid version",
			id:        "TestPackage",
			version:   "1.2.3",
			wantError: nil,
		},
		{
			name:      "Invalid version",
			id:        "TestPackage",
			version:   "invalid_version",
			wantError: errors.New("Invalid Semantic Version"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity, err := NewPackageIdentity(tt.id, tt.version)
			require.Equal(t, err, tt.wantError)
			if err == nil {
				require.NotNil(t, identity)
				assert.Equal(t, tt.id, identity.Id)
			}
		})
	}
}

func TestFrameworkSpecificGroup(t *testing.T) {
	tests := []struct {
		name            string
		items           []string
		targetFramework string
		wantError       error
	}{
		{
			name:            "Valid items",
			items:           []string{"file1.dll", "file2.dll"},
			targetFramework: "net5.0",
			wantError:       nil,
		},
		{
			name:            "Empty items",
			targetFramework: "net5.0",
			items:           nil,
			wantError:       errors.New("items cannot be nil"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := NewFrameworkSpecificGroup(tt.targetFramework, tt.items...)
			require.Equal(t, err, tt.wantError)
			if err == nil {
				assert.NotNil(t, group)
				assert.Equal(t, len(tt.items), len(group.Items))
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
			name: "No options",
			optionsFunc: func() []PackageDependencyInfoFunc {
				return []PackageDependencyInfoFunc{}
			},
			wantDataFunc: func(t *testing.T) *PackageDependencyInfo {
				return &PackageDependencyInfo{}
			},
			wantError: nil,
		},
		{
			name: "With Identity",
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
					PackageIdentity:          identity,
					DependencyGroups:         nil,
					FrameworkReferenceGroups: nil,
				}
			},
			wantError: nil,
		},
		{
			name: "With DependencyGroups",
			optionsFunc: func() []PackageDependencyInfoFunc {
				dependencies := &Dependencies{
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
				}
				return []PackageDependencyInfoFunc{
					WithDependencyGroups(dependencies),
				}
			},
			wantDataFunc: func(t *testing.T) *PackageDependencyInfo {
				versionrange1203, err := ParseVersionRange("12.0.3")
				require.NoError(t, err)

				versionrange500, err := ParseVersionRange("5.0.0")
				require.NoError(t, err)

				return &PackageDependencyInfo{
					PackageIdentity: nil,
					DependencyGroups: []*PackageDependencyGroup{
						{
							TargetFramework: ".NETFramework4.8",
							Packages: []*Dependency{
								{
									Id:           "Newtonsoft.Json",
									VersionRaw:   "12.0.3",
									ExcludeRaw:   "Build,Analyzers",
									VersionRange: versionrange1203,
									//Version:    &NuGetVersion{semver.New(12, 0, 3, "", "")},
									Exclude: []string{"Build", "Analyzers"},
								},
								{
									Id:           "Microsoft.Extensions.Logging",
									VersionRaw:   "5.0.0",
									VersionRange: versionrange500,
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
									VersionRange: versionrange1203,
									//Version:    &NuGetVersion{semver.New(12, 0, 3, "", "")},
									Exclude: []string{"Build", "Analyzers"},
								},
							},
						},
					},
					FrameworkReferenceGroups: nil,
				}
			},
			wantError: nil,
		},
		{
			name: "With FrameworkReferenceGroups",
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
					PackageIdentity:  nil,
					DependencyGroups: nil,
					FrameworkReferenceGroups: []*FrameworkSpecificGroup{
						{
							Items:           []string{"System.Net.Http"},
							HasEmptyFolder:  false,
							TargetFramework: ".NETFramework4.8",
						},
					},
				}
			},
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &PackageDependencyInfo{}
			err := ApplyPackageDependency(input, tt.optionsFunc()...)
			require.Equal(t, err, tt.wantError)
			if err == nil {
				require.Equal(t, input, tt.wantDataFunc(t))
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

	versionrange1203, err := ParseVersionRange("12.0.3")
	require.NoError(t, err)

	versionrange500, err := ParseVersionRange("5.0.0")
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
						VersionRange: versionrange1203,
						//Version:    &NuGetVersion{semver.New(12, 0, 3, "", "")},
						Exclude: []string{"Build", "Analyzers"},
					},
					{
						Id:           "Microsoft.Extensions.Logging",
						VersionRaw:   "5.0.0",
						VersionRange: versionrange500,
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
						VersionRange: versionrange1203,
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
