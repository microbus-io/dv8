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

func TestCompile_UnknownDirective(t *testing.T) {
	x := struct {
		S string `dv8:"requird"`
	}{}
	err := Validate(nil, &x)
	assert.True(t, errors.Is(err, ErrDirective))
	assert.ErrorContains(t, err, "unknown directive")
	assert.ErrorContains(t, err, "requird")
	assert.ErrorContains(t, err, "S: ")
}

func TestCompile_RequiredRetired(t *testing.T) {
	x := struct {
		S string `dv8:"required"`
	}{}
	err := Validate(nil, &x)
	assert.True(t, errors.Is(err, ErrDirective))
	assert.ErrorContains(t, err, "notzero")
}

func TestCompile_NotApplicable(t *testing.T) {
	x := struct {
		I int `dv8:"trim"`
	}{}
	err := Validate(nil, &x)
	assert.True(t, errors.Is(err, ErrDirective))
	assert.ErrorContains(t, err, "not applicable")

	y := struct {
		S string `dv8:"each len>0"`
	}{}
	err = Validate(nil, &y)
	assert.True(t, errors.Is(err, ErrDirective))
	assert.ErrorContains(t, err, "arrays and maps")

	z := struct {
		A []int `dv8:"key val>0"`
	}{}
	err = Validate(nil, &z)
	assert.True(t, errors.Is(err, ErrDirective))
	assert.ErrorContains(t, err, "maps")
}

func TestCompile_BadOperator(t *testing.T) {
	x := struct {
		I int `dv8:"val*=2"`
	}{}
	err := Validate(nil, &x)
	assert.True(t, errors.Is(err, ErrDirective))
	assert.ErrorContains(t, err, "operator")

	y := struct {
		S string `dv8:"len"`
	}{}
	err = Validate(nil, &y)
	assert.True(t, errors.Is(err, ErrDirective))
	assert.ErrorContains(t, err, "incomplete")
}

func TestCompile_BadValue(t *testing.T) {
	x := struct {
		I int `dv8:"val>abc"`
	}{}
	err := Validate(nil, &x)
	assert.True(t, errors.Is(err, ErrDirective))

	y := struct {
		B bool `dv8:"default=xyz"`
	}{}
	err = Validate(nil, &y)
	assert.True(t, errors.Is(err, ErrDirective))

	z := struct {
		B bool `dv8:"val>true"`
	}{}
	err = Validate(nil, &z)
	assert.True(t, errors.Is(err, ErrDirective))
	assert.ErrorContains(t, err, "operator")
}

func TestCompile_BadRegexp(t *testing.T) {
	x := struct {
		S string `dv8:"regexp ^[a-z$"`
	}{}
	err := Validate(nil, &x)
	assert.True(t, errors.Is(err, ErrDirective))
}

func TestCompile_OnMissingField(t *testing.T) {
	type inner struct {
		ID int
	}
	x := struct {
		N inner `dv8:"notzero,on Feild"`
	}{}
	err := Validate(nil, &x)
	assert.True(t, errors.Is(err, ErrDirective))
	assert.ErrorContains(t, err, "Feild")
}

func TestCompile_OnAndDelegateTargets(t *testing.T) {
	// A directive inapplicable to the struct passes when its "on" target accepts it
	type inner struct {
		ID int
	}
	x := struct {
		N inner `dv8:"notzero,val>2,on ID"`
	}{
		N: inner{ID: 3},
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)

	// Same via a delegate field
	type wrapper struct {
		Name string `dv8:"delegate"`
	}
	y := struct {
		W wrapper `dv8:"default=Unknown"`
	}{}
	err = Validate(nil, &y)
	assert.NoError(t, err)
	assert.Equal(t, "Unknown", y.W.Name)

	// Without a delegate, the same directive is inapplicable
	type plain struct {
		Name string
	}
	z := struct {
		P plain `dv8:"default=Unknown"`
	}{}
	err = Validate(nil, &z)
	assert.True(t, errors.Is(err, ErrDirective))
	assert.ErrorContains(t, err, "not applicable")
}

func TestCompile_NestedTypes(t *testing.T) {
	type deep struct {
		S string `dv8:"lenght>0"`
	}
	x := struct {
		M map[string][]*deep
	}{}
	err := Validate(nil, &x)
	assert.True(t, errors.Is(err, ErrDirective))
	assert.ErrorContains(t, err, "M: ")
	assert.ErrorContains(t, err, "S: ")
}

func TestCompile_Recursive(t *testing.T) {
	type node struct {
		Name string `dv8:"notzero"`
		Next *node
	}
	n := node{Name: "a", Next: &node{Name: "b"}}
	err := Validate(nil, &n)
	assert.NoError(t, err)
}

func TestCompile_EachKeyNesting(t *testing.T) {
	x := struct {
		M map[string]map[string]int `dv8:"each key len>0,each each val>=0"`
	}{
		M: map[string]map[string]int{"a": {"b": 1}},
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)

	y := struct {
		M map[string]map[string]int `dv8:"each key trim"`
	}{}
	err = Validate(nil, &y)
	assert.NoError(t, err) // trim on a string key is valid

	z := struct {
		M map[string][]int `dv8:"each key len>0"`
	}{}
	err = Validate(nil, &z)
	assert.True(t, errors.Is(err, ErrDirective))
	assert.ErrorContains(t, err, "maps")
}
