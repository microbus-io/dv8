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
)

// validateStruct takes in a data struct and validates each of its fields given their dv8 field tags.
func validateStruct(refType reflect.Type, refVal reflect.Value, tags []string) (err error) {
	if tagsContain(tags, "required") && !refVal.IsValid() {
		return errors.New("value is required")
	}
	if !refVal.IsValid() {
		return nil
	}
	// Iterate over fields
	for i := 0; i < refType.NumField(); i++ {
		fld := refType.Field(i)
		tagVal := fld.Tag.Get("dv8")
		if tagVal == "-" {
			continue
		}
		tags := strings.Split(tagVal, ",")
		if tagsContain(tags, "-") {
			continue
		}
		rt := refType.Field(i).Type
		rv := refVal.Field(i)
		err := validateAny(rt, rv, tags)
		if err != nil {
			return fmt.Errorf("%s: %w", fld.Name, err)
		}
	}
	return nil
}

func tagsContain(tags []string, val string) bool {
	for _, t := range tags {
		if t == val {
			return true
		}
	}
	return false
}
