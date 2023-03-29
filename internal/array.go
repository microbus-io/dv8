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
	"strconv"
	"strings"
)

// validateArray validates the value of an array against the tags.
func validateArray(refType reflect.Type, refVal reflect.Value, tags []string) (err error) {
	// Length
	for i, t := range tags {
		if strings.HasPrefix(t, "arrlen") && len(t) > 7 {
			if refVal.IsNil() {
				return errors.New("value is required")
			}
			// Example: arrlen<8
			operator := t[6:7]
			var l int
			if t[7] == '=' {
				operator += "="
				l, err = strconv.Atoi(t[8:])
			} else {
				l, err = strconv.Atoi(t[7:])
			}
			if err != nil {
				return err
			}
			arrayLen := refVal.Len()
			switch {
			case operator == "<=" && arrayLen > l:
				err = fmt.Errorf("length must be less than or equal to %d", l)
			case operator == "<" && arrayLen >= l:
				err = fmt.Errorf("length must be less than %d", l)
			case operator == ">=" && arrayLen < l:
				err = fmt.Errorf("length must be greater than or equal to %d", l)
			case operator == ">" && arrayLen <= l:
				err = fmt.Errorf("length must be greater than %d", l)
			case operator == "!=" && arrayLen == l:
				err = fmt.Errorf("length must not equal %d", l)
			case operator == "==" && arrayLen != l:
				err = fmt.Errorf("length must equal %d", l)
			case operator != "<=" && operator != "<" && operator != ">=" && operator != ">" && operator != "!=" && operator != "==":
				err = fmt.Errorf("unsupported operator '%s'", operator)
			}
			if err != nil {
				return err
			}
			tags[i] = "" // Do not apply to items
		}
	}
	// Nested elements
	arrayType := refType.Elem()
	for j := 0; j < refVal.Len(); j++ {
		val := refVal.Index(j)
		err = validateAny(arrayType, val, tags)
		if err != nil {
			return fmt.Errorf("[%d]: %w", j, err)
		}
	}
	return nil
}
