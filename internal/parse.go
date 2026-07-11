/*
Copyright 2023-2024 Microbus LLC and various contributors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package internal

import (
	"strings"
)

// Directive is one parsed directive of a dv8 struct tag.
type Directive struct {
	Prefixes []string // "each" and "key" prefixes, outermost first
	Name     string   // "notzero", "trim", "tolower", "toupper", "len", "val", "oneof", "regexp", "default", "delegate", "on", "-", or the verbatim text of an unrecognized directive
	Operator string   // the comparison operator of a "len" or "val" directive
	Value    string   // the operand: the bound, pattern, set of values, default value, or field name
}

// ParseTag parses the value of a dv8 struct tag into its directives.
// It performs no validity checks; Compile is the arbiter of validity.
func ParseTag(tag string) []Directive {
	var parsed []Directive
	for _, d := range splitDirectives(tag) {
		if d == "" {
			continue
		}
		dir := Directive{}
		for {
			if rest, ok := strings.CutPrefix(d, "each "); ok {
				dir.Prefixes = append(dir.Prefixes, "each")
				d = rest
				continue
			}
			if rest, ok := strings.CutPrefix(d, "key "); ok {
				dir.Prefixes = append(dir.Prefixes, "key")
				d = rest
				continue
			}
			break
		}
		switch {
		case d == "notzero" || d == "trim" || d == "tolower" || d == "toupper" || d == "delegate" || d == "-":
			dir.Name = d
		case strings.HasPrefix(d, "default="):
			dir.Name = "default"
			dir.Value = d[len("default="):]
		case strings.HasPrefix(d, "oneof "):
			dir.Name = "oneof"
			dir.Value = d[len("oneof "):]
		case strings.HasPrefix(d, "regexp "):
			dir.Name = "regexp"
			dir.Value = d[len("regexp "):]
		case strings.HasPrefix(d, "on "):
			dir.Name = "on"
			dir.Value = d[len("on "):]
		case strings.HasPrefix(d, "len") || strings.HasPrefix(d, "val"):
			operator, value, err := splitOpValue(d, 3)
			if err != nil {
				dir.Name = d
			} else {
				dir.Name = d[:3]
				dir.Operator = operator
				dir.Value = value
			}
		default:
			dir.Name = d
		}
		parsed = append(parsed, dir)
	}
	return parsed
}
