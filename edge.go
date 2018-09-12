package fgbase

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"
)

// EdgeNode contains information on a Node connected to an Edge.
type edgeNode struct {
	node    *Node
	srcFlag bool
}

// Edge of a flowgraph.
type Edge struct {

	// values shared by upstream and downstream Nodes
	Name      string              // for trace
	Data      *[]chan interface{} // slice of data channels
	Ack       chan struct{}       // request (or acknowledge) channel
	edgeNodes *[]edgeNode         // list of Node's associated with this Edge.

	// values unique to upstream and downstream Nodes
	Val    interface{}   // generic empty interface
	RdyCnt int           // readiness of I/O
	Flow   bool          // set true to allow one output, data or ack
	Ack2   chan struct{} // alternate channel for ack steering

}

// Return new Edge to connect one upstream Node to one or more downstream Node's.
// Initialize optional data value to start flow.
func makeEdge(name string, initVal interface{}) Edge {
	var e Edge

	i := atomic.AddInt64(&EdgeID, 1)
	if name == "" {
		e.Name = "e" + strconv.Itoa(int(i-1))
	} else {
		e.Name = name
	}

	e.Val = initVal
	var dc []chan interface{}
	e.Data = &dc
	e.Ack = make(chan struct{}, ChannelSize)
	var nl []edgeNode
	e.edgeNodes = &nl
	return e
}

// MakeEdge initializes optional data value to start flow.
func MakeEdge(name string, initVal interface{}) Edge {
	return makeEdge(name, initVal)
}

// Const sets up an Edge to provide a constant value.
func (e *Edge) Const(d interface{}) {
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

// SrcJSON sets up an Edge with a remote JSON value source.
func (e *Edge) SrcJSON(n *Node, portString string) {

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

			var v interface{}
			err = json.Unmarshal(b, &v)
			if err != nil {
				n.LogError("%v", err)
			}
			if IsSlice(v) {
				// n.Tracef("type of [] is %s\n", reflect.TypeOf(Index(v, 0)))
			}

			c.Send(reflect.ValueOf(v))
		}
	}()

	writer := bufio.NewWriter(conn)
	go func() {
		bufCnt := 0
		for {
			<-e.Ack
			bufCnt++
			_, err := writer.WriteString("\n")
			if err != nil {
				n.LogError("write error: %v", err)
				close(e.Ack)
				e.Ack = nil
				return
			}
			if bufCnt == ChannelSize {
				writer.Flush()
				bufCnt = 0
			}
		}
	}()

}

// DstJSON sets up an Edge with a remote JSON value destination.
func (e *Edge) DstJSON(n *Node, portString string) {

	conn, err := net.Dial("tcp", portString)
	n.Tracef("dialing err if any:  %v\n", err)
	if err != nil {
		StderrLog.Printf("%v\n", err)
		return
	}

	reader := bufio.NewReader(conn)
	go func() {
		for {
			_, err := reader.ReadString('\n')
			if err != nil {
				if err.Error() != "EOF" {
					n.LogError("Dst read error: %v", err)
				}
				return
			}
			e.Ack <- struct{}{}
		}
	}()

	writer := bufio.NewWriter(conn)
	j := len(*e.Data)
	*e.Data = append(*e.Data, make(chan interface{}, ChannelSize))
	ej := (*e.Data)[j]
	go func() {
		bufCnt := 0
		for {
			v := <-ej
			time.Sleep(10000)
			bufCnt++
			b, err := json.Marshal(v)
			// n.Tracef("json output:  %v", string(b))
			if err != nil {
				n.LogError("%v", err)
			}
			_, err = writer.WriteString(string(b) + "\n")
			if err != nil {
				n.LogError("write error:  %v", err)
				close(ej)
				ej = nil
				return
			}
			if bufCnt == ChannelSize {
				writer.Flush()
				bufCnt = 0
			}
		}
	}()

}

// Rdy tests if RdyCnt has returned to zero.
func (e *Edge) RdyZero() bool {
	return e.RdyCnt == 0
}

// srcReadRdy tests if a source Edge is ready for a data read.
func (e *Edge) srcReadRdy(n *Node) bool {
	i := n.edgeToCase[e]
	return n.cases[i].Chan.IsValid() && n.cases[i].Chan.Len() > 0
}

// srcReadHandle handles a source Edge data read.
func (e *Edge) srcReadHandle(n *Node, selectFlag bool) {
	var wrapFlag = false
	if n2, ok := e.Val.(nodeWrap); ok {
		e.Ack2 = n2.ack2
		e.Val = e.Val.(nodeWrap).datum
		wrapFlag = true
		if &n2.node.FireFunc == &n.FireFunc {
			n.flag |= flagRecursed
		} else {
			bitr := ^flagRecursed
			n.flag = (n.flag & ^bitr)
		}
	}
	e.RdyCnt--
	if false {
		n.Tracef("srcReadHandle -- %s.RdyCnt=%d (%s)\n", e.Name, e.RdyCnt,
			func() string {
				if selectFlag {
					return "s"
				}
				return "!s"
			}())
	}
	if e.RdyCnt < 0 {
		n.Tracef("%s.RdyCnt less than zero, time to panic\n", e.Name)
	}
	if TraceLevel >= VV {
		var attrs string
		if selectFlag {
			attrs += " // s"
		} else {
			attrs = " // !s"
		}
		if wrapFlag && TraceLevel >= VV {
			attrs += fmt.Sprintf(",Ack2=%p", e.Ack2)
		}
		if e.Val == nil {
			n.Tracef("<nil> <- %s.Data%s\n", e.Name, attrs)
		} else {
			n.Tracef("%s <- %s.Data%s\n", String(e.Val), e.Name, attrs)
		}
	}
	if e.RdyCnt < 0 {
		panic("Edge.srcReadHandle:  Edge RdyCnt less than zero")
	}
}

// srcWriteRdy tests if a source Edge is ready for an ack write.
func (e *Edge) srcWriteRdy() bool {
	return len(e.Ack) < cap(e.Ack)
}

// SrcRdy tests if a source Edge is ready.
func (e *Edge) SrcRdy(n *Node) bool {
	if !e.RdyZero() {
		if !e.srcReadRdy(n) {
			return false
		}

		i := n.edgeToCase[e]
		if n.cases[i].Chan != reflect.ValueOf(nil) {

			c := n.cases[i].Chan
			var ok bool
			v, ok := c.Recv()
			if !ok {
				panic("Unexpected error in reading channel\n")
			}
			e.Val = v.Interface()
			n.RemoveInputCase(e)
			e.srcReadHandle(n, false)
		}

		return e.RdyZero()
	}
	return true
}

// SrcWait waits for a source Edge to be ready.
func (e *Edge) SrcWait(n *Node) {
	if !e.RdyZero() {

		i := n.edgeToCase[e]
		if n.cases[i].Chan != reflect.ValueOf(nil) {
			c := n.cases[i].Chan
			var ok bool
			v, ok := c.Recv()
			if !ok {
				panic("Unexpected error in reading channel\n")
			}
			e.Val = v.Interface()
			n.RemoveInputCase(e)
			e.srcReadHandle(n, false)
		} else {
			panic("Unexpected nil src data channel\n")
		}
	}
}

// dstReadRdy tests if a destination Edge is ready for an ack read.
func (e *Edge) dstReadRdy() bool {
	return len(e.Ack) > 0
}

// dstReadHandle handles a destination Edge ack read.
func (e *Edge) dstReadHandle(n *Node, selectFlag bool) {

	e.RdyCnt--
	if false {
		n.Tracef("dstReadHandle -- %s.RdyCnt=%d (%s)\n", e.Name, e.RdyCnt,
			func() string {
				if selectFlag {
					return "s"
				}
				return "!s"
			}())
	}
	if e.RdyCnt < 0 {
		n.Tracef("%s.RdyCnt less than zero, time to panic\n", e.Name)
	}
	if TraceLevel >= VV {
		var selectStr string
		if selectFlag {
			selectStr = "// s"
		} else {
			selectStr = "// !s"
		}
		nm := e.Name + ".Ack"
		if true || len(*e.Data) > 1 {
			nm += "{" + strconv.Itoa(e.RdyCnt+1) + "}"
		}
		n.Tracef("<- %s %s\n", nm, selectStr)
	}
	if e.RdyCnt < 0 {
		panic("Edge.dstReadHandle:  Edge RdyCnt less than zero")
	}
}

// dstWriteRdy tests if a destination Edge is ready for a data write.
func (e *Edge) dstWriteRdy() bool {
	for _, c := range *e.Data {
		if cap(c) < len(c)+e.SrcCnt() {
			return false
		}
	}
	return true
}

// DstRdy tests if a destination Edge is ready.
func (e *Edge) DstRdy(n *Node) bool {
	if !e.RdyZero() {
		if !e.dstReadRdy() {
			return e.dstWriteRdy()
		}

		l := len(e.Ack)
		for l > 0 {
			if l > e.RdyCnt {
				n.Tracef("READ ACK with l=%d and e.RdyCnt=%d\n", l, e.RdyCnt)
				panic("Unexpected ack received\n")
			}
			<-e.Ack
			e.dstReadHandle(n, false)
			l--
		}

		if e.dstWriteRdy() {
			return true
		}
	}

	f := e.RdyZero()
	return f

}

// SendData writes to the Data channel
func (e *Edge) SendData(n *Node) bool {
	sendOK := false
	if e.Data != nil {
		if e.Flow {
			for i := range *e.Data {
				(*e.Data)[i] <- e.Val
			}
			e.RdyCnt += len(*e.Data)

			if TraceLevel >= VV {
				nm := e.Name + ".Data"
				if len(*e.Data) > 1 {
					nm += "{" + strconv.Itoa(len(*e.Data)) + "}"
				}
				ev := e.Val
				var attrs string

				// remove from wrapper if in one
				if _, ok := ev.(nodeWrap); ok {
					attrs += fmt.Sprintf(" // Ack2=%p", ev.(nodeWrap).ack2)
					ev = ev.(nodeWrap).datum
				}

				if false {
					// add other attributes for debug purposes
					if attrs == "" {
						attrs += " // "
					} else {
						attrs += ","
					}
					attrs += "len={"
					for i := range *e.Data {
						if i != 0 {
							attrs += ","
						}
						attrs += strconv.Itoa(len((*e.Data)[i]))
					}
					attrs += "},"
					attrs += "cap={"
					for i := range *e.Data {
						if i != 0 {
							attrs += ","
						}
						attrs += strconv.Itoa(cap((*e.Data)[i]))
					}
					attrs += "}"
				}

				if ev == nil {
					n.Tracef("%s <- <nil>%s\n", nm, attrs)
				} else {
					n.Tracef("%s <- %s%s\n", nm, String(ev), attrs)
				}
			}

			e.Val = nil
			sendOK = true
		}
	}
	e.Flow = false
	return sendOK
}

// SendAck writes struct{} to the Ack channel
func (e *Edge) SendAck(n *Node) bool {
	sendOK := false
	if e.Ack != nil {
		if e.Flow {
			if e.Ack2 != nil {
				if TraceLevel >= VV {
					n.Tracef("%s.Ack <- // Ack2=%p\n", e.Name, e.Ack2)
				}
				e.Ack2 <- struct{}{}
				e.Ack2 = nil
			} else {
				if TraceLevel >= VV {
					n.Tracef("%s.Ack <-\n", e.Name)
				}
				e.Ack <- struct{}{}
			}
			e.RdyCnt++
			sendOK = true
		}
	}
	e.Flow = false
	return sendOK
}

// MakeEdges returns a slice of Edge.
func MakeEdges(sz int) []Edge {
	e := make([]Edge, sz)
	for i := 0; i < sz; i++ {
		e[i] = MakeEdge("", nil)
	}
	return e
}

// PoolEdge returns an output Edge that is directed back into the Pool.
func (e *Edge) PoolEdge(src *Edge) *Edge {
	e.Data = src.Data
	e.Name = e.Name + "(" + src.Name + ")"
	return e
}

// SrcCnt is the number of Node's upstream of an Edge
func (e *Edge) SrcCnt() int {
	i := 0
	for ; i < len(*e.edgeNodes) && (*e.edgeNodes)[i].srcFlag; i++ {
	}
	return i
}

// DstCnt is the number of Node's downstream of an Edge
func (e *Edge) DstCnt() int {
	return len(*e.edgeNodes) - e.SrcCnt()
}

// DstOrder returns the order of a Node in an Edge's destinations
func (e *Edge) DstOrder(n *Node) int {
	for i := e.SrcCnt(); i < len(*e.edgeNodes); i++ {
		if (*e.edgeNodes)[i].node == n {
			return i
		}
	}
	return -1
}

// SrcNode returns the ith upstream Node of an Edge
func (e *Edge) SrcNode(i int) *Node {
	if i > e.SrcCnt() || i < 0 {
		return nil
	}
	return (*e.edgeNodes)[i].node
}

// DstNode returns the ith downstream Node of an Edge
func (e *Edge) DstNode(i int) *Node {
	if i > e.DstCnt() || i < 0 {
		return nil
	}
	h := e.SrcCnt()
	return (*e.edgeNodes)[i+h].node
}

// CloseData closes all outgoing data channels.
func (e *Edge) CloseData() {
	for i := range *e.Data {
		close((*e.Data)[i])
		(*e.Data)[i] = nil
	}
}

// NameEdges adds a name to each Edge
func NameEdges(edges []Edge, names []string) {
	l := len(names)
	for i := range edges {
		if i == l {
			break
		}
		edges[i].Name = names[i]

	}
}

// SrcGet returns the empty interface value flowing from the input Edge
func (e *Edge) SrcGet() interface{} {
	e.Flow = true
	return e.Val
}

// DstPut sets the empty interface value flowing to the output Edge
func (e *Edge) DstPut(v interface{}) {
	e.Flow = true
	e.Val = v
}
