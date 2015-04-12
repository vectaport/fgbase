package flowgraph

import (		
	"bufio"
)      			

func srcFire (n *Node) {	 
	x := n.Dsts[0] 		 
	var ok error
	x.Val, ok = x.Aux.(*bufio.Reader).ReadString('\n')
	if ok != nil {
		close(x.Data)
		x.Data = nil
	}
}

// FuncSrc reads a value from a network connection
func FuncSrc(x Edge, r *bufio.Reader) {

	node := MakeNode("src", nil, []*Edge{&x}, nil, srcFire)
	x.Aux = r
	node.Run()
	
}
	
