package flowgraph

import (
)

func arbit_func (n *Node) {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	if(a.Rdy && !b.Rdy || a.Rdy && !a.Aux.(bool)) {
		a.Aux = true
		x.Val = a.Val
		b.Nack = true
	} else if (b.Rdy) {
		a.Aux = false
		x.Val = b.Val
		a.Nack = true
	}
}

func arbit_rdy (n *Node) bool {
	return (n.Srcs[0].Rdy || n.Srcs[1].Rdy) && n.Dsts[0].Rdy
}

// Arbiter goroutine
func FuncArbit(a, b, x Edge) {

	node := MakeNode("arbit", []*Edge{&a, &b}, []*Edge{&x}, arbit_rdy, arbit_func)
	a.Aux = false // aux value that means "a" won the arbitration last
	node.Run()

}
