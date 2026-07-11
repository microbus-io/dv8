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

// compileUint compiles the validation of an unsigned integer against the tags.
func compileUint(tags []string) []step {
	required := false
	hasDefault := false
	var def uint64
	for _, t := range tags {
		if t == "notzero" {
			required = true
		} else if !hasDefault && strings.HasPrefix(t, "default=") {
			v, err := strconv.ParseUint(t[len("default="):], 10, 64)
			if err != nil {
				continue
			}
			def = v
			hasDefault = true
		}
	}
	type valCheck struct {
		operator string
		v        uint64
	}
	var checks []valCheck
	for _, t := range tags {
		if strings.HasPrefix(t, "val") && len(t) > 4 {
			operator, value, err := splitOpValue(t, 3)
			if err != nil {
				continue
			}
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				continue
			}
			checks = append(checks, valCheck{operator, v})
		}
	}
	if !required && !hasDefault && len(checks) == 0 {
		return nil
	}
	return []step{func(_ context.Context, refVal reflect.Value) error {
		u := refVal.Uint()
		if u == 0 && hasDefault && def != 0 {
			if !refVal.CanSet() {
				return errors.New("data must be passed by reference")
			}
			refVal.SetUint(def)
			u = def
		}
		if u == 0 && required {
			return errInvalid("non-zero value is required")
		}
		for _, c := range checks {
			switch {
			case c.operator == "<=" && u > c.v:
				return errInvalid("must be less than or equal to %d", c.v)
			case c.operator == "<" && u >= c.v:
				return errInvalid("must be less than %d", c.v)
			case c.operator == ">=" && u < c.v:
				return errInvalid("must be greater than or equal to %d", c.v)
			case c.operator == ">" && u <= c.v:
				return errInvalid("must be greater than %d", c.v)
			case c.operator == "!=" && u == c.v:
				return errInvalid("must not equal %d", c.v)
			case c.operator == "==" && u != c.v:
				return errInvalid("must equal %d", c.v)
			}
		}
		return nil
	}}
}
