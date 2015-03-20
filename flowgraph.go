package flowgraph

import (
	"sync/atomic"
	"fmt"
)

var node_id int64 = 0
var global_exec_cnt int64 = 0

var Debug bool = false
var GlobalExecCnt bool = false
var Indent bool = false

type Datum interface{}

type rdy_func func(*Node) bool

type Edge struct {

	// values shared by upstream and downstream Node
	Data chan Datum
	Data_init bool
	Init_val Datum
	Ack chan bool
	Ack_init bool
	Name string

	// values unique to upstream and downstream Node
	Rdy bool
	Val Datum

}


type Node struct {
	Id int64
	Name string
	Cnt int64
	Srcs []*Edge
	Dsts []*Edge
	RdyFunc rdy_func
}

func MakeEdge(name string, data_init, ack_init bool, init_val Datum) Edge {
	var e Edge
	e.Data = make(chan Datum)
	e.Data_init = data_init
	e.Init_val = init_val
	e.Ack = make(chan bool)
	e.Ack_init = ack_init
	e.Name = name
	return e
}

func (e *Edge) InitSrc(n *Node) {
	e.Rdy = e.Data_init
}

func (e *Edge) InitDst(n *Node) {
	e.Rdy = e.Ack_init
}

func MakeNode(nm string, srcs, dsts []*Edge, ready rdy_func) Node {
	var n Node
	i := atomic.AddInt64(&node_id, 1)
	n.Id = i-1
	n.Name = nm
	n.Cnt = -1
	n.Srcs = srcs
	n.Dsts = dsts
	for i := range n.Srcs {
		n.Srcs[i].InitSrc(&n)
		n.Srcs[i].Val = srcs[i].Init_val
	}
	for i := range n.Dsts {
		n.Dsts[i].InitDst(&n)
	}
	n.RdyFunc = ready
	return n
}

func prefix_varlist(n Node) (format string, varlist []interface {}) {
	var varl [] interface {}
	varl = append(varl, n.Name)
	varl = append(varl, n.Id)
	if (n.Cnt>=0) {
		varl = append(varl, n.Cnt)
	} else {
		varl = append(varl, "*")
	}
	var f string
	if (Indent) {
		for i := int64(0);i<n.Id;i++ {
			f += "\t"
		}
	}
	f += "%s(%d:%v) "
	return f,varl
}

func (n Node) Printf(format string, v ...interface{}) {
	if (!Debug || format=="select\n") {
		return
	}
	newfmt,varlist := prefix_varlist(n)
	newfmt += format
	varlist = append(varlist, v...)
	fmt.Printf(newfmt, varlist...)
}

func (n Node) PrintStatus(done bool) {
	if (!done && !Debug) {return}
	newfmt,varlist := prefix_varlist(n)
	for i := range n.Srcs {
		varlist = append(varlist, n.Srcs[i].Name)
		var inval interface {}
		if (n.Srcs[i].Rdy) {
			inval = n.Srcs[i].Val
		} else {
			inval = "{}"
		}
		varlist = append(varlist, inval)
		if (i!=0) { newfmt += "," }
		newfmt += "%s=%v"
	}
	newfmt += ":"
	for i := range n.Dsts {
		if (done) {
			varlist = append(varlist, n.Dsts[i].Name)
			if (n.Dsts[i].Val != nil) {
				varlist = append(varlist, n.Dsts[i].Val)
			} else {
				varlist = append(varlist, "{}")
			}
		} else {
			varlist = append(varlist, n.Dsts[i].Name+".Ack")
			varlist = append(varlist, n.Dsts[i].Rdy)
		}
		if (i!=0) { newfmt += "," }
		newfmt += "%s=%v"
	}
	newfmt += "\n"
	fmt.Printf(newfmt, varlist...)
}

func (n Node) PrintVals() { n.PrintStatus(true) }

func (n *Node) ExecCnt() {
	if (GlobalExecCnt) {
		c := atomic.AddInt64(&global_exec_cnt, 1)
		n.Cnt = c-1
	} else {
		n.Cnt = n.Cnt+1
	}
}

func (n *Node) Rdy() bool {
	n.PrintStatus(false)
	for i := range n.Srcs {
		if !n.Srcs[i].Rdy { return false }
	}
	for i := range n.Dsts {
		if !n.Dsts[i].Rdy { return false }
	}
	n.ExecCnt();
	return true
}

func Sink(a Datum) () {
}

