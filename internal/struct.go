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
	"strings"
)

// validateStruct takes in a data struct and validates each of its fields given their dv8 field tags.
func validateStruct(ctx context.Context, refType reflect.Type, refVal reflect.Value, structTags []string) (err error) {
	if tagsContain(structTags, "notzero") {
		zero := reflect.Zero(refType)
		if reflect.DeepEqual(zero.Interface(), refVal.Interface()) {
			return errors.New("value is required")
		}
	}
	// On runs the validation on a nested field
	for _, t := range structTags {
		if strings.HasPrefix(t, "on ") {
			fld, ok := refType.FieldByName(t[3:])
			if ok {
				rt := fld.Type
				rv := refVal.FieldByName(t[3:])
				err = validateAny(ctx, rt, rv, structTags)
				if err != nil {
					return err
				}
			}
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
		rt := fld.Type
		rv := refVal.Field(i)
		// Delegate fields run validations of the parent struct too
		if tagsContain(fldTags, "delegate") {
			err = validateAny(ctx, rt, rv, structTags)
			if err != nil {
				return fmt.Errorf("%s: %w", fld.Name, err)
			}
		}
		err = validateAny(ctx, rt, rv, fldTags)
		if err != nil {
			return fmt.Errorf("%s: %w", fld.Name, err)
		}
	}
	return nil
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
