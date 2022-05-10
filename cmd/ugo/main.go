// Copyright (c) 2020-2022 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

//go:build !js
// +build !js

package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/peterh/liner"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/importers"
	"github.com/ozanh/ugo/token"

	ugofmt "github.com/ozanh/ugo/stdlib/fmt"
	ugojson "github.com/ozanh/ugo/stdlib/json"
	ugostrings "github.com/ozanh/ugo/stdlib/strings"
	ugotime "github.com/ozanh/ugo/stdlib/time"
)

const logo = `
            /$$$$$$   /$$$$$$ 
           /$$__  $$ /$$__  $$
 /$$   /$$| $$  \__/| $$  \ $$
| $$  | $$| $$ /$$$$| $$  | $$
| $$  | $$| $$|_  $$| $$  | $$
| $$  | $$| $$  \ $$| $$  | $$
|  $$$$$$/|  $$$$$$/|  $$$$$$/
 \______/  \______/  \______/ 
                                       
`

const (
	title         = "uGO"
	promptPrefix  = ">>> "
	promptPrefix2 = "... "
)

var (
	noOptimizer    bool
	traceEnabled   bool
	traceParser    bool
	traceOptimizer bool
	traceCompiler  bool
)

var initialSuggLen int

var errExit = errors.New("exit")

var scriptGlobals = &ugo.SyncMap{
	Value: ugo.Map{
		"Gosched": &ugo.Function{
			Name: "Gosched",
			Value: func(args ...ugo.Object) (ugo.Object, error) {
				runtime.Gosched()
				return ugo.Undefined, nil
			},
		},
	},
}

var grepl *repl

type repl struct {
	ctx          context.Context
	eval         *ugo.Eval
	lastBytecode *ugo.Bytecode
	lastResult   ugo.Object
	stdout       io.Writer
	commands     map[string]func(string) error
	script       *bytes.Buffer
	isMultiline  bool
}

func newREPL(ctx context.Context, stdout io.Writer) *repl {
	opts := ugo.CompilerOptions{
		ModulePath:        "(repl)",
		ModuleMap:         defaultModuleMap("."),
		SymbolTable:       ugo.NewSymbolTable(),
		OptimizerMaxCycle: ugo.TraceCompilerOptions.OptimizerMaxCycle,
		TraceParser:       traceParser,
		TraceOptimizer:    traceOptimizer,
		TraceCompiler:     traceCompiler,
		OptimizeConst:     !noOptimizer,
		OptimizeExpr:      !noOptimizer,
	}

	if stdout == nil {
		stdout = os.Stdout
	}

	if traceEnabled {
		opts.Trace = stdout
	}

	r := &repl{
		ctx:    ctx,
		eval:   ugo.NewEval(opts, scriptGlobals),
		stdout: stdout,
		script: bytes.NewBuffer(nil),
	}

	r.commands = map[string]func(string) error{
		".bytecode":      r.cmdBytecode,
		".builtins":      r.cmdBuiltins,
		".gc":            r.cmdGC,
		".globals":       r.cmdGlobals,
		".globals+":      r.cmdGlobalsVerbose,
		".locals":        r.cmdLocals,
		".locals+":       r.cmdLocalsVerbose,
		".return":        r.cmdReturn,
		".return+":       r.cmdReturnVerbose,
		".symbols":       r.cmdSymbols,
		".modules_cache": r.cmdModulesCache,
		".memory_stats":  r.cmdMemoryStats,
		".reset":         r.cmdReset,
		".exit":          func(string) error { return errExit },
	}
	return r
}

func (r *repl) cmdBytecode(_ string) error {
	_, _ = fmt.Fprintf(r.stdout, "%s\n", r.lastBytecode)
	return nil
}

func (r *repl) cmdBuiltins(_ string) error {
	builtins := make([]string, len(ugo.BuiltinsMap))

	for k, v := range ugo.BuiltinsMap {
		builtins[v] = fmt.Sprint(ugo.BuiltinObjects[v].TypeName(), ":", k)
	}
	_, _ = fmt.Fprintln(r.stdout, strings.Join(builtins, "\n"))
	return nil
}

func (*repl) cmdGC(_ string) error {
	runtime.GC()
	return nil
}

func (r *repl) cmdGlobals(_ string) error {
	_, _ = fmt.Fprintf(r.stdout, "%+v\n", r.eval.Globals)
	return nil
}

func (r *repl) cmdGlobalsVerbose(_ string) error {
	_, _ = fmt.Fprintf(r.stdout, "%#v\n", r.eval.Globals)
	return nil
}

func (r *repl) cmdLocals(_ string) error {
	_, _ = fmt.Fprintf(r.stdout, "%+v\n", r.eval.Locals)
	return nil
}

func (r *repl) cmdLocalsVerbose(_ string) error {
	fmt.Fprintf(r.stdout, "%#v\n", r.eval.Locals)
	return nil
}

func (r *repl) cmdReturn(_ string) error {
	_, _ = fmt.Fprintf(r.stdout, "%#v\n", r.lastResult)
	return nil
}

func (r *repl) cmdReturnVerbose(_ string) error {
	if r.lastResult != nil {
		_, _ = fmt.Fprintf(r.stdout,
			"GoType:%[1]T, TypeName:%[2]s, Value:%#[1]v\n",
			r.lastResult, r.lastResult.TypeName())
	} else {
		_, _ = fmt.Fprintln(r.stdout, "<nil>")
	}
	return nil
}

func (r *repl) cmdReset(_ string) error {
	grepl = newREPL(r.ctx, r.stdout)
	return nil
}

func (r *repl) cmdSymbols(_ string) error {
	_, _ = fmt.Fprintf(r.stdout, "%v\n", r.eval.Opts.SymbolTable.Symbols())
	return nil
}

func (r *repl) cmdMemoryStats(_ string) error {
	// writeMemStats writes the formatted current, total and OS memory
	// being used. As well as the number of garbage collection cycles completed.
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	_, _ = fmt.Fprintf(r.stdout, "Go Memory Stats see: "+
		"https://golang.org/pkg/runtime/#MemStats\n\n")
	_, _ = fmt.Fprintf(r.stdout, "HeapAlloc = %s", humanFriendlySize(m.HeapAlloc))
	_, _ = fmt.Fprintf(r.stdout, "\tHeapObjects = %v", m.HeapObjects)
	_, _ = fmt.Fprintf(r.stdout, "\tSys = %s", humanFriendlySize(m.Sys))
	_, _ = fmt.Fprintf(r.stdout, "\tNumGC = %v\n", m.NumGC)
	return nil
}

func (r *repl) cmdModulesCache(_ string) error {
	_, _ = fmt.Fprintf(r.stdout, "%v\n", r.eval.ModulesCache)
	return nil
}

func (r *repl) writeErrorStr(msg string) {
	_, _ = fmt.Fprint(r.stdout, msg)
	_, _ = fmt.Fprintln(r.stdout)
}

func (r *repl) writeStr(msg string) {
	_, _ = fmt.Fprint(r.stdout, msg)
	_, _ = fmt.Fprintln(r.stdout)
}

func (r *repl) execute(line string) error {
	switch {
	case !r.isMultiline && line == "":
		return nil
	case !r.isMultiline && len(line) > 0 && line[0] == '.':
		cmd := strings.Fields(line)[0]
		if fn, ok := r.commands[cmd]; ok {
			return fn(line)
		}
	case strings.HasSuffix(line, "\\"):
		r.isMultiline = true
		r.script.WriteString(line[:len(line)-1])
		r.script.WriteString("\n")
		return nil
	}

	r.script.WriteString(line)

	r.executeScript()

	r.isMultiline = false
	r.script.Reset()
	return nil
}

func (r *repl) executeScript() {

	var err error

	r.lastResult, r.lastBytecode, err = r.eval.Run(r.ctx, r.script.Bytes())
	if err != nil {
		r.writeErrorStr(fmt.Sprintf("\n%+v\n", err))
		return
	}

	if err != nil {
		r.writeErrorStr(fmt.Sprintf("VM:\n     %+v\n", err))
		return
	}

	switch v := r.lastResult.(type) {
	case ugo.String:
		r.writeStr(fmt.Sprintf("%q\n", string(v)))
	case ugo.Char:
		r.writeStr(fmt.Sprintf("%q\n", rune(v)))
	case ugo.Bytes:
		r.writeStr(fmt.Sprintf("%v\n", []byte(v)))
	default:
		r.writeStr(fmt.Sprintf("%v\n", r.lastResult))
	}

	symbols := r.eval.Opts.SymbolTable.Symbols()
	suggestions = suggestions[:initialSuggLen]

	for _, s := range symbols {
		if s.Scope != ugo.ScopeBuiltin {
			suggestions = append(suggestions,
				suggest{
					text:        s.Name,
					description: string(s.Scope) + " variable",
				},
			)
		}
	}
}

func (r *repl) prefix() string {
	if r.isMultiline {
		return promptPrefix2
	}
	return promptPrefix
}

func defaultModuleMap(workdir string) *ugo.ModuleMap {
	return ugo.NewModuleMap().
		AddBuiltinModule("time", ugotime.Module).
		AddBuiltinModule("strings", ugostrings.Module).
		AddBuiltinModule("fmt", ugofmt.Module).
		AddBuiltinModule("json", ugojson.Module).
		SetExtImporter(&importers.FileImporter{WorkDir: workdir})
}

func humanFriendlySize(b uint64) string {
	if b < 1024 {
		return fmt.Sprint(strconv.FormatUint(b, 10), " bytes")
	}

	if b >= 1024 && b < 1024*1024 {
		return fmt.Sprint(strconv.FormatFloat(
			float64(b)/1024, 'f', 1, 64), " KiB")
	}

	return fmt.Sprint(strconv.FormatFloat(
		float64(b)/1024/1024, 'f', 1, 64), " MiB")
}

var suggestions = []suggest{
	// Commands
	{text: ".bytecode", description: "Print Bytecode"},
	{text: ".builtins", description: "Print Builtins"},
	{text: ".reset", description: "Reset"},
	{text: ".locals", description: "Print Locals"},
	{text: ".locals+", description: "Print Locals (verbose)"},
	{text: ".globals", description: "Print Globals"},
	{text: ".globals+", description: "Print Globals (verbose)"},
	{text: ".return", description: "Print Last Return Result"},
	{text: ".return+", description: "Print Last Return Result (verbose)"},
	{text: ".modules_cache", description: "Print Modules Cache"},
	{text: ".memory_stats", description: "Print Memory Stats"},
	{text: ".gc", description: "Run Go GC"},
	{text: ".symbols", description: "Print Symbols"},
	{text: ".exit", description: "Exit"},
}

func init() {
	// add builtins to suggestions
	for k := range ugo.BuiltinsMap {
		suggestions = append(suggestions,
			suggest{
				text:        k,
				description: "Builtin " + k,
			},
		)
	}

	for tok := token.Question + 3; tok.IsKeyword(); tok++ {
		s := tok.String()
		suggestions = append(suggestions, suggest{
			text:        s,
			description: "keyword " + s,
		})
	}
	initialSuggLen = len(suggestions)
}

type suggest struct {
	text        string
	description string
}

type prompt struct {
	out           io.Writer
	executer      func(s string) error
	prefixer      func() string
	completer     liner.Completer
	wordcompleter liner.WordCompleter
	exiter        func(error)
}

func (p *prompt) init() {
	_, _ = fmt.Fprintln(p.out, "Copyright (c) 2020-2022 Ozan Hacıbekiroğlu")
	_, _ = fmt.Fprintln(p.out, "License: MIT")
	_, _ = fmt.Fprintln(p.out, "Press Ctrl+D to exit or use .exit command")
	_, _ = fmt.Fprint(p.out, logo)
}

func (p *prompt) run(history io.Reader) {
	p.init()

	line := liner.NewLiner()
	line.SetMultiLineMode(true)
	if p.completer != nil {
		line.SetCompleter(p.completer)
	}
	if p.wordcompleter != nil {
		line.SetWordCompleter(p.wordcompleter)
	}
	_, err := line.ReadHistory(history)
	if err != nil {
		p.errorf("Failed to read history. Error: %v\n", err)
	}

	var str string
	for {
		str, err = line.Prompt(p.prefixer())
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			err = fmt.Errorf("prompt error: %w", err)
			break
		}
		err = p.executer(str)
		if err != nil {
			break
		}
	}
	if err != nil {
		p.errorf("%v\n", err)
	}
	_ = line.Close()
	p.exiter(err)
}

func (p *prompt) errorf(format string, err error) {
	_, _ = fmt.Fprintf(p.out, format, err)
}

func completer(line string) []string {
	return nil
}

func wordCompleter(
	line string,
	pos int,
) (head string, completions []string, tail string) {
	return
}

func parseFlags(
	flagset *flag.FlagSet,
	args []string,
) (filePath string, timeout time.Duration, err error) {

	var trace string
	flagset.StringVar(&trace, "trace", "",
		`Comma separated units: -trace parser,optimizer,compiler`)
	flagset.BoolVar(&noOptimizer, "no-optimizer", false, `Disable optimization`)
	flagset.DurationVar(&timeout, "timeout", 0,
		"Program timeout. It is applicable if a script file is provided and "+
			"must be non-zero duration")

	flagset.Usage = func() {
		_, _ = fmt.Fprint(flagset.Output(),
			"Usage: ugo [flags] [uGO script file]\n\n",
			"If script file is not provided, REPL terminal application is started\n",
			"Use - to read from stdin\n\n",
			"\nFlags:\n",
		)
		flagset.PrintDefaults()
	}

	if err = flagset.Parse(args); err != nil {
		return
	}

	if trace != "" {
		traceEnabled = true
		trace = "," + trace + ","
		if strings.Contains(trace, ",parser,") {
			traceParser = true
		}
		if strings.Contains(trace, ",optimizer,") {
			traceOptimizer = true
		}
		if strings.Contains(trace, ",compiler,") {
			traceCompiler = true
		}
	}

	if flagset.NArg() != 1 {
		return
	}

	filePath = flagset.Arg(0)
	if filePath == "-" {
		return
	}
	_, err = os.Stat(filePath)
	return
}

func executeScript(
	ctx context.Context,
	workdir string,
	script []byte,
	traceOut io.Writer,
) error {
	opts := ugo.DefaultCompilerOptions
	if traceEnabled {
		opts.Trace = traceOut
		opts.TraceParser = traceParser
		opts.TraceCompiler = traceCompiler
		opts.TraceOptimizer = traceOptimizer
	}

	opts.ModuleMap = defaultModuleMap(workdir)

	bc, err := ugo.Compile(script, opts)
	if err != nil {
		return err
	}

	vm := ugo.NewVM(bc)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, err = vm.Run(scriptGlobals)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		vm.Abort()
		<-done
		if err == nil {
			err = ctx.Err()
		}
	}
	return err
}

func checkErr(err error, fn func()) {
	if err == nil {
		return
	}

	defer os.Exit(1)
	_, _ = fmt.Fprintln(os.Stderr, err.Error())
	if fn != nil {
		fn()
	}
}

func main() {
	filePath, timeout, err := parseFlags(flag.CommandLine, os.Args[1:])
	checkErr(err, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if filePath != "" {
		if timeout > 0 {
			var c func()
			ctx, c = context.WithTimeout(ctx, timeout)
			defer c()
		}

		var (
			workdir = "."
			script  []byte
		)
		if filePath == "-" {
			script, err = ioutil.ReadAll(os.Stdin)
		} else {
			workdir = filepath.Dir(filePath)
			script, err = ioutil.ReadFile(filePath)
		}

		checkErr(err, cancel)
		err = executeScript(ctx, workdir, script, os.Stdout)
		checkErr(err, cancel)
		return
	}

	history := []string{
		"a := 1",
		"sum := func(...a) { total := 0; for v in a { total += v }; return total }",
		"func(a, b){ return a*b }(2, 3)",
		`println("")`,
		`var (x, y, z); if x { y } else { z }`,
		`var (x, y, z); x ? y : z`,
		`for i := 0; i < 3; i++ { }`,
		`m := {}; for k,v in m { printf("%s:%v\n", k, v) }`,
		`try { } catch err { } finally { }`,
	}
	histrd := strings.NewReader(strings.Join(history, "\n"))

	grepl = newREPL(ctx, os.Stdout)

	p := &prompt{
		out:           os.Stdout,
		executer:      grepl.execute,
		prefixer:      grepl.prefix,
		completer:     completer,
		wordcompleter: wordCompleter,
		exiter: func(err error) {
			if err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		},
	}
	p.run(histrd)
}
