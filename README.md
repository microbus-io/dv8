# DV8 - Data Validation for Golang

`DV8` uses Golang's struct tags to validate data of struct fields.
Its primary purpose is validation of data entered by an untrusted source such as an end-user.
It draws inspiration from [Pydantic](https://docs.pydantic.dev).

```go
type Person struct {
    First   string `dv8:"required,len<=32"`
    Last    string `dv8:"required,len<=32"`
    Age     int    `dv8:"val>=18,val<=120"`
    State   string `dv8:"len==2,default=CA"`
    Zip     string `dv8:"required,regexp ^[0-9]{5}$"`
}
```

`DV8` recognizes the following directives:
* `required` indicates that a non-zero value must be provided
* `default` sets a default value when the zero-value is provided
* `val` enforces a constraint on the value using any of the operators `==`, `!=`, `<=`, `<`, `>=` or `>`
* `len` enforces a minimum or maximum length using any of the operators `==`, `!=`, `<=`, `<`, `>=` or `>`. It can be applied to a `string`, array or map. The length of a string is its length in runes, not bytes
* `regexp` enforces a regular expression on a `string`
* `notrim` disables the default trimming of leading and trailing white-spaces of a `string`
* `-` skips the field and stops recursion into nested fields 

`DV8` recognizes the following types:
* `string`
* `int`, `int8`, `int16`, `int32`, `int64`
* `uint`, `uint8`, `uint16`, `uint32`, `uint64`
* `float32`, `float64`
* `bool`
* `time.Time`
* `time.Duration`
* Array `[]any` or map `map[]any` of any of the above

|Directive|Applicable types|
|---|---|
|`required`|`string`, `int`, `float`, `bool`, `Time`, `Duration`, `map[]any`, `[]any`|
|`default`|`string`, `int`, `float`, `bool`, `Time`, `Duration`|
|`val` with `<=`, `<`, `>=` or `>`|`string`, `int`, `float`, `Time`, `Duration`|
|`val` with `==` or `!=`|`string`, `int`, `float`, `bool`, `Time`, `Duration`|
|`len`|`string`, `map[]any`, `[]any`|
|`regexp`|`string`|
|`notrim`|`string`|

The name `DV8` is a word play on both `D`ata `V`alid`ate` and `deviate`.

`DV8` is released by `Microbus LLC` under the [Apache 2.0 license](http://www.apache.org/licenses/LICENSE-2.0).
