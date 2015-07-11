package flowgraph

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"time"
)

type Nada struct {}

// Edge of a flowgraph.
type Edge struct {

	// values shared by upstream and downstream Node
	Name string        // for trace
	Data *[]chan Datum // slice of data channels
	Ack chan Nada      // request (or acknowledge) channel

	// values unique to upstream and downstream Node
	Val Datum          // generic empty interface
	RdyCnt int         // readiness of I/O
	NoOut bool         // set true to inhibit one output, data or ack
	Aux Datum          // auxiliary empty interface to hold state
	Ack2 chan Nada     // alternate channel for ack steering

}

// Return new Edge to connect one upstream Node to one or more downstream Node's.
// Initialize optional data value to start flow.
func makeEdge(name string, initVal Datum) Edge {
	var e Edge
	e.Name = name
	e.Val = initVal
	dc := make([]chan Datum, 0)
	e.Data = &dc
	e.Ack = make(chan Nada, ChannelSize)
	return e
}

// MakeEdge initializes optional data value to start flow.
func MakeEdge(name string, initVal Datum) Edge {
	return makeEdge(name, initVal)
}

// Const sets up an Edge to provide a constant value.
func (e *Edge) Const(d Datum) {
	e.Val = d
	e.Data = nil
	e.Ack = nil
}
	
// IsConst returns true if Edge provides a constant value.
func (e *Edge) IsConst() bool { 
	return e.Data == nil && e.Val != nil
}

// Sink sets up an Edge as a value sink.
func (e *Edge) Sink() {
	e.Val = nil
	e.Data = nil
	e.Ack = nil
}

// IsSink returns true if Edge is a value sink.
func (e *Edge) IsSink() bool { 
	return e.Data == nil && e.Val == nil
}

// Src sets up an Edge as a remote JSON value source.
func (e *Edge) Src(n *Node, portString string) {

	ln, err := net.Listen("tcp", portString)
	if err != nil {
		StderrLog.Printf("%v\n", err)
		return
	}
	conn, err := ln.Accept()
	if err != nil {
		StderrLog.Printf("%v\n", err)
		return
	}

	reader := bufio.NewReader(conn)
	j := n.edgeToCase[e]
	c := n.cases[j].Chan
	go func() {
		for {
			b, err := reader.ReadBytes('\n')
			// n.Tracef("json input:  %v", string(b))
			if err != nil {
				if err.Error() != "EOF" {
					n.LogError("%v", err)
				}
				return
			}

			var v Datum
			err = json.Unmarshal(b,&v)
			if err != nil {
				n.LogError("%v", err)
			}
			if IsSlice(v) {
				// n.Tracef("type of [] is %s\n", reflect.TypeOf(Index(v, 0)))
			}

			c.Send(reflect.ValueOf(v))
		}
	} ()


	writer := bufio.NewWriter(conn)
	go func() {
		bufCnt := 0
		for {
			<- e.Ack
			bufCnt++
			_, err := writer.WriteString("\n")
			if err != nil {
				n.LogError("write error: %v", err)
				close(e.Ack)
				e.Ack = nil
				return
			}
			if bufCnt==ChannelSize {
				writer.Flush()
				bufCnt = 0
			}
		}
	} ()

}

// Dst sets up an Edge as a remote JSON value destination.
func (e *Edge) Dst(n *Node, portString string) {

	conn, err := net.Dial("tcp", portString)
	if err != nil {
		StderrLog.Printf("%v\n", err)
		return
	}

	reader := bufio.NewReader(conn)
	go func() {
		var nada Nada
		for {
			_, err := reader.ReadString('\n')
			if err != nil {
				if err.Error() != "EOF" {
					n.LogError("Dst read error: %v", err)
				}
				return
			}
			e.Ack <- nada
		}
	} ()


	writer := bufio.NewWriter(conn)
	j := len(*e.Data)
	*e.Data = append(*e.Data, make(chan Datum, ChannelSize))
	ej := (*e.Data)[j]
	go func() {
		bufCnt := 0
		for {
			v := <- ej
			time.Sleep(10000)
			bufCnt++
			b,err := json.Marshal(v)
			// n.Tracef("json output:  %v", string(b))
			if err != nil {
				n.LogError("%v", err)
			}
			_, err = writer.WriteString(string(b)+"\n")
			if err != nil {
				n.LogError("write error:  %v", err)
				close(ej)
				ej = nil
				return
			}
			if bufCnt==ChannelSize {
				writer.Flush()
				bufCnt = 0
			}
		}
	} ()

}

// Rdy tests if RdyCnt has returned to zero.
func (e *Edge) Rdy() bool {
	return e.RdyCnt==0
}

// srcReadRdy tests if a source Edge is ready for a data read.
func (e *Edge) srcReadRdy(n *Node) bool {
	i := n.edgeToCase[e]
	return n.cases[i].Chan.IsValid() && n.cases[i].Chan.Len()>0
}

// srcReadHandle handles a source Edge data read.
func (e *Edge) srcReadHandle (n *Node, selectFlag bool) {
	var wrapFlag = false
	if _,ok := e.Val.(nodeWrap); ok {
		n2 := e.Val.(nodeWrap).node
		e.Ack2 = n2.Dsts[0].Ack
		e.Val = e.Val.(nodeWrap).datum
		wrapFlag = true
		if &n2.FireFunc == &n.FireFunc { 
			n.flag |=flagRecursed 
		} else {
			bitr := ^flagRecursed
			n.flag =(n.flag & ^bitr)
		}
	}
	e.RdyCnt--
	if (TraceLevel>=VV) {
		var selectStr string
		if selectFlag {
			selectStr = " (s)"
		} else {
			selectStr = " (!s)"
		}
		var asterisk string
		if wrapFlag && TraceLevel>=VV { asterisk = fmt.Sprintf(" *(Ack2=%p)", e.Ack2)}
		asterisk += selectStr
		if (e.Val==nil) {
			n.Tracef("<nil> <- %s.Data%s%ss\n", e.Name, asterisk)
		} else {
			n.Tracef("%s <- %s.Data%s\n", String(e.Val), e.Name, asterisk)
		}
	}
}

// srcWriteRdy tests if a source Edge is ready for an ack write.
func (e *Edge) srcWriteRdy() bool {
	return len(e.Ack)<cap(e.Ack)
}

// SrcRdy tests if a source Edge is ready.
func (e *Edge) SrcRdy(n *Node) bool {
	if !e.Rdy() {
		if !e.srcReadRdy(n) { 
			return false 
		}

		i := n.edgeToCase[e]
		if n.cases[i].Chan!=reflect.ValueOf(nil) {

			c := n.cases[i].Chan
			var ok bool
			v,ok := c.Recv()
			if !ok {
				panic("Unexpected error in reading channel\n")
			}
			e.Val = v.Interface()
			n.cases[i].Chan = reflect.ValueOf(nil) // don't read this again until after RdyAll
			e.srcReadHandle(n, false)
		}

		return e.Rdy()
	}
	return true
}

// dstReadRdy tests if a destination Edge is ready for an ack read.
func (e *Edge) dstReadRdy() bool {
	return len(e.Ack)>0
}

// dstReadHandle handles a destination Edge ack read.
func (e *Edge) dstReadHandle (n *Node, selectFlag bool) {
	
	e.RdyCnt--
	if (TraceLevel>=VV) {
		var selectStr string
		if selectFlag {
			selectStr = " (s)"
		} else {
			selectStr = " (!s)"
		}
		nm := e.Name + ".Ack"
		if len(*e.Data)>1 {
			nm += "{" + strconv.Itoa(e.RdyCnt+1) + "}"
		}
		n.Tracef("<- %s(%p)%s\n", nm, e.Ack, selectStr)
	}
}

// dstWriteRdy tests if a destination Edge is ready for a data write.
func (e *Edge) dstWriteRdy() bool {
	for _,c := range *e.Data {
		if cap(c)==len(c) { 
			return false 
		}
	}
	return true
}

// DstRdy tests if a destination Edge is ready.
func (e *Edge) DstRdy(n *Node) bool {
	if !e.Rdy() {
		if !e.dstReadRdy() { 
			return e.dstWriteRdy()
		}

		for len(e.Ack)>0 {
			<- e.Ack
			e.dstReadHandle(n, false)
		}

		if e.dstWriteRdy() {
			return true 
		}
	}
			
	f := e.Rdy()
	return f
	
}

// SendData writes to the Data channel
func (e *Edge) SendData(n *Node) {
	if(e.Data !=nil) {
		if (!e.NoOut) {
			if (TraceLevel>=VV) {
				nm := e.Name + ".Data"
				if len(*e.Data)>1 {
					nm += "{" + strconv.Itoa(len(*e.Data)) + "}"
				}
				ev := e.Val
				var asterisk string
				
				// remove from wrapper if in one
				if _,ok := ev.(nodeWrap); ok {
					n2 := ev.(nodeWrap).node
					ev = ev.(nodeWrap).datum
					asterisk += fmt.Sprintf(" *(Ack2=%p)", n2.Srcs[0].Ack)
				}

				if (ev==nil) {
					n.Tracef("%s <- <nil>%s\n", nm, asterisk)
				} else {
					n.Tracef("%s <- %s%s\n", nm, String(ev), asterisk)
				}
			}

			for i := range *e.Data {
				(*e.Data)[i] <- e.Val
			}
			e.RdyCnt += len(*e.Data)
			e.Val = nil
		} else {
			e.NoOut = false
		}
	}
}

// SendAck writes Nada to the Ack channel
func (e *Edge) SendAck(n *Node) {
	if(e.Ack !=nil) {
		if (!e.NoOut) {
			var nada Nada
			if e.Ack2 != nil {
				if (TraceLevel>=VV) {
					n.Tracef("%s.Ack2(%p) <-\n", e.Name, e.Ack2)
				}
				e.Ack2 <- nada
				e.Ack2 = nil
			} else {
				if (TraceLevel>=VV) {
					n.Tracef("%s.Ack(%p) <-\n", e.Name, e.Ack)
				}
				e.Ack <- nada
			}
			e.RdyCnt = 1
			
		} else {
			e.NoOut = false
		}
	}
}

// MakeEdges returns a slice of Edge.
func MakeEdges(sz int) []Edge {
	e := make([]Edge, sz)
	for i:=0; i<sz; i++ {
		nm := "e" + strconv.Itoa(int(i))
		e[i] = MakeEdge(nm, nil)
	}
	return e
}

// PoolEdge returns an output Edge that is directed back into the Pool.
func (dst *Edge) PoolEdge(src *Edge) Edge {
	e := *dst
	e.Data = src.Data
	e.Name = dst.Name+"("+src.Name+")"
	return e
}
	
