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

package dv8

import (
	"context"
	"reflect"

	"github.com/microbus-io/dv8/internal"
)

// ErrDirective indicates a malformed or misapplied dv8 directive: a bug in the tag, not in the data.
// Match it with errors.Is to distinguish programming errors from invalid data.
var ErrDirective = internal.ErrDirective

/*
Validate takes in a reference to one or more data struct (pointer, map of, slice of)
and validates each of its fields against their dv8 field tags.
It recurses into nested structs.
The context is passed to any type in the object graph that implements the Validator interface.
A nil ctx is tolerated and replaced with context.Background.

Example:

	type Person struct {
		First   string `dv8:"notzero,len<=32"`
		Last    string `dv8:"notzero,len<=32"`
		Age     int    `dv8:"val>=18,val<=120"`
		State   string `dv8:"len==2,default=CA"`
		Zip     string `dv8:"trim,notzero,regexp ^[0-9]{5}$"`
	}

	p := &Person{
		First: "Jane",
		Last:  "Simmons",
		State: "",        // Set default to "CA"
		Age:   200,       // Detect bad data
		Zip:   " 12345",  // Trim whitespaces
	}

	err := dv8.Validate(ctx, p)
	if err != nil {
		return err // Age: must be less than or equal to 120
	}
*/
func Validate(ctx context.Context, data ...any) error {
	for i := range data {
		err := internal.Validate(ctx, data[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// Compile validates the dv8 directives declared on one or more types, recursing into nested types,
// without validating any data. Each argument may be a reflect.Type or a specimen value of the type.
// A malformed or misapplied directive is reported as an error wrapping ErrDirective.
// Validate compiles on first use and caches per type; calling Compile eagerly, e.g. at startup,
// surfaces broken directives before any data arrives.
func Compile(types ...any) error {
	for _, t := range types {
		refType, ok := t.(reflect.Type)
		if !ok {
			refType = reflect.TypeOf(t)
		}
		err := internal.Compile(refType)
		if err != nil {
			return err
		}
	}
	return nil
}

// Directive is one parsed directive of a dv8 struct tag.
type Directive = internal.Directive

// ParseTag parses the value of a dv8 struct tag into its directives, without checking their validity.
// It enables tooling, such as an OpenAPI generator, to project the directives onto other formats.
func ParseTag(tag string) []Directive {
	return internal.ParseTag(tag)
}

// Validator implements a single method that returns an error if a struct is invalid.
// DV8 calls this method during validation on any type in the object graph that implements it,
// passing the context given to dv8.Validate.
// A parameterless Validate() error method is honored as a fallback for interop with types not written for DV8.
// Because both variants are named Validate, a type can implement at most one of them.
type Validator = internal.Validator
