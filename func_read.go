package flowgraph

import (		
	"bufio"
	"io"
	"os"
)      			

func readFire (n *Node) {	 
	x := n.Dsts[0] 		 
	r := n.Aux.(*bufio.Reader)
	var err error

	// read data string
	x.Val, err = r.ReadString('\n')
	if err != nil {
		if err==io.EOF {
			os.Exit(0)
		}
		n.LogError("%v", err)
		x.CloseData()
	}

}

// FuncRead reads a data value from a Reader
func FuncRead(x Edge, r io.Reader) Node {

	node := MakeNode("read", nil, []*Edge{&x}, nil, readFire)
	node.Aux = bufio.NewReader(r)
	return node
	
}
	
