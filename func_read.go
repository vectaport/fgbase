package flowgraph

import (		
	"bufio"
	"io"
)      			

func readFire (n *Node) {	 
	x := n.Dsts[0] 		 
	r := n.Aux.(*bufio.Reader)
	var err error

	// read data string
	x.Val, err = r.ReadString('\n')
	if err != nil {
		n.LogError("%v", err)
		for i := range *x.Data {
			close((*x.Data)[i])
			(*x.Data)[i] = nil
		}
	}

}

// FuncRead reads a data value from a Reader
func FuncRead(x Edge, r io.Reader) Node {

	node := MakeNode("read", nil, []*Edge{&x}, nil, readFire)
	node.Aux = bufio.NewReader(r)
	return node
	
}
	
