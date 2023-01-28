# Builtin Objects

## Functions

### append

Appends object(s) to an array or bytes (first argument) and returns a new array
or bytes object. (Like Go's `append` builtin)

**Syntax**

> `append(arrayLike [, ...args])`

**Parameters**

- > `arrayLike`: array or bytes types.
- > `args`: any valid object for array, int|uint|char value for bytes.

**Return Value**

> A new array or bytes.

**Runtime Errors**

- > `WrongNumArgumentsError`
- > `TypeError`

**Examples**

```go
// append to an existing array
v := [1]
v = append(v, 2, 3)    // v == [1, 2, 3]

// append to bytes
b := bytes()
b = append(b, 0, 1, 2)    // b == [0 1 2] Go equivalent []byte{0, 1, 2}

// if first argument is undefined, an array is created and
// rest of arguments is appended to the new array
c := append(undefined, "a", "b", 'c')    // c == ["a", "b", 'c']
```

---

### delete

Deletes the element with the specified key from an object type. First argument
should implement `IndexDeleter` interface and second argument is converted to
string to delete specified string index. `map` and `syncMap` types implement
`IndexDeleter` interface. `delete` returns `undefined` value if successful and
it mutates given object.

**Syntax**

> `delete(object, key)`

**Parameters**

- > `object`: map, syncMap or object implementing `IndexDeleter` to delete given
  > key from.
- > `key`: String value of the key will be used as index.

**Return Value**

> `undefined`

**Runtime Errors**

- > `WrongNumArgumentsError`
- > `TypeError`

**Examples**

```go
v := {key: "value"}
delete(v, "key") // v == {}
```

```go
v := {key: "value"}
delete(v, "missing") // v == {"key": "value"}
```

---

### copy

Creates a copy of the given variable. `copy` function calls `Copy() Object`
method if implemented, which is expected to return a deep-copy of the value it
holds. int, uint, char, float, string, bool types do not implement a
[`Copier`](tutorial.md#interfaces) interface which wraps `Copy() Object` method.
Assignment is sufficient to copy these types. array, bytes, map, syncMap can be
deeply copied with `copy` builtin function.

**Syntax**

> `copy(object)`

**Parameters**

- > `object`: any object

**Return Value**

> deep copy of the given object if `Copier` interface is implemented otherwise
> given value is returned.

**Runtime Errors**

- > `WrongNumArgumentsError`

**Examples**

```go
v1 := [1, 2, 3]
v2 := v1
v3 := copy(v1)
v1[1] = 0
println(v2[1]) // "0"; 'v1' and 'v2' referencing the same array
println(v3[1]) // "2"; 'v3' not affected by 'v1'
```

---

### repeat

Creates new array, string or bytes from given array, string or bytes by
repeating input "count" times.

**Syntax**

> `repeat(sequence, count)`

**Parameters**

- > `sequence`: array, string or bytes
- > `count`: count of repeat as non-negative int or uint value

**Return Value**

> array, string or bytes

**Runtime Errors**

- > `WrongNumArgumentsError`
- > `TypeError`

**Examples**

```go
v1 := repeat([1, 2], 2)      // v1 == [1, 2, 1, 2]
v2 := repeat("abc", 2)       // v2 == "abcabc"
v3 := repeat(bytes(0), 1024) // bytes with 1024 elements all zero

// if count is zero, zero sized object is returned
v4 := repeat([1, 2], 0)        // v4 == []
v5 := repeat("abc", 0)         // v5 == ""
v6 := repeat(bytes("abc"), 0)  // v6 == bytes()
```

---

### contains

Reports whether given element is in object.

**Syntax**

- > `contains(object, element)`

**Parameters**

- > `object`: valid types are following
  - array
  - string
  - bytes
  - map
  - syncMap
  - undefined: contains returns false if object value is undefined
- > `element`:
  - if object's type is array, element can be of any type and sequential search
    is applied in array.
  - if object's type is string, element can be of any type. element's string
    representation is searched as substring.
  - if object's type is bytes, element can be of int, uint, char, string or
    bytes type.
  - if object's type is map or syncMap, element's string representation is
    looked up in the map's keys.
  - if object value is undefined, element is ignored.

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`
- > `TypeError`

**Examples**

```go
v := contains([1, 2, "a"], "b")            // v == false
v = contains([1, 2, "a"], "a")             // v == true
v = contains("abc3", "d")                  // v == false
v = contains("abc3", "a")                  // v == true
v = contains("abc3", 3)                    // v == true
v = contains(bytes(0, 1, 2), 3)            // v == false
v = contains(bytes(0, 1, 2), 0)            // v == true
v = contains(bytes(0, 1, 2), bytes(0, 1))  // v == true
v = contains({foo: "bar"}, "baz")          // v == false
v = contains({foo: "bar"}, "foo")          // v == true
```

---

### len

Returns the number of elements if the given variable implements `LengthGetter`
interface. `array`, `string`, `bytes`, `map`, `syncMap` implements
`LengthGetter`. If object doesn't implement `LengthGetter`, it returns 0.
Note that, `len` returns byte count of string values.

**Syntax**

> `len(object)`

**Parameters**

- > `object`: valid types are following
  - array
  - string
  - bytes
  - map
  - syncMap
  - other types implementing `LengthGetter`

**Return Value**

> int value

**Runtime Errors**

- > `WrongNumArgumentsError`

**Examples**

```go
v := len([1, 2, []])          // v == 3
v = len("x")                  // v == 1
v = len({foo: "", bar: ""})   // v == 2
v = len(bytes("xyz"))         // v == 3
v = len(1)                    // v == 0
```

---

### cap

Returns the capacity of an array or bytes type. It always returns 0 for other
types.

**Syntax**

> `cap(object)`

**Parameters**

- > `object`: valid types are following
  - array
  - bytes

**Return Value**

> int value

**Runtime Errors**

- > `WrongNumArgumentsError`

**Examples**

```go
v := cap([1, 2])
v = cap(bytes("abc"))
```

---

### sort

Returns sorted object in ascending order. Given object is modified if it is not
a string. Note that, string value is converted to Go rune slice before sort and
sorted rune slice is converted back to string.

**Syntax**

> `sort(object)`

**Parameters**

- > `object`: valid types are following
  - array
  - string
  - bytes
  - undefined

**Return Value**

> sorted object

**Runtime Errors**

- > `WrongNumArgumentsError`
- > `TypeError`

**Examples**

```go
v1 := [3u, 2.0, 1]
v2 := sort(v1)        // v1 == v2, v1 == [1, 2.0, 3u]

v3 := "zyx"
v4 := sort(v3)        // v4 == "xyz", v3 == "zyx"

v5 := bytes("cba")
v6 := sort(v5)        // v5 == v6, v5 == bytes("abc")

v7 := sort(undefined) // v7 == undefined

// if array elements are not comparable, a runtime error is thrown.
sort(["a", 1])        // RuntimeError: TypeError
```

---

### sortReverse

Returns sorted object in descending order. Given object is modified if it is not
a string. Note that, string value is converted to Go rune slice before sort and
sorted rune slice is converted back to string.

**Syntax**

> `sortReverse(object)`

**Parameters**

- > `object`: valid types are following
  - array
  - string
  - bytes
  - undefined

**Return Value**

> sorted object

**Runtime Errors**

- > `WrongNumArgumentsError`
- > `TypeError`

**Examples**

```go
v1 := [1, 2.0, 3]
v2 := sortReverse(v1)        // v1 == v2, v1 == [3u, 2.0, 1]

v3 := "xyz"
v4 := sortReverse(v3)        // v4 == "zyx", v3 == "xyz"

v5 := bytes("abc")
v6 := sortReverse(v5)        // v5 == v6, v5 == bytes("cba")

v7 := sortReverse(undefined) // v7 == undefined

// if array elements are not comparable, a runtime error is thrown.
sortReverse(["a", 1])        // RuntimeError: TypeError
```

---

### error

Returns a new [error value](tutorial.md#error-values). Given object's string
representation is used as Message of the new error.

**Syntax**

> `error(object)`

**Parameters**

- > `object`: any type

**Return Value**

> error value

**Runtime Errors**

- > `WrongNumArgumentsError`

```go
err := error("foo error")    // err.Message == "foo error"
```

---

### typeName

Returns the type name of given object. Note that, it calls `TypeName` method of
the object.

**Syntax**

> `typeName(object)`

**Parameters**

- > `object`: any type

**Return Value**

> type name as string value

**Runtime Errors**

- > `WrongNumArgumentsError`

**Examples**

```go
v1 := typeName(1)          // v1 == "int"
v2 := typeName("str")      // v2 == "string"
v3 := typeName([1, 2, 3])  // v3 == "array"
```

---

### bool

Converts the given object to a bool value and returns it. It calls `IsFalsy`
method of the object under the hood. Note that, float value is falsy if its
value is NaN.

**Syntax**

> `bool(object)`

**Parameters**

- > `object`: any type

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

**Examples**

```go
v1 := bool(1)          // v1 == true
v2 := bool(0)          // v2 == false
v3 := bool("str")      // v3 == true
v4 := bool("")         // v4 == false
v5 := bool([1, 2, 3])  // v5 == true
v6 := bool([])         // v5 == false
```

---

### int

Tries to convert the given object to an int value and returns it. Note that,
`int` type is derived from Go's int64 type, see numeric conversions in [Go
spec.](https://golang.org/ref/spec#Conversions) and conversion relies on Go's
"wrap around". See Go's `strconv.ParseInt` function for more information about
string conversion.

**Syntax**

> `int(object)`

**Parameters**

- > `object`: valid types are following
  - string
  - uint
  - float
  - char
  - bool
  - int

**Return Value**

> int value

**Runtime Errors**

- > `WrongNumArgumentsError`
- > `TypeError`
- > Unspecified errors from Go standard library

**Examples**

```go
v1 := int("123")          // v1 == 123
v2 := int("-12")          // v2 == -12
v3 := int("0x10")         // v3 == 16
v4 := int("0b101")        // v4 == 5
v5 := int(1u)             // v5 == 1
v6 := int(1.1)            // v6 == 1
v7 := int('a')            // v7 == 97
v8 := int(true)           // v8 == 1
v9 := int(false)          // v9 == 0
```

---

### uint

Tries to convert the given object to an uint value and returns it. Note that,
`uint` type is derived from Go's uint64 type, see numeric conversions in [Go
spec.](https://golang.org/ref/spec#Conversions). See Go's `strconv.ParseUint`
function for more information about string conversion and conversion relies on
Go's "wrap around".

**Syntax**

> `uint(object)`

**Parameters**

- > `object`: valid types are following
  - string
  - int
  - float
  - char
  - bool
  - uint

**Return Value**

> uint value

**Runtime Errors**

- > `WrongNumArgumentsError`
- > `TypeError`
- > Unspecified errors from Go standard library

**Examples**

```go
v1 := uint("123")          // v1 == 123u
v3 := uint("0x10")         // v3 == 16u
v4 := uint("0b101")        // v4 == 5u
v5 := uint(1)              // v5 == 1u
v6 := uint(1.1)            // v6 == 1u
v7 := uint('a')            // v7 == 97u
v8 := uint(true)           // v8 == 1u
v9 := uint(false)          // v9 == 0u
```

---

### char

Tries to convert the given object to a char value and returns it. Note that, if
string object is provided and encoding is invalid or string is empty, undefined
is returned. Note that, `char` type is derived from Go's rune type, see numeric
conversions in [Go spec.](https://golang.org/ref/spec#Conversions) and
conversion relies on Go's "wrap around".

**Syntax**

> `char(object)`

**Parameters**

- > `object`: valid types are following
  - string
  - int
  - uint
  - float
  - bool
  - char

**Return Value**

> char value / undefined

**Runtime Errors**

- > `WrongNumArgumentsError`
- > `TypeError`

**Examples**

```go
v1 := char("abc")          // v1 == 'a'
v2 := char(1)              // v2 == '\x01'
v3 := char(1u)             // v3 == '\x01'
v4 := char(1.1)            // v4 == '\x01'
v5 := char(true)           // v5 == '\x01'
v6 := char(false)          // v6 == '\x00'
v7 := char("")             // v7 == undefined
```

---

### float

Tries to convert the given object to a float value and returns it. Note that,
`float` type is derived from Go's float64 type, see numeric conversions in [Go
spec.](https://golang.org/ref/spec#Conversions). See Go's `strconv.ParseFloat`
function for more information about string conversion.

**Syntax**

> `float(object)`

**Parameters**

- > `object`: valid types are following
  - string
  - int
  - uint
  - bool
  - char
  - float

**Return Value**

> float value

**Runtime Errors**

- > `WrongNumArgumentsError`
- > `TypeError`
- > Unspecified errors from Go standard library

**Examples**

```go
v1 := float("1.1")        // v1 == 1.1
v2 := float(5)            // v2 == 5.0
v3 := float(true)         // v3 == 1.0
v4 := float(false)        // v4 == 0.0
```

---

### string

Converts the given object to a string value and returns it. It calls `String`
method of the object under the hood. Note that, map or syncMap types are
derived from Go's map type which has randomized iteration. This may cause
different results.

**Syntax**

> `string(object)`

**Parameters**

- > `object`: any object

**Return Value**

> string value

**Runtime Errors**

- > `WrongNumArgumentsError`

**Examples**

```go
v1 := string([1, 2])     // v1 == "[1, 2]"
v2 := string(12)         // v2 == "12"
v3 := string('a')        // v3 == "a"
v4 := string(1.0)        // v4 == "1"
v5 := string(1.1)        // v5 == "1.1"
v6 := string(undefined)  // v6 == "undefined"
v7 := string(true)       // v7 == "true"
```

---

### bytes

Returns a bytes value from given value(s). If argument is not provided, an empty
bytes value is returned. Note that numeric value conversion relies on Go's "wrap
around".

**Syntax**

> `bytes(...args)`

**Parameters**

- > `args`: valid types are following
  - string
  - bytes
  - int
  - uint
  - char

**Return Value**

> bytes value

**Runtime Errors**

- > `TypeError`

**Examples**

```go
v1 := bytes()              // v1 == empty bytes object
v2 := bytes(0, 1, 2u)      // v2 == bytes [0 1 2]
v3 := bytes('a')           // v3 == bytes [97]
v4 := bytes("abc")         // v4 == bytes [97 98 99]
v5 := bytes(256, 257)      // v5 == bytes [0 1] wrapped around
```

---

### chars

Returns an array containing chars of given string or bytes. If given
string/bytes has invalid encoding (incorrect UTF-8), it returns undefined.

**Syntax**

> `chars(object)`

**Parameters**

- > `object`: string or bytes value

**Return Value**

> array of char values or undefined

**Runtime Errors**

- > `WrongNumArgumentsError`
- > `TypeError`

**Examples**

```go
v1 := chars("abc")              // v1 == ['a', 'b', 'c']
v2 := chars(bytes(0, 1, 2))     // v2 == ['\x00', '\x01', '\x02']
v3 := chars("a\xc5")            // v3 == undefined, incorrect UTF-8
```

---

### printf

Writes the given format and arguments to default writer, which is stdout. Note
that, default writer can be updated. It calls Go's `fmt.Fprintf` function after
converting first argument to a string value and optional arguments to
`interface{}`.

**Syntax**

> `printf(format, ...args)`

**Parameters**

- > `format`: any object
- > `args`: any object

**Return Value**

> undefined

**Runtime Errors**

- > `WrongNumArgumentsError`
- > Unspecified write errors

**Examples**

```go
printf("%s%d%v", 'a', 5, [1, 2])    // a5[1, 2]
```

---

### println

Writes the given arguments to default writer, which is stdout, with a newline.
Note that, default writer can be updated. It calls Go's `fmt.Fprintln` function
after converting arguments to `interface{}`.

**Syntax**

> `println(...args)`

**Parameters**

- > `args`: any object

**Return Value**

> undefined

**Runtime Errors**

- > Unspecified write errors

**Examples**

```go
println()                  // \n
println('a', 5, [1, 2])    // a 5 [1, 2]\n
```

---

### sprintf

Formats according to a format specifier and returns the resulting string. It
calls Go's `fmt.Sprintf` function after converting first argument to a string
value and optional arguments to `interface{}`.

**Syntax**

> `sprintf(format, ...args)`

**Parameters**

- > `format`: any object
- > `args`: any object

**Return Value**

> string value

**Runtime Errors**

- > `WrongNumArgumentsError`

**Examples**

```go
v1 := sprintf("%s%d", "x", 5)    // v1 == "x5"
v2 := sprintf("test")            // v2 == "test"
```

---

### isError

Reports whether given value is of error type. Optionally if second argument is
provided, reports whether the error's cause is the provided error.

**Syntax**

> `isError(errorValue [, cause])`

**Parameters**

- > `errorValue`: an error value
- > `cause`: an error value

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

**Examples**

```go
v1 := isError(error("foo error: bar"))               // v1 == true
v2 := isError(error("foo error: bar"), TypeError)    // v2 == false
v3 := isError(TypeError.New("foo"), TypeError)       // v3 == true

try {
    1 / 0
} catch err {
    v4 := isError(err)                       // v4 == true
    v5 := isError(err, ZeroDivisionError)    // v5 == true
}
```

---

### isInt

Reports whether given object is of int type.

**Syntax**

> `isInt(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

---

### isUint

Reports whether given object is of uint type.

**Syntax**

> `isUint(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

---

### isFloat

Reports whether given object is of float type.

**Syntax**

> `isFloat(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

---

### isChar

Reports whether given object is of char type.

**Syntax**

> `isChar(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

---

### isBool

Reports whether given object is of bool type.

**Syntax**

> `isBool(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

---

### isString

Reports whether given object is of string type.

**Syntax**

> `isString(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

---

### isBytes

Reports whether given object is of bytes type.

**Syntax**

> `isBytes(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

---

### isMap

Reports whether given object is of map type.

**Syntax**

> `isMap(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

---

### isSyncMap

Reports whether given object is of syncMap type.

**Syntax**

> `isSyncMap(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

---

### isArray

Reports whether given object is of array type.

**Syntax**

> `isArray(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

---

### isUndefined

Reports whether given object value is undefined.

**Syntax**

> `isUndefined(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

---

### isFunction

Reports whether given object is of function, compiledFunction or
builtinFunction type.

**Syntax**

> `isFunction(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

---

### isCallable

Reports whether given object is a callable object. It reports objects `CanCall`
method result.

**Syntax**

> `isCallable(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`

---

### isIterable

Reports whether given object is an iterable object. It reports objects
`CanIterate` method result.

**Syntax**

> `isIterable(object)`

**Parameters**

- > `object`: any object

**Return Value**

> bool value

**Runtime Errors**

- > `WrongNumArgumentsError`
