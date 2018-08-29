package flowgraph

import ()

func allOfFire(n *Node) {
	var a []interface{}
	a = make([]interface{}, len(n.Srcs))
	t := n.Aux.(Transformer)
	for i, _ := range a {
		a[i] = n.Srcs[i].SrcGet()
	}
	x, _ := t.Transform(n.Owner, a...)
	for i, _ := range x {
		n.Dsts[i].DstPut(x[i])
	}
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
