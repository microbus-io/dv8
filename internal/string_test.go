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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString_Required(t *testing.T) {
	x := struct {
		S string `dv8:"notzero"`
	}{
		S: "Foo",
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)

	x.S = ""
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "required")
}

func TestString_Pointer(t *testing.T) {
	x := struct {
		S *string `dv8:"notzero"`
	}{}
	s := "foo"
	x.S = &s
	err := Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, "foo", *x.S)

	x.S = nil
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "required")
}

func TestString_Default(t *testing.T) {
	x := struct {
		S string `dv8:"notzero,default=Foo"`
	}{
		S: "",
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, "Foo", x.S)

	x.S = "Foo"
	err = Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, "Foo", x.S)
}

func TestString_LenMulti(t *testing.T) {
	x := struct {
		S string `dv8:"len>2,len<=8"`
	}{
		S: "",
	}
	err := Validate(nil, &x)
	assert.ErrorContains(t, err, "greater")

	x.S = "12"
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "greater")

	x.S = "123"
	err = Validate(nil, &x)
	assert.NoError(t, err)

	x.S = "12345678"
	err = Validate(nil, &x)
	assert.NoError(t, err)

	x.S = "123456789"
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "less")
}

func TestString_Len(t *testing.T) {
	gte := struct {
		S string `dv8:"len>=2"`
	}{
		S: "",
	}
	err := Validate(nil, &gte)
	assert.ErrorContains(t, err, "greater")
	gte.S = "12"
	err = Validate(nil, &gte)
	assert.NoError(t, err)

	gt := struct {
		S string `dv8:"len>2"`
	}{
		S: "12",
	}
	err = Validate(nil, &gt)
	assert.ErrorContains(t, err, "greater")
	gt.S = "123"
	err = Validate(nil, &gt)
	assert.NoError(t, err)

	lte := struct {
		S string `dv8:"len<=2"`
	}{
		S: "123",
	}
	err = Validate(nil, &lte)
	assert.ErrorContains(t, err, "less")
	lte.S = "12"
	err = Validate(nil, &lte)
	assert.NoError(t, err)

	lt := struct {
		S string `dv8:"len<2"`
	}{
		S: "12",
	}
	err = Validate(nil, &lt)
	assert.ErrorContains(t, err, "less")
	lt.S = "1"
	err = Validate(nil, &lt)
	assert.NoError(t, err)

	eq := struct {
		S string `dv8:"len==2"`
	}{
		S: "123",
	}
	err = Validate(nil, &eq)
	assert.ErrorContains(t, err, "equal")
	eq.S = "12"
	err = Validate(nil, &eq)
	assert.NoError(t, err)

	ne := struct {
		S string `dv8:"len!=2"`
	}{
		S: "12",
	}
	err = Validate(nil, &ne)
	assert.ErrorContains(t, err, "equal")
	ne.S = "1"
	err = Validate(nil, &ne)
	assert.NoError(t, err)

	bad := struct {
		S string `dv8:"len*=2"`
	}{
		S: "",
	}
	err = Validate(nil, &bad)
	assert.ErrorContains(t, err, "operator")
}

func TestString_Trim(t *testing.T) {
	x := struct {
		S string `dv8:"len>2,trim"`
	}{
		S: "  ",
	}

	err := Validate(nil, &x)
	assert.Error(t, err, "length")
	assert.Equal(t, "", x.S)

	x.S = "  Foo  "
	err = Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, "Foo", x.S)
}

func TestString_NoTrim(t *testing.T) {
	// Not trimming is the default
	x := struct {
		S string `dv8:"len>=7"`
	}{
		S: "  Foo  ",
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, "  Foo  ", x.S)
}

func TestString_Val(t *testing.T) {
	gte := struct {
		S string `dv8:"val>=2"`
	}{
		S: "",
	}
	err := Validate(nil, &gte)
	assert.ErrorContains(t, err, "greater")
	gte.S = "2"
	err = Validate(nil, &gte)
	assert.NoError(t, err)

	gt := struct {
		S string `dv8:"val>2"`
	}{
		S: "2",
	}
	err = Validate(nil, &gt)
	assert.ErrorContains(t, err, "greater")
	gt.S = "21"
	err = Validate(nil, &gt)
	assert.NoError(t, err)

	lte := struct {
		S string `dv8:"val<=2"`
	}{
		S: "21",
	}
	err = Validate(nil, &lte)
	assert.ErrorContains(t, err, "less")
	lte.S = "2"
	err = Validate(nil, &lte)
	assert.NoError(t, err)

	lt := struct {
		S string `dv8:"val<2"`
	}{
		S: "2",
	}
	err = Validate(nil, &lt)
	assert.ErrorContains(t, err, "less")
	lt.S = "19"
	err = Validate(nil, &lt)
	assert.NoError(t, err)

	eq := struct {
		S string `dv8:"val==2"`
	}{
		S: "1",
	}
	err = Validate(nil, &eq)
	assert.ErrorContains(t, err, "equal")
	eq.S = "2"
	err = Validate(nil, &eq)
	assert.NoError(t, err)

	ne := struct {
		S string `dv8:"val!=2"`
	}{
		S: "2",
	}
	err = Validate(nil, &ne)
	assert.ErrorContains(t, err, "equal")
	ne.S = "1"
	err = Validate(nil, &ne)
	assert.NoError(t, err)

	bad := struct {
		S string `dv8:"val*=2"`
	}{
		S: "",
	}
	err = Validate(nil, &bad)
	assert.ErrorContains(t, err, "operator")
}

func TestString_Regexp(t *testing.T) {
	x := struct {
		S string `dv8:"trim,regexp ^[A-Z]*$"`
	}{
		S: " Foo ",
	}
	err := Validate(nil, &x)
	assert.ErrorContains(t, err, "pattern")

	x.S = " FOO "
	err = Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, x.S, "FOO")
}

func TestString_EscapedComma(t *testing.T) {
	// Go's tag syntax consumes one level of escaping, so \\, in source reaches dv8 as \,
	x := struct {
		S string `dv8:"notzero,regexp ^[0-9]{2\\,4}$"`
	}{
		S: "123",
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)

	x.S = "12345"
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "pattern")

	x.S = ""
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "required")

	y := struct {
		S string `dv8:"oneof a\\,b|c"`
	}{
		S: "a,b",
	}
	err = Validate(nil, &y)
	assert.NoError(t, err)

	y.S = "a"
	err = Validate(nil, &y)
	assert.ErrorContains(t, err, "one of")
}

func TestString_RegexpBackslash(t *testing.T) {
	x := struct {
		S string `dv8:"regexp ^\\.$"`
	}{
		S: "m",
	}
	err := Validate(nil, &x)
	assert.ErrorContains(t, err, "pattern")

	x.S = "."
	err = Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, x.S, ".")
}

func TestString_EscapedQuote(t *testing.T) {
	// Go's tag syntax unescapes \" to a literal quote in the directive's value
	x := struct {
		S string `dv8:"notzero,regexp ^\"[a-z]+\"$"`
	}{
		S: `"abc"`,
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)

	x.S = "abc"
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "pattern")
}

func TestString_LiteralBackslash(t *testing.T) {
	// Go's tag syntax unescapes \\\\ to \\, which the regexp matches as a literal backslash.
	// The backslash is not followed by a comma, so it does not act as a dv8 escape.
	x := struct {
		S string `dv8:"regexp ^\\\\$,notzero"`
	}{
		S: `\`,
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)

	x.S = "m"
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "pattern")

	x.S = ""
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "required")
}

func TestString_ToLower(t *testing.T) {
	x := struct {
		S string `dv8:"tolower,default=foo"`
	}{
		S: "",
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, "foo", x.S)

	x.S = "FOO"
	err = Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, "foo", x.S)
}

func TestString_ToUpper(t *testing.T) {
	x := struct {
		S string `dv8:"toupper,default=FOO"`
	}{
		S: "",
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, "FOO", x.S)

	x.S = "foo"
	err = Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, "FOO", x.S)
}

func TestString_OneOf(t *testing.T) {
	x := struct {
		S string `dv8:"oneof Sun|Mon|Tue|Wed|Thu|Fri|Sat"`
	}{
		S: "Foo",
	}
	err := Validate(nil, &x)
	assert.ErrorContains(t, err, "one of")

	x.S = "Mon"
	err = Validate(nil, &x)
	assert.NoError(t, err)

	x.S = ""
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "one of")

	y := struct {
		S string `dv8:"oneof |A|B|C"`
	}{
		S: "",
	}
	err = Validate(nil, &y)
	assert.NoError(t, err)
}
