// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"path/filepath"
	"strings"
)

var (
	invalidSourceCharacters = []rune{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F,
		'"', '<', '>', '|',
	}
	referenceFileInvalidCharacters = append(invalidSourceCharacters, ':', '*', '?', '\\', '/')

	invalidTargetCharacters = func() []rune {
		m := map[rune]bool{'\\': true, '/': true}
		var result []rune
		for _, r := range referenceFileInvalidCharacters {
			if !m[r] {
				result = append(result, r)
			}
		}
		return result
	}()
)

type ManifestFile struct {
	Source  string
	Target  string
	Exclude string
}

func (m *ManifestFile) SetTarget(value string) {
	if value == "" {
		m.Target = value
	} else {
		if filepath.Separator == '/' && strings.Contains(value, "\\") {
			m.Target = strings.ReplaceAll(value, "\\", "/")
		} else {
			m.Target = strings.ReplaceAll(value, "/", "\\")
		}
	}
}
func (m *ManifestFile) Validate() []string {
	var errs []string
	if strings.TrimSpace(m.Source) == "" {
		errs = append(errs, "Missing required metadata: Source")
	} else if strings.ContainsAny(m.Source, string(invalidSourceCharacters)) {
		errs = append(errs, fmt.Sprintf("Source contains invalid characters: %q", m.Source))
	}
	if m.Target != "" && strings.ContainsAny(m.Target, string(invalidTargetCharacters)) {
		errs = append(errs, fmt.Sprintf("Target contains invalid characters: %q", m.Target))
	}
	if m.Exclude != "" && strings.ContainsAny(m.Exclude, string(invalidSourceCharacters)) {
		errs = append(errs, fmt.Sprintf("Exclude contains invalid characters: %q", m.Exclude))
	}
	return errs
}

type ManifestContentFiles struct {
	Include      string
	Exclude      string
	BuildAction  string
	CopyToOutput string
	Flatten      string
}
