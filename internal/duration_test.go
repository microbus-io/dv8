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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDuration_Required(t *testing.T) {
	x := struct {
		D time.Duration `dv8:"required"`
	}{
		D: time.Second,
	}
	err := Validate(&x)
	assert.NoError(t, err)

	x.D = 0
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestDuration_Pointer(t *testing.T) {
	x := struct {
		D *time.Duration `dv8:"required"`
	}{}
	dur := time.Second
	x.D = &dur
	err := Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, time.Second, *x.D)

	x.D = nil
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestDuration_Default(t *testing.T) {
	x := struct {
		D time.Duration `dv8:"required,default=2s"`
	}{
		D: 0,
	}
	err := Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, time.Second*2, x.D)

	x.D = time.Minute
	err = Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, time.Minute, x.D)
}

func TestDuration_Val(t *testing.T) {
	gte := struct {
		D time.Duration `dv8:"val>=2s"`
	}{
		D: 0,
	}
	err := Validate(&gte)
	assert.ErrorContains(t, err, "greater")
	gte.D = time.Second * 2
	err = Validate(&gte)
	assert.NoError(t, err)

	gt := struct {
		D time.Duration `dv8:"val>2s"`
	}{
		D: time.Second * 2,
	}
	err = Validate(&gt)
	assert.ErrorContains(t, err, "greater")
	gt.D = time.Second * 3
	err = Validate(&gt)
	assert.NoError(t, err)

	lte := struct {
		D time.Duration `dv8:"val<=2s"`
	}{
		D: time.Second * 3,
	}
	err = Validate(&lte)
	assert.ErrorContains(t, err, "less")
	lte.D = time.Second * 2
	err = Validate(&lte)
	assert.NoError(t, err)

	lt := struct {
		D time.Duration `dv8:"val<2s"`
	}{
		D: time.Second * 2,
	}
	err = Validate(&lt)
	assert.ErrorContains(t, err, "less")
	lt.D = time.Second * 1
	err = Validate(&lt)
	assert.NoError(t, err)

	eq := struct {
		D time.Duration `dv8:"val==2s"`
	}{
		D: time.Second * 1,
	}
	err = Validate(&eq)
	assert.ErrorContains(t, err, "equal")
	eq.D = time.Second * 2
	err = Validate(&eq)
	assert.NoError(t, err)

	ne := struct {
		D time.Duration `dv8:"val!=2s"`
	}{
		D: time.Second * 2,
	}
	err = Validate(&ne)
	assert.ErrorContains(t, err, "equal")
	ne.D = time.Second * 1
	err = Validate(&ne)
	assert.NoError(t, err)

	bad := struct {
		D time.Duration `dv8:"val*=2s"`
	}{
		D: 0,
	}
	err = Validate(&bad)
	assert.ErrorContains(t, err, "operator")
}
