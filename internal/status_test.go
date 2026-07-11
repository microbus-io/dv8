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
	"context"
	stderrors "errors"
	"net/http"
	"testing"

	"github.com/microbus-io/errors"
	"github.com/stretchr/testify/assert"
)

type plainValidated struct {
	Name string
}

func (p *plainValidated) Validate() error {
	return stderrors.New("plain failure")
}

type tracedValidated struct {
	Name string
}

func (p *tracedValidated) Validate() error {
	return errors.New("traced failure")
}

type forbiddenValidated struct {
	Name string
}

func (p *forbiddenValidated) Validate(ctx context.Context) error {
	return errors.New("forbidden", http.StatusForbidden)
}

func TestStatus_DataErrors(t *testing.T) {
	// A tag violation is attributed to the input data
	x := struct {
		S string `dv8:"len<=2"`
	}{
		S: "abc",
	}
	err := Validate(nil, &x)
	assert.Equal(t, http.StatusBadRequest, errors.StatusCode(err))

	// The status survives nesting prefixes
	type inner struct {
		S string `dv8:"notzero"`
	}
	y := struct {
		A []*inner
	}{
		A: []*inner{{}},
	}
	err = Validate(nil, &y)
	assert.Equal(t, http.StatusBadRequest, errors.StatusCode(err))
	assert.ErrorContains(t, err, "A: ")
	assert.ErrorContains(t, err, "[0]: ")
	assert.ErrorContains(t, err, "S: ")
}

func TestStatus_DirectiveErrors(t *testing.T) {
	// A malformed directive is not the caller's fault
	x := struct {
		S string `dv8:"lenght>0"`
	}{}
	err := Validate(nil, &x)
	assert.True(t, stderrors.Is(err, ErrDirective))
	assert.Equal(t, http.StatusInternalServerError, errors.StatusCode(err))
}

func TestStatus_CustomValidate(t *testing.T) {
	// A plain error from a custom validator is attributed to the input data
	err := Validate(nil, &plainValidated{})
	assert.Equal(t, http.StatusBadRequest, errors.StatusCode(err))

	// So is a traced error carrying the default 500
	err = Validate(nil, &tracedValidated{})
	assert.Equal(t, http.StatusBadRequest, errors.StatusCode(err))

	// A deliberate non-500 status chosen by the validator is respected
	err = Validate(nil, &forbiddenValidated{})
	assert.Equal(t, http.StatusForbidden, errors.StatusCode(err))
}
