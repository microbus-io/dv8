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
	"reflect"
)

// validateAny validates the value of any type against the tags.
func validateAny(ctx context.Context, refType reflect.Type, refVal reflect.Value, tags []string) (err error) {
	switch refType.String() {
	case "time.Duration":
		err = validateDuration(refVal, tags)
	case "time.Time":
		err = validateTime(refVal, tags)
	default:
		switch refType.Kind() {
		case reflect.String:
			err = validateString(refVal, tags)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			err = validateInt(refVal, tags)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			err = validateUint(refVal, tags)
		case reflect.Float32, reflect.Float64:
			err = validateFloat(refVal, tags)
		case reflect.Bool:
			err = validateBool(refVal, tags)
		case reflect.Pointer:
			err = validatePointer(ctx, refType, refVal, tags)
		case reflect.Struct:
			err = validateStruct(ctx, refType, refVal, tags)
		case reflect.Map:
			err = validateMap(ctx, refType, refVal, tags)
		case reflect.Array, reflect.Slice:
			err = validateArray(ctx, refType, refVal, tags)
		}
	}
	if err != nil {
		return err
	}

	// Call the type's Validate method, if implemented
	var ok bool
	var okCtx bool
	var validator Validator
	var validatorCtx ValidatorContext
	if refVal.CanAddr() {
		underlyingPtr := refVal.Addr().Interface()
		validator, ok = underlyingPtr.(Validator)
		validatorCtx, okCtx = underlyingPtr.(ValidatorContext)
	}
	if !ok {
		underlying := refVal.Interface()
		validator, ok = underlying.(Validator)
		validatorCtx, okCtx = underlying.(ValidatorContext)
	}
	if ok {
		err = validator.Validate()
		if err != nil {
			return err
		}
	}
	if okCtx {
		err = validatorCtx.ValidateContext(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
