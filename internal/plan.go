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
	"reflect"
	"strings"
	"sync"
)

// step is one compiled validation action, executed against a value.
type step func(ctx context.Context, refVal reflect.Value) error

// plan is the compiled validation of a type under a set of incoming directives.
// Directive parsing happens once, at compile time; each step captures its pre-parsed operands.
type plan struct {
	steps    []step
	building bool
}

// execute runs the plan's steps in order, stopping at the first error.
func (p *plan) execute(ctx context.Context, refVal reflect.Value) error {
	for _, s := range p.steps {
		err := s(ctx, refVal)
		if err != nil {
			return err
		}
	}
	return nil
}

// isEmpty indicates that the plan has no effect and can be pruned by its caller.
// A plan still being built (a recursive type mid-compilation) is conservatively considered non-empty.
func (p *plan) isEmpty() bool {
	return !p.building && len(p.steps) == 0
}

// planKey identifies a plan node by the type it validates and the directives incoming from its use site.
type planKey struct {
	refType reflect.Type
	tags    string
}

// planResult caches the outcome of compiling a root type: its plan, or the directive error that failed it.
type planResult struct {
	plan *plan
	err  error
}

var cachedPlans sync.Map // reflect.Type -> *planResult

// planOf returns the compiled execution plan of a root type, checking directive validity on first use.
// The result is cached per type.
func planOf(refType reflect.Type) (*plan, error) {
	if refType == nil {
		return nil, nil
	}
	if cached, ok := cachedPlans.Load(refType); ok {
		result := cached.(*planResult)
		return result.plan, result.err
	}
	err := compileType(refType, map[reflect.Type]bool{})
	var p *plan
	if err == nil {
		p = buildPlan(refType, nil, map[planKey]*plan{})
	}
	cachedPlans.Store(refType, &planResult{p, err})
	return p, err
}

// buildPlan compiles the plan of a type under a set of incoming directives, assuming the directives
// were already checked for validity. The memo closes cycles in recursive types: the plan is registered
// before its steps are built, and steps reference sub-plans by pointer.
func buildPlan(refType reflect.Type, tags []string, memo map[planKey]*plan) *plan {
	key := planKey{refType, strings.Join(tags, "\x00")}
	if p, ok := memo[key]; ok {
		return p
	}
	p := &plan{building: true}
	memo[key] = p
	p.steps = buildSteps(refType, tags, memo)
	p.building = false
	return p
}
