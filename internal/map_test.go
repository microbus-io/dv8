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
		M *map[int]int `dv8:"required"`
	}{}
	m := map[int]int{}
	x.M = &m
	err := Validate(&x)
	assert.NoError(t, err)

	x.M = nil
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestMap_Nesting(t *testing.T) {
	type nested struct {
		I int `dv8:"required"`
	}
	x := struct {
		M map[int]*nested
	}{
		M: map[int]*nested{
			1: {I: 1},
			2: {I: 4},
		},
	}
	err := Validate(&x)
	assert.NoError(t, err)

	x.M[1].I = 0
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestMap_DeepNesting(t *testing.T) {
	type nested struct {
		I int `dv8:"required"`
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
	err := Validate(&x)
	assert.NoError(t, err)

	x.M[999][1].I = 0
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestMap_Len(t *testing.T) {
	gte := struct {
		M map[int]int `dv8:"maplen>=2"`
	}{
		M: map[int]int{},
	}
	err := Validate(&gte)
	assert.ErrorContains(t, err, "greater")
	gte.M = map[int]int{1: 1, 2: 4}
	err = Validate(&gte)
	assert.NoError(t, err)

	gt := struct {
		M map[int]int `dv8:"maplen>2"`
	}{
		M: map[int]int{1: 1, 2: 4},
	}
	err = Validate(&gt)
	assert.ErrorContains(t, err, "greater")
	gt.M = map[int]int{1: 1, 2: 4, 3: 9}
	err = Validate(&gt)
	assert.NoError(t, err)

	lte := struct {
		M map[int]int `dv8:"maplen<=2"`
	}{
		M: map[int]int{1: 1, 2: 4, 3: 9},
	}
	err = Validate(&lte)
	assert.ErrorContains(t, err, "less")
	lte.M = map[int]int{1: 1, 2: 4}
	err = Validate(&lte)
	assert.NoError(t, err)

	lt := struct {
		M map[int]int `dv8:"maplen<2"`
	}{
		M: map[int]int{1: 1, 2: 4},
	}
	err = Validate(&lt)
	assert.ErrorContains(t, err, "less")
	lt.M = map[int]int{1: 1}
	err = Validate(&lt)
	assert.NoError(t, err)

	eq := struct {
		M map[int]int `dv8:"maplen==2"`
	}{
		M: map[int]int{1: 1, 2: 4, 3: 9},
	}
	err = Validate(&eq)
	assert.ErrorContains(t, err, "equal")
	eq.M = map[int]int{1: 1, 2: 4}
	err = Validate(&eq)
	assert.NoError(t, err)

	ne := struct {
		M map[int]int `dv8:"maplen!=2"`
	}{
		M: map[int]int{1: 1, 2: 4},
	}
	err = Validate(&ne)
	assert.ErrorContains(t, err, "equal")
	ne.M = map[int]int{1: 1, 2: 4, 3: 9}
	err = Validate(&ne)
	assert.NoError(t, err)

	bad := struct {
		M map[int]int `dv8:"maplen*=2"`
	}{
		M: map[int]int{},
	}
	err = Validate(&bad)
	assert.ErrorContains(t, err, "operator")

	zero := struct {
		A map[int]int `dv8:"maplen>=0"`
	}{}
	err = Validate(&zero)
	assert.ErrorContains(t, err, "required")
	zero.A = map[int]int{}
	err = Validate(&zero)
	assert.NoError(t, err)
}

func TestMap_Items(t *testing.T) {
	items := struct {
		M map[int]string `dv8:"len>1,toupper"`
	}{
		M: map[int]string{1: "Foo"},
	}
	err := Validate(&items)
	assert.NoError(t, err)
	assert.Equal(t, "FOO", items.M[1])
	items.M[2] = "x"
	err = Validate(&items)
	assert.ErrorContains(t, err, "length")
}
