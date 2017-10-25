package grid

import (
	"github.com/vectaport/flowgraph"
)

func gridRdy (n *flowgraph.Node) bool {
	srcn := n.Srcs[0] 		 
	srce := n.Srcs[1] 		 
	srcs := n.Srcs[2] 		 
	srcw := n.Srcs[3]
	/*
	dstn := n.Dsts[0] 		 
	dste := n.Dsts[1] 		 
	dsts := n.Dsts[2] 		 
	dstw := n.Dsts[3]
	*/
	
        return srcn.SrcRdy(n) || srce.SrcRdy(n) || srcs.SrcRdy(n) || srcw.SrcRdy(n)
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

	if srcn.SrcRdy(n) && dsts.DstRdy(n) { dsts.DstPut(srcn.SrcGet()) }
	if srce.SrcRdy(n) && dstw.DstRdy(n) { dstw.DstPut(srce.SrcGet()) }
	if srcs.SrcRdy(n) && dsts.DstRdy(n) { dstn.DstPut(srcs.SrcGet()) }
	if srcw.SrcRdy(n) && dstw.DstRdy(n) { dste.DstPut(srcw.SrcGet()) }
	return
}

// FuncGrid coordinates with its neighbors
func FuncGrid(srcn,srce,srcs,srcw flowgraph.Edge, dstn,dste,dsts,dstw flowgraph.Edge) flowgraph.Node {
	
	node := flowgraph.MakeNode("grid", []*flowgraph.Edge{&srcn, &srce, &srcs, &srcw}, []*flowgraph.Edge{&dstn, &dste, &dsts, &dstw},
		gridRdy, gridFire)
	return node
	
}
	
