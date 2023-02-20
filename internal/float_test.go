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

func TestFloat_Required(t *testing.T) {
	x := struct {
		F float64 `dv8:"required"`
	}{
		F: 1.5,
	}
	err := Validate(&x)
	assert.NoError(t, err)

	x.F = 0
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestFloat_Pointer(t *testing.T) {
	x := struct {
		F *float64 `dv8:"required"`
	}{}
	f := float64(1)
	x.F = &f
	err := Validate(&x)
	assert.NoError(t, err)

	x.F = nil
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestFloat_Default(t *testing.T) {
	x := struct {
		F float64 `dv8:"required,default=2.5"`
	}{
		F: 0,
	}
	err := Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, 2.5, x.F)

	x.F = 1.5
	err = Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, 1.5, x.F)
}

func TestFloat_Val(t *testing.T) {
	gte := struct {
		F float64 `dv8:"val>=2.5"`
	}{
		F: 0,
	}
	err := Validate(&gte)
	assert.ErrorContains(t, err, "greater")
	gte.F = 2.5
	err = Validate(&gte)
	assert.NoError(t, err)

	gt := struct {
		F float64 `dv8:"val>2.5"`
	}{
		F: 2.5,
	}
	err = Validate(&gt)
	assert.ErrorContains(t, err, "greater")
	gt.F = 3.5
	err = Validate(&gt)
	assert.NoError(t, err)

	lte := struct {
		F float64 `dv8:"val<=2.5"`
	}{
		F: 3.5,
	}
	err = Validate(&lte)
	assert.ErrorContains(t, err, "less")
	lte.F = 2.5
	err = Validate(&lte)
	assert.NoError(t, err)

	lt := struct {
		F float64 `dv8:"val<2.5"`
	}{
		F: 2.5,
	}
	err = Validate(&lt)
	assert.ErrorContains(t, err, "less")
	lt.F = 1.5
	err = Validate(&lt)
	assert.NoError(t, err)

	eq := struct {
		F float64 `dv8:"val==2.5"`
	}{
		F: 1.5,
	}
	err = Validate(&eq)
	assert.ErrorContains(t, err, "equal")
	eq.F = 2.5
	err = Validate(&eq)
	assert.NoError(t, err)

	ne := struct {
		F float64 `dv8:"val!=2.5"`
	}{
		F: 2.5,
	}
	err = Validate(&ne)
	assert.ErrorContains(t, err, "equal")
	ne.F = 1.5
	err = Validate(&ne)
	assert.NoError(t, err)

	bad := struct {
		F float64 `dv8:"val*=2.5"`
	}{
		F: 0,
	}
	err = Validate(&bad)
	assert.ErrorContains(t, err, "operator")
}

func TestFloat_Types(t *testing.T) {
	x := struct {
		F32 float32 `dv8:"val>2.5"`
		F64 float64 `dv8:"val>2.5"`
	}{
		F32: 1.5,
		F64: 1.5,
	}
	err := Validate(&x)
	assert.Error(t, err, "greater")

	x.F32 = 3.5
	x.F64 = 3.5
	err = Validate(&x)
	assert.NoError(t, err)
}
