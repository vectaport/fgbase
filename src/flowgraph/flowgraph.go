package flowgraph

import (
)

var PipeId int = 0

type Datum interface{}
type Dataconn chan Datum
type Ackconn chan bool

type Conn struct {
	Data Dataconn
	Data_init bool
	Init_val Datum
	Ack Ackconn
	Ack_init bool
}
// type Pipe ???

func MakeConn(data_init, ack_init bool, init_val Datum) Conn {
	var c Conn
	c.Data = make(Dataconn)
	c.Data_init = data_init
	c.Init_val = init_val
	c.Ack = make(Ackconn)
	c.Ack_init = ack_init
	return c
}

func MakePipe() int {
	PipeId = PipeId + 1
	return PipeId-1
}

func Sink(a Datum) () {
}
