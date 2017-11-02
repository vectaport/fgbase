package grid

import (
	"github.com/vectaport/flowgraph"
)

type auxStruct struct {
        Ncnt, Ecnt, Scnt, Wcnt int
	rdy []bool
}

func gridRdy (n *flowgraph.Node) bool {
	srcn := n.Srcs[0] 		 
	srce := n.Srcs[1] 		 
	srcs := n.Srcs[2] 		 
	srcw := n.Srcs[3]
	dstn := n.Dsts[0] 		 
	dste := n.Dsts[1] 		 
	dsts := n.Dsts[2] 		 
	dstw := n.Dsts[3]

        var as auxStruct
        if n.Aux == nil {
	        as = auxStruct{}
		as.rdy = make([]bool, 4)
	} else {
		as = n.Aux.(auxStruct)
	}

        as.rdy[0] = srcn.SrcRdy(n)
	as.rdy[1] = srce.SrcRdy(n)
	as.rdy[2] = srcs.SrcRdy(n)
	as.rdy[3] = srcw.SrcRdy(n)
        n.Aux = as
	
        return as.rdy[0]&&dsts.DstRdy(n) || as.rdy[1]&&dstw.DstRdy(n) || as.rdy[2]&&dstn.DstRdy(n) || as.rdy[3]&&dste.DstRdy(n)
}

func gridFire (n *flowgraph.Node) {	 
	srcn := n.Srcs[0] 		 
	srce := n.Srcs[1] 		 
	srcs := n.Srcs[2] 		 
	srcw := n.Srcs[3] 		 
	dstn := n.Dsts[0] 		 
	dste := n.Dsts[1] 		 
	dsts := n.Dsts[2] 		 
	dstw := n.Dsts[3]

	as := n.Aux.(auxStruct)

	if as.rdy[0] && dsts.DstRdy(n) { dsts.DstPut(srcn.SrcGet()); as.Scnt++ }
	if as.rdy[1] && dstw.DstRdy(n) { dstw.DstPut(srce.SrcGet()); as.Wcnt++ }
	if as.rdy[2] && dstn.DstRdy(n) { dstn.DstPut(srcs.SrcGet()); as.Ncnt++ }
	if as.rdy[3] && dste.DstRdy(n) { dste.DstPut(srcw.SrcGet()); as.Ecnt++ }

        n.Aux = as
	return
}

// FuncGrid coordinates with its neighbors
func FuncGrid(srcn,srce,srcs,srcw flowgraph.Edge, dstn,dste,dsts,dstw flowgraph.Edge) flowgraph.Node {
	
	node := flowgraph.MakeNode("grid", []*flowgraph.Edge{&srcn, &srce, &srcs, &srcw}, []*flowgraph.Edge{&dstn, &dste, &dsts, &dstw},
		gridRdy, gridFire)
	return node
	
}
	
