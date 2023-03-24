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
	"reflect"
)

// validateAny validates the value of any type against the tags.
func validateAny(refType reflect.Type, refVal reflect.Value, tags []string) (err error) {
	switch refType.String() {
	case "time.Duration":
		return validateDuration(refVal, tags)
	case "time.Time":
		return validateTime(refVal, tags)
	}

	switch refType.Kind() {
	case reflect.String:
		return validateString(refVal, tags)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return validateInt(refVal, tags)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return validateUint(refVal, tags)
	case reflect.Float32, reflect.Float64:
		return validateFloat(refVal, tags)
	case reflect.Bool:
		return validateBool(refVal, tags)
	case reflect.Pointer:
		return validatePointer(refType, refVal, tags)
	case reflect.Struct:
		return validateStruct(refType, refVal, tags)
	case reflect.Map:
		return validateMap(refType, refVal, tags)
	case reflect.Array, reflect.Slice:
		return validateArray(refType, refVal, tags)
	}

	return nil
}
