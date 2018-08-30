package fgbase

import (
	"fmt"
)

func printFire(n *Node) {
	a := n.Srcs[0]
	x := n.Dsts[0]
	form := n.Aux.(string)
	v := a.SrcGet()

	// print data string
	_, _ = fmt.Printf(form, v)

	x.DstPut(v)

}

// FuncPrint prints a data value using fmt.Printf
func FuncPrint(a Edge, x Edge, format string) Node {

	node := MakeNode("print", []*Edge{&a}, []*Edge{&x}, nil, printFire)
	node.Aux = format
	return node

}
