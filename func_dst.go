package flowgraph

import (		
	"bufio"
	"fmt"
	"io"
)      			

type irw struct {
	Initialized bool
	RW *bufio.ReadWriter
} 

func dstFire (n *Node) {	 
	a := n.Srcs[0] 		 
	s := a.Aux.(*irw)
	rw := s.RW
	var err error

	// read ack
	if s.Initialized  {
		_, err = rw.ReadString('\n')
		if err != nil {
			n.Errorf("%v", err)
			close(a.Ack)
			a.Ack = nil
			return
		}
	} else {
		s.Initialized = true
	}
	
	// write data
	_, err = rw.WriteString(fmt.Sprintf("%v\n", a.Val))
	if err != nil {
		n.Errorf("%v", err)
		close(a.Ack)
		a.Ack = nil
		return
	}
	rw.Flush()
}

// FuncDst writes data to an io.ReadWriter and waits for an acknowledging '\n'.
func FuncDst(a Edge, rw io.ReadWriter) {
	
	node := MakeNode("dst", []*Edge{&a}, nil, nil, dstFire)
	reader := bufio.NewReader(rw)
	writer := bufio.NewWriter(rw)
	a.Aux = &irw{RW: bufio.NewReadWriter(reader, writer)}
	node.Run()
	
}
	
