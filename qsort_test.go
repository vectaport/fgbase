package flowgraph

import (
	"math/rand"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

var bushelCnt int64

type bushel struct {
	Slic []int
	Orig []int
	depth int64
	bushelID int64
}

// borrowed from Golang 1.4.2 sort example, copyright notice in flowgraph/GO-LICENSE
func (a bushel) Len() int           { return len(a.Slic) }
func (a bushel) Swap(i, j int)      { a.Slic[i], a.Slic[j] = a.Slic[j], a.Slic[i] }
func (a bushel) Less(i, j int) bool { return a.Slic[i] < a.Slic[j] }

func (a bushel) SubSlice(n, m int) Datum {
	a.Slic = a.Slic[n:m]
	a.depth += 1
	return a
}

func (a bushel) Slice() []int {
	return a.Slic
}

func (a bushel) SliceSorted() bool {
	l := len(a.Slic)
	for i:= 0; i<l-1; i++ {
		if a.Slic[i] > a.Slic[i+1] {
			return false
	}
	}
	return true
}

func (a bushel) Original() []int {
	return a.Orig
}

func (a bushel) OriginalSorted() bool {
	l := len(a.Orig)
	for i:= 0; i<l-1; i++ {
		if a.Orig[i] > a.Orig[i+1] {
			return false
		}
	}
	return true
}

func (a bushel) Depth() int64 { 
	return a.depth
}

func (a bushel) ID() int64 {
	return a.bushelID
}

func tbiRand(pow2 uint) RecursiveSort {
	var s bushel
	s.bushelID = atomic.AddInt64(&bushelCnt, 1)-1
	n := rand.Intn(1<<pow2)+1
	l := rand.Intn(n)
	for i:=0; i<l; i++ {
		s.Orig = append(s.Orig, rand.Intn(l))
	}
	s.Slic = s.Orig
	return s
}

func tbi(x Edge, pow2 uint) Node {

	node := MakeNode("tbi", nil, []*Edge{&x}, nil,
		func(n *Node) { x.Val = tbiRand(pow2) })
	return node
}

func tbo(a Edge) Node {

	node := MakeNode("tbo", []*Edge{&a}, nil, nil, 
		func(n *Node) {
			switch v := a.Val.(type) {
			case RecursiveSort: {
				if v.OriginalSorted() { n.Tracef("END for id=%d, depth=%d, len=%d\n", v.ID(), v.Depth(), v.Len()) }
				n.Tracef("Original(%p) sorted %t, Slice sorted %t, depth=%d, id=%d, len=%d, poolsz=%d, ratio = %d\n", v.Original(), v.OriginalSorted(), v.SliceSorted(), v.Depth(), v.ID(), len(v.Original()), PoolQsort.Size(), len(v.Original())/(1+int(v.Depth())))
			}
			default: {
				n.Tracef("not of type RecursiveSort\n")
			}
			}})
	return node
}

func TestQsort(t *testing.T) {

	poolSz := int32(64)
	numCore := runtime.NumCPU()-1
	sec := 1
	pow2 := uint(20)
	post := false
	runtime.GOMAXPROCS(numCore)
	PostDump = post

	TraceLevel = V
	TraceSeconds = true

	e,n := MakeGraph(2, int32(poolSz+2))

	n[0] = tbi(e[0], pow2)
	n[1] = tbo(e[1])

	p := FuncQsort(e[0], e[1], poolSz, 1)
	copy(n[2:poolSz+2], p.Nodes())

	RunAll(n, time.Duration(sec)*time.Second)

}
