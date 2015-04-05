/*
Package flowgraph layers a ready-send flow mechanism on top of goroutines.
*/

package flowgraph

import (
	"log"
	"os"
	"reflect"
	"sync/atomic"
)

var nodeID int64
var globalExecCnt int64

// Log for tracing flowgraph execution
var StdoutLog = log.New(os.Stdout, "", 0)

// Enable debug tracing
var Debug = false

// Indent trace by node id
var Indent = false

// Use global execution count
var GlobalExecCnt = false

func prefixVarlist(n *Node) (format string, varlist []interface {}) {
	var varl [] interface {}
	varl = append(varl, n.Name)
	varl = append(varl, n.ID)
	if (n.Cnt>=0) {
		varl = append(varl, n.Cnt)
	} else {
		varl = append(varl, "*")
	}
	var f string
	if (Indent) {
		for i := int64(0);i<n.ID;i++ {
			f += "\t"
		}
	}
	f += "%s(%d:%v) "
	return f,varl
}

func addSliceToVarlist(d Datum, format string, varlist []interface {}) (newfmt string, newvarlist []interface {}) {
	m := 8
	l := Len(d)
	if l < m { m = l }
	varlist = append(varlist, d)
	format += "%T(["
	for i := 0; i<m; i++ {
		if i!=0 {format += " "}
		varlist = append(varlist, Index(d,i))
		format += "%+v"
	}
	if m<l {format += " ..."}
	format += "])"
	return format,varlist
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (n *Node) Tracef(format string, v ...interface{}) {
	if (!Debug) {
		return
	}
	newfmt,varlist := prefixVarlist(n)
	newfmt += format
	varlist = append(varlist, v...)
	StdoutLog.Printf(newfmt, varlist...)
}

// TraceValRdy lists Node input values and output readiness
func (n *Node) TraceValRdy(valOnly bool) {

	if (!valOnly && !Debug) {return}
	newfmt,varlist := prefixVarlist(n)
	if !valOnly { newfmt += "<<" }
	for i := range n.Srcs {
		if (i!=0) { newfmt += "," }
		varlist = append(varlist, n.Srcs[i].Name)
		newfmt += "%s="
		if (n.Srcs[i].Rdy) {
			if IsSlice(n.Srcs[i].Val) {
				newfmt,varlist = addSliceToVarlist(n.Srcs[i].Val, newfmt, varlist)
			} else {
				if true { 
					varlist = append(varlist, n.Srcs[i].Val)
					newfmt += "%v"
				} else {
					varlist = append(varlist, n.Srcs[i])
					newfmt += "%+v"
				}
			}
		} else {
			varlist = append(varlist, "{}")
			newfmt += "%s"
		}
	}
	newfmt += ":"
	for i := range n.Dsts {
		if (i!=0) { newfmt += "," }
		if (valOnly) {
			varlist = append(varlist, n.Dsts[i].Name)
			newfmt += "%s="
			if IsSlice(n.Dsts[i].Val) {
				newfmt,varlist = addSliceToVarlist(n.Dsts[i].Val, newfmt, varlist)
			} else {
				if (n.Dsts[i].Val != nil) {
					varlist = append(varlist, n.Dsts[i].Val)
					varlist = append(varlist, n.Dsts[i].Val)
					newfmt += "%T(%v)"
				} else {
					varlist = append(varlist, "{}")
					newfmt += "%v"
				}
			}
		} else {
			if true {
				varlist = append(varlist, n.Dsts[i].Name+".Rdy")
				varlist = append(varlist, n.Dsts[i].Rdy)
				newfmt += "%s=%v"
			} else {
				varlist = append(varlist, n.Dsts[i].Name)
				varlist = append(varlist, n.Dsts[i])
				newfmt += "%s=%+v"
			}
		}
	}
	if !valOnly { newfmt += ">>" }
	newfmt += "\n"
	StdoutLog.Printf(newfmt, varlist...)
}

// TraceVals lists input and output values for a Node.
func (n *Node) TraceVals() { n.TraceValRdy(true) }

// IncrExecCnt increments execution count of Node
func (n *Node) IncrExecCnt() {
	if (GlobalExecCnt) {
		c := atomic.AddInt64(&globalExecCnt, 1)
		n.Cnt = c-1
	} else {
		n.Cnt = n.Cnt+1
	}
}

// RdyAll tests readiness of Node to execute.
func (n *Node) RdyAll() bool {
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
	n.IncrExecCnt();
	return true
}

// Fire node using function pointer
func (n *Node) Fire() {
	if (n.FireFunc!=nil) { n.FireFunc(n) }
}


// SendAll writes all data and acks after new result is computed.
func (n *Node) SendAll() {
	n.TraceVals()
	for i := range n.Srcs {
		n.Srcs[i].SendAck(n)
	}
	for i := range n.Dsts {
		n.Dsts[i].SendData(n)
	}
}

// RecvOne reads one data or ack and marks that input as ready.
func (n *Node) RecvOne() {
	l := len(n.Srcs)
	n.TraceValRdy(false)
	chosen,recv,recvOK := reflect.Select(n.Cases)
	if (recvOK) {
		if chosen<l {
			n.Srcs[chosen].Val = recv.Interface()
			n.Srcs[chosen].Rdy = true
			n.Tracef("%T(%v) <- %s.Data\n", n.Srcs[chosen].Val, n.Srcs[chosen].Val, n.Srcs[chosen].Name)
		} else {
			n.Dsts[chosen-l].Rdy = true
			n.Tracef("true <- %s.Ack\n", n.Dsts[chosen-l].Name)
		}
	}
}

// Run is an event loop that runs forever.
func (n *Node) Run() {
	for {
		if(n.RdyAll()) {
			n.Fire()	
		n.SendAll()
		}

		n.RecvOne()
	}
}

