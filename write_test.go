package flowgraph

import (
	"flag"
	"os"
	"testing"
	"time"
)

func tbiWrite(x Edge) Node {

	node := MakeNode("tbi", nil, []*Edge{&x}, nil, 
		func (n *Node) {
			x.Val = x.Aux
			x.Aux = x.Aux.(int) + 1
		})

	x.Aux = 0
	return node

}

func TestWrite(t *testing.T) {

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

	f, err := os.Create(fileName)
	check(err)

	e,n := MakeGraph(1,2)

	n[0] = tbiWrite(e[0])
	n[1] = FuncWrite(e[0], f)

	RunAll(n, 2*time.Second)

}

