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

func TestUint_Required(t *testing.T) {
	x := struct {
		I uint `dv8:"required"`
	}{
		I: 1,
	}
	err := Validate(&x)
	assert.NoError(t, err)

	x.I = 0
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestUint_Pointer(t *testing.T) {
	x := struct {
		I *uint `dv8:"required"`
	}{}
	i := uint(1)
	x.I = &i
	err := Validate(&x)
	assert.NoError(t, err)

	x.I = nil
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestUint_Default(t *testing.T) {
	x := struct {
		I uint `dv8:"required,default=2"`
	}{
		I: 0,
	}
	err := Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, uint(2), x.I)

	x.I = 8
	err = Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, uint(8), x.I)
}

func TestUint_Val(t *testing.T) {
	gte := struct {
		I uint `dv8:"val>=2"`
	}{
		I: 0,
	}
	err := Validate(&gte)
	assert.ErrorContains(t, err, "greater")
	gte.I = 2
	err = Validate(&gte)
	assert.NoError(t, err)

	gt := struct {
		I uint `dv8:"val>2"`
	}{
		I: 2,
	}
	err = Validate(&gt)
	assert.ErrorContains(t, err, "greater")
	gt.I = 3
	err = Validate(&gt)
	assert.NoError(t, err)

	lte := struct {
		I uint `dv8:"val<=2"`
	}{
		I: 3,
	}
	err = Validate(&lte)
	assert.ErrorContains(t, err, "less")
	lte.I = 2
	err = Validate(&lte)
	assert.NoError(t, err)

	lt := struct {
		I uint `dv8:"val<2"`
	}{
		I: 2,
	}
	err = Validate(&lt)
	assert.ErrorContains(t, err, "less")
	lt.I = 1
	err = Validate(&lt)
	assert.NoError(t, err)

	eq := struct {
		I uint `dv8:"val==2"`
	}{
		I: 1,
	}
	err = Validate(&eq)
	assert.ErrorContains(t, err, "equal")
	eq.I = 2
	err = Validate(&eq)
	assert.NoError(t, err)

	ne := struct {
		I uint `dv8:"val!=2"`
	}{
		I: 2,
	}
	err = Validate(&ne)
	assert.ErrorContains(t, err, "equal")
	ne.I = 1
	err = Validate(&ne)
	assert.NoError(t, err)

	bad := struct {
		I uint `dv8:"val*=2"`
	}{
		I: 0,
	}
	err = Validate(&bad)
	assert.ErrorContains(t, err, "operator")
}

func TestUint_Types(t *testing.T) {
	x := struct {
		I   uint   `dv8:"val>2"`
		I8  uint8  `dv8:"val>2"`
		I16 uint16 `dv8:"val>2"`
		I32 uint32 `dv8:"val>2"`
		I64 uint64 `dv8:"val>2"`
	}{
		I:   0,
		I8:  0,
		I16: 0,
		I32: 0,
		I64: 0,
	}
	err := Validate(&x)
	assert.Error(t, err, "greater")

	x.I = 3
	x.I8 = 3
	x.I16 = 3
	x.I32 = 3
	x.I64 = 3
	err = Validate(&x)
	assert.NoError(t, err)
}
