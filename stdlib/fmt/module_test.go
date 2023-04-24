package fmt_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/ozanh/ugo"
	. "github.com/ozanh/ugo/stdlib/fmt"
)

func Example() {
	exampleRun(`
	fmt := import("fmt")
	fmt.Print("print_")
	fmt.Println("line")
	fmt.Println("a", "b", 3)
	fmt.Printf("%v\n", [1, 2])
	fmt.Println(fmt.Sprint("x", "y", 4))

	a1 := fmt.ScanArg("string")
	a2 := fmt.ScanArg("int")
	r := fmt.Sscanf("abc 123", "%s%d", a1, a2)
	fmt.Println(r)
	fmt.Println(bool(a1), a1.Value)
	fmt.Println(bool(a2), a2.Value)
	`)
	// Output:
	// print_line
	// a b 3
	// [1, 2]
	// xy4
	// 2
	// true abc
	// true 123
}

func TestScript(t *testing.T) {

	testCases := []struct {
		s string
		r Object
	}{
		// scan
		{
			s: `return string(fmt.ScanArg())`,
			r: String("<scanArg>"),
		},
		{
			s: `return typeName(fmt.ScanArg())`,
			r: String("scanArg"),
		},
		{
			s: `
		a1 := fmt.ScanArg()
		ret := fmt.Sscan("abc", a1)
		return ret, bool(a1), a1.Value
			`,
			r: Array{Int(1), True, String("abc")},
		},
		{
			s: `
		a1 := fmt.ScanArg()
		ret := fmt.Sscan("abc xyz", a1)
		return ret, bool(a1), a1.Value
			`,
			r: Array{Int(1), True, String("abc")},
		},
		{
			s: `
		a1 := fmt.ScanArg()
		a2 := fmt.ScanArg()
		ret := fmt.Sscan("abc xyz", a1, a2)
		return [
			ret,
			[bool(a1), a1.Value],
			[bool(a2), a2.Value],
		]
			`,
			r: Array{
				Int(2),
				Array{True, String("abc")},
				Array{True, String("xyz")},
			},
		},
		{
			s: `
		a1 := fmt.ScanArg("string")
		a2 := fmt.ScanArg("int")
		a3 := fmt.ScanArg("uint")
		a4 := fmt.ScanArg("float")
		a5 := fmt.ScanArg("char")
		a6 := fmt.ScanArg("bool")
		a7 := fmt.ScanArg("bytes")
		ret := fmt.Sscan("abc 1 2 3.4 5 t bytes", 
			a1, a2, a3, a4, a5, a6, a7)
		return [
			ret,
			[bool(a1), a1.Value],
			[bool(a2), a2.Value],
			[bool(a3), a3.Value],
			[bool(a4), a4.Value],
			[bool(a5), a5.Value],
			[bool(a6), a6.Value],
			[bool(a7), a7.Value],
		]
			`,
			r: Array{
				Int(7),
				Array{True, String("abc")},
				Array{True, Int(1)},
				Array{True, Uint(2)},
				Array{True, Float(3.4)},
				Array{True, Char(5)},
				Array{True, True},
				Array{True, Bytes("bytes")},
			},
		},
		{
			s: `
		a1 := fmt.ScanArg(string)
		a2 := fmt.ScanArg(int)
		a3 := fmt.ScanArg(uint)
		a4 := fmt.ScanArg(float)
		a5 := fmt.ScanArg(char)
		a6 := fmt.ScanArg(bool)
		a7 := fmt.ScanArg(bytes)
		ret := fmt.Sscan("abc 1 2 3.4 5 t bytes", 
			a1, a2, a3, a4, a5, a6, a7)
		return [
			ret,
			[bool(a1), a1.Value],
			[bool(a2), a2.Value],
			[bool(a3), a3.Value],
			[bool(a4), a4.Value],
			[bool(a5), a5.Value],
			[bool(a6), a6.Value],
			[bool(a7), a7.Value],
		]
			`,
			r: Array{
				Int(7),
				Array{True, String("abc")},
				Array{True, Int(1)},
				Array{True, Uint(2)},
				Array{True, Float(3.4)},
				Array{True, Char(5)},
				Array{True, True},
				Array{True, Bytes("bytes")},
			},
		},
		{
			s: `
		a1 := fmt.ScanArg()
		a2 := fmt.ScanArg()
		a3 := fmt.ScanArg()
		ret := fmt.Sscan("abc xyz", a1, a2, a3)
		return [
			string(ret),
			[bool(a1), a1.Value],
			[bool(a2), a2.Value],
			[bool(a3), a3.Value],
		]
			`,
			r: Array{
				String("error: EOF"),
				Array{True, String("abc")},
				Array{True, String("xyz")},
				Array{False, Undefined},
			},
		},
		{
			s: `
		a1 := fmt.ScanArg("string")
		a2 := fmt.ScanArg("int")
		a3 := fmt.ScanArg("int")
		ret := fmt.Sscanf("abc 3 15", "%s%d", a1, a2, a3)
		return [
			string(ret),
			[bool(a1), a1.Value],
			[bool(a2), a2.Value],
			[bool(a3), a3.Value],
		]
			`,
			r: Array{
				String("error: too many operands"),
				Array{True, String("abc")},
				Array{True, Int(3)},
				Array{False, Undefined},
			},
		},
		{
			s: `
		a1 := fmt.ScanArg("string")
		a2 := fmt.ScanArg("int")
		a3 := fmt.ScanArg("float")
		ret := fmt.Sscanln("abc 3\n1.5", a1, a2, a3)
		return [
			string(ret),
			[bool(a1), a1.Value],
			[bool(a2), a2.Value],
			[bool(a3), a3.Value],
		]
			`,
			r: Array{
				String("error: unexpected newline"),
				Array{True, String("abc")},
				Array{True, Int(3)},
				Array{False, Undefined},
			},
		},
		// sprint
		{
			s: `return fmt.Sprint(1, 2, "c", 'd')`,
			r: String("1 2c100"),
		},
		{
			s: `return fmt.Sprintf("%.1f%s%c%d", 1.2, "abc", 'e', 18u)`,
			r: String("1.2abce18"),
		},
		{
			s: `return fmt.Sprintln(1.2, "abc", 'e', 18u)`,
			r: String("1.2 abc 101 18\n"),
		},
		// runtime errors
		{
			s: `
		try {
			fmt.Printf()
		} catch err {
			return string(err)
		}
			`,
			r: String("WrongNumberOfArgumentsError: want>=1 got=0"),
		},
		{
			s: `
		try {
			fmt.Sprintf()
		} catch err {
			return string(err)
		}
			`,
			r: String("WrongNumberOfArgumentsError: want>=1 got=0"),
		},
		{
			s: `
		try {
			arg := fmt.ScanArg("unknown")
		} catch err {
			return string(err)
		}
			`,
			r: String("TypeError: \"unknown\" not implemented"),
		},
		{
			s: `
		try {
			arg := fmt.Sscan()
		} catch err {
			return string(err)
		}
			`,
			r: String("WrongNumberOfArgumentsError: want>=2 got=0"),
		},
		{
			s: `
		try {
			arg := fmt.Sscanf()
		} catch err {
			return string(err)
		}
			`,
			r: String("WrongNumberOfArgumentsError: want>=3 got=0"),
		},
		{
			s: `
		try {
			arg := fmt.Sscanln()
		} catch err {
			return string(err)
		}
			`,
			r: String("WrongNumberOfArgumentsError: want>=2 got=0"),
		},
		{
			s: `
		try {
			arg := fmt.Sscanf("", "", 1)
		} catch err {
			return string(err)
		}
			`,
			r: String("TypeError: invalid type for argument '2': expected ScanArg interface, found int"),
		},
		{
			s: `
		try {
			arg := fmt.Sscanln("", 1)
		} catch err {
			return string(err)
		}
			`,
			r: String("TypeError: invalid type for argument '1': expected ScanArg interface, found int"),
		},
	}
	for _, tC := range testCases {
		expectRun(t, tC.s, tC.r)
	}
}

func expectRun(t *testing.T, script string, expected Object) {
	t.Helper()

	script = `
		fmt := import("fmt")
	` + script

	mm := NewModuleMap()
	mm.AddBuiltinModule("fmt", Module)
	c := DefaultCompilerOptions
	c.ModuleMap = mm
	bc, err := Compile([]byte(script), c)
	require.NoError(t, err, script)
	ret, err := NewVM(bc).Run(nil)
	require.NoError(t, err, script)
	require.Equal(t, expected, ret, script)
}

func exampleRun(script string) {
	mm := NewModuleMap()
	mm.AddBuiltinModule("fmt", Module)
	c := DefaultCompilerOptions
	c.ModuleMap = mm
	bc, err := Compile([]byte(script), c)
	if err != nil {
		panic(err)
	}
	_, err = NewVM(bc).Run(nil)
	if err != nil {
		panic(err)
	}
}
