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

// compileBool compiles the validation of a boolean against the tags.
func compileBool(tags []string) []step {
	required := false
	hasDefault := false
	var def bool
	for _, t := range tags {
		if t == "notzero" {
			required = true
		} else if !hasDefault && strings.HasPrefix(t, "default=") {
			v, err := strconv.ParseBool(t[len("default="):])
			if err != nil {
				continue
			}
			def = v
			hasDefault = true
		}
	}
	type valCheck struct {
		operator string
		v        bool
	}
	var checks []valCheck
	for _, t := range tags {
		if strings.HasPrefix(t, "val") && len(t) > 4 {
			operator, value, err := splitOpValue(t, 3)
			if err != nil {
				continue
			}
			v, err := strconv.ParseBool(value)
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
		b := refVal.Bool()
		if !b && hasDefault && def {
			if !refVal.CanSet() {
				return errors.New("data must be passed by reference")
			}
			refVal.SetBool(def)
			b = def
		}
		if !b && required {
			return errInvalid("non-zero value is required")
		}
		for _, c := range checks {
			switch {
			case c.operator == "!=" && b == c.v:
				return errInvalid("must not equal %v", c.v)
			case c.operator == "==" && b != c.v:
				return errInvalid("must equal %v", c.v)
			}
		}
		return nil
	}}
}
