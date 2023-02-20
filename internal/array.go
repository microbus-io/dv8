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
	if tagsContain(tags, "required") && (!refVal.IsValid() || refVal.IsNil()) {
		return errors.New("value is required")
	}
	if !refVal.IsValid() {
		return nil
	}
	// Length
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
		}
	}
	arrayType := refType.Elem()
	switch arrayType.Kind() {
	case reflect.Pointer, reflect.Struct, reflect.Array, reflect.Slice, reflect.Map:
		for j := 0; j < refVal.Len(); j++ {
			val := refVal.Index(j)
			err = validateAny(arrayType, val, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
