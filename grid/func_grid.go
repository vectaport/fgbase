package grid

import (
	"github.com/vectaport/flowgraph"
)

type auxStruct struct {
        Ncnt, Ecnt, Scnt, Wcnt int
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
	
        return srcn.SrcRdy(n)&&dsts.DstRdy(n) || srce.SrcRdy(n)&&dstw.DstRdy(n) || srcs.SrcRdy(n)&&dstn.DstRdy(n) || srcw.SrcRdy(n)&&dste.DstRdy(n)
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

        if n.Aux == nil {
	        n.Aux = auxStruct{}
	}
	as := n.Aux.(auxStruct)

	if srcn.SrcRdy(n) && dsts.DstRdy(n) { dsts.DstPut(srcn.SrcGet()); as.Scnt++ }
	if srce.SrcRdy(n) && dstw.DstRdy(n) { dstw.DstPut(srce.SrcGet()); as.Wcnt++ }
	if srcs.SrcRdy(n) && dstn.DstRdy(n) { dstn.DstPut(srcs.SrcGet()); as.Ncnt++ }
	if srcw.SrcRdy(n) && dste.DstRdy(n) { dste.DstPut(srcw.SrcGet()); as.Ecnt++ }

        n.Aux = as
	return
}

// FuncGrid coordinates with its neighbors
func FuncGrid(srcn,srce,srcs,srcw flowgraph.Edge, dstn,dste,dsts,dstw flowgraph.Edge) flowgraph.Node {
	
	node := flowgraph.MakeNode("grid", []*flowgraph.Edge{&srcn, &srce, &srcs, &srcw}, []*flowgraph.Edge{&dstn, &dste, &dsts, &dstw},
		gridRdy, gridFire)
	return node
	
}
	
