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

package dv8_test

import (
	"context"
	"testing"
	"time"

	"github.com/microbus-io/dv8"
)

type benchAddress struct {
	Street string `dv8:"trim,len<=128"`
	City   string `dv8:"trim,notzero,len<=64"`
	State  string `dv8:"toupper,len==2,default=CA"`
	Zip    string `dv8:"regexp ^[0-9]{5}$"`
}

type benchPerson struct {
	First    string            `dv8:"trim,notzero,len<=32"`
	Last     string            `dv8:"trim,notzero,len<=32"`
	Age      int               `dv8:"val>=0,val<=120"`
	Balance  float64           `dv8:"val>=0"`
	Size     string            `dv8:"oneof S|M|L,default=M"`
	Waited   time.Duration     `dv8:"val>=0s,val<=24h"`
	Address  *benchAddress     `dv8:"notzero"`
	Tags     []string          `dv8:"len<=8,each trim,each len>0"`
	Attribs  map[string]string `dv8:"len<=8,key len>=2,each notzero"`
	Friends  []benchAddress    ``
	Verified bool              ``
}

func (p *benchPerson) Validate(ctx context.Context) error {
	return nil
}

func newBenchPerson() *benchPerson {
	return &benchPerson{
		First:   " Jane ",
		Last:    "Simmons",
		Age:     36,
		Balance: 102.5,
		Size:    "",
		Waited:  5 * time.Minute,
		Address: &benchAddress{
			Street: "123 Main St",
			City:   "Anytown",
			State:  "ny",
			Zip:    "12345",
		},
		Tags: []string{"alpha", " beta ", "gamma"},
		Attribs: map[string]string{
			"color": "blue",
			"shape": "round",
		},
		Friends: []benchAddress{
			{Street: "1 First Ave", City: "Springfield", State: "il", Zip: "54321"},
			{Street: "2 Second Ave", City: "Shelbyville", State: "il", Zip: "54322"},
		},
	}
}

func BenchmarkValidate(b *testing.B) {
	ctx := context.Background()
	person := newBenchPerson()
	err := dv8.Validate(ctx, person)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		err = dv8.Validate(ctx, person)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidate_Flat(b *testing.B) {
	ctx := context.Background()
	addr := &benchAddress{
		Street: "123 Main St",
		City:   "Anytown",
		State:  "NY",
		Zip:    "12345",
	}
	err := dv8.Validate(ctx, addr)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		err = dv8.Validate(ctx, addr)
		if err != nil {
			b.Fatal(err)
		}
	}
}
