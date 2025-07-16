// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package version

import (
	"fmt"
	"strings"
	"sync"
)

var (
	All                       *VersionRange
	parsedVersionRangeMapping versionRangeMapping
	// parsedVersionRangeMappingMaxEntries If dictionary exceeds this size, parsedVersionRangeMapping will be cleared.
	parsedVersionRangeMappingMaxEntries = 500
)

func init() {
	versionRange, _ := NewVersionRange(nil, nil, true, true, nil, "")
	All = versionRange
	parsedVersionRangeMapping = versionRangeMapping{
		versionMap: make(map[versionKey]VersionRange),
	}
}

type versionKey struct {
	value         string
	allowFloating bool
}

type versionRangeMapping struct {
	mu         sync.RWMutex
	versionMap map[versionKey]VersionRange
}

func (c *versionRangeMapping) setVersion(value string, allowFloating bool, version VersionRange) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.versionMap) >= parsedVersionRangeMappingMaxEntries {
		c.versionMap = make(map[versionKey]VersionRange)
	}
	c.versionMap[versionKey{
		value:         value,
		allowFloating: allowFloating,
	}] = version
}

func (c *versionRangeMapping) getVersion(value string, allowFloating bool) (VersionRange, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	version, ok := c.versionMap[versionKey{
		value:         value,
		allowFloating: allowFloating,
	}]
	return version, ok
}

// ParseRange The version string is either a simple version or an arithmetic range
//
//	e.g.
//	1.0         --> 1.0 ≤ x
//	(,1.0]      --> x ≤ 1.0
//
// (,1.0)      --> x &lt; 1.0
// [1.0]       --> x == 1.0
// (1.0,)      --> 1.0 &lt; x
// (1.0, 2.0)   --> 1.0 &lt; x &lt; 2.0
// [1.0, 2.0]   --> 1.0 ≤ x ≤ 2.0
func ParseRange(value string) (*VersionRange, error) {
	return ParseRangeWithRequired(value, true)
}

func ParseRangeWithRequired(value string, allowFloating bool) (*VersionRange, error) {
	if v, ok := TryParseRange(value, allowFloating); !ok {
		return nil, fmt.Errorf("'%s' is not a valid version string", value)
	} else {
		return v, nil
	}
}

// TryParseRange  Parses a VersionRange from its string representation.
func TryParseRange(value string, allowFloating bool) (*VersionRange, bool) {
	trimmedValue := strings.TrimSpace(value)
	if strings.TrimSpace(trimmedValue) == "" {
		return nil, false
	}
	if v, ok := parsedVersionRangeMapping.getVersion(trimmedValue, allowFloating); ok {
		return &v, true
	}
	runes := []rune(trimmedValue)

	// * is the only 1 char range
	if allowFloating && len(runes) == 1 && runes[0] == '*' {
		floatRange, err := ParseFloatRange(trimmedValue)
		if err != nil {
			return nil, false
		}
		if v, err := NewVersionRange(NewVersionFrom(0, 0, 0, "", ""), nil, true, true, floatRange, value); err != nil {
			return nil, false
		} else {
			parsedVersionRangeMapping.setVersion(value, allowFloating, *v)
			return v, true
		}
	}
	var minVersionString, maxVersionString string
	var isMinInclusive, isMaxInclusive bool
	var minVersion, maxVersion *Version
	var floatRange *FloatRange
	if runes[0] == '(' || runes[0] == '[' {
		// The first character must be [ to (
		switch runes[0] {
		case '[':
			isMinInclusive = true
		case '(':
			isMinInclusive = false
		default:
			return nil, false
		}
		// The last character must be ] ot )
		switch runes[len(runes)-1] {
		case ']':
			isMaxInclusive = true
		case ')':
			isMaxInclusive = false
		default:
			return nil, false
		}
		// Get rid of the two brackets
		trimmedValue = trimmedValue[1 : len(trimmedValue)-1]

		// Split by comma, and make sure we don't get more than two pieces
		parts := strings.Split(trimmedValue, ",")
		if len(parts) > 2 {
			return nil, false
		} else {
			allEmpty := true
			for i := 0; i < len(parts); i++ {
				if strings.TrimSpace(parts[i]) != "" {
					allEmpty = false
					break
				}
			}
			// If all parts are empty, then neither of upper or lower bounds were specified. Version spec is of the
			// format (,]
			if allEmpty {
				return nil, false
			}
		}
		// (1.0.0] and [1.0.0),(1.0.0) are invalid.
		if len(parts) == 1 && !(isMinInclusive && isMaxInclusive) {
			return nil, false
		}

		// If there is only one piece, we use it for both min and max
		minVersionString = parts[0]
		if len(parts) == 2 {
			maxVersionString = parts[1]
		} else {
			maxVersionString = parts[0]
		}
	} else {
		// default to min inclusive when there are no braces
		isMinInclusive = true

		// use the entire value as the version
		minVersionString = trimmedValue
	}
	if strings.TrimSpace(minVersionString) != "" {
		if allowFloating && strings.Contains(minVersionString, "*") {
			if float, ok := TryParseFloatRange(minVersionString); ok && float.HasMinVersion() {
				minVersion = float.MinVersion
			} else {
				// invalid float
				return nil, false
			}
		} else {
			// single non-floating version
			if v, err := Parse(minVersionString); err != nil {
				return nil, false
			} else {
				minVersion = v
			}
		}
	}
	// parse the max version string, the max cannot float
	if strings.TrimSpace(maxVersionString) != "" {
		if v, err := Parse(maxVersionString); err != nil {
			return nil, false
		} else {
			maxVersion = v
		}
	}
	if minVersion != nil && maxVersion != nil {
		result := minVersion.Semver.Compare(maxVersion.Semver)
		// minVersion > maxVersion
		if result > 0 {
			return nil, false
		}
		if result == 0 && (isMinInclusive != isMaxInclusive) {
			return nil, false
		}
	}
	if v, err := NewVersionRange(minVersion, maxVersion, isMinInclusive, isMaxInclusive, floatRange, value); err != nil {
		return nil, false
	} else {
		// Successful parse!
		parsedVersionRangeMapping.setVersion(value, allowFloating, *v)
		return v, true
	}
}
