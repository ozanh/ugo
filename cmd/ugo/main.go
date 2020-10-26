package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/token"
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
	initialSugLen int
)

func init() {
	var trace string
	flag.StringVar(&trace, "trace", "", `comma separated units: -trace parser,optimizer,compiler`)
	flag.BoolVar(&noOptimizer, "no-optimizer", false, `disable optimization`)
	flag.Parse()
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
}

type repl struct {
	ctx          context.Context
	eval         *ugo.Eval
	lastBytecode *ugo.Bytecode
	lastReturn   ugo.Object
	multiline    string
	werr         prompt.ConsoleWriter
	wout         prompt.ConsoleWriter
}

func newREPL(ctx context.Context) *repl {
	moduleMap := ugo.NewModuleMap()
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
	if traceEnabled {
		opts.Trace = os.Stdout
	}
	r := &repl{
		ctx:  ctx,
		eval: ugo.NewEval(opts, nil),
		werr: prompt.NewStdoutWriter(),
		wout: prompt.NewStdoutWriter(),
	}
	return r
}

func (r *repl) writeErrorStr(msg string) {
	r.werr.SetColor(prompt.Red, prompt.DefaultColor, true)
	r.werr.WriteStr(msg)
	r.werr.Flush()
}

func (r *repl) writeStr(msg string) {
	r.wout.SetColor(prompt.Green, prompt.DefaultColor, false)
	r.wout.WriteStr(msg)
	r.wout.Flush()
}

func (r *repl) executor(line string) {
	switch {
	case line == "":
		return
	case line[0] == '.' && len(line) > 1:
		switch line[1] {
		case 'b':
			switch line {
			case ".bytecode":
				fmt.Printf("%s\n", r.lastBytecode)
				return
			case ".builtins":
				builtins := make([]string, len(ugo.BuiltinsMap))
				for k, v := range ugo.BuiltinsMap {
					builtins[v] = fmt.Sprint(ugo.BuiltinObjects[v].TypeName(), ":", k)
				}
				fmt.Print(strings.Join(builtins, "\n"), "\n")
				return
			}
		case 'g':
			switch line {
			case ".gc":
				runtime.GC()
				return
			case ".globals":
				fmt.Printf("%+v\n", r.eval.Globals)
				return
			case ".globals+":
				fmt.Printf("%#v\n", r.eval.Globals)
				return
			}
		case 'l':
			switch line {
			case ".locals":
				fmt.Printf("%+v\n", r.eval.Locals)
				return
			case ".locals+":
				fmt.Printf("%#v\n", r.eval.Locals)
				return
			}
		case 'r':
			switch line {
			case ".return":
				fmt.Printf("%#v\n", r.lastReturn)
				return
			case ".return+":
				if r.lastReturn != nil {
					fmt.Printf("GoType:%[1]T, TypeName:%[2]s, Value:%#[1]v\n",
						r.lastReturn, r.lastReturn.TypeName())
				} else {
					fmt.Println("<nil>")
				}
				return
			case ".reset":
				*r = *newREPL(r.ctx)
				return
			}
		case 's':
			switch line {
			case ".symbols":
				fmt.Printf("%v\n", r.eval.Opts.SymbolTable.Symbols())
				return
			}
		case 'm':
			switch line {
			case ".memory_stats":
				writeMemStats(os.Stdout)
				return
			case ".modules_cache":
				fmt.Printf("%v\n", r.eval.ModulesCache)
				return
			}
		case 'e':
			switch line {
			case ".exit":
				os.Exit(0)
			}
		}
	case strings.HasSuffix(line, "\\"):
		isMultiline = true
		r.multiline += line[:len(line)-1] + "\n"
		return
	}
	defer func() {
		isMultiline = false
		r.multiline = ""
	}()
	var err error
	r.lastReturn, r.lastBytecode, err = r.eval.Run(r.ctx, []byte(r.multiline+line))
	if err != nil {
		r.writeErrorStr(fmt.Sprintf("\n%+v\n", err))
		return
	}
	if err != nil {
		r.writeErrorStr(fmt.Sprintf("VM:\n     %+v\n", err))
		return
	}
	switch v := r.lastReturn.(type) {
	case ugo.String:
		r.writeStr(fmt.Sprintf("%q\n", string(v)))
	case ugo.Char:
		r.writeStr(fmt.Sprintf("%q\n", rune(v)))
	case ugo.Bytes:
		r.writeStr(fmt.Sprintf("%v\n", []byte(v)))
	default:
		r.writeStr(fmt.Sprintf("%v\n", r.lastReturn))
	}

	symbols := r.eval.Opts.SymbolTable.Symbols()
	suggestions = suggestions[:initialSugLen]
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

// writeMemStats writes the formatted current, total and OS memory being used. As well as the number
// of garbage collection cycles completed.
func writeMemStats(w io.Writer) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	_, _ = fmt.Fprintf(w, "Go Memory Stats see: https://golang.org/pkg/runtime/#MemStats\n\n")
	_, _ = fmt.Fprintf(w, "HeapAlloc = %s", humanFriendlySize(m.HeapAlloc))
	_, _ = fmt.Fprintf(w, "\tHeapObjects = %v", m.HeapObjects)
	_, _ = fmt.Fprintf(w, "\tSys = %s", humanFriendlySize(m.Sys))
	_, _ = fmt.Fprintf(w, "\tNumGC = %v\n", m.NumGC)
}

func humanFriendlySize(b uint64) string {
	if b < 1024 {
		return fmt.Sprint(strconv.FormatUint(b, 10), " bytes")
	}
	if b >= 1024 && b < 1024*1024 {
		return fmt.Sprint(strconv.FormatFloat(float64(b)/1024, 'f', 1, 64), " KiB")
	}
	return fmt.Sprint(strconv.FormatFloat(float64(b)/1024/1024, 'f', 1, 64), " MiB")
}

func completer(in prompt.Document) []prompt.Suggest {
	w := in.GetWordBeforeCursorWithSpace()
	return prompt.FilterContains(suggestions, w, true)
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
	{Text: ".return", Description: "Print Last Return"},
	{Text: ".return+", Description: "Print Last Return (verbose)"},
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
	initialSugLen = len(suggestions)
}

func livePrefix() (string, bool) {
	if isMultiline {
		return promptPrefix2, true
	}
	return "", false
}

func main() {
	fmt.Println("Copyright (c) 2020 Ozan Hacıbekiroğlu")
	fmt.Println("License: MIT")
	fmt.Println("Press Ctrl+D to exit or use .exit command")
	fmt.Println(logo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := newREPL(ctx)
	p := prompt.New(
		r.executor,
		completer,
		prompt.OptionPrefix(promptPrefix),
		prompt.OptionHistory([]string{
			"a := 1",
			"sum := func(a...) { total:=0; for v in a { total+=v }; return total }",
			"func(a, b){ return a*b }(2, 3)",
			`println("")`,
			`var (x, y, z); if x { y } else { z }`,
			`var (x, y, z); x ? y : z`,
			`for i := 0; i < 3; i++ { }`,
			`m := {}; for k,v in m { printf("%s:%v\n", k, v) }`,
			`try { } catch err { } finally { }`,
		}),
		prompt.OptionLivePrefix(livePrefix),
		prompt.OptionTitle(title),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
	)
	p.Run()
}
