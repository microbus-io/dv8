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

// compilePointer compiles the validation of a pointer against the tags.
// The tags pass through to the pointed-to value.
func compilePointer(refType reflect.Type, tags []string, memo map[planKey]*plan) []step {
	notzero := tagsContain(tags, "notzero")
	sub := buildPlan(refType.Elem(), tags, memo)
	if !notzero && sub.isEmpty() {
		return nil
	}
	return []step{func(ctx context.Context, refVal reflect.Value) error {
		if refVal.IsNil() {
			if notzero {
				return errInvalid("value is required")
			}
			return nil
		}
		return sub.execute(ctx, refVal.Elem())
	}}
}
