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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt_Required(t *testing.T) {
	x := struct {
		I int `dv8:"notzero"`
	}{
		I: 1,
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)

	x.I = 0
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "required")
}

func TestInt_Pointer(t *testing.T) {
	x := struct {
		I *int `dv8:"notzero"`
	}{}
	i := 1
	x.I = &i
	err := Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, 1, *x.I)

	x.I = nil
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "required")
}

func TestInt_Default(t *testing.T) {
	x := struct {
		I int `dv8:"notzero,default=2"`
	}{
		I: 0,
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, 2, x.I)

	x.I = 8
	err = Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, 8, x.I)
}

func TestInt_Val(t *testing.T) {
	gte := struct {
		I int `dv8:"val>=2"`
	}{
		I: 0,
	}
	err := Validate(nil, &gte)
	assert.ErrorContains(t, err, "greater")
	gte.I = 2
	err = Validate(nil, &gte)
	assert.NoError(t, err)

	gt := struct {
		I int `dv8:"val>2"`
	}{
		I: 2,
	}
	err = Validate(nil, &gt)
	assert.ErrorContains(t, err, "greater")
	gt.I = 3
	err = Validate(nil, &gt)
	assert.NoError(t, err)

	lte := struct {
		I int `dv8:"val<=2"`
	}{
		I: 3,
	}
	err = Validate(nil, &lte)
	assert.ErrorContains(t, err, "less")
	lte.I = 2
	err = Validate(nil, &lte)
	assert.NoError(t, err)

	lt := struct {
		I int `dv8:"val<2"`
	}{
		I: 2,
	}
	err = Validate(nil, &lt)
	assert.ErrorContains(t, err, "less")
	lt.I = 1
	err = Validate(nil, &lt)
	assert.NoError(t, err)

	eq := struct {
		I int `dv8:"val==2"`
	}{
		I: 1,
	}
	err = Validate(nil, &eq)
	assert.ErrorContains(t, err, "equal")
	eq.I = 2
	err = Validate(nil, &eq)
	assert.NoError(t, err)

	ne := struct {
		I int `dv8:"val!=2"`
	}{
		I: 2,
	}
	err = Validate(nil, &ne)
	assert.ErrorContains(t, err, "equal")
	ne.I = 1
	err = Validate(nil, &ne)
	assert.NoError(t, err)

	bad := struct {
		I int `dv8:"val*=2"`
	}{
		I: 0,
	}
	err = Validate(nil, &bad)
	assert.ErrorContains(t, err, "operator")
}

func TestInt_Types(t *testing.T) {
	x := struct {
		I   int   `dv8:"val>2"`
		I8  int8  `dv8:"val>2"`
		I16 int16 `dv8:"val>2"`
		I32 int32 `dv8:"val>2"`
		I64 int64 `dv8:"val>2"`
	}{
		I:   0,
		I8:  0,
		I16: 0,
		I32: 0,
		I64: 0,
	}
	err := Validate(nil, &x)
	assert.Error(t, err, "greater")

	x.I = 3
	x.I8 = 3
	x.I16 = 3
	x.I32 = 3
	x.I64 = 3
	err = Validate(nil, &x)
	assert.NoError(t, err)
}

func TestInt_PrimitiveType(t *testing.T) {
	type Primitive int
	x := struct {
		I Primitive `dv8:"val>2"`
	}{
		I: 1,
	}
	err := Validate(nil, &x)
	assert.Error(t, err, "greater")

	x.I = 3
	err = Validate(nil, &x)
	assert.NoError(t, err)
}

type Month int

func (m Month) Validate() error {
	if int(m) >= 1 && int(m) <= 12 {
		return nil
	}
	return errors.New("invalid value")
}

func TestInt_Validator(t *testing.T) {
	jan := Month(1)
	err := Validate(nil, &jan)
	assert.NoError(t, err)

	bad := Month(13)
	err = Validate(nil, &bad)
	assert.ErrorContains(t, err, "invalid value")
}
