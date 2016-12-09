package regexp

import (		
       "github.com/vectaport/flowgraph"
)      			

func barFire (n *flowgraph.Node) {	 
	a := n.Srcs[0] 		 
	b := n.Srcs[1] 		 
	x := n.Dsts[0]
	sink := n.Aux.(bool)

        as := a.Val.(Search).Curr
	ast := a.Val.(Search).State
        bs := b.Val.(Search).Curr
	bst := b.Val.(Search).State
	
        if ast==Fail && bst==Fail {
		if sink {
			x.NoOut = true
			return
		}
                x.Val = Search{}
                return
        }

        if a.Val==nil {
	        x.Val = b.Val
		return
        }

        if b.Val==nil {
	        x.Val = a.Val
		return
        }

	if len(as)>len(bs) {
  		x.Val = a.Val
        }
        x.Val = b.Val
}

// FuncBar waits for both inputs and returns the one that matches the shortest string.
// Returns nil if no match at all.
func FuncBar(a,b flowgraph.Edge, x flowgraph.Edge, sink bool) flowgraph.Node {
	
	node := flowgraph.MakeNode("bar", []*flowgraph.Edge{&a, &b}, []*flowgraph.Edge{&x}, nil, barFire)
	node.Aux = sink
	return node
	
}
	
