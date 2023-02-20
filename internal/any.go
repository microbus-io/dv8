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
	refType, refVal = followPointers(refType, refVal)

	switch refType.String() {
	case "string":
		return validateString(refVal, tags)
	case "int", "int8", "int16", "int32", "int64":
		return validateInt(refVal, tags)
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return validateUint(refVal, tags)
	case "float32", "float64":
		return validateFloat(refVal, tags)
	case "bool":
		return validateBool(refVal, tags)
	case "time.Duration":
		return validateDuration(refVal, tags)
	case "time.Time":
		return validateTime(refVal, tags)
	}

	switch refType.Kind() {
	case reflect.Struct:
		return validateStruct(refType, refVal, tags)
	case reflect.Map:
		return validateMap(refType, refVal, tags)
	case reflect.Array, reflect.Slice:
		return validateArray(refType, refVal, tags)
	}

	return nil
}

// followPointers follows pointers to the type and value of the element
func followPointers(refType reflect.Type, refVal reflect.Value) (reflect.Type, reflect.Value) {
	for refType.Kind() == reflect.Pointer {
		refVal = refVal.Elem()
		refType = refType.Elem()
	}
	return refType, refVal
}
