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

// validateUint validates the value of an unsigned integer against the tags.
func validateUint(refVal reflect.Value, tags []string) (err error) {
	i := refVal.Uint()
	// Default value and required
	required := false
	changed := false
	for _, t := range tags {
		if t == "required" {
			required = true
		} else if i == 0 && strings.HasPrefix(t, "default=") {
			def, err := strconv.ParseUint(t[len("default="):], 10, 64)
			if err != nil {
				return err
			}
			if def != i {
				i = def
				changed = true
			}
		}
	}
	if changed {
		if !refVal.CanSet() {
			return errors.New("data must be passed by reference")
		}
		refVal.SetUint(i)
	}
	if i == 0 && required {
		return errors.New("non-zero value is required")
	}
	// Range constraints
	for _, t := range tags {
		if strings.HasPrefix(t, "val") && len(t) > 4 {
			// Example: val<M
			operator := t[3:4]
			var v uint64
			if t[4] == '=' {
				operator += "="
				v, err = strconv.ParseUint(t[5:], 10, 64)
			} else {
				v, err = strconv.ParseUint(t[4:], 10, 64)
			}
			if err != nil {
				return err
			}
			switch {
			case operator == "<=" && i > v:
				err = fmt.Errorf("must be less than or equal to %d", v)
			case operator == "<" && i >= v:
				err = fmt.Errorf("must be less than %d", v)
			case operator == ">=" && i < v:
				err = fmt.Errorf("must be greater than or equal to %d", v)
			case operator == ">" && i <= v:
				err = fmt.Errorf("must be greater than %d", v)
			case operator == "!=" && i == v:
				err = fmt.Errorf("must not equal %d", v)
			case operator == "==" && i != v:
				err = fmt.Errorf("must equal %d", v)
			case operator != "<=" && operator != "<" && operator != ">=" && operator != ">" && operator != "!=" && operator != "==":
				err = fmt.Errorf("unsupported operator '%s'", operator)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
