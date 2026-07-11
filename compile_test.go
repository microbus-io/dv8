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

package dv8_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/microbus-io/dv8"
	"github.com/stretchr/testify/assert"
)

type goodIn struct {
	Name string `dv8:"notzero,len<=32"`
	Age  int    `dv8:"val>=0,val<=120"`
}

type badIn struct {
	Name string `dv8:"requird"`
}

func TestCompile(t *testing.T) {
	// Specimen values and reflect.Type arguments are both accepted
	err := dv8.Compile(goodIn{})
	assert.NoError(t, err)
	err = dv8.Compile(reflect.TypeOf(goodIn{}), &goodIn{})
	assert.NoError(t, err)

	err = dv8.Compile(badIn{})
	assert.True(t, errors.Is(err, dv8.ErrDirective))
	assert.ErrorContains(t, err, "requird")

	// Validate reports the same directive error on first use
	err = dv8.Validate(nil, &badIn{Name: "x"})
	assert.True(t, errors.Is(err, dv8.ErrDirective))

	// A data error is not a directive error
	err = dv8.Validate(nil, &goodIn{Age: 200})
	assert.Error(t, err)
	assert.False(t, errors.Is(err, dv8.ErrDirective))
}
