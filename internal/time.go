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
	"strings"
	"time"

	"github.com/microbus-io/errors"
)

// compileTime compiles the validation of a time against the tags.
func compileTime(tags []string) []step {
	required := false
	hasDefault := false
	var def time.Time
	for _, t := range tags {
		if t == "notzero" {
			required = true
		} else if !hasDefault && strings.HasPrefix(t, "default=") {
			v, err := parseTime(t[len("default="):])
			if err != nil {
				continue
			}
			def = v
			hasDefault = true
		}
	}
	type valCheck struct {
		operator string
		v        time.Time
	}
	var checks []valCheck
	for _, t := range tags {
		if strings.HasPrefix(t, "val") && len(t) > 4 {
			operator, value, err := splitOpValue(t, 3)
			if err != nil {
				continue
			}
			v, err := parseTime(value)
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
		i := refVal.Interface().(time.Time)
		if i.IsZero() && hasDefault && !def.IsZero() {
			if !refVal.CanSet() {
				return errors.New("data must be passed by reference")
			}
			refVal.Set(reflect.ValueOf(def))
			i = def
		}
		if i.IsZero() && required {
			return errInvalid("non-zero value is required")
		}
		for _, c := range checks {
			switch {
			case c.operator == "<=" && i.After(c.v):
				return errInvalid("must be earlier than or equal to %v", c.v)
			case c.operator == "<" && !i.Before(c.v):
				return errInvalid("must be earlier than %v", c.v)
			case c.operator == ">=" && i.Before(c.v):
				return errInvalid("must be later than or equal to %v", c.v)
			case c.operator == ">" && !i.After(c.v):
				return errInvalid("must be later than %v", c.v)
			case c.operator == "!=" && i.Equal(c.v):
				return errInvalid("must not equal %v", c.v)
			case c.operator == "==" && !i.Equal(c.v):
				return errInvalid("must equal %v", c.v)
			}
		}
		return nil
	}}
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
