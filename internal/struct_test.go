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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStruct_Required(t *testing.T) {
	type nested struct {
		I int
	}
	x := struct {
		N *nested `dv8:"required"`
	}{
		N: &nested{
			I: 5,
		},
	}
	// err := Validate(&x)
	// assert.NoError(t, err)

	x.N = nil
	err := Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestStruct_Pointer(t *testing.T) {
	x := struct {
		S *struct{} `dv8:"required"`
	}{}
	s := struct{}{}
	x.S = &s
	err := Validate(&x)
	assert.NoError(t, err)

	x.S = nil
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestStruct_Nesting(t *testing.T) {
	type nested struct {
		I int `dv8:"required"`
	}
	x := struct {
		N *nested
	}{
		N: &nested{
			I: 5,
		},
	}
	err := Validate(&x)
	assert.NoError(t, err)

	x.N.I = 0
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")

	y := struct {
		N nested
	}{
		N: nested{
			I: 5,
		},
	}
	err = Validate(&y)
	assert.NoError(t, err)

	y.N.I = 0
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}
