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
	"context"
	"reflect"
	"strconv"
	"strings"

	"github.com/microbus-io/errors"
)

// compileArray compiles the validation of an array or slice against the tags.
// Unprefixed directives apply to the array itself; "each" directives distribute to its elements.
func compileArray(refType reflect.Type, tags []string, memo map[planKey]*plan) []step {
	var eachTags []string
	// Array-level checks run in tag order
	var checks []func(refVal reflect.Value) error
	for _, t := range tags {
		if strings.HasPrefix(t, "each ") {
			eachTags = append(eachTags, t[len("each "):])
			continue
		}
		if t == "notzero" && refType.Kind() == reflect.Slice {
			checks = append(checks, func(refVal reflect.Value) error {
				if refVal.IsNil() {
					return errInvalid("value is required")
				}
				return nil
			})
		}
		if strings.HasPrefix(t, "len") && len(t) > 4 {
			operator, value, err := splitOpValue(t, 3)
			if err != nil {
				continue
			}
			l, err := strconv.Atoi(value)
			if err != nil {
				continue
			}
			checks = append(checks, compileLenCheck(operator, l))
		}
	}
	elemPlan := buildPlan(refType.Elem(), eachTags, memo)
	if elemPlan.isEmpty() {
		elemPlan = nil
	}
	if len(checks) == 0 && elemPlan == nil {
		return nil
	}
	return []step{func(ctx context.Context, refVal reflect.Value) error {
		for _, check := range checks {
			err := check(refVal)
			if err != nil {
				return err
			}
		}
		if elemPlan == nil {
			return nil
		}
		// Nested elements
		for j := 0; j < refVal.Len(); j++ {
			err := elemPlan.execute(ctx, refVal.Index(j))
			if err != nil {
				return errors.New("[%d]", j, err)
			}
		}
		return nil
	}}
}
