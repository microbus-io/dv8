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

// buildSteps compiles the validation steps of any type against the incoming directives.
func buildSteps(refType reflect.Type, tags []string, memo map[planKey]*plan) []step {
	var steps []step
	switch refType.String() {
	case "time.Duration":
		steps = compileDuration(tags)
	case "time.Time":
		steps = compileTime(tags)
	default:
		switch refType.Kind() {
		case reflect.String:
			steps = compileString(tags)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			steps = compileInt(tags)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			steps = compileUint(tags)
		case reflect.Float32, reflect.Float64:
			steps = compileFloat(tags)
		case reflect.Bool:
			steps = compileBool(tags)
		case reflect.Pointer:
			steps = compilePointer(refType, tags, memo)
		case reflect.Struct:
			steps = compileStruct(refType, tags, memo)
		case reflect.Map:
			steps = compileMap(refType, tags, memo)
		case reflect.Array, reflect.Slice:
			steps = compileArray(refType, tags, memo)
		}
	}
	steps = append(steps, compileValidatorCall(refType)...)
	return steps
}

var (
	validatorType          = reflect.TypeOf((*Validator)(nil)).Elem()
	validatorNoContextType = reflect.TypeOf((*validatorNoContext)(nil)).Elem()
)

// compileValidatorCall compiles the call to the type's Validate method, if implemented.
// Both variants share the method name Validate, so a type implements at most one of them.
func compileValidatorCall(refType reflect.Type) []step {
	if refType.Kind() == reflect.Interface {
		// The dynamic type is unknown at compile time
		return []step{func(ctx context.Context, refVal reflect.Value) error {
			var underlying any
			if refVal.CanAddr() {
				underlying = refVal.Addr().Interface()
			} else {
				underlying = refVal.Interface()
			}
			var err error
			switch v := underlying.(type) {
			case Validator:
				err = v.Validate(ctx)
			case validatorNoContext:
				err = v.Validate()
			}
			return attributeValidatorError(err)
		}}
	}
	ptrType := reflect.PointerTo(refType)
	ptrCtx := ptrType.Implements(validatorType)
	ptrNoCtx := !ptrCtx && ptrType.Implements(validatorNoContextType)
	valCtx := refType.Implements(validatorType)
	valNoCtx := !valCtx && refType.Implements(validatorNoContextType)
	if !ptrCtx && !ptrNoCtx && !valCtx && !valNoCtx {
		return nil
	}
	return []step{func(ctx context.Context, refVal reflect.Value) error {
		var err error
		if refVal.CanAddr() {
			if ptrCtx {
				err = refVal.Addr().Interface().(Validator).Validate(ctx)
			} else if ptrNoCtx {
				err = refVal.Addr().Interface().(validatorNoContext).Validate()
			}
		} else {
			if valCtx {
				err = refVal.Interface().(Validator).Validate(ctx)
			} else if valNoCtx {
				err = refVal.Interface().(validatorNoContext).Validate()
			}
		}
		return attributeValidatorError(err)
	}}
}

// attributeValidatorError attributes a custom validator's failure to the input data,
// unless the validator chose a status code of its own.
func attributeValidatorError(err error) error {
	if err != nil && errors.StatusCode(err) == http.StatusInternalServerError {
		err = errors.Trace(err, http.StatusBadRequest)
	}
	return err
}
