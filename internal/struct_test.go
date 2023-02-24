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
	"errors"
	"testing"
	"time"

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
		S *struct{ I int } `dv8:"required"`
	}{}
	s := struct{ I int }{I: 1}
	x.S = &s
	err := Validate(&x)
	assert.NoError(t, err)

	x.S.I = 0
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")

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

func TestStruct_On1(t *testing.T) {
	type child struct {
		I int
	}
	type parent struct {
		C *child `dv8:"required,val>2,on I"`
	}
	x := parent{
		C: &child{
			I: 1,
		},
	}
	err := Validate(&x)
	assert.Error(t, err, "greater")

	x.C.I = 3
	err = Validate(&x)
	assert.NoError(t, err)

	x.C.I = 0
	err = Validate(&x)
	assert.Error(t, err, "required")

	x.C = nil
	err = Validate(&x)
	assert.Error(t, err, "required")
}

func TestStruct_On2(t *testing.T) {
	type Timestamp struct {
		time.Time
	}
	type Key struct {
		ID int
	}
	type Person struct {
		Name string
	}
	type MyData struct {
		K       Key       `dv8:"required,on ID"`
		Expires Timestamp `dv8:"required,on Time"`
		Owner   Person    `dv8:"default=Unknown,on Name"`
	}

	x := MyData{}
	err := Validate(&x)
	assert.Error(t, err, "required")
	assert.Error(t, err, "K:")

	x.K.ID = 1
	err = Validate(&x)
	assert.Error(t, err, "required")
	assert.Error(t, err, "Expires:")

	x.Expires.Time = time.Now()
	err = Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, x.Owner.Name, "Unknown")
}

func TestStruct_Main(t *testing.T) {
	type Timestamp struct {
		time.Time `dv8:"main"`
	}
	type Key struct {
		ID int `dv8:"main"`
	}
	type Person struct {
		Name string `dv8:"main"`
	}
	type MyData struct {
		K       Key       `dv8:"required"`
		Expires Timestamp `dv8:"required"`
		Owner   Person    `dv8:"default=Unknown"`
	}

	x := MyData{}
	err := Validate(&x)
	assert.Error(t, err, "required")
	assert.Error(t, err, "K:")

	x.K.ID = 1
	err = Validate(&x)
	assert.Error(t, err, "required")
	assert.Error(t, err, "Expires:")

	x.Expires.Time = time.Now()
	err = Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, x.Owner.Name, "Unknown")
}

type SmallRect struct {
	W int `dv8:"val>=0"`
	H int `dv8:"val>=0"`
}

func (r SmallRect) Validate() error {
	if r.H*r.W > 100 {
		return errors.New("too big")
	}
	return nil
}

func TestStruct_Validator(t *testing.T) {
	small := SmallRect{W: 5, H: 5}
	err := Validate(small)
	assert.NoError(t, err)

	big := SmallRect{W: 50, H: 50}
	err = Validate(big)
	assert.ErrorContains(t, err, "too big")
}

func TestStruct_ValidatorOfAnonymous(t *testing.T) {
	x := struct {
		SmallRect
	}{
		SmallRect{W: 50, H: 50},
	}
	err := Validate(x)
	assert.ErrorContains(t, err, "too big")
}
