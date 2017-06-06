package flowgraph

import (
)

// FuncReduce reduces a stream of data into a single empty interface.
func FuncReduce(a,x Edge, reducer func(n *Node, datum,collection interface{}) interface{}) Node {

        var reduceRdy = func (n *Node) bool {
		a := n.Srcs[0]
		x := n.Dsts[0]
		return a.SrcRdy(n) || x.DstRdy(n)
	}

	var reduceFire = func (n *Node) {
	        if a.SrcRdy(n) {
  		        n.Aux = reducer(n, a.SrcGet(), n.Aux)
		}
		if x.DstRdy(n) {
		        x.DstPut(n.Aux)
		}
	}


	node := MakeNode("reduce", []*Edge{&a}, []*Edge{&x}, reduceRdy, reduceFire)
	node.Aux = make([]string, 0)
	return node
}
