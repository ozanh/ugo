# uGO Tutorial

uGO is another script language for Go applications to make Go more dynamic with
scripts. uGO is inspired by script language [Tengo](https://github.com/d5/tengo)
by Daniel Kang.

uGO source code is compiled to bytecode and run in a Virtual Machine (VM).
Compiled bytecode can be serialized/deserialized for wire to remove compilation
step before execution for remote processes, and deserialization will solve
version differences to a certain degree.

Main script and uGO source modules in uGO are all functions which have
`compiledFunction` type name. Parameters can be defined for main function with
[`param`](#param) statement and main function returns a value with `return`
statement as well. If return statement is missing, `undefined` value is returned
by default. All functions return single value but thanks to
[destructuring](destructuring.md) feature uGO allows to return multiple values
as an array and set returning array elements to multiple variables.

uGO relies on Go's garbage collector and there is no allocation limit for
objects like Tengo. To run scripts which are not safe must be run in a sandboxed
environment which is the only way for Go applications. Note: Builtin objects can
be disabled before compilation.

uGO can handle runtime errors and Go panic with `try-catch-finally` statements
which is similar to Ecmascript implementation with a few minor differences.
Although Go developers are not fan of `try-catch-finally`, they are well known
statements to work with.

uGO does not use reflect package during execution and avoids unsafe function
calls. uGO objects most of the time escape to heap inevitably because they are
of interface types but it is minimized.

uGO is developed to be an embedded script language for Go applications, and
importing source modules from files will not be added in near future but one can
implement a custom module to return file content to the compiler.

uGO currently has a simple optimizer for constant folding and evaluating
expressions which do not have side effects. Optimizer greedily evaluates
expressions having only literals (int, uint, char, float, bool, string).
Optimizer can be disabled with compiler options.

## Run Scripts

To run a script, it must be compiled to create a `Bytecode` object then it is
provided to Virtual Machine (VM). uGO has a simple optimizer enabled by default
in the compiler. Optimizer evaluates simple expressions not having side effects
to replace expressions with constant values. Note that, optimizer can be
disabled to speed up compilation process.

```go
package main

import (
  "fmt"

  "github.com/ozanh/ugo"
)

func main() {
  script := `
  param num

  var fib
  fib = func(n, a, b) {
    if n == 0 {
      return a
    } else if n == 1 {
      return b
    }
    return fib(n-1, b, a+b)
    }
  return fib(num, 0, 1)
  `
  bytecode, err := ugo.Compile([]byte(script), ugo.DefaultCompilerOptions)

  if err != nil {
    panic(err)
  }

  retValue, err := ugo.NewVM(bytecode).Run(nil,  ugo.Int(35))

  if err != nil {
    panic(err)
  }

  fmt.Println(retValue) // 9227465
}
```

Script above is pretty self explanatory, which calculates the fibonacci number
of given number.

Compiler options hold all customizable options for the compiler.
`TraceCompilerOptions` is used to trace parse-optimize-compile steps for
debugging and testing purposes like below;

```go
bytecode, err := ugo.Compile([]byte(script), ugo.TraceCompilerOptions)
// or change output and disable tracing parser
// opts := ugo.TraceCompilerOptions
// opts.Trace = os.Stderr
// opts.TraceParser = false
// bytecode, err := ugo.Compile([]byte(script), opts)
```

VM execution can be aborted by using `Abort` method which cause `Run` method to
return an error wrapping `ErrVMAborted` error. `Abort` must be called from a
different goroutine and it is safe to call multiple times.

Errors returned from `Run` method can be checked for specific error values with
Go's `errors.Is` function in `errors` package.

`VM` instances are reusable. `Clear` method of `VM` clears all references held
and ensures stack and module cache is cleaned.

```go
vm := ugo.NewVM(bytecode)
retValue, err := vm.Run(nil,  ugo.Int(35))
/* ... */
// vm.Clear()
retValue, err := vm.Run(nil,  ugo.Int(34))
/* ... */
```

Global variables can be provided to VM which are declared with
[`global`](#global) keyword. Globals are accessible to source modules as well.
Map like objects should be used to get/set global variables as below.

```go
script := `
param num
global upperBound
return num > upperBound ? "big" : "small"
`
bytecode, err := ugo.Compile([]byte(script), ugo.DefaultCompilerOptions)

if err != nil {
  panic(err)
}

g := ugo.Map{"upperBound": ugo.Int(1984)}
retValue, err := ugo.NewVM(bytecode).Run(g,  ugo.Int(2018))
// retValue == ugo.String("big")
```

There is a special type `SyncMap` in uGO to make goroutine safe map object where
scripts/Go might need to interact with each other concurrently, e.g. one can
collect statistics or data within maps. Underlying map of `SyncMap` is guarded
with a `sync.RWMutex`.

```go
module := `
global stats

return func() {
  stats.fn2++
  /* ... */
}
`
script := `
global stats

fn1 := func() {
  stats.fn1++
  /* ... */
}

fn1()

fn2 := import("module")
fn2()
`
mm := ugo.NewModuleMap()
mm.AddSourceModule("module", []byte(module))

opts := ugo.DefaultCompilerOptions
opts.ModuleMap = mm

bytecode, err := ugo.Compile([]byte(script), opts)

if err != nil {
  panic(err)
}

g := &ugo.SyncMap{
    Map: ugo.Map{"stats": ugo.Map{"fn1": ugo.Int(0), "fn2": ugo.Int(0)}},
}
_, err = ugo.NewVM(bytecode).Run(g)
/* ... */
```

As can be seen from examples above, VM's `Run` method takes arguments and its
signature is as below. A map like `globals` argument or `nil` value can be
provided for globals parameter. `args` variadic parameter enables providing
arbitrary number of arguments to VM which are accessed via [`param`](#param)
statement.

```go
func (vm *VM) Run(globals Object, args ...Object) (Object, error)
```

## Variables Declaration and Scopes

### param

`param` keyword is used to declare a parameter for main function (main script).
Parenthesis is required for multiple declarations. Last argument can also be
variadic. Unlike `var` keyword, initializing value is illegal. Variadic
argument initialized as an empty array `[]`, and others are initialized as
`undefined` if not provided. `param` keyword can be used only once in main
function.

```go
param (arg0, arg1, ...vargs)
```

```go
param foo
param bar    // illegal, multiple param keyword is not allowed
```

```go
if condition  {
  param arg    // illegal, not allowed in this scope
}

func(){
    param (a, b)    // illegal, not allowed in this scope
}
```

### global

`global` keyword is to declare global variables. Note that `var` statements or
short variable declaration `:=` always creates local variables not global.
Parenthesis is required for multiple declarations. Unlike `var`, initializing
value is illegal. `global` statements can appear multiple times in the scripts.
`global` gives access to indexable `globals` argument with a variable name
provided to Virtual Machine (VM).

If `nil` is passed to VM as globals, a temporary `map` assigned to globals.

Any assignment to a global variable creates or updates the globals element.

Note that global variables can be accessed by imported source modules which
enables to export objects to scripts like `extern` in C.

```go
global foo
global (bar, baz)
```

```go
// "globals" builtin function returns "globals" provided to VM.
g := globals()
v := g["foo"]    // same as `global foo; v := foo`
```

```go
if condition {
  global x     // illegal, not allowed in this scope
}

func() {
  global y     // illegal, not allowed in this scope
}
```

### var

`var` keyword is used to declare a local variable. Parenthesis is required for
multiple declaration. Note: Tuple assignment is not supported with var
statements.

```go
var foo               // foo == undefined
var (bar, baz = 1)    // bar == undefined, baz == 1
var (bar,
     baz = 1)         // valid
var (
    foo = 1
    bar
    baz = "baz"
)                     // valid
```

A value can be assigned to a variable using short variable declaration `:=` and
assignment `=` operators.

* `:=` operator defines a new variable in the scope and assigns a value.
* `=` operator assigns a new value to an existing variable in the scope.

```go
                 // function scope A
a := "foo"       // define 'a' in local scope

func() {         // function scope B
  b := 52        // define 'b' in function scope B
  
  func() {       // function scope C
    c := 19.84   // define 'c' in function scope C

    a = "bee"    // ok: assign new value to 'a' from function scope A
    b = 20       // ok: assign new value to 'b' from function scope B

    b := true    // ok: define new 'b' in function scope C
                 //     (shadowing 'b' from function scope B)
  }
  
  a = "bar"      // ok: assign new value to 'a' from function scope A
  b = 10         // ok: assign new value to 'b'
  a := -100      // ok: define new 'a' in function scope B
                 //     (shadowing 'a' from function scope A)
  
  c = -9.1       // illegal: 'c' is not defined
  var b = [1, 2] // illegal: 'b' is already defined in the same scope
}

b = 25           // illegal: 'b' is not defined
var a = {d: 2}   // illegal: 'a' is already defined in the same scope
```

Following is illegal because variable is not defined when function is created.
In assignment statements right hand side is compiled before left hand side.

```go
f := func() {
  f()    // illegal: unresolved symbol "f"
}
```

```go
var f
f = func() {
  f()    // ok: "f" is declared before assignment.
}
```

Unlike Go, a variable can be assigned a value of different types.

```go
a := 123        // assigned    'int'
a = "123"       // reassigned 'string'
a = [1, 2, 3]   // reassigned 'array'
```

Capturing loop variables returns the last value of the variable set after last
post statement of the for loop, like Go.

```go
var f

for i := 0; i < 3; i++ {
  f = func(){
    return i
  }  
}

println(f())  // 3
```

Like Go, to capture the variable define a new variable using same name or
different.

```go
var f

for i := 0; i < 3; i++ {
  i := i
  f = func(){
    return i
  }  
}

println(f())  // 2
```

### const

`const` keyword is used to declare a local constant variable. Parenthesis is
required for multiple declaration. Note: Tuple assignment is not supported.
The value of a constant can't be changed through reassignment.
Reassignment is checked during compilation and an error is thrown.
An initializer for a constant is required while declaring. The const declaration
creates a read-only reference to a value. It does not mean the value it holds is
immutable.

```go
const (
  a = 1
  b = {foo: "bar"}
)

const c       // illegal, no initializer

a = 2         // illegal, reassignment
b.foo = "baz" // legal
```

`iota` is supported as well.

```go
const (
  x = iota
  y
  z
)
println(x, y, z) // 0 1 2
```

```go
const (
  x = 1<<iota
  y
  z
)
println(x, y, z) // 1 2 4
```

```go
const (
  _ = 1<<iota
  x
  y
  z
)
println(x, y, z) // 2 4 8
```

```go
const (
  x = 1+iota
  _
  z
)
println(x, z) // 1 3
```

```go
const (
  x = func() { return iota }() // illegal, compile error
)
```

```go
const (
  iota = 1 // illegal, compile error
)
```

RHS of the assignment can be any expression so `iota` can be used with them as well.

```go
const (
  x = [iota]
  y
)
println(x) // [0]
println(y) // [1]
```

```go
const (
  _ = iota
  x = "string" + iota
  y
)
println(x) // string1
println(y) // string2
```

**Warning:** if a variable named `iota` is created before `const` assignments,
`iota` is not used for enumeration and it is treated as normal variable.

```go
iota := "foo"

const (
  x = iota
  y
)
println(x) // foo
println(y) // foo
```

## Values and Value Types

In uGO, everything is a value, and, all values are associated with a type.

```go
19 + 84                 // int values
1u + 5u                 // uint values
"foo" + `bar`           // string values
-9.22 + 1e10            // float values
true || false           // bool values
'ç' > '9'               // char values
[1, false, "foo"]       // array value
{a: 12.34, "b": "bar"}  // map value
func() { /*...*/ }      // function value
```

Here's a list of all available value types in uGO. See [runtime
types](runtime-types.md) for more information.

| uGO Type          | Description                          | Equivalent Type in Go |
|:------------------|:-------------------------------------|:----------------------|
| int               | signed 64-bit integer value          | `int64`               |
| uint              | unsigned 64-bit integer value        | `uint64`              |
| float             | 64-bit floating point value          | `float64`             |
| bool              | boolean value                        | `bool`                |
| char              | unicode character                    | `rune`                |
| string            | unicode string                       | `string`              |
| bytes             | byte array                           | `[]byte`              |
| error             | [error](#error-values) value         | -                     |
| array             | value array                          | `[]Object`            |
| map               | value map with string keys           | `map[string]Object`   |
| undefined         | [undefined](#undefined-values) value | -                     |
| compiledFunction  | [function](#function-values) value   | -                     |

### Error Values

In uGO, an error can be represented using "error" typed values. An error value
is created using `error` builtin function, and, it has an underlying message.
The underlying message of an error can be accessed using `.Message` selector.
Error has also a name which is accessed using `.Name`. Errors created with
`error` builtin have default name `error` but builtin errors have different
names like `NotIterableError`, `ZeroDivisionError`.

First argument passed to `error` builtin function is converted to string as
message.

```go
err1 := error("oops")
err2 := error(1+2+3)         // equivalent to err2 := error("6")
if isError(err1) {           // 'isError' is a builtin function
  name := err1.Name          // get underlying name
  message := err1.Message    // get underlying message
}  
```

#### Builtin Errors

Builtin errors do not have message but have name. With `.New(message)` function
call on an error value creates a new error by wrapping the error.

Note: See [error handling](error-handling.md) for more information about errors.

* WrongNumArgumentsError
* InvalidOperatorError
* IndexOutOfBoundsError
* NotIterableError
* NotIndexableError
* NotIndexAssignableError
* NotCallableError
* NotImplementedError
* ZeroDivisionError
* TypeError

### Undefined Values

In uGO, an `undefined` value can be used to represent an unexpected or
non-existing value:

* A function that does not return a value explicitly considered to return
`undefined` value.
* Indexer or selector on composite value types may return `undefined` if the
key or index does not exist.  
* Builtin functions may return `undefined`.

```go
a := func() { b := 4 }()    // a == undefined
c := {a: "foo"}["b"]        // c == undefined
d := sort(undefined)        // d == undefined
e := delete({}, "foo")      // "delete" always returns undefined
```

Builtin function `isUndefined` or `==` operator can be used to check value is
undefined.

### Array Values

In uGO, array is an ordered list of values of any types. Elements of an array
can be accessed using indexer `[]`.

```go
[1, 2, 3][0]       // == 1
[1, 2, 3][2]       // == 3
[1, 2, 3][3]       // RuntimeError: IndexOutOfBoundsError

["foo", 'x', [1, 2, 3], {bar: 2u}, true, undefined, bytes()]   // ok
```

### Map Values

In uGO, map is a set of key-value pairs where key is string and the value is
of any value types. Value of a map can be accessed using indexer `[]` or
selector '.' operators.

```go
m := { a: 1, "b": false, c: "foo" }
m["b"]                                // == false
m.c                                   // == "foo"
m.x                                   // == undefined

{a: [1, 2, 3], b: {c: "foo", d: "bar"}} // ok
```  

### Function Values

In uGO, function is a callable value with a number of function arguments and
a return value. Just like any other values, functions can be passed into or
returned from another function.

```go
sum := func(arg1, arg2) {
  return arg1 + arg2
}

var mul = func(arg1, arg2) {
  return arg1 * arg2
}

adder := func(base) {
  return func(x) { return base + x }  // capturing 'base'
}
add5 := adder(5)
nine := add5(4)    // == 9
```

Unlike Go, uGO does not have function declarations. All functions are anonymous
functions. So the following code is illegal:

```go
func foo(arg1, arg2) {  // illegal
  return arg1 + arg2
}
```

uGO also supports variadic functions:

```go
variadic := func (a, b, ...c) {
  return [a, b, c]
}
variadic(1, 2, 3, 4) // [1, 2, [3, 4]]

variadicClosure := func(a) {
  return func(b, ...c) {
    return [a, b, c]
  }
}
variadicClosure(1)(2, 3, 4) // [1, 2, [3, 4]]
```

Only the last parameter can be variadic. The following code is illegal:

```go
// illegal, because "a" is variadic and is not the last parameter
illegal := func(...a, b) {}
```

When calling a function, the number of passing arguments must match that of
function definition.

```go
f := func(a, b) {}
f(1, 2, 3)    // RuntimeError: WrongNumArgumentsError
```

Like Go, you can use ellipsis `...` to pass value of array type as its last
parameter:

```go
f1 := func(a, b, c) { return a + b + c }
f1(...[1, 2, 3])    // => 6
f1(1, ...[2, 3])    // => 6
f1(1, 2, ...[3])    // => 6
f1(...[1, 2])       // RuntimeError: WrongNumArgumentsError

f2 := func(a, ...b) {}
f2(1)               // valid; a == 1, b == []
f2(1, 2)            // valid; a == 1, b == [2]
f2(1, 2, 3)         // valid; a == 1, b == [2, 3]
f2(...[1, 2, 3])    // valid; a == 1, b == [2, 3]
```

## Type Conversions

Although the type is not directly specified in uGO, one can use type conversion
[builtin functions](builtins.md) to convert between value types and see
[conversion/coersion table](runtime-types.md) for more information.

```go
s1 := string(1984)    // "1984"
i2 := int("-999")     // -999
f3 := float(-51)      // -51.0
b4 := bool(1)         // true
c5 := char("X")       // 'X'
```

See [Operators](operators.md) for more details on type conversions/coercions as
well.

## Order of evaluation

Expressions are evaluated from left to right but in assignments, right hand side
of the assignment is evaluated before left hand side.

```go
a := 1
f := func() {
  a*=10
  return a
}
g := func() {
  a++
  return a
}
h := func() {
  a+=2
  return a
}
d := {}
d[f()] = [g(), h()]
return d    // d == {"40": [2, 4]}
```

## Operators

### Unary Operators

| Operator | Operation               | Types(Results)                                            |
|:--------:|:-----------------------:|:---------------------------------------------------------:|
| `+`      | `0 + x`                 | int(int), uint(uint), char(char), float(float), bool(int) |
| `-`      | `0 - x`                 | int(int), uint(uint), char(int), float(float), bool(int)  |
| `^`      | bitwise complement `^x` | int(int), uint(uint), char(char), bool(int)               |
| `!`      | logical NOT             | all types*                                                |

_* In uGO, all values can be either
[truthy or falsy](runtime-types.md#objectisfalsy)._

### Binary Operators

| Operator | Usage                    |
|:--------:|:------------------------:|
| `==`     | equal                    |
| `!=`     | not equal                |
| `&&`     | logical AND              |
| `\|\|`   | logical OR               |
| `+`      | add/concat               |
| `-`      | subtract                 |
| `*`      | multiply                 |
| `/`      | divide                   |
| `&`      | bitwise AND              |
| `\|`     | bitwise OR               |
| `^`      | bitwise XOR              |
| `&^`     | bitclear (AND NOT)       |
| `<<`     | shift left               |
| `>>`     | shift right              |
| `<`      | less than                |
| `<=`     | less than or equal to    |
| `>`      | greater than             |
| `>=`     | greater than or equal to |

_See [Operators](operators.md) for more details._

### Ternary Operators

uGO has a ternary conditional operator
`(condition expression) ? (true expression) : (false expression)`.

```go
a := true ? 1 : -1    // a == 1

min := func(a, b) {
  return a < b ? a : b
}
b := min(5, 10)      // b == 5
```

### Assignment and Increment Operators

| Operator | Usage                     |
|:--------:|:-------------------------:|
| `+=`     | `(lhs) = (lhs) + (rhs)`   |
| `-=`     | `(lhs) = (lhs) - (rhs)`   |
| `*=`     | `(lhs) = (lhs) * (rhs)`   |
| `/=`     | `(lhs) = (lhs) / (rhs)`   |
| `%=`     | `(lhs) = (lhs) % (rhs)`   |
| `&=`     | `(lhs) = (lhs) & (rhs)`   |
| `\|=`    | `(lhs) = (lhs) \| (rhs)`  |
| `&^=`    | `(lhs) = (lhs) &^ (rhs)`  |
| `^=`     | `(lhs) = (lhs) ^ (rhs)`   |
| `<<=`    | `(lhs) = (lhs) << (rhs)`  |
| `>>=`    | `(lhs) = (lhs) >> (rhs)`  |
| `++`     | `(lhs) = (lhs) + 1`       |
| `--`     | `(lhs) = (lhs) - 1`       |

### Operator Precedences

Unary operators have the highest precedence, and, ternary operator has the
lowest precedence. There are five precedence levels for binary operators.
Multiplication operators bind strongest, followed by addition operators,
comparison operators, `&&` (logical AND), and finally `||` (logical OR):

| Precedence | Operator                             |
|:----------:|:------------------------------------:|
| 5          | `*`  `/`  `%`  `<<`  `>>`  `&`  `&^` |
| 4          | `+`  `-`  `\|`  `^`                  |
| 3          | `==`  `!=`  `<`  `<=`  `>`  `>=`     |
| 2          | `&&`                                 |
| 1          | `\|\|`                               |

Like Go, `++` and `--` operators form statements, not expressions, they fall
outside the operator hierarchy.

### Selector and Indexer

One can use selector (`.`) and indexer (`[]`) operators to read or write
elements of composite types (array, map, string, bytes).

```go
["one", "two", "three"][1]  // == "two"

bytes(0, 1, 2, 3)[1]    // == 1

// Like Go, indexing string returns byte value of index as int value.
"foobarbaz"[4]    // == 97

m := {
  a: 1,
  b: [2, 3, 4],
  c: func() { return 10 }
}
m.a              // == 1
m["b"][1]        // == 3
m.c()            // == 10
m.x.y.z          // == undefined
m.x.y.z = 1      // RuntimeError: NotIndexAssignableError
m.x = 5          // add 'x' to map 'm'
```

Like Go, one can use slice operator `[:]` for sequence value types such as
array, string, bytes. Negative indexes are illegal.

```go
a := [1, 2, 3, 4, 5][1:3]    // == [2, 3]
b := [1, 2, 3, 4, 5][3:]     // == [4, 5]
c := [1, 2, 3, 4, 5][:3]     // == [1, 2, 3]
d := "hello world"[2:10]     // == "llo worl"
e := [1, 2, 3, 4, 5][:]      // == [1, 2, 3, 4, 5]
f := [1, 2, 3, 4, 5][-1:]    // RuntimeError: InvalidIndexError
g := [1, 2, 3, 4, 5][10:]    // RuntimeError: IndexOutOfBoundsError
```

**Note: Keywords cannot be used as selectors.**

```go
a := {}
a.func = ""     // Parse Error: expected selector, found 'func'
```

Use double quotes and indexer to use keywords with maps.

```go
a := {}
a["func"] = ""
```

## Statements

### If Statement

"If" statement is very similar to Go.

```go
if a < 0 {
  // execute if 'a' is negative
} else if a == 0 {
  // execute if 'a' is zero
} else {
  // execute if 'a' is positive
}
```

Like Go, the condition expression may be preceded by a simple statement,
which executes before the expression is evaluated.

```go
if a := foo(); a < 0 {
  // execute if 'a' is negative
}
```

### For Statement

"For" statement is very similar to Go.

```go
// for (init); (condition); (post) {}
for a:=0; a<10; a++ {
  // ...
}

// for (condition) {}
for a < 10 {
  // ...
}

// for {}
for {
  // ...
}
```

### For-In Statement

It's similar to Go's `for range` statement.
"For-In" statement can iterate any iterable value types (array, map, bytes,
string).  

```go
for v in [1, 2, 3] {          // array: element
  // 'v' is array element value
}
for i, v in [1, 2, 3] {       // array: index and element
  // 'i' is index
  // 'v' is array element value
}
for k, v in {k1: 1, k2: 2} {  // map: key and value
  // 'k' is key
  // 'v' is map element value
}
for i, v in "foo" {           // array: index and element
  // 'i' is index
  // 'v' is char
}
```

## Modules

Module is the basic compilation unit in uGO. A module can import another module
using `import` expression. There 3 types of modules. Source modules, builtin
modules and custom modules. Source module is in the form uGO code. Builtin
module type is `map[string]Object`. Lastly, any value implementing Go
`Importable` interface can be a module. `Import` method must return a valid uGO
Object or `[]byte`. Source module is called like a compiled function and
returned value is stored for future use. Other module values are copied while
importing in VM if `Copier` interface is implemented.

```go
type Importable interface {
  Import(moduleName string) (interface{}, error)
}
```

```go
type Copier interface {
  Copy() Object
}
```

Main module:

```go
sum := import("sum")    // load a module
println(sum(10))        // module function
```

Source module as `sum`:

```go
base := 5

return func(x) {
  return x + base
}
```

In uGO, modules are very similar to functions.

* `import` expression loads the module code and execute it like a function.
* Module should return a value using `return` statement.
  * Module can return a value of any types: int, map, function, etc.
  * `return` in a module stops execution and return a value to the importing
    code.
  * If the module does not have any `return` statement, `import` expression
  simply returns `undefined`. _(Just like the function that has no `return`.)_  
* importing same module multiple times at different places or in different
  modules returns the same object so it preserves the state of imported object.
* Arguments cannot be provided to source modules while importing although it is
  allowed to use `param` statement in module.
* Modules can use `global` statements to access globally shared object.

## Comments

Like Go, uGO supports line comments (`//...`) and block comments
(`/* ... */`).

```go
/*
  multi-line block comments
*/

a := 5    // line comments
```

## Differences from Go

Unlike Go, uGO does not have the following:

* Imaginary values
* Structs
* Pointers
* Channels
* Goroutines
* Tuple assignment (uGO supports [destructuring](destructuring.md) array)
* Switch statement
* Goto statement
* Defer statement
* Panic and recover
* Type assertion

## Interfaces

uGO types implement `Object` interface. Any Go type implementing `Object`
interface can be provided to uGO VM.

### Object interface

```go

// Object represents an object in the VM.
type Object interface {
  // TypeName should return the name of the type.
  TypeName() string

  // String should return a string of the type's value.
  String() string

  // BinaryOp handles +,-,*,/,%,<<,>>,<=,>=,<,> operators.
  // Returned error stops VM execution if not handled with an error handler
  // and VM.Run returns the same error as wrapped.
  BinaryOp(tok token.Token, right Object) (Object, error)

  // IsFalsy returns true if value is falsy otherwise false.
  IsFalsy() bool

  // Equal checks equality of objects.
  Equal(right Object) bool

  // Call is called from VM if CanCall() returns true. Check the number of
  // arguments provided and their types in the method. Returned error stops VM
  // execution if not handled with an error handler and VM.Run returns the
  // same error as wrapped.
  Call(args ...Object) (Object, error)

  // CanCall returns true if type can be called with Call() method.
  // VM returns an error if one tries to call a noncallable object.
  CanCall() bool

  // Iterate should return an Iterator for the type.
  Iterate() Iterator

  // CanIterate should return whether the Object can be Iterated.
  CanIterate() bool

  // IndexGet should take an index Object and return a result Object or an
  // error for indexable objects. Indexable is an object that can take an
  // index and return an object. Returned error stops VM execution if not
  // handled with an error handler and VM.Run returns the same error as
  // wrapped. If Object is not indexable, ErrNotIndexable should be returned
  // as error.
  IndexGet(index Object) (value Object, err error)

  // IndexSet should take an index Object and a value Object for index
  // assignable objects. Index assignable is an object that can take an index
  // and a value on the left-hand side of the assignment statement. If Object
  // is not index assignable, ErrNotIndexAssignable should be returned as
  // error. Returned error stops VM execution if not handled with an error
  // handler and VM.Run returns the same error as wrapped.
  IndexSet(index, value Object) error
}
```

### Iterator interface

If an object's `CanIterate` method returns `true`, its `Iterate` method must
return a value implementing `Iterator` interface to use in `for-in` loops.

```go
// Iterator wraps the methods required to iterate Objects in VM.
type Iterator interface {
  // Next returns true if there are more elements to iterate.
  Next() bool

  // Key returns the key or index value of the current element.
  Key() Object

  // Value returns the value of the current element.
  Value() Object
}
```

### Copier interface

Assignments to uGO values copy the values except array, map or bytes like Go.
`copy` builtin function returns the copy of a value if Copier interface is
implemented by object. If not implemented, same object is returned which copies
the value under the hood by Go.

```go
// Copier wraps the Copy method to create a deep copy of an object.
type Copier interface {
  Copy() Object
}
```

### IndexDeleter interface

`delete` builtin checks if the given object implements `IndexDeleter` interface
to delete an element from the object. `map` and `syncMap` implement this
interface.

```go
// IndexDeleter wraps the IndexDelete method to delete an index of an object.
type IndexDeleter interface {
    IndexDelete(Object) error
}
```

### LengthGetter interface

`len` builtin checks if the given object implements `IndexDeleter` interface
to get the length of an object. `array`, `bytes`, `string`, `map` and `syncMap`
implement this interface.

```go
// LengthGetter wraps the Len method to get the number of elements of an object.
type LengthGetter interface {
    Len() int
}
```

### Object Interface Extensions

Note that `ExCallerObject` will replace the existing Object interface in the
future.

```go
// ExCallerObject is an interface for objects that can be called with CallEx
// method. It is an extended version of the Call method that can be used to
// call an object with a Call struct. Objects implementing this interface is
// called with CallEx method instead of Call method.
// Note that CanCall() should return true for objects implementing this
// interface.
type ExCallerObject interface {
    Object
    CallEx(c Call) (Object, error)
}

// NameCallerObject is an interface for objects that can be called with CallName
// method to call a method of an object. Objects implementing this interface can
// reduce allocations by not creating a callable object for each method call.
type NameCallerObject interface {
    Object
    CallName(name string, c Call) (Object, error)
}
```