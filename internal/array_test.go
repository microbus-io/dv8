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
		A *[]int `dv8:"notzero"`
	}{}
	a := []int{}
	x.A = &a
	err := Validate(nil, &x)
	assert.NoError(t, err)

	x.A = nil
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "required")
}

func TestArray_Nesting(t *testing.T) {
	type nested struct {
		I int `dv8:"notzero"`
	}
	x := struct {
		A []*nested
	}{
		A: []*nested{
			{I: 1},
			{I: 4},
		},
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)

	x.A[0].I = 0
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "required")
}

func TestArray_DeepNesting(t *testing.T) {
	type nested struct {
		I int `dv8:"notzero"`
	}
	x := struct {
		A [][]*nested
	}{
		A: [][]*nested{{
			{I: 1},
			{I: 4},
		}},
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)

	x.A[0][0].I = 0
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "required")
}

func TestArray_ArrLen(t *testing.T) {
	gte := struct {
		A []int `dv8:"len>=2"`
	}{
		A: []int{},
	}
	err := Validate(nil, &gte)
	assert.ErrorContains(t, err, "greater")
	gte.A = []int{1, 2}
	err = Validate(nil, &gte)
	assert.NoError(t, err)

	gt := struct {
		A []int `dv8:"len>2"`
	}{
		A: []int{1, 2},
	}
	err = Validate(nil, &gt)
	assert.ErrorContains(t, err, "greater")
	gt.A = []int{1, 2, 3}
	err = Validate(nil, &gt)
	assert.NoError(t, err)

	lte := struct {
		A []int `dv8:"len<=2"`
	}{
		A: []int{1, 2, 3},
	}
	err = Validate(nil, &lte)
	assert.ErrorContains(t, err, "less")
	lte.A = []int{1, 2}
	err = Validate(nil, &lte)
	assert.NoError(t, err)

	lt := struct {
		A []int `dv8:"len<2"`
	}{
		A: []int{1, 2},
	}
	err = Validate(nil, &lt)
	assert.ErrorContains(t, err, "less")
	lt.A = []int{1}
	err = Validate(nil, &lt)
	assert.NoError(t, err)

	eq := struct {
		A []int `dv8:"len==2"`
	}{
		A: []int{1, 2, 3},
	}
	err = Validate(nil, &eq)
	assert.ErrorContains(t, err, "equal")
	eq.A = []int{1, 2}
	err = Validate(nil, &eq)
	assert.NoError(t, err)

	ne := struct {
		A []int `dv8:"len!=2"`
	}{
		A: []int{1, 2},
	}
	err = Validate(nil, &ne)
	assert.ErrorContains(t, err, "equal")
	ne.A = []int{1, 2, 3}
	err = Validate(nil, &ne)
	assert.NoError(t, err)

	bad := struct {
		A []int `dv8:"len*=2"`
	}{
		A: []int{},
	}
	err = Validate(nil, &bad)
	assert.ErrorContains(t, err, "operator")

	zero := struct {
		A []int `dv8:"notzero,len>=0"`
	}{}
	err = Validate(nil, &zero)
	assert.ErrorContains(t, err, "required")
	zero.A = []int{}
	err = Validate(nil, &zero)
	assert.NoError(t, err)
}

func TestArray_Items(t *testing.T) {
	items := struct {
		A []string `dv8:"each len>1,each toupper"`
	}{
		A: []string{"Foo"},
	}
	err := Validate(nil, &items)
	assert.NoError(t, err)
	assert.Equal(t, "FOO", items.A[0])
	items.A = append(items.A, "x")
	err = Validate(nil, &items)
	assert.ErrorContains(t, err, "length")
	assert.ErrorContains(t, err, "[1]")
}

func TestArray_ContainerAndItems(t *testing.T) {
	x := struct {
		A []string `dv8:"len!=0,each len!=0,each trim"`
	}{}
	err := Validate(nil, &x)
	assert.ErrorContains(t, err, "length")

	x.A = []string{" Foo "}
	err = Validate(nil, &x)
	assert.NoError(t, err)
	assert.Equal(t, "Foo", x.A[0])

	x.A = append(x.A, "  ")
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "length")
	assert.ErrorContains(t, err, "[1]")
}

func TestArray_EachChaining(t *testing.T) {
	x := struct {
		A [][]string `dv8:"len==1,each len>0,each each len<=3"`
	}{
		A: [][]string{{"ab", "cd"}},
	}
	err := Validate(nil, &x)
	assert.NoError(t, err)

	x.A[0][1] = "toolong"
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "length")

	x.A = [][]string{{}}
	err = Validate(nil, &x)
	assert.ErrorContains(t, err, "length")
}

func TestArray_KeyRejected(t *testing.T) {
	x := struct {
		A []string `dv8:"key len>0"`
	}{
		A: []string{"a"},
	}
	err := Validate(nil, &x)
	assert.ErrorContains(t, err, "maps")
}
