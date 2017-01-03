package regexp

import (		
       "github.com/vectaport/flowgraph"
)      			

func barFire (n *flowgraph.Node) {	 
	a := n.Srcs[0] 		 
	b := n.Srcs[1] 		 
	x := n.Dsts[0]
	sink := n.Aux.(bool)

	av := a.SrcGet()
        as := av.(Search).Curr
	ast := av.(Search).State
	bv := b.SrcGet()
        bs := bv.(Search).Curr
	bst := bv.(Search).State
	
        if ast==Done && bst==Done {
		if sink {
			return
		}
                x.DstPut(Search{})
                return
        }

        if ast==Done {
	        x.DstPut(bv)
		return
        }

        if bst==Done {
	        x.DstPut(av)
		return
        }

	if len(as)>len(bs) {
  		x.DstPut(av)
        }
        x.DstPut(bv)
}

// FuncBar waits for both inputs and returns the one that matches the shortest string.
// Returns nil if no match at all.
func FuncBar(a,b flowgraph.Edge, x flowgraph.Edge, sink bool) flowgraph.Node {
	
	node := flowgraph.MakeNode("bar", []*flowgraph.Edge{&a, &b}, []*flowgraph.Edge{&x}, nil, barFire)
	node.Aux = sink
	return node
	
}
	
