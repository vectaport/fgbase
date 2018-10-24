package fgbase

import ()

// Sinker consumes wavefronts of values one at a time forever
type Sinker interface {
	Sink(source []interface{})
}

// SinkStats to use with Sinker interface
type SinkStats struct {
	Cnt int
	Sum int
}

func (s *SinkStats) Sink(v []interface{}) {
	for i := range v {
		s.Cnt++
		s.Sum += Int(v[i])
	}

}

func SinkFire(n *Node) error {
	a := n.Srcs[0]
	v := a.SrcGet()

	if v, ok := v.(error); ok && v.Error() == "EOF" {
		a.Flow = false
		return v
	}

	if s, ok := n.Aux.(Sinker); ok {
		s.Sink([]interface{}{v})
	}
	return nil
}

// FuncSink sinks a single value.
func FuncSink(a Edge) Node {

	node := MakeNode("sink", []*Edge{&a}, nil, nil, SinkFire)
	node.Aux = &SinkStats{0, 0}
	return node
}
