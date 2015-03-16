package flowgraph

import (
	"fmt"
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
	Name string
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

func MakeNode(nm string) Node {
	var n Node
        n.Id = NodeId
	NodeId = NodeId + 1
	n.Name = nm
	return n
}

func Sink(a Datum) () {
}

func (n Node) Printf(format string, v ...interface{}) {
	var newv [] interface{}
	newv = append(newv, n.Name)
	newv = append(newv, n.Id)
	newv = append(newv, v...)
	fmt.Printf("%s(%d):  "+format, newv...)
}
