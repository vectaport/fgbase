package flowgraph

import (
	"flag"
	"math/rand"
	"runtime"
	"testing"
	"sort"
	"sync/atomic"
	"time"
)

var bushelCnt int64

type bushel struct {
	depth int64
	bushelID int64
	Original []int
	Sliced []int
}

// borrowed from Golang 1.4.2 sort example, copyright notice in flowgraph/GO-LICENSE
func (a bushel) Len() int           { return len(a.Sliced) }
func (a bushel) Swap(i, j int)      { a.Sliced[i], a.Sliced[j] = a.Sliced[j], a.Sliced[i] }
func (a bushel) Less(i, j int) bool { return a.Sliced[i] < a.Sliced[j] }

func (a bushel) Sorted() bool {
	l := len(a.Sliced)
	for i:= 0; i<l-1; i++ {
		if a.Sliced[i] > a.Sliced[i+1] {
			return false
	}
	}
	return true
}

func (a bushel) SubSlice(n, m int) Datum {
	a.Sliced = a.Sliced[n:m]
	a.depth += 1
	return a
}

func (a bushel) OrigSorted() bool {
	l := len(a.Original)
	for i:= 0; i<l-1; i++ {
		if a.Original[i] > a.Original[i+1] {
			return false
		}
	}
	return true
}

func (a bushel) Orig() []int {
	return a.Original
}

func (a bushel) Slic() []int {
	return a.Sliced
}

func (a bushel) Depth() int64 { 
	return a.depth
}

func (a *bushel) DepthIncr() { 
	
	a.depth += 1
}

func (a bushel) ID() int64 {
	return a.bushelID
}

func tbiQsortRand() sort.Interface {
	var s bushel
	s.bushelID = atomic.AddInt64(&bushelCnt, 1)-1
	n := 1024*1024
	l := rand.Intn(n)
	for i:=0; i<l; i++ {
		s.Original = append(s.Original, rand.Intn(n))
	}
	s.Sliced = s.Original
	return s
}

func tbiQsort(x Edge) Node {

	node := MakeNode("tbi", nil, []*Edge{&x}, nil,
		func(n *Node) { n.Dsts[0].Val = tbiQsortRand() })
	return node
}

func tboQsort(a Edge) Node {

	node := MakeNode("tbo", []*Edge{&a}, nil, nil, 
		func(n *Node) {
			switch v := a.Val.(type) {
			case SortInterface: {
				n.Tracef("Original(%p) sorted %t, Sliced sorted %t, depth=%d, id=%d, len=%d\n", v.Orig(), v.OrigSorted(), v.Sorted(), v.Depth(), v.ID(), v.Len())
			}
			default: {
				n.Tracef("not of type SortInterface\n")
			}
			}})
	return node
}

func TestQsort(t *testing.T) {

	poolSzp := flag.Int("poolsz", 64, "qsort pool size")
	numCorep := flag.Int("numcore", 1, "num cores to use")
	flag.Parse()
	poolSz := *poolSzp
	runtime.GOMAXPROCS(*numCorep)

	TraceLevel = V

	e,n := MakeGraph(2, poolSz+2)

	n[0] = tbiQsort(e[0])
	n[1] = tboQsort(e[1])

	p := n[2:poolSz+2]
	copy(p, FuncQsort(e[0], e[1], poolSz))

	RunAll(n, 1*time.Second)

}
