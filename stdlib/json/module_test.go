package json_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/ozanh/ugo"
	. "github.com/ozanh/ugo/stdlib/json"
)

func TestModuleTypes(t *testing.T) {
	ret, err := ToObject(json.RawMessage(nil))
	require.NoError(t, err)
	require.Equal(t, &RawMessage{Value: Bytes{}}, ret)

	ret, err = ToObject(json.RawMessage([]byte("null")))
	require.NoError(t, err)
	require.Equal(t, &RawMessage{Value: Bytes([]byte("null"))}, ret)

	iface := ToInterface(ret)
	require.Equal(t, json.RawMessage([]byte("null")), iface)
}

func TestScript(t *testing.T) {
	catchf := func(s string, args ...interface{}) string {
		return fmt.Sprintf(`
		json := import("json")
		try {
			return %s
		} catch err {
			return string(err)
		}
		`, fmt.Sprintf(s, args...))
	}
	scriptf := func(s string, args ...interface{}) string {
		return fmt.Sprintf(`
		json := import("json")
		return %s
		`, fmt.Sprintf(s, args...))
	}
	errnarg := func(want, got int) String {
		return String(ErrWrongNumArguments.NewError(
			fmt.Sprintf("want=%d got=%d", want, got),
		).String())
	}

	expectRun(t, scriptf(""), nil, Undefined)

	for key, val := range Module {
		expectRun(t, scriptf("typeName(json.%s)", key), nil, String("function"))
		expectRun(t, scriptf("string(json.%s)", key), nil, String(fmt.Sprintf("<function:%s>", key)))
		require.NotNil(t, val)
		require.NotNil(t, val.(*Function).Value)
	}

	expectRun(t, catchf(`json.Marshal()`), nil, errnarg(1, 0))
	expectRun(t, catchf(`typeName(json.Marshal(undefined))`), nil, String("bytes"))
	expectRun(t, catchf(`string(json.Marshal(undefined))`), nil, String("null"))
	expectRun(t, catchf(`string(json.Marshal(error("test")))`), nil, String("")) // ignore error
	expectRun(t, catchf(`string(json.Marshal(true))`), nil, String("true"))
	expectRun(t, catchf(`string(json.Marshal(false))`), nil, String("false"))
	expectRun(t, catchf(`string(json.Marshal(1))`), nil, String("1"))
	expectRun(t, catchf(`string(json.Marshal(2u))`), nil, String("2"))
	expectRun(t, catchf(`string(json.Marshal(3.4))`), nil, String("3.4"))
	expectRun(t, catchf(`string(json.Marshal('x'))`), nil, String("120"))
	expectRun(t, catchf(`string(json.Marshal("test"))`), nil, String(`"test"`))
	expectRun(t, catchf(`string(json.Marshal(bytes(0,1)))`), nil, String(`"AAE="`))
	expectRun(t, catchf(`string(json.Marshal([]))`), nil, String("[]"))
	expectRun(t, catchf(`string(json.Marshal([1, "a", 2u, 'x',3.4,true,false,
	{a:[],"b":0,ç:undefined},bytes(0,1),
	]))`), nil, String(`[1,"a",2,120,3.4,true,false,{"a":[],"b":0,"ç":null},"AAE="]`))
	expectRun(t, catchf(`string(json.Marshal({}))`), nil, String("{}"))
	expectRun(t, catchf(`string(json.Marshal({_: 1, k2:[3,true,"a"]}))`),
		nil, String(`{"_":1,"k2":[3,true,"a"]}`))

	expectRun(t, catchf(`json.Indent()`), nil, errnarg(3, 0))
	expectRun(t, catchf(`string(json.Indent("[1,2]", "", " "))`), nil, String("[\n 1,\n 2\n]"))

	expectRun(t, catchf(`json.MarshalIndent()`), nil, errnarg(3, 0))
	expectRun(t, catchf(`string(json.MarshalIndent({a: 1, b: [2, true, "<"]},"", " "))`),
		nil, String("{\n \"a\": 1,\n \"b\": [\n  2,\n  true,\n  \"\\u003c\"\n ]\n}"))

	expectRun(t, catchf(`json.Compact()`), nil, errnarg(2, 0))
	expectRun(t, catchf(`string(json.Compact(json.Marshal(json.NoEscape(["<",">"])), true))`),
		nil, String(`["\u003c","\u003e"]`))
	expectRun(t, catchf(`string(json.Compact(json.MarshalIndent({a: 1, b: [2, true, "<"]},"", " "), true))`),
		nil, String(`{"a":1,"b":[2,true,"\u003c"]}`))
	expectRun(t, catchf(`string(json.Compact(json.MarshalIndent(json.NoEscape(["<",">"]), "", " "), false))`),
		nil, String(`["<",">"]`))

	expectRun(t, catchf(`json.RawMessage()`), nil, errnarg(1, 0))
	expectRun(t, catchf(`string(json.RawMessage(json.Marshal([1, 2])))`),
		nil, String("[1,2]"))
	expectRun(t, catchf(`string(json.Marshal(json.RawMessage(json.Marshal([1, 2]))))`),
		nil, String("[1,2]"))

	expectRun(t, catchf(`json.Quote()`), nil, errnarg(1, 0))
	expectRun(t, catchf(`json.NoQuote()`), nil, errnarg(1, 0))
	expectRun(t, catchf(`json.NoEscape()`), nil, errnarg(1, 0))
	expectRun(t, catchf(`string(json.Marshal(json.Quote([1,2,"a"])))`),
		nil, String(`["1","2","\"a\""]`))
	expectRun(t, catchf(`string(json.Marshal(json.Quote([1,2,{a:"x"}])))`),
		nil, String(`["1","2",{"a":"\"x\""}]`))
	expectRun(t, catchf(`string(json.Marshal(json.Quote([1,2,{a:json.NoQuote("x")}])))`),
		nil, String(`["1","2",{"a":"x"}]`))

	expectRun(t, catchf(`json.Unmarshal()`), nil, errnarg(1, 0))
	expectRun(t, catchf(`json.Unmarshal("[1,true,false,\"x\",{\"a\":\"b\"}]")`),
		nil, Array{Float(1), True, False, String("x"), Map{"a": String("b")}})

	expectRun(t, catchf(`json.Valid()`), nil, errnarg(1, 0))
	expectRun(t, catchf(`json.Valid("{}")`), nil, True)
	expectRun(t, catchf(`json.Valid("{")`), nil, False)

	expectRun(t, catchf(`string(json.Marshal(json.NoEscape(json.Quote("<"))))`), nil, String(`"\"<\""`))
	expectRun(t, catchf(`string(json.Marshal(json.NoQuote(json.NoEscape("<"))))`), nil, String(`"<"`))
	expectRun(t, catchf(`string(json.Marshal(json.Quote(json.NoEscape("<"))))`), nil, String(`"\"<\""`))

	expectRun(t, catchf(`string(json.Unmarshal(bytes(0)))`),
		nil, String(`error: invalid character '\x00' looking for beginning of value`))
	expectRun(t, catchf(`string(json.Indent(bytes(0), "", " "))`),
		nil, String(`error: invalid character '\x00' looking for beginning of value`))
	expectRun(t, catchf(`string(json.Compact(bytes(0), true))`),
		nil, String(`error: invalid character '\x00' looking for beginning of value`))
}

func TestCycle(t *testing.T) {
	expectRun(t, `json:=import("json");a:=[1,2];a[1]=a;return string(json.Marshal(a))`,
		nil, String(`error: json: unsupported value: encountered a cycle via array`))
	expectRun(t, `json:=import("json");a:=[1,2];a[1]=a;return string(json.MarshalIndent(a,""," "))`,
		nil, String(`error: json: unsupported value: encountered a cycle via array`))
	expectRun(t, `json:=import("json");m:={a:1};m.b=m;return string(json.Marshal(m))`,
		nil, String(`error: json: unsupported value: encountered a cycle via map`))
	expectRun(t, `param m;json:=import("json");m.b=m;return string(json.Marshal(m))`,
		newOpts().Args(&SyncMap{Value: Map{}}),
		String(`error: json: unsupported value: encountered a cycle via syncMap`))

	ptr := &ObjectPtr{}
	var m Object = Map{}
	m.(Map)["a"] = ptr
	ptr.Value = &m
	_, err := Marshal(ptr)
	require.Error(t, err)
	require.Contains(t, err.Error(), `json: unsupported value: encountered a cycle via objectPtr`)
}

type Opts struct {
	global Object
	args   []Object
}

func newOpts() *Opts {
	return &Opts{}
}

func (o *Opts) Args(args ...Object) *Opts {
	o.args = args
	return o
}

func (o *Opts) Globals(g Object) *Opts {
	o.global = g
	return o
}

func expectRun(t *testing.T, script string, opts *Opts, expected Object) {
	t.Helper()
	if opts == nil {
		opts = newOpts()
	}
	mm := NewModuleMap()
	mm.AddBuiltinModule("json", Module)
	c := DefaultCompilerOptions
	c.ModuleMap = mm
	bc, err := Compile([]byte(script), c)
	require.NoError(t, err)
	ret, err := NewVM(bc).Run(opts.global, opts.args...)
	require.NoError(t, err)
	require.Equal(t, expected, ret)
}
