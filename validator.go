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

package dv8

import (
	"github.com/microbus-io/dv8/internal"
)

/*
Validate takes in a reference to one or more data struct (pointer, map of, slice of)
and validates each of its fields against their dv8 field tags.
It recurses into nested structs.

Example:

	type Person struct {
		First   string `dv8:"required,len<=32"`
		Last    string `dv8:"required,len<=32"`
		Age     int    `dv8:"val>=18,val<=120"`
		State   string `dv8:"len==2,default=CA"`
		Zip     string `dv8:"required,regexp ^[0-9]{5}$"`
	}

	p := &Person{
		First: "Jane",
		Last:  "Simmons",
		State: "",        // Set default to "CA"
		Age:   200,       // Detect bad data
		Zip:   " 12345",  // Trim whitespaces
	}

	err := dv8.Validate(p)
	if err != nil {
		return err // Age: must be less than or equal to 120
	}
*/
func Validate(data ...any) error {
	for i := range data {
		err := internal.Validate(data[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// Validator implements a single method that returns an error if a struct is invalid.
// DV8 calls this function during validation on types that implements it.
type Validator interface {
	Validate() error
}
