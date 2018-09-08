package fgbase

import (
)

type SinkStats struct {
	Cnt int
	Sum int
}

func sinkFire(n *Node) error {
	a := n.Srcs[0]
	v := a.SrcGet()

	if v, ok := v.(error); ok && v.Error()=="EOF" {
		return v
	}

	s := n.Aux.(SinkStats)
	if v, ok := v.(int); ok {
		n.Aux = SinkStats{s.Cnt + 1, s.Sum + v}
	} else {
		n.Aux = SinkStats{s.Cnt + 1, 0}
	}
	return nil
}

// FuncSink sinks a single value.
func FuncSink(a Edge) Node {

	node := MakeNode("sink", []*Edge{&a}, nil, nil, sinkFire)
	node.Aux = SinkStats{0, 0}
	return node
}
