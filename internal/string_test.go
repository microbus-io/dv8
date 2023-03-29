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

func TestString_Required(t *testing.T) {
	x := struct {
		S string `dv8:"required"`
	}{
		S: "Foo",
	}
	err := Validate(&x)
	assert.NoError(t, err)

	x.S = ""
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestString_Pointer(t *testing.T) {
	x := struct {
		S *string `dv8:"required"`
	}{}
	s := "foo"
	x.S = &s
	err := Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, "foo", *x.S)

	x.S = nil
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestString_Default(t *testing.T) {
	x := struct {
		S string `dv8:"required,default=Foo"`
	}{
		S: "",
	}
	err := Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, "Foo", x.S)

	x.S = "Foo"
	err = Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, "Foo", x.S)
}

func TestString_LenMulti(t *testing.T) {
	x := struct {
		S string `dv8:"len>2,len<=8"`
	}{
		S: "",
	}
	err := Validate(&x)
	assert.ErrorContains(t, err, "greater")

	x.S = "12"
	err = Validate(&x)
	assert.ErrorContains(t, err, "greater")

	x.S = "123"
	err = Validate(&x)
	assert.NoError(t, err)

	x.S = "12345678"
	err = Validate(&x)
	assert.NoError(t, err)

	x.S = "123456789"
	err = Validate(&x)
	assert.ErrorContains(t, err, "less")
}

func TestString_Len(t *testing.T) {
	gte := struct {
		S string `dv8:"len>=2"`
	}{
		S: "",
	}
	err := Validate(&gte)
	assert.ErrorContains(t, err, "greater")
	gte.S = "12"
	err = Validate(&gte)
	assert.NoError(t, err)

	gt := struct {
		S string `dv8:"len>2"`
	}{
		S: "12",
	}
	err = Validate(&gt)
	assert.ErrorContains(t, err, "greater")
	gt.S = "123"
	err = Validate(&gt)
	assert.NoError(t, err)

	lte := struct {
		S string `dv8:"len<=2"`
	}{
		S: "123",
	}
	err = Validate(&lte)
	assert.ErrorContains(t, err, "less")
	lte.S = "12"
	err = Validate(&lte)
	assert.NoError(t, err)

	lt := struct {
		S string `dv8:"len<2"`
	}{
		S: "12",
	}
	err = Validate(&lt)
	assert.ErrorContains(t, err, "less")
	lt.S = "1"
	err = Validate(&lt)
	assert.NoError(t, err)

	eq := struct {
		S string `dv8:"len==2"`
	}{
		S: "123",
	}
	err = Validate(&eq)
	assert.ErrorContains(t, err, "equal")
	eq.S = "12"
	err = Validate(&eq)
	assert.NoError(t, err)

	ne := struct {
		S string `dv8:"len!=2"`
	}{
		S: "12",
	}
	err = Validate(&ne)
	assert.ErrorContains(t, err, "equal")
	ne.S = "1"
	err = Validate(&ne)
	assert.NoError(t, err)

	bad := struct {
		S string `dv8:"len*=2"`
	}{
		S: "",
	}
	err = Validate(&bad)
	assert.ErrorContains(t, err, "operator")
}

func TestString_Trim(t *testing.T) {
	x := struct {
		S string `dv8:"len>2"`
	}{
		S: "  ",
	}

	err := Validate(&x)
	assert.Error(t, err, "length")
	assert.Equal(t, "", x.S)

	x.S = "  Foo  "
	err = Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, "Foo", x.S)
}

func TestString_NoTrim(t *testing.T) {
	x := struct {
		S string `dv8:"len>=7,notrim"`
	}{
		S: "  Foo  ",
	}
	err := Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, "  Foo  ", x.S)
}

func TestString_Val(t *testing.T) {
	gte := struct {
		S string `dv8:"val>=2"`
	}{
		S: "",
	}
	err := Validate(&gte)
	assert.ErrorContains(t, err, "greater")
	gte.S = "2"
	err = Validate(&gte)
	assert.NoError(t, err)

	gt := struct {
		S string `dv8:"val>2"`
	}{
		S: "2",
	}
	err = Validate(&gt)
	assert.ErrorContains(t, err, "greater")
	gt.S = "21"
	err = Validate(&gt)
	assert.NoError(t, err)

	lte := struct {
		S string `dv8:"val<=2"`
	}{
		S: "21",
	}
	err = Validate(&lte)
	assert.ErrorContains(t, err, "less")
	lte.S = "2"
	err = Validate(&lte)
	assert.NoError(t, err)

	lt := struct {
		S string `dv8:"val<2"`
	}{
		S: "2",
	}
	err = Validate(&lt)
	assert.ErrorContains(t, err, "less")
	lt.S = "19"
	err = Validate(&lt)
	assert.NoError(t, err)

	eq := struct {
		S string `dv8:"val==2"`
	}{
		S: "1",
	}
	err = Validate(&eq)
	assert.ErrorContains(t, err, "equal")
	eq.S = "2"
	err = Validate(&eq)
	assert.NoError(t, err)

	ne := struct {
		S string `dv8:"val!=2"`
	}{
		S: "2",
	}
	err = Validate(&ne)
	assert.ErrorContains(t, err, "equal")
	ne.S = "1"
	err = Validate(&ne)
	assert.NoError(t, err)

	bad := struct {
		S string `dv8:"val*=2"`
	}{
		S: "",
	}
	err = Validate(&bad)
	assert.ErrorContains(t, err, "operator")
}

func TestString_Regexp(t *testing.T) {
	x := struct {
		S string `dv8:"regexp ^[A-Z]*$"`
	}{
		S: " Foo ",
	}
	err := Validate(&x)
	assert.Error(t, err, "pattern")

	x.S = " FOO "
	err = Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, x.S, "FOO")
}

func TestString_RegexpBackslash(t *testing.T) {
	x := struct {
		S string `dv8:"regexp ^\\.$"`
	}{
		S: "m",
	}
	err := Validate(&x)
	assert.Error(t, err, "pattern")

	x.S = "."
	err = Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, x.S, ".")
}

func TestString_ToLower(t *testing.T) {
	x := struct {
		S string `dv8:"tolower,default=foo"`
	}{
		S: "",
	}
	err := Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, "foo", x.S)

	x.S = "FOO"
	err = Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, "foo", x.S)
}

func TestString_ToUpper(t *testing.T) {
	x := struct {
		S string `dv8:"toupper,default=FOO"`
	}{
		S: "",
	}
	err := Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, "FOO", x.S)

	x.S = "foo"
	err = Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, "FOO", x.S)
}
