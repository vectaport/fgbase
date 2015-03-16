package flowgraph

import (
	"sync/atomic"
	"fmt"
)

var node_id int64 = 0
var global_cnt int64 = 0

type Datum interface{}

type Edge struct {
	Data chan Datum
	Data_init bool
	Init_val Datum
	Ack chan bool
	Ack_init bool
}

type Node struct {
	Id int64
	Name string
	Cnt int64
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
	i := atomic.AddInt64(&node_id, 1)
	n.Id = i-1
	n.Name = nm
	n.Cnt = -1
	return n
}

func Sink(a Datum) () {
}

func (n Node) Printf(format string, v ...interface{}) {
	if (format=="select\n") {
		return
	}
	var newv [] interface{}
	newv = append(newv, n.Name)
	newv = append(newv, n.Id)
	if (n.Cnt>=0) {
		newv = append(newv, n.Cnt)
	} else {
		newv = append(newv, "*")
	}

	newv = append(newv, v...)
	fmt.Printf("%s(%d:%v):  "+format, newv...)
}

func (n *Node) ExecCnt() {
	c := atomic.AddInt64(&global_cnt, 1)
	n.Cnt = c-1
}
