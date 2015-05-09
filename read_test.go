package flowgraph

import (
	"flag"
	"os"
	"testing"
	"time"
)

func tboRead(a Edge) Node {

	node := MakeNode("tbo", []*Edge{&a}, nil, nil, nil)
	return node
}

func TestRead(t *testing.T) {

	var check = func (e error) {
		if e != nil {
			StderrLog.Printf("%v\n", e)
			os.Exit(1)
		}
	}
		
	flag.Parse()
	if len(flag.Args()) == 0  { 
		flag.Usage()
		os.Exit(1)
	}
	fileName := flag.Arg(0)

	TraceLevel = V

	f, err := os.Open(fileName)
	check(err)

	e,n := MakeGraph(1,2)

	n[0] = FuncRead(e[0], f)
	n[1] = tboRead(e[0])

	RunAll(n, 2*time.Second)

}

