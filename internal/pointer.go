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
	"reflect"
)

// validatePointer validates the value of a pointer against the tags.
func validatePointer(refType reflect.Type, refVal reflect.Value, tags []string) (err error) {
	if refVal.IsNil() {
		if tagsContain(tags, "required") {
			return errors.New("value is required")
		}
		return nil
	}
	return validateAny(refType.Elem(), refVal.Elem(), tags)
}
