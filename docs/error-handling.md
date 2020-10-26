# Error Handling

uGO's `try-catch-finally` statements can handle all kinds of runtime errors even
runtime `panic` in Go. It is similar to Ecmascript's error handling statements.
Although `try-catch-finally` statements are not a must to handle all kinds of
errors, one often needs finalizers for cleanup operations while interacting with
Go code, which makes error handling very important for long running or
repetitive executions.

**Syntax**

```go
try {
    var x

    // braces are required to create blocks even if they are empty
    // catch statement can be elided
    try {} finally {}

} catch err {
    // no parenthesis is allowed around "err" variable
    // thrown error is assigned to "err" variable
    // "x" variable is accessible from catch block
} finally {
    // "x" variable is accessible from finally block
    // "err" variable is accessible from finally block
    try {} catch {
        // variable name can be elided after catch keyword
        // finally statement can be elided
    }
}
```

```go
// illegal; catch or finally statement is missing
try {}
```

```go
// illegal; catch must be above finally
try {

} finally {

} catch  {

}
```

```go
// illegal; catch or finally keyword must follow the
// previous closing brace without a newline or semicolon.
try {}
catch {}
```

```go
// illegal; missing try block
catch {} finally {}
```

Runtime errors stop Virtual Machine (VM) execution if they are not handled. Some
[Object interface](tutorial.md#interfaces) methods have `error` typed return
parameters which generates runtime error. uGO tries to minimize the runtime
errors in builtin functions and modules but dynamic nature of uGO requires
proper error handling. Runtime errors can also be generated with `throw`
statements in scripts.

List of Builtin Errors:

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

Error names are self explanatory. `.Name` selector of error values returns the
same name with builtin name `TypeError.Name == "TypeError"`. Errors are
immutable so fields of an error cannot be modified.

```go
// create error from a builtin error
err := TypeError.New("want X type, but got Y type")
err.Name == "TypeError"
err.Message == "want X type, but got Y type"
```

```go
// create a custom error
err := error("custom error message")
err.Name == "error"
err.Message == "custom error message"
```

Creating error from a builtin error using `.New` method enables to check error
types using `isError` [builtin function](builtins.md#iserror).

```go
var ErrNotAnInt = error("not an integer")

fn := func(x) {
    if !isInt(x) {
        throw ErrNotAnInt
        /* OR set a new message
        msg := sprintf("%s %s", typeName(x), ErrNotAnInt.Message)
        throw ErrNotAnInt.New(msg)
        */
    }
    return 10 / x
}

try {

   result := fn("x")

} catch myerr {

    if isError(myerr, ErrNotAnInt) {
        /* ... */
    } else if isError(myerr, ZeroDivisionError) {
        /* ... */
    }

} finally {
    if myerr != undefined {
        return -1
    }
    return result
}
```

## throw Statement

`throw <expression>` statement enables to generate runtime errors. If thrown
expression is an error, it is wrapped in a `RuntimeError` object and propagated
to upper levels until an error handler is found. If expression is not of error
type, expression's string value is used as a message to build a new error.

Errors can be returned as values from functions like Go but under some
circumstances using `throw` is inevitable.

## panic

To handle Go runtime `panic`, use VM's `SetRecover(true)`. One can also use
`SetRecover(true)` to get panic messages as a Go error from VM's `Run` method.

Stack overflow errors is not handled even if they are thrown in a `try` block,
because of zero stack size. Unhandled panic is propagated if it is not handled.

```go
bytecode, _ := ugo.Compile(script, options)
// handle error here

vm := ugo.NewVM(bytecode).SetRecover(true)
```

## Stack Trace

Generated runtime errors holds stack trace information which can be printed
using `%+v` format specifier.

```go
val, err := ugo.NewVM(bytecode).Run(nil)
if err != nil {
    fmt.Printf("%+v\n", err)
    /*
    e := err.(*ugo.RuntimeError)
    _ = e.StackTrace()
    */
}
```
