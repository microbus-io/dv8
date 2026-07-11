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
	"strconv"
	"strings"

	"github.com/microbus-io/errors"
)

// compileMap compiles the validation of a map against the tags.
// Unprefixed directives apply to the map itself; "each" directives distribute to its values
// and "key" directives to its keys. A mutated key is reinserted under its new value;
// two keys folding into one is an error.
func compileMap(refType reflect.Type, tags []string, memo map[planKey]*plan) []step {
	var eachTags []string
	var keyTags []string
	// Map-level checks run in tag order
	var checks []func(refVal reflect.Value) error
	for _, t := range tags {
		if strings.HasPrefix(t, "each ") {
			eachTags = append(eachTags, t[len("each "):])
			continue
		}
		if strings.HasPrefix(t, "key ") {
			keyTags = append(keyTags, t[len("key "):])
			continue
		}
		if t == "notzero" {
			checks = append(checks, func(refVal reflect.Value) error {
				if refVal.IsNil() {
					return errInvalid("value is required")
				}
				return nil
			})
		}
		if strings.HasPrefix(t, "len") && len(t) > 4 {
			operator, value, err := splitOpValue(t, 3)
			if err != nil {
				continue
			}
			l, err := strconv.Atoi(value)
			if err != nil {
				continue
			}
			checks = append(checks, compileLenCheck(operator, l))
		}
	}
	keyType := refType.Key()
	elemType := refType.Elem()
	var keyPlan *plan
	if len(keyTags) > 0 {
		keyPlan = buildPlan(keyType, keyTags, memo)
		if keyPlan.isEmpty() {
			keyPlan = nil
		}
	}
	valPlan := buildPlan(elemType, eachTags, memo)
	if valPlan.isEmpty() {
		valPlan = nil
	}
	if len(checks) == 0 && keyPlan == nil && valPlan == nil {
		return nil
	}
	return []step{func(ctx context.Context, refVal reflect.Value) error {
		for _, check := range checks {
			err := check(refVal)
			if err != nil {
				return err
			}
		}
		if keyPlan == nil && valPlan == nil {
			return nil
		}
		// Nested elements
		var renamedKeys [][2]reflect.Value // old, new
		iter := refVal.MapRange()
		for iter.Next() {
			if keyPlan != nil {
				key := iter.Key()
				if refVal.CanSet() {
					// Create an addressable copy of the key
					key = reflect.New(keyType).Elem()
					key.Set(iter.Key())
				}
				err := keyPlan.execute(ctx, key)
				if err != nil {
					return errors.New("[%v] key", iter.Key(), err)
				}
				if refVal.CanSet() && !key.Equal(iter.Key()) {
					renamedKeys = append(renamedKeys, [2]reflect.Value{iter.Key(), key})
				}
			}
			if valPlan != nil {
				val := iter.Value()
				if refVal.CanSet() {
					// Create an addressable copy of the value item
					val = reflect.New(elemType).Elem()
					val.Set(iter.Value())
				}
				err := valPlan.execute(ctx, val)
				if err != nil {
					return errors.New("[%v]", iter.Key(), err)
				}
				if refVal.CanSet() {
					refVal.SetMapIndex(iter.Key(), val)
				}
			}
		}
		// Reinsert values under keys mutated by "key" directives
		for _, r := range renamedKeys {
			val := refVal.MapIndex(r[0])
			refVal.SetMapIndex(r[0], reflect.Value{})
			if refVal.MapIndex(r[1]).IsValid() {
				return errInvalid("[%v] key: [%v] already exists", r[0], r[1])
			}
			refVal.SetMapIndex(r[1], val)
		}
		return nil
	}}
}

// compileLenCheck compiles a length constraint on a map, array, or slice.
func compileLenCheck(operator string, l int) func(refVal reflect.Value) error {
	return func(refVal reflect.Value) error {
		length := refVal.Len()
		switch {
		case operator == "<=" && length > l:
			return errInvalid("length must be less than or equal to %d", l)
		case operator == "<" && length >= l:
			return errInvalid("length must be less than %d", l)
		case operator == ">=" && length < l:
			return errInvalid("length must be greater than or equal to %d", l)
		case operator == ">" && length <= l:
			return errInvalid("length must be greater than %d", l)
		case operator == "!=" && length == l:
			return errInvalid("length must not equal %d", l)
		case operator == "==" && length != l:
			return errInvalid("length must equal %d", l)
		}
		return nil
	}
}
