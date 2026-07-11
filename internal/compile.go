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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ErrDirective indicates a malformed or misapplied dv8 directive: a bug in the tag, not in the data.
var ErrDirective = errors.New("malformed directive")

// errNotApplicable is an internal marker that a directive is recognized but not valid for a given type.
var errNotApplicable = errors.New("not applicable")

var compiledTypes sync.Map // reflect.Type -> error

// Compile validates the dv8 directives declared on a type, recursing into nested types.
// Errors wrap ErrDirective. The result is cached per type.
func Compile(refType reflect.Type) error {
	if refType == nil {
		return nil
	}
	if cached, ok := compiledTypes.Load(refType); ok {
		if cached == nil {
			return nil
		}
		return cached.(error)
	}
	err := compileType(refType, map[reflect.Type]bool{})
	if err == nil {
		compiledTypes.Store(refType, nil)
	} else {
		compiledTypes.Store(refType, err)
	}
	return err
}

// compileType walks a type looking for struct fields with dv8 tags and checks their directives.
func compileType(refType reflect.Type, visited map[reflect.Type]bool) error {
	switch refType.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		return compileType(refType.Elem(), visited)
	case reflect.Map:
		err := compileType(refType.Key(), visited)
		if err != nil {
			return err
		}
		return compileType(refType.Elem(), visited)
	case reflect.Struct:
		if visited[refType] {
			return nil
		}
		visited[refType] = true
		if refType.String() == "time.Time" {
			return nil
		}
		for i := 0; i < refType.NumField(); i++ {
			fld := refType.Field(i)
			tagVal := fld.Tag.Get("dv8")
			if tagVal != "" && tagVal != "-" {
				err := checkDirectives(fld.Type, splitDirectives(tagVal))
				if err != nil {
					return fmt.Errorf("%s: %w", fld.Name, err)
				}
			}
			err := compileType(fld.Type, visited)
			if err != nil {
				return fmt.Errorf("%s: %w", fld.Name, err)
			}
		}
	}
	return nil
}

// checkDirectives checks a field's directives against the field's type.
// A value directive must be valid for the field itself or for one of its "on" or "delegate" targets.
func checkDirectives(refType reflect.Type, dirs []string) error {
	base := refType
	for base.Kind() == reflect.Pointer {
		base = base.Elem()
	}
	// Value directives may target the field itself, its "on" fields, or its "delegate" field
	targets := []reflect.Type{base}
	for _, d := range dirs {
		if strings.HasPrefix(d, "on ") {
			if base.Kind() != reflect.Struct {
				return fmt.Errorf("%w: 'on' applies only to structs", ErrDirective)
			}
			fld, ok := base.FieldByName(d[len("on "):])
			if !ok {
				return fmt.Errorf("%w: no field '%s' for 'on'", ErrDirective, d[len("on "):])
			}
			targets = append(targets, fld.Type)
		}
	}
	if base.Kind() == reflect.Struct && base.String() != "time.Time" {
		for i := 0; i < base.NumField(); i++ {
			fld := base.Field(i)
			if tagsContain(splitDirectives(fld.Tag.Get("dv8")), "delegate") {
				targets = append(targets, fld.Type)
			}
		}
	}
	for _, d := range dirs {
		if d == "" || d == "-" || d == "delegate" || strings.HasPrefix(d, "on ") {
			continue
		}
		if strings.HasPrefix(d, "each ") {
			switch base.Kind() {
			case reflect.Slice, reflect.Array:
				err := checkDirectives(base.Elem(), []string{d[len("each "):]})
				if err != nil {
					return err
				}
			case reflect.Map:
				err := checkDirectives(base.Elem(), []string{d[len("each "):]})
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("%w: 'each' applies only to arrays and maps", ErrDirective)
			}
			continue
		}
		if strings.HasPrefix(d, "key ") {
			if base.Kind() != reflect.Map {
				return fmt.Errorf("%w: 'key' applies only to maps", ErrDirective)
			}
			err := checkDirectives(base.Key(), []string{d[len("key "):]})
			if err != nil {
				return err
			}
			continue
		}
		// A value directive passes if it is valid for at least one target
		var parseErr error
		applies := false
		for _, target := range targets {
			err := checkValueDirective(target, d)
			if err == nil {
				applies = true
				break
			}
			if !errors.Is(err, errNotApplicable) && parseErr == nil {
				parseErr = err
			}
		}
		if !applies {
			if parseErr != nil {
				return parseErr
			}
			return fmt.Errorf("%w: '%s' is not applicable to %s", ErrDirective, d, refType)
		}
	}
	return nil
}

// checkValueDirective checks a single value directive against a single target type.
// It returns nil if valid, errNotApplicable if the directive doesn't apply to the type,
// or an ErrDirective-wrapped error if the directive is malformed.
func checkValueDirective(refType reflect.Type, d string) error {
	for refType.Kind() == reflect.Pointer {
		refType = refType.Elem()
	}
	if refType.Kind() == reflect.Interface {
		// The dynamic type is unknown at compile time
		return nil
	}
	cat := typeCategory(refType)
	if d == "required" {
		return fmt.Errorf("%w: 'required' was replaced by 'notzero'", ErrDirective)
	}
	if d == "notzero" {
		return nil
	}
	if d == "trim" || d == "tolower" || d == "toupper" {
		if cat == "string" {
			return nil
		}
		return errNotApplicable
	}
	if strings.HasPrefix(d, "default=") {
		return checkParsableValue(cat, d[len("default="):])
	}
	if strings.HasPrefix(d, "oneof ") {
		if cat == "string" {
			return nil
		}
		return errNotApplicable
	}
	if strings.HasPrefix(d, "regexp ") {
		if cat != "string" {
			return errNotApplicable
		}
		_, err := compileRegexp(d[len("regexp "):])
		if err != nil {
			return fmt.Errorf("%w: %s", ErrDirective, err)
		}
		return nil
	}
	if strings.HasPrefix(d, "val") {
		switch cat {
		case "string", "int", "uint", "float", "bool", "duration", "time":
			operator, value, err := splitOpValue(d, len("val"))
			if err != nil {
				return err
			}
			if cat == "bool" && operator != "==" && operator != "!=" {
				return fmt.Errorf("%w: unsupported operator '%s'", ErrDirective, operator)
			}
			return checkParsableValue(cat, value)
		}
		return errNotApplicable
	}
	if strings.HasPrefix(d, "len") {
		switch cat {
		case "string", "array", "map":
			_, value, err := splitOpValue(d, len("len"))
			if err != nil {
				return err
			}
			_, err = strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("%w: %s", ErrDirective, err)
			}
			return nil
		}
		return errNotApplicable
	}
	return fmt.Errorf("%w: unknown directive '%s'", ErrDirective, d)
}

// splitOpValue splits a comparison directive such as "val>=2" into its operator and value.
func splitOpValue(d string, prefixLen int) (operator string, value string, err error) {
	if len(d) <= prefixLen+1 {
		return "", "", fmt.Errorf("%w: incomplete directive '%s'", ErrDirective, d)
	}
	operator = d[prefixLen : prefixLen+1]
	value = d[prefixLen+1:]
	if d[prefixLen+1] == '=' {
		operator += "="
		value = d[prefixLen+2:]
	}
	switch operator {
	case "<", "<=", ">", ">=", "==", "!=":
	default:
		return "", "", fmt.Errorf("%w: unsupported operator '%s'", ErrDirective, operator)
	}
	if value == "" {
		return "", "", fmt.Errorf("%w: incomplete directive '%s'", ErrDirective, d)
	}
	return operator, value, nil
}

// checkParsableValue checks that a directive's value parses as the target type.
func checkParsableValue(cat string, value string) error {
	var err error
	switch cat {
	case "string":
	case "int":
		_, err = strconv.ParseInt(value, 10, 64)
	case "uint":
		_, err = strconv.ParseUint(value, 10, 64)
	case "float":
		_, err = strconv.ParseFloat(value, 64)
	case "bool":
		_, err = strconv.ParseBool(value)
	case "duration":
		_, err = time.ParseDuration(value)
	case "time":
		_, err = parseTime(value)
	default:
		return errNotApplicable
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrDirective, err)
	}
	return nil
}

// typeCategory maps a type to the directive vocabulary it supports.
func typeCategory(refType reflect.Type) string {
	switch refType.String() {
	case "time.Duration":
		return "duration"
	case "time.Time":
		return "time"
	}
	switch refType.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "uint"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Bool:
		return "bool"
	case reflect.Struct:
		return "struct"
	case reflect.Map:
		return "map"
	case reflect.Array, reflect.Slice:
		return "array"
	default:
		return "other"
	}
}
