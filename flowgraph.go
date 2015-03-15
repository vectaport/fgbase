package flowgraph

import (
)

var NodeId int = 0

type Datum interface{}

type Edge struct {
	Data chan Datum
	Data_init bool
	Init_val Datum
	Ack chan bool
	Ack_init bool
}

type Node struct {
	Id int
}

func MakeEdge(data_init, ack_init bool, init_val Datum) Edge {
	var e Edge
	e.Data = make(chan Datum)
	e.Data_init = data_init
	e.Init_val = init_val
	e.Ack = make(chan bool)
	e.Ack_init = ack_init
	return e
}

func MakeNode() Node {
	var n Node
        n.Id = NodeId
	NodeId = NodeId + 1
	return n
}

func Sink(a Datum) () {
}
