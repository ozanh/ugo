# Standard Library

## Module List

* [time](stdlib-time.md) module at `github.com/ozanh/ugo/stdlib/time`

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
    "github.com/ozanh/ugo/stdlib/time"
)

func main() {
    script := `
    time := import("time")
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
    `
    moduleMap := ugo.NewModuleMap()
    moduleMap.AddBuiltinModule("time", time.Module)
    opts := ugo.DefaultCompilerOptions
    opts.ModuleMap = moduleMap
    bc, err := ugo.Compile([]byte(script), opts)
    ret, err := ugo.NewVM(bc).Run(nil)
    /* ... */
}
```
