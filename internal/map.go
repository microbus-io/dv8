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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// validateMap validates the value of a map against the tags.
func validateMap(ctx context.Context, refType reflect.Type, refVal reflect.Value, tags []string) (err error) {
	// Length
	for i, t := range tags {
		if strings.HasPrefix(t, "maplen") && len(t) > 7 {
			if refVal.IsNil() {
				return errors.New("value is required")
			}
			// Example: maplen<8
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
			mapLen := refVal.Len()
			switch {
			case operator == "<=" && mapLen > l:
				err = fmt.Errorf("length must be less than or equal to %d", l)
			case operator == "<" && mapLen >= l:
				err = fmt.Errorf("length must be less than %d", l)
			case operator == ">=" && mapLen < l:
				err = fmt.Errorf("length must be greater than or equal to %d", l)
			case operator == ">" && mapLen <= l:
				err = fmt.Errorf("length must be greater than %d", l)
			case operator == "!=" && mapLen == l:
				err = fmt.Errorf("length must not equal %d", l)
			case operator == "==" && mapLen != l:
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
	mapType := refType.Elem()
	iter := refVal.MapRange()
	for iter.Next() {
		val := iter.Value()
		if refVal.CanSet() {
			// Create an addressable copy of the value item
			val = reflect.New(mapType).Elem()
			val.Set(iter.Value())
		}
		err = validateAny(ctx, mapType, val, tags)
		if err != nil {
			return fmt.Errorf("[%v]: %w", iter.Key(), err)
		}
		if refVal.CanSet() {
			refVal.SetMapIndex(iter.Key(), val)
		}
	}
	return nil
}
