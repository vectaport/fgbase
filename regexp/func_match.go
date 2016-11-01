package regexp

import (		
       "github.com/vectaport/flowgraph"
)      			

func matchFire (n *flowgraph.Node) {	 
	a := n.Srcs[0] 		 
	x := n.Dsts[0] 		 
	match := n.Aux.(string)


        if a.Val==nil {
                x.Val = nil
                return
        }

	slice := a.Val.([]byte)
	for i := range match {
                if slice[i]!=match[i] {
		        x.Val = nil
                        return
                }
        }

	x.Val = slice[len(match):]
}

// FuncMatch advances a byte slice if it matches a string, otherwise returns the empty slice
func FuncMatch(a flowgraph.Edge, x flowgraph.Edge, match string) flowgraph.Node {
	
	node := flowgraph.MakeNode("match", []*flowgraph.Edge{&a}, []*flowgraph.Edge{&x}, nil, matchFire)
	node.Aux = match
	return node
	
}
	
