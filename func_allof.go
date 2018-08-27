package flowgraph

import ()

func allOfFire(n *Node) {
	a := n.Srcs[0]
	x := n.Dsts[0]
	t := n.Aux.(Transformer)
	vs, _ := t.Transform(a.SrcGet())
	x.DstPut(vs[0])
}

// FuncAllOf waits for all inputs to be ready before transforming them into all outputs
func FuncAllOf(a, x []Edge, name string, transformer Transformer) Node {

	var abuf []*Edge
	for i, _ := range a {
		abuf = append(abuf, &a[i])
	}
	var xbuf []*Edge
	for i, _ := range x {
		xbuf = append(xbuf, &x[i])
	}
	node := MakeNode(name, abuf, xbuf, nil, allOfFire)
	node.Aux = transformer
	return node
}
