package fgbase

import ()

type SinkStats struct {
	Cnt int
	Sum int
}

func sinkFire(n *Node) {
	a := n.Srcs[0]
	v := a.SrcGet()

	s := n.Aux.(SinkStats)
	if v, ok := v.(int); ok {
		n.Aux = SinkStats{s.Cnt + 1, s.Sum + v}
	} else {
		n.Aux = SinkStats{s.Cnt + 1, 0}
	}
}

// FuncSink sinks a single value.
func FuncSink(a Edge) Node {

	node := MakeNode("sink", []*Edge{&a}, nil, nil, sinkFire)
	node.Aux = SinkStats{0, 0}
	return node
}
