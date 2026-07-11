# `DV8` - Data Validation for Golang

## Overview

`DV8` uses Golang's struct tags to validate data of struct fields.
Its primary purpose is validation of data entered by an untrusted source such as an end-user.
It draws inspiration from [Pydantic](https://docs.pydantic.dev).

```go
type Person struct {
    First   string `dv8:"trim,notzero,len<=32"`
    Last    string `dv8:"notzero,len<=32"`
    Age     int    `dv8:"val>=0,val<=120"`
    State   string `dv8:"len==2,default=CA,toupper"`
    Zip     string `dv8:"notzero,regexp ^[0-9]{5}$"`
    Country string `dv8:"notzero,len==2,oneof US|MX,default=US,toupper"`
}

p := &Person{
    First:   " Julie",  // Trim whitespaces
    Last:    "Supercalifragilisticexpialidocious", // Enforce length limits
    State:   "",        // Set default to "CA"
    Age:     200,       // Enforce value constraints
    Zip:     "12x45",   // Enforce a regexp pattern
    Country: "USA",     // Check against a set of valid values
}

err := dv8.Validate(ctx, p)
if err != nil {
    return err
}
```

## Directives

`DV8` recognizes the following directives:

|Directive|Applicable types|Effect|
|---|---|---|
|`notzero`|`any`|Requires a value that is not the type's zero value: non-`nil` for pointers, arrays and maps, non-empty for strings, `true` for booleans|
|`default`|`string`, `int`, `float`, `bool`, `time.Time`, `time.Duration`|Sets a default value when the zero-value is provided|
|`val` with `==` or `!=`|`string`, `int`, `float`, `bool`, `time.Time`, `time.Duration`|Enforces an equality constraint on the value|
|`val` with `<=`, `<`, `>=` or `>`|`string`, `int`, `float`, `time.Time`, `time.Duration`|Enforces an ordering constraint on the value. Strings compare lexicographically, so `"9" > "10"`|
|`len` with `==`, `!=`, `<=`, `<`, `>=` or `>`|`string`, `[]any`, `map[any]any`|Enforces a constraint on the length of the string (in runes, not bytes), array or map. A `nil` array or map has length 0; use `notzero` to check for `nil`|
|`oneof`|`string`|Check against a set of valid values separated by a `\|`|
|`regexp`|`string`|Requires the string to match a regular expression|
|`each`|`[]any`, `map[any]any`|Applies the directive that follows it to each of the elements of the array, or values of the map, e.g. `each len>0` (see below)|
|`key`|`map[any]any`|Applies the directive that follows it to each of the keys of the map, e.g. `key len>0` (see below)|
|`on`|`struct`, `*struct`|Applies the directives on the named field of the struct instead of the struct itself (see below)|
|`delegate`|`any`|Applies the directives set on the parent struct to the field (see below)|
|`trim`|`string`|Trims leading and trailing whitespaces before validation|
|`tolower`|`string`|Transforms the string to lowercase|
|`toupper`|`string`|Transforms the string to uppercase|
|`-`|`any`|Skips the field and stops recursion into nested fields|

Directives are separated by commas. To include a comma inside a directive's value, such as in a regular expression
quantifier or a `oneof` option, escape it with a backslash. Go's struct tag syntax consumes one level of escaping,
so it is written `\\,` in the tag:

```go
type Card struct {
    Number string `dv8:"notzero,regexp ^[0-9]{13\\,19}$"`
}
```

## `on` and `delegate`

`on` and `delegate` are two sides of the same wrapper-type mechanism: they route directives aimed at a struct to one
of its fields, typically when the struct is a thin wrapper around a single meaningful value. Declare the routing
either on the wrapper itself (`delegate`) or at its point of use (`on`).

The `on` directive allows pushing directives one level down into a nested field of a struct. It can be useful when the struct definition is not under your control and you cannot add field tags to it. You can push validation on only one of the fields. In more complex situations, a custom `Validator` interface is needed.

```go
type Timestamp struct {
    time.Time
}
type Key struct {
    ID int
}
type Person struct {
    Name string
}
type MyData struct {
    // Require a Key with a non-zero ID
    Index   Key       `dv8:"notzero,on ID"`
    // Require a Timestamp with a non-zero Time 
    Expires Timestamp `dv8:"notzero,on Time"`
    // Set default Name of Person to "Unknown"
    Owner   Person    `dv8:"default=Unknown,on Name"`
}
```

The `delegate` directive is the mirror image of `on` and allows a struct to define a field on which to apply the validations that are set on the struct itself. It is useful when the struct is under your control and you can edit its field tags.

```go
type Timestamp struct {
    time.Time `dv8:"delegate"`
}
type Key struct {
    ID int `dv8:"delegate"`
}
type Person struct {
    Name string `dv8:"delegate"`
}
type MyData struct {
    // Require a Key with a non-zero ID
    Index   Key       `dv8:"notzero"`
    // Require a Timestamp with a non-zero Time 
    Expires Timestamp `dv8:"notzero"`
    // Set default Name of Person to "Unknown"
    Owner   Person    `dv8:"default=Unknown"`
}
```

## Arrays and maps

Directives set on an array or map apply to the array or map themselves.
To apply a directive to each of the elements of an array or the values of a map, prefix it with `each`.
To apply a directive to each of the keys of a map, prefix it with `key`.
A key mutated by a directive such as `key trim` or `key tolower` is reinserted under its new value;
two keys folding into the same mutated key is a validation error.
Prefixes compose for nested containers: `each each len>0` reaches the strings of a `[][]string`, and
`each key len>0` reaches the keys of the inner maps of a `[]map[string]int`.

```go
type Group struct {
    // The array must not be empty, and each of its (string) elements is limited in length
    Names []string `dv8:"len>0,each len>0,each len<=32"`
}
g := Group{
    Names: []string{"John", "Paul", ""},
}
err := dv8.Validate(ctx, &g)
if err != nil {
    return err // Names: [2]: length must be greater than 0
}
```

```go
type Directory struct {
    // Keys must be non-empty and each (string) value is limited in length
    Index map[string]string `dv8:"key len>0,each len>0,each len<=32"`
}
d := Directory{
    Index: map[string]string{
        "john": "John",
        "paul": "Paul",
        "geo":  "",
    },
}
err := dv8.Validate(ctx, &d)
if err != nil {
    return err // Index: [geo]: length must be greater than 0
}
```

## `Validator` interface

The `Validator` interface enables types to define custom validations.
`DV8` calls `Validate(ctx)` on any type in the object graph that implements the `Validator` interface
and considers any error received as a validation error.
The context is the one passed to `dv8.Validate`, or `context.Background` when `nil` was passed.

```go
type Validator interface {
    Validate(ctx context.Context) error
}
```

```go
type Rect struct {
    Top    int `dv8:"val>=0"`
    Left   int `dv8:"val>=0"`
    Right  int `dv8:"val>=0"`
    Bottom int `dv8:"val>=0"`
}
func (r *Rect) Validate(ctx context.Context) error {
    if r.Left >= r.Right {
        return errors.New("right must be greater than left")
    }
    if r.Top >= r.Bottom {
        return errors.New("bottom must be greater than top")
    }
    return nil
}
```

A parameterless `Validate() error` method is honored as a fallback, enabling interop with types that were
not written for `DV8`. Both variants share the method name `Validate`, so a type can implement at most one of them.

## Strict directives and `Compile`

Directives are compiled and strictly checked once per type, on first use. A typo (`requird`), a directive that
doesn't apply to the field's type (`trim` on an `int`), a malformed operator or value (`len*=2`, `val>abc`), an
uncompilable regular expression, or an `on` naming a nonexistent field all fail validation with an error wrapping
`dv8.ErrDirective`. Match it with `errors.Is` to tell a bug in the tags (the programmer's fault) apart from
invalid data (the caller's fault):

```go
err := dv8.Validate(ctx, &in)
if errors.Is(err, dv8.ErrDirective) {
    // a bug in the struct tags, not bad input
}
```

To surface broken directives before any data arrives, compile types eagerly at startup. `Compile` accepts
specimen values or `reflect.Type`s, recurses into nested types, and caches the result:

```go
err := dv8.Compile(CreateIn{}, UpdateIn{}, DeleteIn{})
if err != nil {
    log.Fatal(err)
}
```

## `DV8`, so your data doesn't!

The name `DV8` is a word play on both `D`ata `V`alid`ate` and `deviate`.

## Legal

`DV8` is released by `Microbus LLC` under the [Apache 2.0 license](http://www.apache.org/licenses/LICENSE-2.0).
