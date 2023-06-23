package main

import (
	"fmt"

	"github.com/ozanh/ugo"
)

func main() {
	script := `
param ...args
global fn
mapEach := func(a, e, ...c;t=(1+5),b=56) {
	return t+b +1
}
x := mapEach(1,...[2,3,4];b=3,...{t:2})
return x
`

	bytecode, err := ugo.Compile([]byte(script), ugo.DefaultCompilerOptions)
	if err != nil {
		panic(err)
	}
	fn, _ := ugo.ToObject(func(caller *ugo.CallContext) (ugo.Object, error) {
		return ugo.Int(999), nil
	})
	globals := ugo.Map{
		"fn": fn,
	}
	ret, err := ugo.NewVM(bytecode).Run(
		globals,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(ret) // [2, 4, 6, 8]
}
