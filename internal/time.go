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

// validateTime validates the value of a time against the tags.
func validateTime(refVal reflect.Value, tags []string) (err error) {
	var i time.Time
	if refVal.IsValid() {
		i = refVal.Interface().(time.Time)
	}
	// Default value and required
	required := false
	changed := false
	for _, t := range tags {
		if t == "required" {
			required = true
		} else if i.IsZero() && strings.HasPrefix(t, "default=") {
			def, err := parseTime(t[len("default="):])
			if err != nil {
				return err
			}
			if !def.Equal(i) {
				i = def
				changed = true
			}
		}
	}
	if changed {
		if !refVal.CanSet() {
			return errors.New("data must be passed by reference")
		}
		refVal.Set(reflect.ValueOf(i))
	}
	if i.IsZero() && required {
		return errors.New("non-zero value is required")
	}
	// Range constraints
	for _, t := range tags {
		if strings.HasPrefix(t, "val") && len(t) > 4 {
			// Example: val<2006-01-02
			operator := t[3:4]
			var v time.Time
			if t[4] == '=' {
				operator += "="
				v, err = parseTime(t[5:])
			} else {
				v, err = parseTime(t[4:])
			}
			if err != nil {
				return err
			}
			switch {
			case operator == "<=" && i.After(v):
				err = fmt.Errorf("must be earlier than or equal to %v", v)
			case operator == "<" && !i.Before(v):
				err = fmt.Errorf("must be earlier than %v", v)
			case operator == ">=" && i.Before(v):
				err = fmt.Errorf("must be later than or equal to %v", v)
			case operator == ">" && !i.After(v):
				err = fmt.Errorf("must be later than %v", v)
			case operator == "!=" && i.Equal(v):
				err = fmt.Errorf("must not equal %v", v)
			case operator == "==" && !i.Equal(v):
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

func parseTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	layout := ""
	if len(value) == 10 &&
		value[4] == '-' && value[7] == '-' {
		layout = "2006-01-02"
	} else if len(value) == 19 &&
		value[4] == '-' && value[7] == '-' &&
		value[10] == 'T' && value[13] == ':' && value[16] == ':' {
		layout = "2006-01-02T15:04:05"
	} else if len(value) == 19 &&
		value[4] == '-' && value[7] == '-' &&
		value[10] == ' ' && value[13] == ':' && value[16] == ':' {
		layout = "2006-01-02 15:04:05"
	} else if len(value) >= 20 &&
		value[4] == '-' && value[7] == '-' &&
		value[10] == 'T' && value[13] == ':' && value[16] == ':' {
		layout = time.RFC3339Nano
		if value[19] != '.' {
			layout = time.RFC3339
		}
	}
	return time.Parse(layout, value)
}

func mustParseTime(value string) time.Time {
	t, err := parseTime(value)
	if err != nil {
		panic(err)
	}
	return t
}
