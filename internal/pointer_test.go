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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type BigRect struct {
	W int `dv8:"val>=0"`
	H int `dv8:"val>=0"`
}

func (r *BigRect) Validate() error {
	if r.H*r.W < 100 {
		return errors.New("too small")
	}
	return nil
}

func TestPointer_Validator(t *testing.T) {
	small := BigRect{W: 5, H: 5}
	err := Validate(&small)
	assert.Error(t, err, "too small")

	big := BigRect{W: 50, H: 50}
	err = Validate(&big)
	assert.NoError(t, err)
}
