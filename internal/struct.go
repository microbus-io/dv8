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

	"github.com/microbus-io/errors"
)

// compileStruct compiles the validation of a struct against the incoming tags
// and the dv8 field tags of each of its fields.
func compileStruct(refType reflect.Type, tags []string, memo map[planKey]*plan) []step {
	var steps []step
	if tagsContain(tags, "notzero") {
		zero := reflect.Zero(refType).Interface()
		steps = append(steps, func(_ context.Context, refVal reflect.Value) error {
			if reflect.DeepEqual(zero, refVal.Interface()) {
				return errInvalid("value is required")
			}
			return nil
		})
	}
	// On runs the validation on a nested field
	for _, t := range tags {
		if strings.HasPrefix(t, "on ") {
			fld, ok := refType.FieldByName(t[3:])
			if !ok {
				continue
			}
			sub := buildPlan(fld.Type, tags, memo)
			if sub.isEmpty() {
				continue
			}
			index := fld.Index
			steps = append(steps, func(ctx context.Context, refVal reflect.Value) error {
				return sub.execute(ctx, refVal.FieldByIndex(index))
			})
		}
	}
	// Iterate over fields
	for i := 0; i < refType.NumField(); i++ {
		fld := refType.Field(i)
		tagVal := fld.Tag.Get("dv8")
		if tagVal == "-" {
			continue
		}
		fldTags := splitDirectives(tagVal)
		if tagsContain(fldTags, "-") {
			continue
		}
		// Delegate fields run validations of the parent struct too
		var delegate *plan
		if tagsContain(fldTags, "delegate") {
			delegate = buildPlan(fld.Type, tags, memo)
			if delegate.isEmpty() {
				delegate = nil
			}
		}
		fieldPlan := buildPlan(fld.Type, fldTags, memo)
		if delegate == nil && fieldPlan.isEmpty() {
			continue
		}
		index := i
		name := fld.Name
		steps = append(steps, func(ctx context.Context, refVal reflect.Value) error {
			rv := refVal.Field(index)
			if delegate != nil {
				err := delegate.execute(ctx, rv)
				if err != nil {
					return errors.New("%s", name, err)
				}
			}
			err := fieldPlan.execute(ctx, rv)
			if err != nil {
				return errors.New("%s", name, err)
			}
			return nil
		})
	}
	return steps
}

// splitDirectives splits a dv8 tag on commas. A comma preceded by a backslash is part of the
// directive's value rather than a separator, and the backslash is removed.
func splitDirectives(tagVal string) []string {
	segments := strings.Split(tagVal, ",")
	directives := []string{segments[0]}
	for _, seg := range segments[1:] {
		last := directives[len(directives)-1]
		if strings.HasSuffix(last, `\`) {
			directives[len(directives)-1] = last[:len(last)-1] + "," + seg
		} else {
			directives = append(directives, seg)
		}
	}
	return directives
}

func tagsContain(tags []string, val string) bool {
	for _, t := range tags {
		if t == val {
			return true
		}
	}
	return false
}
