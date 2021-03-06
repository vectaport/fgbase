package fgbase

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"sync/atomic"
)

// edgeNode contains information on a Node connected to an Edge.
type edgeNode struct {
	node    *Node
	srcFlag bool
}

// edgeNodePlus adds the Edge as well
type edgeNodePlus struct {
	edgeNode
	edge *Edge
}

// block state
type block int

const (
	noBlock block = iota
	dataBlock
	ackBlock
	readBlock
)

// Edge of a flowgraph.
type Edge struct {

	// values shared by upstream and downstream Nodes
	Name           string              // for trace
	Data           *[]chan interface{} // slice of data channels
	Ack            chan struct{}       // request (or acknowledge) channel
	edgeNodes      *[]edgeNode         // list of Node's associated with this Edge.
	srcCnt, dstCnt *int                // count of upstream/downstream nodes

	// values unique to upstream and downstream Nodes
	Val      interface{}   // generic empty interface
	RdyCnt   int           // readiness of I/O
	Flow     bool          // set true to allow one output, data or ack
	Ack2     chan struct{} // alternate channel for ack steering
	blocked  block         // blocked status:  dataBlock, ackBlock, readBLock, noBlock
	dotAttrs *[]string     // attributes for dot outputs

}

// Return new Edge to connect one upstream Node to one or more downstream Node's.
// Initialize optional data value to start flow.
func makeEdge(name string, initVal interface{}) Edge {

	var e Edge

	i := atomic.AddInt64(&EdgeID, 1)
	if name == "" {
		e.Name = fmt.Sprintf("e%d", i-1)
	} else {
		e.Name = name
	}

	e.Val = initVal
	var dc []chan interface{}
	e.Data = &dc
	e.Ack = make(chan struct{}, ChannelSize)
	var nl = make([]edgeNode, 0)
	e.edgeNodes = &nl
	srcCount := 0
	e.srcCnt = &srcCount
	dstCount := 0
	e.dstCnt = &dstCount
	da := make([]string, 0)
	e.dotAttrs = &da
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
			// time.Sleep(10000)
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

// RdyZero tests if RdyCnt has returned to zero.
func (e *Edge) RdyZero() bool {
	return e.RdyCnt == 0
}

// srcReadRdy tests if a source Edge is ready for a data read.
func (e *Edge) srcReadRdy(n *Node) bool {
	// nodes in a pools that are sharing src channels could get
	// blocked on a read because another node might read the channel first
	// but that just puts in them in a queue ready to read the next data
	// the problem is these pool nodes can't respond until later to any
	// other incoming data or message
	i := n.edgeToCase[e]
	return n.cases[i].Chan.IsValid() && n.cases[i].Chan.Len() > 0
}

// srcReadHandle handles a source Edge data read.
func (e *Edge) srcReadHandle(n *Node, selectFlag bool) {
	e.RdyCnt--

	// unpack steered ack wrapping
	wrapFlag := false
	if n2, wrapflag := e.Val.(ackWrap); wrapflag {
		e.Ack2 = n2.ack2
		e.Val = e.Val.(ackWrap).datum
		if &n2.node.FireFunc == &n.FireFunc {
			n.flag |= flagRecursed
		} else {
			bitr := ^flagRecursed
			n.flag = (n.flag & bitr)
		}
	}

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
		if TraceLevel >= VVVV {
			if selectFlag {
				attrs += "\t// select"
			} else {
				attrs = "\t// !select"
			}
		}
		if wrapFlag && TraceLevel >= VV {
			attrs += fmt.Sprintf(",Ack2=%p", e.Ack2)
		}
		if TraceLevel >= VVVV {
			attrs += fmt.Sprintf(",Data=%v", e.Data)
		}
		if e.Val == nil {
			n.Tracef("<nil> <- %s.Data%s\n", e.Name, attrs)
		} else {
			n.Tracef("%s <- %s.Data%s\n", String(e.Val), e.Name, attrs)
		}
	}
	if e.RdyCnt < 0 {
		n.Tracef("Edge %q RdyCnt less than zero\n", e.Name)
		panic(fmt.Sprintf("Edge.srcReadHandle:  Edge %q RdyCnt less than zero", e.Name))
	}
}

// srcWriteRdy tests if a source Edge is ready for an ack write.
func (e *Edge) srcWriteRdy() bool {
	if e.Ack2 != nil {
		return len(e.Ack2) < cap(e.Ack2)
	}
	return len(e.Ack) < cap(e.Ack)
}

// SrcRdy tests if a source Edge is ready.
func (e *Edge) SrcRdy(n *Node) bool {
	if !n.isSrc(e) {
		panic(fmt.Sprintf("Unexpected destination edge %q checked for source readiness on node %q\n", e.Name, n.Name))
	}
	if e == nil {
		return false
	}
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

// srcWait waits for a source Edge to be ready.
func (e *Edge) srcWait(n *Node) {
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

func emptystruct() string {
	if TraceStyle == New {
		return "struct{}"
	}
	return ""
}

// dstReadHandle handles a destination Edge ack read.
func (e *Edge) dstReadHandle(n *Node, selectFlag bool) {

	e.RdyCnt--

	if false {
		n.Tracef("dstReadHandle -- %s.RdyCnt=%d (%s)\n", e.Name, e.RdyCnt,
			func() string {
				if selectFlag {
					return "select"
				}
				return "!select"
			}())
	}

	if e.RdyCnt < 0 {
		n.Tracef("%s.RdyCnt less than zero, time to panic\n", e.Name)

	}
	if TraceLevel >= VV {
		var selectStr string
		if TraceLevel >= VVVV {
			if selectFlag {
				selectStr = "\t// select"
			} else {
				selectStr = "\t// !select"
			}
		}
		nm := e.Name + ".Ack"
		if true || len(*e.Data) > 1 {
			nm += "{" + strconv.Itoa(e.RdyCnt+1) + "}"
		}
		n.Tracef("%s<- %s%s\n", emptystruct()+" ", nm, selectStr)
	}

	if e.RdyCnt < 0 {
		n.Tracef("Edge %q RdyCnt less than zero\n", e.Name)
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
	if n.isSrc(e) {
		panic(fmt.Sprintf("Unexpected source edge %q checked for destination readiness on node %q\n", e.Name, n.Name))
	}
	if e == nil {
		return false
	}
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

			e.RdyCnt += len(*e.Data)
			if TraceLevel >= VV {
				nm := e.Name + ".Data"
				if len(*e.Data) > 1 {
					nm += "{" + strconv.Itoa(len(*e.Data)) + "}"
				}
				ev := e.Val
				var attrs string

				if TraceLevel >= VVVV {
					// add other attributes for debug purposes
					if attrs == "" {
						attrs += "\t// "
					} else {
						attrs += ","
					}
					if false {
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
					if true {
						attrs += "chan={"
						for i := range *e.Data {
							if i != 0 {
								attrs += ","
							}
							attrs += fmt.Sprintf("%v", (*e.Data)[i])
						}
						attrs += "}"
					}
				}

				if ev == nil {
					n.Tracef("%s <- <nil>%s\n", nm, attrs)
				} else {
					n.Tracef("%s <- %s%s\n", nm, String(ev), attrs)
				}
			}

			// more than one source on this edge requires ack steering
			if e.SrcCnt() > 1 && !n.IsPool() {
				if false && e.DstCnt() > 1 {
					n.Panicf("Unexpected fan-out that ends on arbirated fan-in for \"%s/%s\"\n", e.Name, e.linkName())
				}
				e.Val = n.AckWrap(e.Val, e.Ack)
			}

			e.blocked = dataBlock
			for i := range *e.Data {
				(*e.Data)[i] <- e.Val
			}
			e.blocked = noBlock

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
			e.RdyCnt++
			if e.Ack2 != nil {
				attrs := ""
				if TraceLevel >= VVVV {
					attrs += fmt.Sprintf("\t// Ack2=%p", e.Ack2)
				}
				if TraceLevel >= VV {
					n.Tracef("%s.Ack <- struct {}%s\n", e.Name, attrs)
				}
				e.blocked = ackBlock
				e.Ack2 <- struct{}{}
				e.blocked = noBlock
				e.Ack2 = nil
			} else {
				if TraceLevel >= VV {
					n.Tracef("%s.Ack <-%s\n", e.Name, " "+emptystruct())
				}
				e.blocked = ackBlock
				e.Ack <- struct{}{}
				e.blocked = noBlock
			}
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

	return *e.srcCnt
	/*
		i := 0
		for ; i < len(*e.edgeNodes) && (*e.edgeNodes)[i].srcFlag; i++ {
		}
		return i
	*/
}

// DstCnt is the number of Node's downstream of an Edge
func (e *Edge) DstCnt() int {
	return *e.dstCnt
	// return len(*e.edgeNodes) - e.SrcCnt()
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
	if i >= e.SrcCnt() || i < 0 {
		return nil
	}
	return (*e.edgeNodes)[i].node
}

// DstNode returns the ith downstream Node of an Edge
func (e *Edge) DstNode(i int) *Node {
	if i >= e.DstCnt() || i < 0 {
		return nil
	}
	i += e.SrcCnt()
	return (*e.edgeNodes)[i].node
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

// Dump prints the edge details
func (e *Edge) Dump() {
	fmt.Printf("Edge:  %+v\n", e)
	fmt.Printf("Edge:  and edgeNodes is nil?  %t\n", e.edgeNodes == nil)
}

// Same returns true if two edges are really the same
func (e *Edge) Same(e2 *Edge) bool {
	return e.Data == e2.Data
}

// SetName sets the edge name in every copy
func (e *Edge) SetName(name string) {
	el := e.allEdgesPlus()
	for _, v := range el {
		v.edge.Name = name
	}
}

// allEdgesPlus returns a slice of all the edges linked with this edge
func (e *Edge) allEdgesPlus() []*edgeNodePlus {
	el := make([]*edgeNodePlus, 0)
	for _, v := range *e.edgeNodes {
		n := v.node
		if v.srcFlag {
			// search node destinations for matching Data pointer
			for j := 0; j < n.DstCnt(); j++ {
				if n.Dsts[j] != nil && n.Dsts[j].Data == e.Data {
					el = append(el, &edgeNodePlus{edgeNode{node: n, srcFlag: true}, n.Dsts[j]})
				}
			}

		} else {
			// search node sources for matching Data pointer
			for j := 0; j < n.SrcCnt(); j++ {
				if n.Srcs[j] != nil && n.Srcs[j].Data == e.Data {
					el = append(el, &edgeNodePlus{edgeNode{node: n, srcFlag: false}, n.Srcs[j]})
				}
			}
		}
	}
	return el
}

// srcRegister registers the node with its src edge
func (e *Edge) srcRegister(n *Node) {
	(*e.dstCnt)++
	*e.edgeNodes = append(*e.edgeNodes, edgeNode{node: n, srcFlag: false})
}

// dstRegister registers the node with its dst edge
func (e *Edge) dstRegister(n *Node) {
	(*e.srcCnt)++
	k := 0
	for ; k < len(*e.edgeNodes) && (*e.edgeNodes)[k].srcFlag; k++ {
	}
	*e.edgeNodes = append(*e.edgeNodes, edgeNode{})
	copy((*e.edgeNodes)[k+1:], (*e.edgeNodes)[k:])
	(*e.edgeNodes)[k] = edgeNode{node: n, srcFlag: true}
}

// SetDotAttrs set the attribute string used for outputting this edge in dot format
// If more than one they are spread across dot edges.
func (e *Edge) SetDotAttrs(attrs []string) {
	*e.dotAttrs = attrs
}

// DotAttrs returns the attribute strings used for outputting this edge in dot format
func (e *Edge) DotAttrs() []string {
	return *e.dotAttrs
}

// linkName returns an alternate name for this edge
func (e *Edge) linkName() string {
	homeName := e.Name
	awayName := ""
	el := e.allEdgesPlus()
	for _, v := range el {
		if v.edge.Name != homeName {
			awayName = v.edge.Name
			break
		}
	}
	return awayName
}

// DumpEdgeNodes dumps all the edgeNodes for this Edge
func (e *Edge) DumpEdgeNodes() {
	f := func(srcFlag bool) string {
		if srcFlag {
			return "src"
		}
		return "dst"
	}

	for _, v := range *e.edgeNodes {
		fmt.Printf("%s:%s\n", v.node.Name, f(v.srcFlag))
	}
}

// Disconnect a node from an edge
func (e *Edge) Disconnect(n *Node) {
	for i := range *e.edgeNodes {
		if (*e.edgeNodes)[i].node == n {
			if i == 0 {
				*e.edgeNodes = (*e.edgeNodes)[1:]
				break
			}
			if i+1 < len(*e.edgeNodes) {
				*e.edgeNodes = append((*e.edgeNodes)[0:i], (*e.edgeNodes)[i+1:]...)
				break
			}
			*e.edgeNodes = (*e.edgeNodes)[:i]
		}
	}
}
