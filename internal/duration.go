/*
Copyright 2023 Microbus LLC and various contributors
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
	"strings"
	"time"
)

// validateDuration validates the value of a duration against the tags.
func validateDuration(refVal reflect.Value, tags []string) (err error) {
	var d time.Duration
	if refVal.IsValid() {
		d = time.Duration(refVal.Int())
	}
	// Default value and required
	required := false
	changed := false
	for _, t := range tags {
		if t == "required" {
			required = true
		} else if d == 0 && strings.HasPrefix(t, "default=") {
			def, err := time.ParseDuration(t[len("default="):])
			if err != nil {
				return err
			}
			if def != d {
				d = def
				changed = true
			}
		}
	}
	if changed {
		if !refVal.CanSet() {
			return errors.New("data must be passed by reference")
		}
		refVal.SetInt(int64(d))
	}
	if d == 0 && required {
		return errors.New("non-zero value is required")
	}
	// Range constraints
	for _, t := range tags {
		if strings.HasPrefix(t, "val") && len(t) > 4 {
			// Example: val<2s
			operator := t[3:4]
			var v time.Duration
			if t[4] == '=' {
				operator += "="
				v, err = time.ParseDuration(t[5:])
			} else {
				v, err = time.ParseDuration(t[4:])
			}
			if err != nil {
				return err
			}
			switch {
			case operator == "<=" && d > v:
				err = fmt.Errorf("must be less than or equal to %v", v)
			case operator == "<" && d >= v:
				err = fmt.Errorf("must be less than %v", v)
			case operator == ">=" && d < v:
				err = fmt.Errorf("must be greater than or equal to %v", v)
			case operator == ">" && d <= v:
				err = fmt.Errorf("must be greater than %v", v)
			case operator == "!=" && d == v:
				err = fmt.Errorf("must not equal %v", v)
			case operator == "==" && d != v:
				err = fmt.Errorf("must equal %v", v)
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
