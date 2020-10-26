# The uGO Language

uGO is a fast, dynamic scripting language to embed in Go applications.
uGO is compiled and executed as bytecode on stack-based VM that's written
in native Go.

uGO is inspired by awesome script language [Tengo](https://github.com/d5/tengo)
by Daniel Kang. Some modules/packages are modified versions of Tengo which are
depicted in the source files. A special thanks to Tengo's creater and
contributors.

Tengo's parser and compiler are ported to uGO but uGO has different runtime with
a similar Object interface.

**_uGO is currently in beta phase, use it at your own risk_**

**Fibonacci Example**

```go
param arg0

var fib

fib = func(x) {
    if x == 0 {
        return 0
    } else if x == 1 {
        return 1
    }
    return fib(x-1) + fib(x-2)
}
return fib(arg0)
```

## Features

* Written in native Go (no cgo).
* Error handling with `try-catch-finally`.
* Go like syntax with additions.

## Why uGO

`uGO` name comes from the initials of my daughter's, wife's and my name. It is
not related with Go.

I needed a faster embedded scripting language with runtime error handling for
distributed embedded database applications.

## Quick Start

`go get github.com/ozanh/ugo`

uGO has a REPL application to learn and test uGO language thanks to
`github.com/c-bata/go-prompt` library.

`go install github.com/ozanh/ugo/cmd/ugo`

`./ugo`

![repl-gif](https://github.com/ozanh/ugo/blob/main/docs/repl.gif)

This example is to show some features of uGO.

```go
package main

import (
    "fmt"

    "github.com/ozanh/ugo"
)

func main() {
    script := `
param ...args

mapEach := func(seq, fn) {

    if !isArray(seq) {
        return error("want array, got " + typeName(seq))
    }

    var out = []

    if sz := len(seq); sz > 0 {
        out = repeat([0], sz)
    } else {
        return out
    }

    try {
        for i, v in seq {
            out[i] = fn(v)
        }
    } catch err {
        println(err)
    } finally {
        return err == undefined ? out : err
    }
}

global multiplier

return mapEach(args, func(x) { return x*multiplier })
`

    bytecode, err := ugo.Compile([]byte(src), ugo.DefaultCompilerOptions)
    if err != nil {
        panic(err)
    }
    globals := ugo.Map{"multiplier": ugo.Int(2)}
    ret, err := ugo.NewVM(bytecode).Run(
        globals,
        ugo.Int(1), ugo.Int(2), ugo.Int(3), ugo.Int(4),
    )
    if err != nil {
        panic(err)
    }
    fmt.Println(ret)    // [2, 4, 6, 8]
}
```

## Roadmap

Note: Until stable version 1, there will be no major language and runtime
change.

Currently, there is no standard library although uGO supports modules but it
will be developed gradually.

More tests will be added.

Optimizer will be improved to handle constant propagation after introducing
`const` keyword.

Dead code elimination will be added.

## Documentation

* [Tutorial](https://github.com/ozanh/ugo/blob/main/docs/tutorial.md)
* [Runtime Types](https://github.com/ozanh/ugo/blob/main/docs/runtime-types.md)
* [Builtins](https://github.com/ozanh/ugo/blob/main/docs/builtins.md)
* [Operators](https://github.com/ozanh/ugo/blob/main/docs/operators.md)
* [Error Handling](https://github.com/ozanh/ugo/blob/main/docs/error-handling.md)

## LICENSE

uGO is licensed under the MIT License.

See [LICENSE](LICENSE) for the full license text.
