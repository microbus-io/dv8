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
	"net/http"
	"reflect"

	"github.com/microbus-io/errors"
)

// errInvalid creates a validation error attributed to the input data, with a 400 status code.
func errInvalid(pattern string, args ...any) error {
	return errors.New(pattern, append(args, http.StatusBadRequest)...)
}

// Validate takes in a reference to a data struct (pointer, map of, slice of)
// and validates each of its fields against their dv8 field tags.
// It recurses into nested structs.
// A nil ctx is tolerated and replaced with context.Background.
func Validate(ctx context.Context, data any) error {
	if ctx == nil {
		ctx = context.Background()
	}
	p, err := planOf(reflect.TypeOf(data))
	if err != nil {
		return err
	}
	if p == nil {
		return nil
	}
	return p.execute(ctx, reflect.ValueOf(data))
}

// Validator implements a single method that returns an error if a struct is invalid.
// DV8 calls this method during validation on any type in the object graph that implements it.
type Validator interface {
	Validate(ctx context.Context) error
}

// validatorNoContext is the parameterless fallback, honored for interop with types not written for DV8.
// Because both variants are named Validate, a type can implement at most one of them.
type validatorNoContext interface {
	Validate() error
}
