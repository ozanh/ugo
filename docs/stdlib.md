# Standard Library

## Module List

* [fmt](stdlib-fmt.md) module at `github.com/ozanh/ugo/stdlib/fmt`
* [strings](stdlib-strings.md) module at `github.com/ozanh/ugo/stdlib/strings`
* [time](stdlib-time.md) module at `github.com/ozanh/ugo/stdlib/time`
* [json](stdlib-json.md) module at `github.com/ozanh/ugo/stdlib/json`

## How-To

### Import Module

Each standard library module is imported separately. `Module` variable as
`map[string]Object` in modules holds module values to pass to module map which
is deeply copied then.

**Example**

```go
package main

import (
    "github.com/ozanh/ugo"
    "github.com/ozanh/ugo/stdlib/fmt"
    "github.com/ozanh/ugo/stdlib/json"
    "github.com/ozanh/ugo/stdlib/strings"
    "github.com/ozanh/ugo/stdlib/time"
)

func main() {
    script := `
    const fmt = import("fmt")
    const strings = import("strings")
    const time = import("time")
    const json = import("json")

    total := 0
    fn := func() {
        start := time.Now()
        try {
            /* ... */
        } finally {
            total += time.Since(start)
        }
    }
    fn()
    /* ... */
    `
    moduleMap := ugo.NewModuleMap()
    moduleMap.AddBuiltinModule("fmt", fmt.Module)
    moduleMap.AddBuiltinModule("strings", strings.Module)
    moduleMap.AddBuiltinModule("time", time.Module)
    moduleMap.AddBuiltinModule("json", json.Module)

    opts := ugo.DefaultCompilerOptions
    opts.ModuleMap = moduleMap

    byteCode, err := ugo.Compile([]byte(script), opts)
    ret, err := ugo.NewVM(byteCode).Run(nil)
    /* ... */
}
```
