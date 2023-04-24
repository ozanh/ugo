# The uGO Language

[![Go Reference](https://pkg.go.dev/badge/github.com/ozanh/ugo.svg)](https://pkg.go.dev/github.com/ozanh/ugo)
[![Go Report Card](https://goreportcard.com/badge/github.com/ozanh/ugo)](https://goreportcard.com/report/github.com/ozanh/ugo)
[![uGO Test](https://github.com/ozanh/ugo/actions/workflows/workflow.yml/badge.svg)](https://github.com/ozanh/ugo/actions/workflows/workflow.yml)
[![uGO Dev Test](https://github.com/ozanh/ugodev/workflows/ugodev-test/badge.svg)](https://github.com/ozanh/ugodev/actions)
[![Maintainability](https://api.codeclimate.com/v1/badges/a358e050217385db8002/maintainability)](https://codeclimate.com/github/ozanh/ugo/maintainability)

uGO is a fast, dynamic scripting language to embed in Go applications.
uGO is compiled and executed as bytecode on stack-based VM that's written
in native Go.

uGO is actively used in production to evaluate Sigma Rules' conditions, and to
perform compromise assessment dynamically.

To see how fast uGO is, please have a look at fibonacci
[benchmarks](https://github.com/ozanh/ugobenchfib) (not updated frequently).

> Play with uGO via [Playground](https://play.verigraf.com) built for
> WebAssembly.

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
* Supports Go 1.15 and above.
* `if else` statements.
* `for` and `for in` statements.
* `try catch finally` statements.
* `param`, `global`, `var` and `const` declarations.
* Rich builtins.
* Pure uGO and Go Module support.
* Go like syntax with additions.
* Call uGO functions from Go.
* Import uGO modules from any source (file system, HTTP, etc.).
* Create wrapper functions for Go functions using code generation.

## Why uGO

`uGO` name comes from the initials of my daughter's, wife's and my name. It is
not related with Go.

I needed a faster embedded scripting language with runtime error handling.

## Quick Start

`go get github.com/ozanh/ugo@latest`

uGO has a REPL application to learn and test uGO scripts.

`go install github.com/ozanh/ugo/cmd/ugo@latest`

`./ugo`

![repl-gif](https://github.com/ozanh/ugo/blob/main/docs/repl.gif)

This example is to show some features of uGO.

<https://play.golang.org/p/1Tj6joRmLiX>

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
        return out, err
    }
}

global multiplier

v, err := mapEach(args, func(x) { return x*multiplier })
if err != undefined {
    return err
}
return v
`

    bytecode, err := ugo.Compile([]byte(script), ugo.DefaultCompilerOptions)
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
    fmt.Println(ret) // [2, 4, 6, 8]
}
```

## Roadmap

Examples for best practices (2023).

Better Playground (2023).

More standard library modules (2023).

Configurable Stdin, Stdout and Stderr per Virtual Machine (2023).

Deferring function calls (2024).

Concurrency support (2024).

## Documentation

* [Tutorial](https://github.com/ozanh/ugo/blob/main/docs/tutorial.md)
* [Runtime Types](https://github.com/ozanh/ugo/blob/main/docs/runtime-types.md)
* [Builtins](https://github.com/ozanh/ugo/blob/main/docs/builtins.md)
* [Operators](https://github.com/ozanh/ugo/blob/main/docs/operators.md)
* [Error Handling](https://github.com/ozanh/ugo/blob/main/docs/error-handling.md)
* [Standard Library](https://github.com/ozanh/ugo/blob/main/docs/stdlib.md)
* [Optimizer](https://github.com/ozanh/ugo/blob/main/docs/optimizer.md)
* [Destructuring](https://github.com/ozanh/ugo/blob/main/docs/destructuring.md)

## LICENSE

uGO is licensed under the MIT License.

See [LICENSE](LICENSE) for the full license text.

## Acknowledgements

uGO is inspired by script language [Tengo](https://github.com/d5/tengo)
by Daniel Kang. A special thanks to Tengo's creater and contributors.
