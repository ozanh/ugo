# uGO Runtime Types

- **int**: signed 64bit integer (`int64` in Go)
- **uint**: unsigned 64bit integer (`uint64` in Go)
- **float**: 64bit floating point (`float64` in Go)
- **bool**: boolean (`bool` in Go)
- **char**: character (`rune` in Go)
- **string**: string (`string` in Go)
- **bytes**: byte array (`[]byte` in Go)
- **array**: objects array (`[]Object` in Go)
- **map**: objects map with string keys (`map[string]Object` in Go)
- **error**: an error with a string Name and Message
- **undefined**: undefined

Note: uGO does not have `byte` type. `uint`, `int` or `string` values can be
used to create/modify `bytes` values.

## Go Type Definitions

- `int`

```go
type Int int64
```

- `uint`

```go
type Uint uint64
```

Note: uint values can be represented by adding `u` suffix to integer values.

- `float`

```go
type Float float64
```

- `bool`

```go
type Bool bool
```

- `char`

```go
type Char rune
```

- `string`

```go
type String string
```

- `bytes`

```go
type Bytes []byte
```

- `error`

```go
type Error struct {
  Name    string
  Message string
  Cause   error
}
```

- `array`

```go
type Array []Object
```

- `map`

```go
type Map map[string]Object
```

- `sync-map`

```go
type SyncMap struct {
  mu sync.RWMutex
  Map
}
```

## Type Conversion/Coercion Table

|           |    int    |    uint   |    float   |    bool    |             char            |      string      |   bytes   | array |  map  |   error  | undefined |
|-----------|:---------:|:---------:|:----------:|:----------:|:---------------------------:|:----------------:|:---------:|:-----:|:-----:|:--------:|:---------:|
| int       |     -     | uint64(v) | float64(v) | !IsFalsy() |           rune(v)           |     _strconv_    |   **X**   | **X** | **X** | String() |   **X**   |
| uint      |  int64(v) |     -     | float64(v) | !IsFalsy() |           rune(v)           |     _strconv_    |   **X**   | **X** | **X** | String() |   **X**   |
| float     |  int64(v) | uint64(v) |      -     | !IsFalsy() |           rune(v)           |     _strconv_    |   **X**   | **X** | **X** | String() |   **X**   |
| bool      |   1 / 0   |   1 / 0   |  1.0 / 0.0 |      -     |            1 / 0            | "true" / "false" |   **X**   | **X** | **X** | String() |   **X**   |
| char      |  int64(v) | uint64(v) | float64(v) | !IsFalsy() |              -              |     string(v)    |   **X**   | **X** | **X** | String() |   **X**   |
| string    | _strconv_ | _strconv_ |  _strconv_ | !IsFalsy() | utf8. DecodeRuneInString(v) |         -        | []byte(v) | **X** | **X** | String() |   **X**   |
| bytes     |   **X**   |   **X**   |    **X**   | !IsFalsy() |            **X**            |     string(v)    |     -     | **X** | **X** | String() |   **X**   |
| array     |   **X**   |   **X**   |    **X**   | !IsFalsy() |            **X**            |     String()     |   **X**   |   -   | **X** | String() |   **X**   |
| map       |   **X**   |   **X**   |    **X**   | !IsFalsy() |            **X**            |     String()     |   **X**   | **X** |   -   | String() |   **X**   |
| error     |   **X**   |   **X**   |    **X**   |    **X**   |            **X**            |     String()     |   **X**   | **X** | **X** |     -    |   **X**   |
| undefined |   **X**   |   **X**   |    **X**   | !IsFalsy() |            **X**            |     String()     |   **X**   | **X** | **X** |   **X**  |     -     |

- **X**: No conversion. Conversion function will throw runtime error, TypeError.
- strconv: converted using Go's conversion functions from `strconv` package.
- IsFalsy(): use [Object.IsFalsy()](#objectisfalsy) function.
- String(): use `Object.String()` function.

## Object.IsFalsy()

`Object.IsFalsy()` interface method is used to determine if a given value
should evaluate to `false` (e.g. for condition expression of `if` statement).

- **int**: `v == 0`
- **uint**: `v == 0`
- **float**: `math.IsNaN(v)`
- **bool**: `!v`
- **char**: `v == 0`
- **string**: `len(v) == 0`
- **bytes**: `len(v) == 0`
- **array**: `len(v) == 0`
- **map**: `len(v) == 0`
- **error**: `true` _(error is always falsy)_
- **undefined**: `true` _(undefined is always falsy)_

_See [builtins](builtins.md) for conversion and type checking functions_
