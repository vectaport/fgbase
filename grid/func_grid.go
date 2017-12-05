package grid

import (
        "math/rand"
	
	"github.com/vectaport/flowgraph"
)

type compassDir int
const (
       nor = iota
       eas
       sou
       wes
)

type auxStruct struct {
        Cnt [wes+1] int
	rdy [wes+1] bool
	dir [wes+1] compassDir
}

func randDir() compassDir {
	return compassDir(rand.Intn(4))
}

func gridRdy (n *flowgraph.Node) bool {
        var as auxStruct
        if n.Aux == nil {
	        as = auxStruct{}
	} else {
		as = n.Aux.(auxStruct)
	}

        for i:= range as.rdy {
	        newDir := randDir()
	        as.rdy[i] = n.Srcs[i].SrcRdy(n) && n.Dsts[newDir].DstRdy(n)
		as.dir[i] = newDir
	}

        n.Aux = as
	
        return as.rdy[0] || as.rdy[1] || as.rdy[2] || as.rdy[3]
}

func gridFire (n *flowgraph.Node) {	 
	as := n.Aux.(auxStruct)

	for i:= range as.rdy {
		if as.rdy[i] {
		        n.Dsts[as.dir[i]].DstPut(n.Srcs[i].SrcGet()); as.Cnt[as.dir[i]]++
		}
	}

        n.Aux = as
	return
}

// FuncGrid coordinates with its neighbors
func FuncGrid(srcn,srce,srcs,srcw flowgraph.Edge, dstn,dste,dsts,dstw flowgraph.Edge) flowgraph.Node {
	
	node := flowgraph.MakeNode("grid", []*flowgraph.Edge{&srcn, &srce, &srcs, &srcw}, []*flowgraph.Edge{&dstn, &dste, &dsts, &dstw},
		gridRdy, gridFire)
	return node
	
}
	
