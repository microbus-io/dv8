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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ctxValueKey = struct{}{}

type Person struct {
	Name string `dv8:"notzero"`
	Zip  string `dv8:"notzero,regexp ^[0-9]{5}$"`
	Age  int    `dv8:"val>=18"`
}

func (p Person) Validate(ctx context.Context) error {
	if ctx.Value(ctxValueKey) != nil {
		return errors.New("failed by context")
	}
	if p.Name == "Fail Validate" {
		return errors.New("failed by name")
	}
	return nil
}

type Robot struct {
	Name string `dv8:"notzero"`
}

func (r *Robot) Validate() error {
	if r.Name == "Fail Validate" {
		return errors.New("Validate")
	}
	return nil
}

type Directory struct {
	Persons []*Person `dv8:"notzero,len>0"`
}

type Animal struct {
	Name string `dv8:"notzero"`
	Kind string `dv8:"default=Mammal"`
}

func Test_Directory(t *testing.T) {
	d := Directory{}

	err := Validate(nil, &d)
	assert.ErrorContains(t, err, "required")

	d.Persons = []*Person{}
	err = Validate(nil, &d)
	assert.ErrorContains(t, err, "length")

	// All good
	d.Persons = append(d.Persons, &Person{
		Name: "Jane",
		Zip:  "12345",
		Age:  19,
	})

	err = Validate(nil, &d)
	assert.NoError(t, err)

	// Name required
	d.Persons = append(d.Persons, &Person{
		Name: "",
		Zip:  "12345",
		Age:  19,
	})

	err = Validate(nil, &d)
	assert.ErrorContains(t, err, "required")
	assert.ErrorContains(t, err, "Name: ")

	d.Persons[len(d.Persons)-1].Name = "John"
	err = Validate(nil, &d)
	assert.NoError(t, err)

	// Bad zip code pattern
	d.Persons = append(d.Persons, &Person{
		Name: "Max",
		Zip:  "123456",
		Age:  19,
	})

	err = Validate(nil, &d)
	assert.ErrorContains(t, err, "pattern")
	assert.ErrorContains(t, err, "Zip: ")

	d.Persons[len(d.Persons)-1].Zip = "12345"
	err = Validate(nil, &d)
	assert.NoError(t, err)

	// Too young
	d.Persons = append(d.Persons, &Person{
		Name: "Max",
		Zip:  "12345",
		Age:  16,
	})

	err = Validate(nil, &d)
	assert.ErrorContains(t, err, "greater")
	assert.ErrorContains(t, err, "Age: ")

	d.Persons[len(d.Persons)-1].Age = 21
	err = Validate(nil, &d)
	assert.NoError(t, err)
}

func Test_ValidatorInterface(t *testing.T) {
	// OK case
	p := Person{
		Name: "Saul Goodman",
		Zip:  "12345",
		Age:  18,
	}
	err := Validate(nil, p)
	assert.NoError(t, err)
	err = Validate(context.Background(), p)
	assert.NoError(t, err)

	// Fail Validate, via the pointer-receiver parameterless variant
	r := &Robot{
		Name: "Fail Validate",
	}
	err = Validate(nil, r)
	assert.Error(t, err)
	err = Validate(context.Background(), r)
	assert.Error(t, err)

	// Pointer-receiver Validate on an addressable element, without a pointer at the top level
	rr := []Robot{{Name: "Fail Validate"}}
	err = Validate(nil, &rr)
	assert.Error(t, err)

	// Fail the context-taking Validate by name
	p = Person{
		Name: "Fail Validate",
		Zip:  "12345",
		Age:  18,
	}
	err = Validate(nil, p)
	assert.Error(t, err)
	err = Validate(context.Background(), p)
	assert.Error(t, err)

	// Custom context
	p = Person{
		Name: "Saul Goodman",
		Zip:  "12345",
		Age:  18,
	}
	failCtx := context.WithValue(context.Background(), ctxValueKey, "fail")
	err = Validate(nil, p)
	assert.NoError(t, err) // Doesn't fail because using context.Background
	err = Validate(failCtx, p)
	assert.Error(t, err) // Fails with failCtx
}

func Test_ReferenceTypes(t *testing.T) {
	p := Animal{
		Name: "Zebra",
	}
	err := Validate(nil, p)
	assert.ErrorContains(t, err, "reference")

	p = Animal{
		Name: "Zebra",
	}
	err = Validate(nil, &p)
	assert.NoError(t, err)
	assert.Equal(t, "Mammal", p.Kind)

	p = Animal{
		Name: "Zebra",
	}
	err = Validate(nil, map[int]Animal{1: p})
	assert.ErrorContains(t, err, "reference")

	p = Animal{
		Name: "Zebra",
	}
	err = Validate(nil, map[int]*Animal{1: &p})
	assert.NoError(t, err)
	assert.Equal(t, "Mammal", p.Kind)

	p = Animal{
		Name: "Zebra",
	}
	err = Validate(nil, []Animal{p})
	assert.NoError(t, err, "reference")

	p = Animal{
		Name: "Zebra",
	}
	err = Validate(nil, []*Animal{&p})
	assert.NoError(t, err)
	assert.Equal(t, "Mammal", p.Kind)
}
