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
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/microbus-io/errors"
)

var (
	regexpCache   = map[string]*regexp.Regexp{}
	regexpCacheMu sync.RWMutex
)

// compileRegexp compiles a regular expression, caching it for reuse across validations.
func compileRegexp(pattern string) (*regexp.Regexp, error) {
	regexpCacheMu.RLock()
	re, ok := regexpCache[pattern]
	regexpCacheMu.RUnlock()
	if ok {
		return re, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	regexpCacheMu.Lock()
	regexpCache[pattern] = re
	regexpCacheMu.Unlock()
	return re, nil
}

// compileString compiles the validation of a string against the tags.
func compileString(tags []string) []step {
	// Mutations run first: trim, then case folding and defaults in tag order
	var mutations []func(s string) string
	if tagsContain(tags, "trim") {
		mutations = append(mutations, strings.TrimSpace)
	}
	required := false
	for _, t := range tags {
		if t == "notzero" {
			required = true
		} else if t == "toupper" {
			mutations = append(mutations, strings.ToUpper)
		} else if t == "tolower" {
			mutations = append(mutations, strings.ToLower)
		} else if strings.HasPrefix(t, "default=") {
			def := t[len("default="):]
			mutations = append(mutations, func(s string) string {
				if s == "" {
					return def
				}
				return s
			})
		}
	}
	// Constraint checks run after mutations, in tag order
	var checks []func(s string) error
	for _, t := range tags {
		if strings.HasPrefix(t, "len") && len(t) > 4 {
			operator, value, err := splitOpValue(t, 3)
			if err != nil {
				continue
			}
			l, err := strconv.Atoi(value)
			if err != nil {
				continue
			}
			checks = append(checks, func(s string) error {
				strLen := len([]rune(s))
				switch {
				case operator == "<=" && strLen > l:
					return errInvalid("length must be less than or equal to %d", l)
				case operator == "<" && strLen >= l:
					return errInvalid("length must be less than %d", l)
				case operator == ">=" && strLen < l:
					return errInvalid("length must be greater than or equal to %d", l)
				case operator == ">" && strLen <= l:
					return errInvalid("length must be greater than %d", l)
				case operator == "!=" && strLen == l:
					return errInvalid("length must not equal %d", l)
				case operator == "==" && strLen != l:
					return errInvalid("length must equal %d", l)
				}
				return nil
			})
		} else if strings.HasPrefix(t, "val") && len(t) > 4 {
			operator, v, err := splitOpValue(t, 3)
			if err != nil {
				continue
			}
			checks = append(checks, func(s string) error {
				switch {
				case operator == "<=" && s > v:
					return errInvalid("must be less than or equal to '%s'", v)
				case operator == "<" && s >= v:
					return errInvalid("must be less than '%s'", v)
				case operator == ">=" && s < v:
					return errInvalid("must be greater than or equal to '%s'", v)
				case operator == ">" && s <= v:
					return errInvalid("must be greater than '%s'", v)
				case operator == "!=" && s == v:
					return errInvalid("must not equal '%s'", v)
				case operator == "==" && s != v:
					return errInvalid("must equal '%s'", v)
				}
				return nil
			})
		} else if strings.HasPrefix(t, "regexp ") && len(t) > 7 {
			re, err := compileRegexp(t[7:])
			if err != nil {
				continue
			}
			checks = append(checks, func(s string) error {
				if !re.MatchString(s) {
					return errInvalid("value doesn't match required pattern")
				}
				return nil
			})
		} else if strings.HasPrefix(t, "oneof ") && len(t) > 6 {
			set := t[6:]
			validVals := strings.Split(set, "|")
			checks = append(checks, func(s string) error {
				for _, v := range validVals {
					if s == v {
						return nil
					}
				}
				return errInvalid("value must be one of %s", set)
			})
		}
	}
	if len(mutations) == 0 && !required && len(checks) == 0 {
		return nil
	}
	return []step{func(_ context.Context, refVal reflect.Value) error {
		s := refVal.String()
		changed := false
		for _, m := range mutations {
			mutated := m(s)
			if mutated != s {
				s = mutated
				changed = true
			}
		}
		if changed {
			if !refVal.CanSet() {
				return errors.New("data must be passed by reference")
			}
			refVal.SetString(s)
		}
		if s == "" && required {
			return errInvalid("value is required")
		}
		for _, check := range checks {
			err := check(s)
			if err != nil {
				return err
			}
		}
		return nil
	}}
}
