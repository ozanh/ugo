// +build !js

package main

import (
	"bytes"
	"context"
	"flag"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/c-bata/go-prompt"
	"github.com/ozanh/ugo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FIXME: prompt package requires /dev/tty so it is not testable with Go tests
// although Renderer can be given with options, because contructor tries to
// open /dev/tty before setting custom renderer.

func TestREPL(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stdout := bytes.NewBuffer(nil)
	cw := &mockConsoleWriter{}
	var exited bool
	r := newTestREPL(ctx, stdout, cw, func() {
		exited = true
	})

	r.executor("test")
	testHasPrefix(t, string(cw.consume()),
		"\nCompile Error: unresolved reference \"test\"")

	r.executor("test := 1")
	testHasPrefix(t, string(cw.consume()), "undefined\n")

	r.executor(".bytecode")
	testHasPrefix(t, string(testReadAll(t, stdout)), "Bytecode\n")

	r.executor(".builtins")
	testHasPrefix(t, string(testReadAll(t, stdout)),
		"builtin-function:append\n")

	r.executor(".gc")
	require.Equal(t, "", string(testReadAll(t, stdout)))

	r.executor(".globals")
	testHasPrefix(t, string(testReadAll(t, stdout)), `{"Gosched": <function:Gosched>}`)

	r.executor(".globals+")
	testHasPrefix(t, string(testReadAll(t, stdout)), "&ugo.SyncMap{")

	r.executor(".locals")
	testHasPrefix(t, string(testReadAll(t, stdout)), "[1]\n")

	r.executor(".locals+")
	testHasPrefix(t, string(testReadAll(t, stdout)), "[]ugo.Object{1}\n")

	r.executor("return test")
	testHasPrefix(t, string(cw.consume()), "1\n")
	r.executor(".return")
	testHasPrefix(t, string(testReadAll(t, stdout)), "1\n")

	r.executor(".return+")
	testHasPrefix(t, string(testReadAll(t, stdout)), "GoType:ugo.Int,")

	r.executor(".symbols")
	testHasPrefix(t, string(testReadAll(t, stdout)),
		"[Symbol{Name:test Index:0 Scope:LOCAL Assigned:true Original:<nil> Constant:false}]\n")

	r.executor(".modules_cache")
	testHasPrefix(t, string(testReadAll(t, stdout)), "[]\n")

	r.executor(`import("time")`)
	testHasPrefix(t, string(cw.consume()), "{")
	r.executor(".modules_cache")
	testHasPrefix(t, string(testReadAll(t, stdout)), "[{")

	r.executor(".memory_stats")
	testHasPrefix(t, string(testReadAll(t, stdout)), "Go Memory Stats")

	g := grepl
	r.executor(".reset")
	require.Empty(t, cw.consume())
	require.Empty(t, testReadAll(t, stdout))
	require.NotSame(t, g, grepl)

	r.executor(".exit")
	require.Empty(t, cw.consume())
	require.Empty(t, testReadAll(t, stdout))
	require.True(t, exited)

	require.Empty(t, cw.consume())
	require.Empty(t, testReadAll(t, stdout))
}

func TestFlags(t *testing.T) {

	testCases1 := []struct {
		args            []string
		expectEnabled   bool
		expectParser    bool
		expectOptimizer bool
		expectCompiler  bool
	}{
		{[]string{"-trace", "parser"}, true, true, false, false},
		{[]string{"-trace", "optimizer"}, true, false, true, false},
		{[]string{"-trace", "compiler"}, true, false, false, true},

		{[]string{"-trace", "parser,optimizer"}, true, true, true, false},
		{[]string{"-trace", "parser,compiler"}, true, true, false, true},
		{[]string{"-trace", "compiler,optimizer"}, true, false, true, true},
	}
	for _, tC := range testCases1 {
		t.Run("", func(t *testing.T) {
			// trace flags are global variables, set to defaults at each run
			traceEnabled = false
			traceParser = false
			traceOptimizer = false
			traceCompiler = false

			fs := flag.NewFlagSet("test tracers", flag.ExitOnError)
			fp, to, err := parseFlags(fs, tC.args)
			require.NoError(t, err)
			require.Empty(t, fp)
			require.Empty(t, to)
			require.Equal(t, tC.expectEnabled, traceEnabled)
			require.Equal(t, tC.expectParser, traceParser)
			require.Equal(t, tC.expectOptimizer, traceOptimizer)
			require.Equal(t, tC.expectCompiler, traceCompiler)
		})
	}

	fs := flag.NewFlagSet("script file", flag.ExitOnError)
	fp, to, err := parseFlags(fs, []string{"testdata/fibtc.ugo"})
	require.NoError(t, err)
	require.Empty(t, to)
	require.Equal(t, "testdata/fibtc.ugo", fp)

	fs = flag.NewFlagSet("stdin", flag.ExitOnError)
	fp, to, err = parseFlags(fs, []string{"-"})
	require.NoError(t, err)
	require.Empty(t, to)
	require.Equal(t, "-", fp)

	fs = flag.NewFlagSet("file does not exist", flag.ExitOnError)
	_, _, err = parseFlags(fs, []string{"testdata/doesnotexist"})
	require.Error(t, err)
}

func TestExecuteScript(t *testing.T) {
	scr, err := ioutil.ReadFile("testdata/fibtc.ugo")
	require.NoError(t, err)
	require.NoError(t, executeScript(context.Background(), scr, nil))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = executeScript(ctx, scr, nil)
	require.Error(t, err)
	if err != context.Canceled && err != ugo.ErrVMAborted {
		t.Fatalf("unexpected error: %+v", err)
	}
}

func testHasPrefix(t *testing.T, s, pref string) {
	t.Helper()
	v := strings.HasPrefix(s, pref)
	if !assert.True(t, v) {
		t.Fatalf("input: %q\nprefix: %q", s, pref)
	}
}

func testReadAll(t *testing.T, r io.Reader) []byte {
	t.Helper()
	b, err := ioutil.ReadAll(r)
	require.NoError(t, err)
	return b
}

func newTestREPL(ctx context.Context,
	stdout io.Writer,
	cw prompt.ConsoleWriter,
	exitFunc func(),
) *repl {
	r := newREPL(ctx, stdout, cw)
	r.commands[".exit"] = exitFunc
	return r
}

type mockConsoleWriter struct {
	buffer  []byte
	flushed []byte
}

func (w *mockConsoleWriter) consume() []byte {
	f := w.flushed
	w.flushed = nil
	return f
}

func (w *mockConsoleWriter) Flush() error {
	w.flushed = append(w.flushed, w.buffer...)
	w.buffer = nil
	return nil
}

func (w *mockConsoleWriter) WriteRaw(data []byte) {
	w.buffer = append(w.buffer, data...)
}

func (w *mockConsoleWriter) Write(data []byte) {
	w.buffer = append(w.buffer, data...)
}

func (w *mockConsoleWriter) WriteRawStr(data string) {
	w.WriteRaw([]byte(data))
}

func (w *mockConsoleWriter) WriteStr(data string) {
	w.Write([]byte(data))
}

func (w *mockConsoleWriter) EraseScreen()                            {}
func (w *mockConsoleWriter) EraseUp()                                {}
func (w *mockConsoleWriter) EraseDown()                              {}
func (w *mockConsoleWriter) EraseStartOfLine()                       {}
func (w *mockConsoleWriter) EraseEndOfLine()                         {}
func (w *mockConsoleWriter) EraseLine()                              {}
func (w *mockConsoleWriter) ShowCursor()                             {}
func (w *mockConsoleWriter) HideCursor()                             {}
func (w *mockConsoleWriter) CursorGoTo(row, col int)                 {}
func (w *mockConsoleWriter) CursorUp(n int)                          {}
func (w *mockConsoleWriter) CursorDown(n int)                        {}
func (w *mockConsoleWriter) CursorForward(n int)                     {}
func (w *mockConsoleWriter) CursorBackward(n int)                    {}
func (w *mockConsoleWriter) AskForCPR()                              {}
func (w *mockConsoleWriter) SaveCursor()                             {}
func (w *mockConsoleWriter) UnSaveCursor()                           {}
func (w *mockConsoleWriter) ScrollDown()                             {}
func (w *mockConsoleWriter) ScrollUp()                               {}
func (w *mockConsoleWriter) SetTitle(title string)                   {}
func (w *mockConsoleWriter) ClearTitle()                             {}
func (w *mockConsoleWriter) SetColor(fg, bg prompt.Color, bold bool) {}
func (w *mockConsoleWriter) SetDisplayAttributes(fg, bg prompt.Color,
	attrs ...prompt.DisplayAttribute) {
}
