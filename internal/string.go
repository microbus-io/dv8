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
	"regexp"
	"strconv"
	"strings"
)

// validateString validates the value of a string against the tags.
func validateString(refVal reflect.Value, tags []string) (err error) {
	s := refVal.String()
	// Trim spaces
	changed := false
	if !tagsContain(tags, "notrim") {
		trimmed := strings.TrimSpace(s)
		if trimmed != s {
			s = trimmed
			changed = true
		}
	}
	// Default value and required
	required := false
	for _, t := range tags {
		if t == "required" {
			required = true
		} else if t == "toupper" && s != strings.ToUpper(s) {
			s = strings.ToUpper(s)
			changed = true
		} else if t == "tolower" && s != strings.ToLower(s) {
			s = strings.ToLower(s)
			changed = true
		} else if s == "" && strings.HasPrefix(t, "default=") {
			def := t[len("default="):]
			if def != s {
				s = def
				changed = true
			}
		}
	}
	if changed {
		if !refVal.CanSet() {
			return errors.New("data must be passed by reference")
		}
		refVal.SetString(s)
	}
	if s == "" && required {
		return errors.New("value is required")
	}
	// Other constraints
	for _, t := range tags {
		if strings.HasPrefix(t, "len") && len(t) > 4 {
			// Example: len<8
			operator := t[3:4]
			var l int
			if t[4] == '=' {
				operator += "="
				l, err = strconv.Atoi(t[5:])
			} else {
				l, err = strconv.Atoi(t[4:])
			}
			if err != nil {
				return err
			}
			strLen := len([]rune(s))
			switch {
			case operator == "<=" && strLen > l:
				err = fmt.Errorf("length must be less than or equal to %d", l)
			case operator == "<" && strLen >= l:
				err = fmt.Errorf("length must be less than %d", l)
			case operator == ">=" && strLen < l:
				err = fmt.Errorf("length must be greater than or equal to %d", l)
			case operator == ">" && strLen <= l:
				err = fmt.Errorf("length must be greater than %d", l)
			case operator == "!=" && strLen == l:
				err = fmt.Errorf("length must not equal %d", l)
			case operator == "==" && strLen != l:
				err = fmt.Errorf("length must equal %d", l)
			case operator != "<=" && operator != "<" && operator != ">=" && operator != ">" && operator != "!=" && operator != "==":
				err = fmt.Errorf("unsupported operator '%s'", operator)
			}
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(t, "val") && len(t) > 4 {
			// Example: val<M
			operator := t[3:4]
			var v string
			if t[4] == '=' {
				operator += "="
				v = t[5:]
			} else {
				v = t[4:]
			}
			if err != nil {
				return err
			}
			switch {
			case operator == "<=" && s > v:
				err = fmt.Errorf("must be less than or equal to '%s'", v)
			case operator == "<" && s >= v:
				err = fmt.Errorf("must be less than '%s'", v)
			case operator == ">=" && s < v:
				err = fmt.Errorf("must be greater than or equal to '%s'", v)
			case operator == ">" && s <= v:
				err = fmt.Errorf("must be greater than '%s'", v)
			case operator == "!=" && s == v:
				err = fmt.Errorf("must not equal '%s'", v)
			case operator == "==" && s != v:
				err = fmt.Errorf("must equal '%s'", v)
			case operator != "<=" && operator != "<" && operator != ">=" && operator != ">" && operator != "!=" && operator != "==":
				err = fmt.Errorf("unsupported operator '%s'", operator)
			}
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(t, "regexp ") && len(t) > 7 {
			re, err := regexp.Compile(t[7:])
			if err != nil {
				return err
			}
			if !re.Match([]byte(s)) {
				return errors.New("value doesn't match required pattern")
			}
		} else if strings.HasPrefix(t, "oneof ") && len(t) > 6 {
			validVals := strings.Split(t[6:], "|")
			found := false
			for _, v := range validVals {
				if s == v {
					found = true
					break
				}
			}
			if !found {
				return errors.New("value must be one of " + t[6:])
			}
		}
	}
	return nil
}
