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

func TestArray_Pointer(t *testing.T) {
	x := struct {
		A *[]int `dv8:"required"`
	}{}
	a := []int{}
	x.A = &a
	err := Validate(&x)
	assert.NoError(t, err)

	x.A = nil
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestArray_Nesting(t *testing.T) {
	type nested struct {
		I int `dv8:"required"`
	}
	x := struct {
		A []*nested
	}{
		A: []*nested{
			{I: 1},
			{I: 4},
		},
	}
	err := Validate(&x)
	assert.NoError(t, err)

	x.A[0].I = 0
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestArray_DeepNesting(t *testing.T) {
	type nested struct {
		I int `dv8:"required"`
	}
	x := struct {
		A [][]*nested
	}{
		A: [][]*nested{{
			{I: 1},
			{I: 4},
		}},
	}
	err := Validate(&x)
	assert.NoError(t, err)

	x.A[0][0].I = 0
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestArray_ArrLen(t *testing.T) {
	gte := struct {
		A []int `dv8:"arrlen>=2"`
	}{
		A: []int{},
	}
	err := Validate(&gte)
	assert.ErrorContains(t, err, "greater")
	gte.A = []int{1, 2}
	err = Validate(&gte)
	assert.NoError(t, err)

	gt := struct {
		A []int `dv8:"arrlen>2"`
	}{
		A: []int{1, 2},
	}
	err = Validate(&gt)
	assert.ErrorContains(t, err, "greater")
	gt.A = []int{1, 2, 3}
	err = Validate(&gt)
	assert.NoError(t, err)

	lte := struct {
		A []int `dv8:"arrlen<=2"`
	}{
		A: []int{1, 2, 3},
	}
	err = Validate(&lte)
	assert.ErrorContains(t, err, "less")
	lte.A = []int{1, 2}
	err = Validate(&lte)
	assert.NoError(t, err)

	lt := struct {
		A []int `dv8:"arrlen<2"`
	}{
		A: []int{1, 2},
	}
	err = Validate(&lt)
	assert.ErrorContains(t, err, "less")
	lt.A = []int{1}
	err = Validate(&lt)
	assert.NoError(t, err)

	eq := struct {
		A []int `dv8:"arrlen==2"`
	}{
		A: []int{1, 2, 3},
	}
	err = Validate(&eq)
	assert.ErrorContains(t, err, "equal")
	eq.A = []int{1, 2}
	err = Validate(&eq)
	assert.NoError(t, err)

	ne := struct {
		A []int `dv8:"arrlen!=2"`
	}{
		A: []int{1, 2},
	}
	err = Validate(&ne)
	assert.ErrorContains(t, err, "equal")
	ne.A = []int{1, 2, 3}
	err = Validate(&ne)
	assert.NoError(t, err)

	bad := struct {
		A []int `dv8:"arrlen*=2"`
	}{
		A: []int{},
	}
	err = Validate(&bad)
	assert.ErrorContains(t, err, "operator")

	zero := struct {
		A []int `dv8:"arrlen>=0"`
	}{}
	err = Validate(&zero)
	assert.ErrorContains(t, err, "required")
	zero.A = []int{}
	err = Validate(&zero)
	assert.NoError(t, err)
}

func TestArray_Items(t *testing.T) {
	items := struct {
		A []string `dv8:"len>1,toupper"`
	}{
		A: []string{"Foo"},
	}
	err := Validate(&items)
	assert.NoError(t, err)
	assert.Equal(t, "FOO", items.A[0])
	items.A = append(items.A, "x")
	err = Validate(&items)
	assert.ErrorContains(t, err, "length")
	assert.ErrorContains(t, err, "[1]")
}
