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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// validateBool validates the value of a boolean against the tags.
func validateBool(refVal reflect.Value, tags []string) (err error) {
	b := refVal.Bool()
	// Default value and required
	required := false
	changed := false
	for _, t := range tags {
		if t == "required" {
			required = true
		} else if !b && strings.HasPrefix(t, "default=") {
			def, err := strconv.ParseBool(t[len("default="):])
			if err != nil {
				return err
			}
			if def != b {
				b = def
				changed = true
			}
		}
	}
	if changed {
		if !refVal.CanSet() {
			return errors.New("data must be passed by reference")
		}
		refVal.SetBool(b)
	}
	if !b && required {
		return errors.New("non-zero value is required")
	}
	for _, t := range tags {
		if strings.HasPrefix(t, "val") && len(t) > 4 {
			// Example: val<M
			operator := t[3:4]
			var v bool
			if t[4] == '=' {
				operator += "="
				v, err = strconv.ParseBool(t[5:])
			} else {
				v, err = strconv.ParseBool(t[4:])
			}
			if err != nil {
				return err
			}
			switch {
			case operator == "!=" && b == v:
				err = fmt.Errorf("must not equal %v", v)
			case operator == "==" && b != v:
				err = fmt.Errorf("must equal %v", v)
			case operator != "!=" && operator != "==":
				err = fmt.Errorf("unsupported operator '%s'", operator)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
