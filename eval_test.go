package ugo_test

import (
	"context"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/ozanh/ugo"
	ugotime "github.com/ozanh/ugo/stdlib/time"
)

func TestEval(t *testing.T) {
	type scriptResult struct {
		script string
		result Object
	}
	testCases := []struct {
		name   string
		opts   CompilerOptions
		global Object
		args   []Object
		ctx    context.Context
		sr     []scriptResult
	}{
		{
			name: "simple",
			sr: []scriptResult{
				{`var a`, Undefined},
				{`1`, Int(1)},
				{`return 10`, Int(10)},
				{`a = 10`, Undefined},
				{`return a`, Int(10)},
				{`return a*a`, Int(100)},
			},
		},
		{
			name: "import",
			opts: CompilerOptions{
				ModuleMap: NewModuleMap().
					AddBuiltinModule("time", ugotime.Module),
			},
			sr: []scriptResult{
				{`time := import("time")`, Undefined},
				{`time.Second`, ugotime.Module["Second"]},
				{`tmp := time.Second`, Undefined},
				{`tmp`, ugotime.Module["Second"]},
				{`time.Second = ""`, Undefined},
				{`time.Second`, String("")},
				{`time.Second = tmp`, Undefined},
				{`time.Second`, ugotime.Module["Second"]},
			},
		},
		{
			name:   "globals",
			global: Map{"g": String("test")},
			sr: []scriptResult{
				{`global g`, Undefined},
				{`return g`, String("test")},
				{`globals()["g"]`, String("test")},
			},
		},
		{
			name: "locals",
			args: []Object{Int(1), Int(2)},
			sr: []scriptResult{
				{`var (a, b, c)`, Undefined},
				{`a`, Undefined},
				{`b`, Undefined},
				{`c`, Undefined},
			},
		},
		{
			name: "params",
			args: []Object{Int(1), Int(2)},
			sr: []scriptResult{
				{`param (a, b, c)`, Undefined},
				{`a`, Int(1)},
				{`b`, Int(2)},
				{`c`, Undefined},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			eval := NewEval(tC.opts, tC.global, tC.args...)
			for _, sr := range tC.sr {
				ret, _, err := eval.Run(tC.ctx, []byte(sr.script))
				require.NoError(t, err, sr.script)
				require.Equal(t, sr.result, ret, sr.script)
			}
		})
	}

	// test context
	t.Run("context", func(t *testing.T) {
		globals := Map{
			"Gosched": &Function{
				Value: func(args ...Object) (Object, error) {
					runtime.Gosched()
					return Undefined, nil
				},
			},
		}
		eval := NewEval(DefaultCompilerOptions, globals)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		ret, bc, err := eval.Run(ctx, []byte(`
		global Gosched; Gosched(); foo := "bar"; return foo`))
		require.Nilf(t, ret, "return value:%v", ret)
		require.Equal(t, context.Canceled, err, err)
		require.NotNil(t, bc)
	})

	// test error
	t.Run("parser error", func(t *testing.T) {
		eval := NewEval(DefaultCompilerOptions, nil)
		ret, bc, err := eval.Run(context.Background(), []byte(`...`))
		require.Nil(t, ret)
		require.Nil(t, bc)
		require.Contains(t, err.Error(),
			`Parse Error: expected statement, found '...'`)
	})
}
