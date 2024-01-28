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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTime_Required(t *testing.T) {
	x := struct {
		T time.Time `dv8:"required"`
	}{
		T: time.Now(),
	}
	err := Validate(&x)
	assert.NoError(t, err)

	x.T = time.Time{}
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestTime_Pointer(t *testing.T) {
	x := struct {
		T *time.Time `dv8:"required"`
	}{}
	now := time.Now()
	x.T = &now
	err := Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, now, *x.T)

	x.T = nil
	err = Validate(&x)
	assert.ErrorContains(t, err, "required")
}

func TestTime_Default(t *testing.T) {
	x := struct {
		T time.Time `dv8:"required,default=2006-01-02"`
	}{
		T: time.Time{},
	}
	err := Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, mustParseTime("2006-01-02"), x.T)

	x.T = mustParseTime("2023-01-01")
	err = Validate(&x)
	assert.NoError(t, err)
	assert.Equal(t, mustParseTime("2023-01-01"), x.T)
}

func TestTime_Val(t *testing.T) {
	gte := struct {
		T time.Time `dv8:"val>=2006-01-02T15:04:05"`
	}{
		T: mustParseTime("2006-01-02"),
	}
	err := Validate(&gte)
	assert.ErrorContains(t, err, "later")
	gte.T = mustParseTime("2006-01-02T15:04:05")
	err = Validate(&gte)
	assert.NoError(t, err)

	gt := struct {
		T time.Time `dv8:"val>2006-01-02T15:04:05"`
	}{
		T: mustParseTime("2006-01-02T15:04:05"),
	}
	err = Validate(&gt)
	assert.ErrorContains(t, err, "later")
	gt.T = mustParseTime("2006-01-03")
	err = Validate(&gt)
	assert.NoError(t, err)

	lte := struct {
		T time.Time `dv8:"val<=2006-01-02T15:04:05"`
	}{
		T: mustParseTime("2006-01-03"),
	}
	err = Validate(&lte)
	assert.ErrorContains(t, err, "earlier")
	lte.T = mustParseTime("2006-01-02T15:04:05")
	err = Validate(&lte)
	assert.NoError(t, err)

	lt := struct {
		T time.Time `dv8:"val<2006-01-02T15:04:05"`
	}{
		T: mustParseTime("2006-01-02T15:04:05"),
	}
	err = Validate(&lt)
	assert.ErrorContains(t, err, "earlier")
	lt.T = mustParseTime("2006-01-02")
	err = Validate(&lt)
	assert.NoError(t, err)

	eq := struct {
		T time.Time `dv8:"val==2006-01-02T15:04:05"`
	}{
		T: mustParseTime("2006-01-02"),
	}
	err = Validate(&eq)
	assert.ErrorContains(t, err, "equal")
	eq.T = mustParseTime("2006-01-02T15:04:05")
	err = Validate(&eq)
	assert.NoError(t, err)

	ne := struct {
		T time.Time `dv8:"val!=2006-01-02T15:04:05"`
	}{
		T: mustParseTime("2006-01-02T15:04:05"),
	}
	err = Validate(&ne)
	assert.ErrorContains(t, err, "equal")
	ne.T = mustParseTime("2006-01-02")
	err = Validate(&ne)
	assert.NoError(t, err)

	bad := struct {
		T time.Time `dv8:"val*=2006-01-02T15:04:05"`
	}{
		T: mustParseTime("2006-01-02"),
	}
	err = Validate(&bad)
	assert.ErrorContains(t, err, "operator")
}
