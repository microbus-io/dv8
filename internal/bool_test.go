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

func TestBool_Required(t *testing.T) {
	x := struct {
		B bool `dv8:"required"`
	}{
		B: true,
	}
	err := Validate(&x)
	assert.NoError(t, err)

	x.B = false
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestBool_Pointer(t *testing.T) {
	x := struct {
		B *bool `dv8:"required"`
	}{}
	flag := true
	x.B = &flag
	err := Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, true, *x.B)

	x.B = nil
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestBool_Default(t *testing.T) {
	x := struct {
		B bool `dv8:"required,default=true"`
	}{
		B: false,
	}
	err := Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, true, x.B)

	x.B = true
	err = Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, true, x.B)
}

func TestBool_Val(t *testing.T) {
	eq := struct {
		B bool `dv8:"val==true"`
	}{
		B: false,
	}
	err := Validate(&eq)
	assert.ErrorContains(t, err, "equal")
	eq.B = true
	err = Validate(&eq)
	assert.NoError(t, err)

	ne := struct {
		B bool `dv8:"val!=true"`
	}{
		B: true,
	}
	err = Validate(&ne)
	assert.ErrorContains(t, err, "equal")
	ne.B = false
	err = Validate(&ne)
	assert.NoError(t, err)

	bad := struct {
		B bool `dv8:"val*=true"`
	}{
		B: true,
	}
	err = Validate(&bad)
	assert.ErrorContains(t, err, "operator")
}
