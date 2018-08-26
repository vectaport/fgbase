package flowgraph

import (
	"bufio"
	"fmt"
	"io"
)

type irw struct {
	Initialized bool
	RW          *bufio.ReadWriter
}

func dstFire(n *Node) {
	a := n.Srcs[0]
	s := n.Aux.(*irw)
	rw := s.RW
	var err error

	// read ack
	a.Flow = true
	if s.Initialized {
		_, err = rw.ReadString('\n')
		if err != nil {
			n.LogError("%v", err)
			close(a.Ack)
			a.Ack = nil
			return
		}
	} else {
		s.Initialized = true
	}

	// write data
	_, err = rw.WriteString(fmt.Sprintf("%v\n", a.SrcGet()))
	if err != nil {
		n.LogError("%v", err)
		close(a.Ack)
		a.Ack = nil
		return
	}
	rw.Flush()
}

// FuncDst writes data and waits for an acknowledging '\n'.
func FuncDst(a Edge, rw io.ReadWriter) Node {

	node := MakeNode("dst", []*Edge{&a}, nil, nil, dstFire)
	reader := bufio.NewReader(rw)
	writer := bufio.NewWriter(rw)
	node.Aux = &irw{RW: bufio.NewReadWriter(reader, writer)}
	return node

}
