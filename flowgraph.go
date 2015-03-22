package flowgraph

import (
	"sync/atomic"
	"fmt"
)

var node_id int64 = 0
var global_exec_cnt int64 = 0

// enable debug tracing
var Debug bool = false

// indent trace by node id
var Indent bool = false

// use global execution count
var GlobalExecCnt bool = false


// empty interface for generic data flow
type Datum interface{}

type rdy_func func(*Node) bool

// flowgraph Edge (augmented channel)
type Edge struct {

	// values shared by upstream and downstream Node
	Data chan Datum
	Data_rdy_init bool
	Init_val Datum
	Ack chan bool
	Ack_rdy_init bool
	Name string

	// values unique to upstream and downstream Node
	Rdy bool
	Val Datum

}

// flowgraph Node (augmented goroutine)
type Node struct {
	Id int64
	Name string
	Cnt int64
	Srcs []*Edge
	Dsts []*Edge
	RdyFunc rdy_func
}

// return new Edge to connect two Node's
func NewEdge(name string, init_val Datum) Edge {
	var e Edge
	e.Data = make(chan Datum)
	e.Ack = make(chan bool)
	e.Init_val = init_val
	e.Data_rdy_init = init_val != nil
	e.Ack_rdy_init = init_val == nil
	e.Name = name
	return e
}

// return new Node with slices of input and output Edge's and customizable ready-testing function
func NewNode(nm string, srcs, dsts []*Edge, ready rdy_func) Node {
	var n Node
	i := atomic.AddInt64(&node_id, 1)
	n.Id = i-1
	n.Name = nm
	n.Cnt = -1
	n.Srcs = srcs
	n.Dsts = dsts
	for i := range n.Srcs {
		n.Srcs[i].Rdy = srcs[i].Data_rdy_init
		n.Srcs[i].Val = srcs[i].Init_val
	}
	for i := range n.Dsts {
		n.Dsts[i].Rdy = dsts[i].Ack_rdy_init
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

// debug printing
func (n Node) Printf(format string, v ...interface{}) {
	if (!Debug /*|| format=="select\n"*/) {
		return
	}
	newfmt,varlist := prefix_varlist(n)
	newfmt += format
	varlist = append(varlist, v...)
	fmt.Printf(newfmt, varlist...)
}

// tracing output
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

// tracing execution of Node
func (n Node) PrintVals() { n.PrintStatus(true) }

// increment execution count of Node
func (n *Node) ExecCnt() {
	if (GlobalExecCnt) {
		c := atomic.AddInt64(&global_exec_cnt, 1)
		n.Cnt = c-1
	} else {
		n.Cnt = n.Cnt+1
	}
}

// test readiness of Node to execute
func (n *Node) Rdy() bool {
	n.PrintStatus(false)
	if (n.RdyFunc == nil) {
		for i := range n.Srcs {
			if !n.Srcs[i].Rdy { return false }
		}
		for i := range n.Dsts {
			if !n.Dsts[i].Rdy { return false }
		}
	} else {
		if !n.RdyFunc(n) { return false }
	}
	n.ExecCnt();
	return true
}

// sink value (to avoid unused error)
func Sink(a Datum) () {
}

// test value for zero
func ZeroTest(a Datum) bool {

	switch a.(type) {
        case int8: { return a.(int8)==0 }
        case uint8: { return a.(uint8)==0 }
        case int16: { return a.(int16)==0 }
        case uint16: { return a.(uint16)==0 }
        case int32: { return a.(int32)==0 }
        case uint32: { return a.(uint32)==0 }
	case int64: { return a.(int64)==0 }
        case uint64: { return a.(uint64)==0 }
	case int: { return a.(int)==0 }
	case uint: { return a.(uint)==0 }
	case float32: { return a.(float32)==0.0 }
	case float64: { return a.(float64)==0.0 }
	case complex64: { return a.(complex64)==0.0+0.0i }
	case complex128: { return a.(complex128)==0.0+0.0i }
	default: { return true }
	}
}

