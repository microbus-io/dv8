# `DV8` - Data Validation for Golang

## Overview

`DV8` uses Golang's struct tags to validate data of struct fields.
Its primary purpose is validation of data entered by an untrusted source such as an end-user.
It draws inspiration from [Pydantic](https://docs.pydantic.dev).

```go
type Person struct {
    First   string `dv8:"required,len<=32"`
    Last    string `dv8:"required,len<=32"`
    Age     int    `dv8:"val>=0,val<=120"`
    State   string `dv8:"len==2,default=CA,toupper"`
    Zip     string `dv8:"required,regexp ^[0-9]{5}$"`
    Country string `dv8:"required,len==2,oneof US|MX,default=US,toupper"`
}

p := &Person{
    First:   " Julie",  // Trim whitespaces
    Last:    "Supercalifragilisticexpialidocious", // Enforce length limits
    State:   "",        // Set default to "CA"
    Age:     200,       // Enforce value constraints
    Zip:     "12x45",   // Enforce a regexp pattern
    Country: "USA",     // Check against a set of valid values
}

err := dv8.Validate(p)
if err != nil {
    return err
}
```

## Directives

`DV8` recognizes the following directives:

|Directive|Applicable types|Effect|
|---|---|---|
|`required`|`string`, `int`, `float`, `bool`, `time.Time`, `time.Duration`, `struct`|Requires a non-zero value to be provided|
|`required`|`*any`|Requires a non-`nil` value to be provided|
|`default`|`string`, `int`, `float`, `bool`, `time.Time`, `time.Duration`|Sets a default value when the zero-value is provided|
|`val` with `==` or `!=`|`string`, `int`, `float`, `bool`, `time.Time`, `time.Duration`|Enforces an equality constraint on the value|
|`val` with `<=`, `<`, `>=` or `>`|`string`, `int`, `float`, `time.Time`, `time.Duration`|Enforces an ordering constraint on the value|
|`len` with `==`, `!=`, `<=`, `<`, `>=` or `>`|`string`|Enforces a constraint on the length of the string (in runes, not bytes)
|`oneof`|`string`|Check against a set of valid values separated by a `\|`|
|`arrlen` with `==`, `!=`, `<=`, `<`, `>=` or `>`|`[]any`|Enforces a constraint on the length of the array. A `nil` array will fail the condition `arrlen>=0`|
|`maplen` with `==`, `!=`, `<=`, `<`, `>=` or `>`|`map[any]any`|Enforces a constraint on the length of the map. A `nil` map will fail the condition `maplen>=0`|
|`regexp`|`string`|Requires the string to match a regular expression|
|`on`|`struct`, `*struct`|Applies the directives on the named field of the struct instead of the struct itself (see below)|
|`main`|`any`|Applies the directives set on the parent struct to the field (see below)|
|`notrim`|`string`|Disables the default trimming of leading and trailing whitespaces|
|`tolower`|`string`|Transforms the string to lowercase|
|`toupper`|`string`|Transforms the string to uppercase|
|`-`|`any`|Skips the field and stops recursion into nested fields|

## `on` and `main`

The `on` directive allows pushing directives one level down into a nested struct. It can be useful when the nested struct is not under your control.

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
    Index   Key       `dv8:"required,on ID"`
    // Require a Timestamp with a non-zero Time 
    Expires Timestamp `dv8:"required,on Time"`
    // Set default Name of Person to "Unknown"
    Owner   Person    `dv8:"default=Unknown,on Name"`
}
```

The `main` directive is the mirror image of `on` and allows a struct to define a field on which to apply the validations that are set on the struct itself. It is useful when the struct is under your control and you can edit its field tags.

```go
type Timestamp struct {
    time.Time `dv8:"main"`
}
type Key struct {
    ID int `dv8:"main"`
}
type Person struct {
    Name string `dv8:"main"`
}
type MyData struct {
    // Require a Key with a non-zero ID
    Index   Key       `dv8:"required"`
    // Require a Timestamp with a non-zero Time 
    Expires Timestamp `dv8:"required"`
    // Set default Name of Person to "Unknown"
    Owner   Person    `dv8:"default=Unknown"`
}
```

## Arrays and maps

Except for the `arrlen` and `maplen` directives that apply to the array or map themselves, directives set
on an array or map apply to their value items.
Directives are not enforced on the key values of a map.

```go
type Group struct {
    // Enforced on each of the (string) value items of the array
    Names []string `dv8:"len>0,len<=32"`
}
g := Group{
    Names: []string{"John", "Paul", ""},
}
err := dv8.Validate(&g)
if err != nil {
    return err // Names: [2]: length must be greater than 0
}
```

```go
type Directory struct {
    // Enforced on each of the (string) value items of the map
    Index map[int]string `dv8:"len>0,len<=32"`
}
d := Directory{
    Index: map[int]string{
        0: "John",
        1: "Paul",
        2: "",
    },
}
err := dv8.Validate(&d)
if err != nil {
    return err // Index: [2]: length must be greater than 0
}
```

## `Validator` interface

The `Validator` interface enables types to define custom validations.
`DV8` calls `Validate()` on structs that implement the `Validator` interface and considers any error received as a validation error.

```go
type Validator interface {
    Validate() error
}
```

```go
type Rect struct {
    Top    int `dv8:"val>=0"`
    Left   int `dv8:"val>=0"`
    Right  int `dv8:"val>=0"`
    Bottom int `dv8:"val>=0"`
}
func (r *Rect) Validate() error {
    if r.Left >= r.Right {
        return errors.New("right must be greater than left")
    }
    if r.Top >= r.Bottom {
        return errors.New("bottom must be greater than top")
    }
}
```

## `DV8`, so your data doesn't!

The name `DV8` is a word play on both `D`ata `V`alid`ate` and `deviate`.

## Legal

`DV8` is released by `Microbus LLC` under the [Apache 2.0 license](http://www.apache.org/licenses/LICENSE-2.0).
