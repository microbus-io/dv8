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

// Validate takes in a reference to a data struct (pointer, map of, slice of)
// and validates each of its fields against their dv8 field tags.
// It recurses into nested structs.
func Validate(data any) error {
	return validateAny(reflect.TypeOf(data), reflect.ValueOf(data), nil)
}

// Validator implements a single method that returns an error if a struct is invalid.
// DV8 calls this function during validation on types that implements it.
type Validator interface {
	Validate() error
}
