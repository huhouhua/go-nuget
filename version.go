// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"fmt"
	"math"
)

var emptyVersion = &Version{0, 0, 0, 0}
var maxVersion = &Version{math.MaxInt, 0, 0, 0}

type Version struct {
	Major    int
	Minor    int
	Build    int
	Revision int
}

func NewVersion(major, minor, build, revision int) (*Version, error) {
	if major < 0 {
		return nil, fmt.Errorf("argument out of range version: %d", major)
	}
	if minor < 0 {
		return nil, fmt.Errorf("argument out of range version: %d", minor)
	}
	if build < 0 {
		return nil, fmt.Errorf("argument out of range version: %d", build)
	}
	if revision < 0 {
		return nil, fmt.Errorf("argument out of range version: %d", revision)
	}
	return &Version{
		Major:    major,
		Minor:    minor,
		Build:    build,
		Revision: revision,
	}, nil
}
