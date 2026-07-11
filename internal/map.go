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
// Unprefixed directives apply to the map itself; "each" directives distribute to its values
// and "key" directives to its keys. A mutated key is reinserted under its new value;
// two keys folding into one is an error.
func validateMap(ctx context.Context, refType reflect.Type, refVal reflect.Value, tags []string) (err error) {
	var eachTags []string
	var keyTags []string
	for _, t := range tags {
		if strings.HasPrefix(t, "each ") {
			eachTags = append(eachTags, t[len("each "):])
			continue
		}
		if strings.HasPrefix(t, "key ") {
			keyTags = append(keyTags, t[len("key "):])
			continue
		}
		if t == "notzero" && refVal.IsNil() {
			return errors.New("value is required")
		}
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
				err = fmt.Errorf("%w: unsupported operator '%s'", ErrDirective, operator)
			}
			if err != nil {
				return err
			}
		}
	}
	// Nested elements
	keyType := refType.Key()
	mapType := refType.Elem()
	var renamedKeys [][2]reflect.Value // old, new
	iter := refVal.MapRange()
	for iter.Next() {
		if len(keyTags) > 0 {
			key := iter.Key()
			if refVal.CanSet() {
				// Create an addressable copy of the key
				key = reflect.New(keyType).Elem()
				key.Set(iter.Key())
			}
			err = validateAny(ctx, keyType, key, keyTags)
			if err != nil {
				return fmt.Errorf("[%v] key: %w", iter.Key(), err)
			}
			if refVal.CanSet() && !key.Equal(iter.Key()) {
				renamedKeys = append(renamedKeys, [2]reflect.Value{iter.Key(), key})
			}
		}
		val := iter.Value()
		if refVal.CanSet() {
			// Create an addressable copy of the value item
			val = reflect.New(mapType).Elem()
			val.Set(iter.Value())
		}
		err = validateAny(ctx, mapType, val, eachTags)
		if err != nil {
			return fmt.Errorf("[%v]: %w", iter.Key(), err)
		}
		if refVal.CanSet() {
			refVal.SetMapIndex(iter.Key(), val)
		}
	}
	// Reinsert values under keys mutated by "key" directives
	for _, r := range renamedKeys {
		val := refVal.MapIndex(r[0])
		refVal.SetMapIndex(r[0], reflect.Value{})
		if refVal.MapIndex(r[1]).IsValid() {
			return fmt.Errorf("[%v] key: [%v] already exists", r[0], r[1])
		}
		refVal.SetMapIndex(r[1], val)
	}
	return nil
}
