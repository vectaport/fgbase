package flowgraph

import (		
	"bufio"
	"fmt"
	"io"
)      			

func writeFire (n *Node) {	 
	a := n.Srcs[0] 		 
	w := n.Aux.(*bufio.Writer)
	var err error

	// write data string
	_, err = w.WriteString(fmt.Sprintf("%v\n", a.Val))
	if err != nil {
		n.LogError("%v", err)
		close(a.Ack)
		a.Ack = nil
		return
	}
	w.Flush()

}

// FuncWrite writes a data value from a Writer
func FuncWrite(a Edge, w io.Writer) Node {

	node := MakeNode("write", []*Edge{&a}, nil, nil, writeFire)
	node.Aux = bufio.NewWriter(w)
	return node
	
}
	
