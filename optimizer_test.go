package ugo_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ozanh/ugo/parser"
	"github.com/stretchr/testify/require"

	. "github.com/ozanh/ugo"
)

func TestOptimizer(t *testing.T) {
	opt, err := newOptimizer([]byte(`
	var (aa = 1<<2, bb=13*2, str="ozan")
	var (xx=1*2, yy)
	var zz
	var mm = 58*2
	a:=13*(1+-5)+0*2-5
	if 10>2 {

	} else if true+1 {

	}
	f := func() {
		return 5+6
	}
	y := ""&&0?"abc":"dec"
	arr := [5+f(), 10*2]
	o := 1<<2+3+9
	`))
	require.NoError(t, err)
	opt.Optimize()
	fmt.Printf("Duration:%s\n", opt.Duration())
}

func newOptimizer(script []byte) (*SimpleOptimizer, error) {
	fileSet := parser.NewFileSet()
	srcFile := fileSet.AddFile("main", -1, len(script))
	p := parser.NewParser(srcFile, script, nil)
	pf, err := p.ParseFile()
	if err != nil {
		return nil, err
	}
	return NewOptimizer(context.Background(), pf, true, true, 1<<8-1, nil), nil
}
