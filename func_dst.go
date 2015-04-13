package flowgraph

import (		
	"bufio"
	"fmt"
	"io"
)      			

func dstFire (n *Node) {	 
	a := n.Srcs[0] 		 
	rw := a.Aux.(*bufio.ReadWriter)
	var err error

	// wait until a newline response has been read
	if a.Val == nil  {
		_, err = rw.ReadString('\n')
		if err != nil {
			n.Errorf("%v", err)
			close(a.Ack)
			a.Ack = nil
			return
		}
	} else {
		a.Val = nil
	}
	
	// write this string and flush it out of the buffer
	_, err = rw.WriteString(fmt.Sprintf("%v\n", a.Val))
	if err != nil {
		n.Errorf("%v", err)
		close(a.Ack)
		a.Ack = nil
		return
	}
	rw.Flush()
}

// FuncDst writes a value to an io.Reader and waits for a new-line terminated response.
func FuncDst(a Edge, rw io.ReadWriter) {
	
	node := MakeNode("dst", []*Edge{&a}, nil, nil, dstFire)
	reader := bufio.NewReader(rw)
	writer := bufio.NewWriter(rw)
	a.Aux = bufio.NewReadWriter(reader, writer)
	a.Val = true // initial state
	node.Run()
	
}
	
