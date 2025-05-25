package creation

import (
	"fmt"
	"strings"

	"github.com/huhouhua/go-nuget"
)

var (
	// SchemaVersionV1 Baseline schema
	SchemaVersionV1 = "http://schemas.microsoft.com/packaging/2010/07/nuspec.xsd"
	// SchemaVersionV2 Added copyrights, references and release notes
	SchemaVersionV2 = "http://schemas.microsoft.com/packaging/2011/08/nuspec.xsd"
	// SchemaVersionV3 Used if the version is a semantic version.
	SchemaVersionV3 = "http://schemas.microsoft.com/packaging/2011/10/nuspec.xsd"
	// SchemaVersionV4 Added 'targetFramework' attribute for 'dependency' elements.
	// Allow framework folders under 'content' and 'tools' folders.
	SchemaVersionV4 = "http://schemas.microsoft.com/packaging/2012/06/nuspec.xsd"
	// SchemaVersionV5 Added 'targetFramework' attribute for 'references' elements.
	// Added 'minClientVersion' attribute
	SchemaVersionV5 = "http://schemas.microsoft.com/packaging/2013/01/nuspec.xsd"
	// SchemaVersionV6 Allows XDT transformation
	SchemaVersionV6 = "http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd"
)

type SchemaVersionTypes []string

var VersionToSchemaMaps = SchemaVersionTypes{
	SchemaVersionV1,
	SchemaVersionV2,
	SchemaVersionV3,
	SchemaVersionV4,
	SchemaVersionV5,
	SchemaVersionV6,
}

func (s SchemaVersionTypes) GetVersionFromNamespace(namespace string) int {
	for i, v := range s {
		if v == namespace {
			return i + 1
		}
	}
	return 1
}

func (s SchemaVersionTypes) GetSchemaNamespace(version int) (string, error) {
	// Versions are internally 0-indexed but stored with a 1 index so decrement it by 1
	if version <= 0 || version > len(s) {
		return "", fmt.Errorf("unknown schema version '%v'", version)
	}
	return s[version-1], nil
}

func (s SchemaVersionTypes) IsKnownSchema(namespace string) bool {
	return nuget.Some(s, func(s string) bool {
		return strings.EqualFold(s, namespace)
	})
}
