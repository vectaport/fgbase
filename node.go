package flowgraph

import (
	"reflect"
	"sync/atomic"
)

// Node of a flowgraph
type Node struct {
	ID int64                   // unique id
	Name string                // for tracing
	Cnt int64                  // execution count
	Srcs []*Edge               // upstream links
	Dsts []*Edge               // downstream links
	RdyFunc RdyTest            // func to test Edge readiness
	FireFunc FireNode          // func to fire Node execution
	Cases []reflect.SelectCase // select cases to read from Edge's
}

// RdyTest is the function signature for evaluating readiness of Node to fire.
type RdyTest func(*Node) bool

// FireNode is the function signature for firing off flowgraph stub.
type FireNode func(*Node)

// MakeNode returns a new Node with slices of input and output Edge's and functions for testing readiness then firing.
func MakeNode(
	name string, 
	srcs, dsts []*Edge, 
	ready RdyTest, 
	fire FireNode) Node {
	var n Node
	i := atomic.AddInt64(&nodeID, 1)
	n.ID = i-1
	n.Name = name
	n.Cnt = -1
	n.Srcs = srcs
	n.Dsts = dsts
	var casel [] reflect.SelectCase
	for i := range n.Srcs {
		n.Srcs[i].Rdy = n.Srcs[i].Val!=nil
		casel = append(casel, reflect.SelectCase{Dir:reflect.SelectRecv, Chan:reflect.ValueOf(n.Srcs[i].Data)})
	}
	for i := range n.Dsts {
		n.Dsts[i].Rdy = n.Dsts[i].Val==nil
		casel = append(casel, reflect.SelectCase{Dir:reflect.SelectRecv, Chan:reflect.ValueOf(n.Dsts[i].Ack)})
	}
	n.Cases = casel
	n.RdyFunc = ready
	n.FireFunc = fire
	return n
}

