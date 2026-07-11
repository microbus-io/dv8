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

func TestMap_Pointer(t *testing.T) {
	x := struct {
		M *map[int]int `dv8:"notzero"`
	}{}
	m := map[int]int{}
	x.M = &m
	err := Validate(nil, &x)
	assert.NoError(t, err)

	x.M = nil
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "required")
}

func TestMap_Nesting(t *testing.T) {
	type nested struct {
		I int `dv8:"notzero"`
	}
	x := struct {
		M map[int]*nested
	}{
		M: map[int]*nested{
			1: {I: 1},
			2: {I: 4},
		},
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)

	x.M[1].I = 0
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "required")
}

func TestMap_DeepNesting(t *testing.T) {
	type nested struct {
		I int `dv8:"notzero"`
	}
	x := struct {
		M map[int]map[int]*nested
	}{
		M: map[int]map[int]*nested{
			999: {
				1: {I: 1},
				2: {I: 4},
			},
		},
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)

	x.M[999][1].I = 0
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "required")
}

func TestMap_Len(t *testing.T) {
	gte := struct {
		M map[int]int `dv8:"len>=2"`
	}{
		M: map[int]int{},
	}
	err := Validate(nil, &gte)
	assert.ErrorContains(t, err, "greater")
	gte.M = map[int]int{1: 1, 2: 4}
	err = Validate(nil, &gte)
	assert.NoError(t, err)

	gt := struct {
		M map[int]int `dv8:"len>2"`
	}{
		M: map[int]int{1: 1, 2: 4},
	}
	err = Validate(nil, &gt)
	assert.ErrorContains(t, err, "greater")
	gt.M = map[int]int{1: 1, 2: 4, 3: 9}
	err = Validate(nil, &gt)
	assert.NoError(t, err)

	lte := struct {
		M map[int]int `dv8:"len<=2"`
	}{
		M: map[int]int{1: 1, 2: 4, 3: 9},
	}
	err = Validate(nil, &lte)
	assert.ErrorContains(t, err, "less")
	lte.M = map[int]int{1: 1, 2: 4}
	err = Validate(nil, &lte)
	assert.NoError(t, err)

	lt := struct {
		M map[int]int `dv8:"len<2"`
	}{
		M: map[int]int{1: 1, 2: 4},
	}
	err = Validate(nil, &lt)
	assert.ErrorContains(t, err, "less")
	lt.M = map[int]int{1: 1}
	err = Validate(nil, &lt)
	assert.NoError(t, err)

	eq := struct {
		M map[int]int `dv8:"len==2"`
	}{
		M: map[int]int{1: 1, 2: 4, 3: 9},
	}
	err = Validate(nil, &eq)
	assert.ErrorContains(t, err, "equal")
	eq.M = map[int]int{1: 1, 2: 4}
	err = Validate(nil, &eq)
	assert.NoError(t, err)

	ne := struct {
		M map[int]int `dv8:"len!=2"`
	}{
		M: map[int]int{1: 1, 2: 4},
	}
	err = Validate(nil, &ne)
	assert.ErrorContains(t, err, "equal")
	ne.M = map[int]int{1: 1, 2: 4, 3: 9}
	err = Validate(nil, &ne)
	assert.NoError(t, err)

	bad := struct {
		M map[int]int `dv8:"len*=2"`
	}{
		M: map[int]int{},
	}
	err = Validate(nil, &bad)
	assert.ErrorContains(t, err, "operator")

	zero := struct {
		A map[int]int `dv8:"notzero,len>=0"`
	}{}
	err = Validate(nil, &zero)
	assert.ErrorContains(t, err, "required")
	zero.A = map[int]int{}
	err = Validate(nil, &zero)
	assert.NoError(t, err)
}

func TestMap_Items(t *testing.T) {
	items := struct {
		M map[int]string `dv8:"each len>1,each toupper"`
	}{
		M: map[int]string{1: "Foo"},
	}
	err := Validate(nil, &items)
	assert.NoError(t, err)
	assert.Equal(t, "FOO", items.M[1])
	items.M[2] = "x"
	err = Validate(nil, &items)
	assert.ErrorContains(t, err, "length")
}

func TestMap_Keys(t *testing.T) {
	x := struct {
		M map[string]int `dv8:"key len<=3,key regexp ^[a-z]+$"`
	}{
		M: map[string]int{"abc": 1},
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)

	x.M["toolong"] = 2
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "length")
	assert.ErrorContains(t, err, "key")
	delete(x.M, "toolong")

	x.M["AB"] = 3
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "pattern")
}

func TestMap_KeyMutation(t *testing.T) {
	x := struct {
		M map[string]int `dv8:"key trim,key tolower"`
	}{
		M: map[string]int{" Foo ": 1, "bar": 2},
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, map[string]int{"foo": 1, "bar": 2}, x.M)

	// Two keys folding into one is an error
	x.M = map[string]int{" a ": 1, "a": 2}
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "already exists")

	// Mutating keys of a map not passed by reference is an error
	x.M = map[string]int{" a ": 1}
	err = Validate(nil, x)
	assert.ErrorContains(t, err, "reference")
}

func TestMap_KeysAndValues(t *testing.T) {
	x := struct {
		M map[string]string `dv8:"len>0,key len>0,each len>0,each trim"`
	}{
		M: map[string]string{"a": " b "},
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, "b", x.M["a"])

	x.M["c"] = " "
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "length")
}
