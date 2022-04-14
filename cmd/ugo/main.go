// Copyright (c) 2020-2022 Ozan Hacıbekiroğlu.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

//go:build !js
// +build !js

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"

	"github.com/ozanh/ugo"
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
	isMultiline    bool
	noOptimizer    bool
	traceEnabled   bool
	traceParser    bool
	traceOptimizer bool
	traceCompiler  bool
)

var (
	initialSuggLen int
)

var scriptGlobals = &ugo.SyncMap{
	Map: ugo.Map{
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
	multiline    string
	werr         prompt.ConsoleWriter
	wout         prompt.ConsoleWriter
	stdout       io.Writer
	commands     map[string]func()
}

func newREPL(ctx context.Context, stdout io.Writer, cw prompt.ConsoleWriter) *repl {
	moduleMap := ugo.NewModuleMap()
	moduleMap.AddBuiltinModule("time", ugotime.Module).
		AddBuiltinModule("strings", ugostrings.Module).
		AddBuiltinModule("fmt", ugofmt.Module).
		AddBuiltinModule("json", ugojson.Module)

	opts := ugo.CompilerOptions{
		ModulePath:        "(repl)",
		ModuleMap:         moduleMap,
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
		werr:   cw,
		wout:   cw,
		stdout: stdout,
	}

	r.commands = map[string]func(){
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
		".exit":          func() { os.Exit(0) },
	}
	return r
}

func (r *repl) cmdBytecode() {
	_, _ = fmt.Fprintf(r.stdout, "%s\n", r.lastBytecode)
}

func (r *repl) cmdBuiltins() {
	builtins := make([]string, len(ugo.BuiltinsMap))

	for k, v := range ugo.BuiltinsMap {
		builtins[v] = fmt.Sprint(ugo.BuiltinObjects[v].TypeName(), ":", k)
	}
	_, _ = fmt.Fprintln(r.stdout, strings.Join(builtins, "\n"))
}

func (*repl) cmdGC() { runtime.GC() }

func (r *repl) cmdGlobals() {
	_, _ = fmt.Fprintf(r.stdout, "%+v\n", r.eval.Globals)
}

func (r *repl) cmdGlobalsVerbose() {
	_, _ = fmt.Fprintf(r.stdout, "%#v\n", r.eval.Globals)
}

func (r *repl) cmdLocals() {
	_, _ = fmt.Fprintf(r.stdout, "%+v\n", r.eval.Locals)
}

func (r *repl) cmdLocalsVerbose() {
	fmt.Fprintf(r.stdout, "%#v\n", r.eval.Locals)
}

func (r *repl) cmdReturn() {
	_, _ = fmt.Fprintf(r.stdout, "%#v\n", r.lastResult)
}

func (r *repl) cmdReturnVerbose() {
	if r.lastResult != nil {
		_, _ = fmt.Fprintf(r.stdout,
			"GoType:%[1]T, TypeName:%[2]s, Value:%#[1]v\n",
			r.lastResult, r.lastResult.TypeName())
	} else {
		_, _ = fmt.Fprintln(r.stdout, "<nil>")
	}
}

func (r *repl) cmdReset() {
	grepl = newREPL(r.ctx, r.stdout, r.wout)
}

func (r *repl) cmdSymbols() {
	_, _ = fmt.Fprintf(r.stdout, "%v\n", r.eval.Opts.SymbolTable.Symbols())
}

func (r *repl) cmdMemoryStats() {
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
}

func (r *repl) cmdModulesCache() {
	_, _ = fmt.Fprintf(r.stdout, "%v\n", r.eval.ModulesCache)
}

func (r *repl) writeErrorStr(msg string) {
	r.werr.SetColor(prompt.Red, prompt.DefaultColor, true)
	r.werr.WriteStr(msg)
	_ = r.werr.Flush()
}

func (r *repl) writeStr(msg string) {
	r.wout.SetColor(prompt.Green, prompt.DefaultColor, false)
	r.wout.WriteStr(msg)
	_ = r.wout.Flush()
}

func (r *repl) executor(line string) {
	switch {
	case line == "":
		if !isMultiline {
			return
		}
	case line[0] == '.':
		if fn, ok := r.commands[line]; ok {
			fn()
			return
		}
	case strings.HasSuffix(line, "\\"):
		isMultiline = true
		r.multiline += line[:len(line)-1] + "\n"
		return
	}
	r.executeScript(line)
}

func (r *repl) executeScript(line string) {
	defer func() {
		isMultiline = false
		r.multiline = ""
	}()

	var err error
	r.lastResult, r.lastBytecode, err = r.eval.Run(r.ctx, []byte(r.multiline+line))
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
				prompt.Suggest{
					Text:        s.Name,
					Description: string(s.Scope) + " variable",
				},
			)
		}
	}
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

func completer(in prompt.Document) []prompt.Suggest {
	w := in.GetWordBeforeCursorWithSpace()
	return prompt.FilterHasPrefix(suggestions, w, true)
}

var suggestions = []prompt.Suggest{
	// Commands
	{Text: ".bytecode", Description: "Print Bytecode"},
	{Text: ".builtins", Description: "Print Builtins"},
	{Text: ".reset", Description: "Reset"},
	{Text: ".locals", Description: "Print Locals"},
	{Text: ".locals+", Description: "Print Locals (verbose)"},
	{Text: ".globals", Description: "Print Globals"},
	{Text: ".globals+", Description: "Print Globals (verbose)"},
	{Text: ".return", Description: "Print Last Return Result"},
	{Text: ".return+", Description: "Print Last Return Result (verbose)"},
	{Text: ".modules_cache", Description: "Print Modules Cache"},
	{Text: ".memory_stats", Description: "Print Memory Stats"},
	{Text: ".gc", Description: "Run Go GC"},
	{Text: ".symbols", Description: "Print Symbols"},
	{Text: ".exit", Description: "Exit"},
}

func init() {
	// add builtins to suggestions
	for k := range ugo.BuiltinsMap {
		suggestions = append(suggestions,
			prompt.Suggest{
				Text:        k,
				Description: "Builtin " + k,
			},
		)
	}

	for tok := token.Question + 3; tok.IsKeyword(); tok++ {
		s := tok.String()
		suggestions = append(suggestions, prompt.Suggest{
			Text:        s,
			Description: "keyword " + s,
		})
	}
	initialSuggLen = len(suggestions)
}

func newPrompt(
	executor func(s string),
	w io.Writer,
	poptions ...prompt.Option,
) *prompt.Prompt {

	_, _ = fmt.Fprintln(w, "Copyright (c) 2020 Ozan Hacıbekiroğlu")
	_, _ = fmt.Fprintln(w, "License: MIT")
	_, _ = fmt.Fprintln(w, "Press Ctrl+D to exit or use .exit command")
	_, _ = fmt.Fprint(w, logo)

	options := []prompt.Option{
		prompt.OptionPrefix(promptPrefix),
		prompt.OptionHistory([]string{
			"a := 1",
			"sum := func(...a) { total:=0; for v in a { total+=v }; return total }",
			"func(a, b){ return a*b }(2, 3)",
			`println("")`,
			`var (x, y, z); if x { y } else { z }`,
			`var (x, y, z); x ? y : z`,
			`for i := 0; i < 3; i++ { }`,
			`m := {}; for k,v in m { printf("%s:%v\n", k, v) }`,
			`try { } catch err { } finally { }`,
		}),
		prompt.OptionLivePrefix(func() (string, bool) {
			if isMultiline {
				return promptPrefix2, true
			}
			return "", false
		}),
		prompt.OptionTitle(title),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
	}

	options = append(options, poptions...)
	return prompt.New(executor, completer, options...)
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

	if _, err = os.Stat(filePath); err != nil {
		return
	}
	return
}

func executeScript(ctx context.Context, scr []byte, traceOut io.Writer) error {
	opts := ugo.DefaultCompilerOptions
	if traceEnabled {
		opts.Trace = traceOut
		opts.TraceParser = traceParser
		opts.TraceCompiler = traceCompiler
		opts.TraceOptimizer = traceOptimizer
	}

	opts.ModuleMap = ugo.NewModuleMap().
		AddBuiltinModule("time", ugotime.Module).
		AddBuiltinModule("strings", ugostrings.Module).
		AddBuiltinModule("fmt", ugofmt.Module)

	bc, err := ugo.Compile(scr, opts)
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

func checkErr(err error, f func()) {
	if err == nil {
		return
	}

	defer os.Exit(1)
	_, _ = fmt.Fprintln(os.Stderr, err.Error())
	if f != nil {
		f()
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

		var script []byte
		if filePath == "-" {
			script, err = ioutil.ReadAll(os.Stdin)
		} else {
			script, err = ioutil.ReadFile(filePath)
		}

		checkErr(err, cancel)
		err = executeScript(ctx, script, os.Stdout)
		checkErr(err, cancel)
		return
	}

	defer handlePromptExit()

	cw := prompt.NewStdoutWriter()
	grepl = newREPL(ctx, os.Stdout, cw)
	newPrompt(
		func(s string) { grepl.executor(s) },
		os.Stdout,
		prompt.OptionWriter(cw),
	).Run()
}

// Workaround for following issue.
// https://github.com/c-bata/go-prompt/issues/228
func handlePromptExit() {
	if runtime.GOOS != "linux" {
		return
	}

	if _, err := exec.LookPath("/bin/stty"); err != nil {
		return
	}

	rawModeOff := exec.Command("/bin/stty", "-raw", "echo")
	rawModeOff.Stdin = os.Stdin
	_ = rawModeOff.Run()
}
