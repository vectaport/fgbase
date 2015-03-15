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
// type Node ???

func MakeEdge(data_init, ack_init bool, init_val Datum) Edge {
	var c Edge
	c.Data = make(chan Datum)
	c.Data_init = data_init
	c.Init_val = init_val
	c.Ack = make(chan bool)
	c.Ack_init = ack_init
	return c
}

func MakeNode() int {
	NodeId = NodeId + 1
	return NodeId-1
}

func Sink(a Datum) () {
}
