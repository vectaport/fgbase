package fgbase

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// Node of a flowgraph.
type Node struct {
	ID       int64       // unique id
	Name     string      // for tracing
	Cnt      int64       // execution count
	Srcs     []*Edge     // upstream Edge's
	Dsts     []*Edge     // downstream Edge's
	RdyFunc  NodeRdy     // func to test Edge readiness
	FireFunc NodeFire    // func to fire off the Node
	RunFunc  NodeRun     // func to repeatedly run Node
	Aux      interface{} // auxiliary empty interface to hold state
	RdyState int         // state of latest readiness
	Owner    interface{} // owner of this node

	cases         []reflect.SelectCase // select cases to read from Edge's
	caseToEdgeDir map[int]edgeDir      // map from index of selected case to associated Edge
	edgeToCase    map[*Edge]int        // map from *Edge to index of associated select case
	dataBackup    []reflect.Value      // backup data channels for inputs
	flag          uintptr              // flags for package internal use

	srcNames       []string       // source names
	dstNames       []string       // destination names
	srcIndexByName map[string]int // map of index of source Edge's by name
	dstIndexByName map[string]int // map of index of destination Edge's by name
	dotAttr        string         // attributes for dot outputs
}

type edgeDir struct {
	edge    *Edge
	srcFlag bool
}

const (
	flagPool = uintptr(1 << iota)
	flagRecurse
	flagRecursed
	flagSelecting
)

// NodeRdy is the function signature for evaluating readiness of a Node to fire.
type NodeRdy func(*Node) bool

// NodeFire is the function signature for executing a Node.
// Any error message should be written using Node.LogError and
// nil written to any output Edge.
type NodeFire func(*Node) error

// NodeRun is the function signature for an alternate Node event loop.
type NodeRun func(*Node) error

func newNode(name string, ready NodeRdy, fire NodeFire) Node {
	var n Node
	i := atomic.AddInt64(&NodeID, 1)
	n.ID = i - 1
	n.Name = name
	n.Cnt = -1
	n.RdyFunc = ready
	n.FireFunc = fire
	n.caseToEdgeDir = make(map[int]edgeDir)
	n.edgeToCase = make(map[*Edge]int)
	return n
}

func makeNode(name string, srcs, dsts []*Edge, ready NodeRdy, fire NodeFire, pool, recurse bool) Node {
	n := newNode(name, ready, fire)

	if pool {
		n.flag = n.flag | flagPool
	}
	if recurse {
		n.flag = n.flag | flagRecurse
	}

	n.Srcs = srcs
	n.Dsts = dsts
	for _, v := range n.Srcs {
		if v != nil {
			v.srcRegister(&n)
		}
	}
	for _, v := range n.Dsts {
		if v != nil {
			v.dstRegister(&n)
		}
	}

	// n.Init()

	return n

}

// Init initializes node internals after edges have been added
func (n *Node) Init() {
	var cnt = 0
	pool := (n.flag & flagPool) == flagPool
	recurse := (n.flag & flagRecurse) == flagRecurse
	for i := range n.Srcs {
		srci := n.Srcs[i]
		if srci == nil {
			break
		}
		srci.RdyCnt = func() int {
			if srci.Val != nil {
				return 0
			}
			return 1
		}()
		if srci.Data != nil {
			j := len(*srci.Data)
			if j == 0 || !pool {
				df := func() int {
					if pool && recurse {
						return 0
					}
					return ChannelSize
				}()
				*srci.Data = append(*srci.Data, make(chan interface{}, df))
			} else {
				j = 0
			}
			n.cases = append(n.cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf((*srci.Data)[j])})
			n.dataBackup = append(n.dataBackup, n.cases[cnt].Chan) // backup copy
			n.caseToEdgeDir[cnt] = edgeDir{srci, true}
			n.edgeToCase[srci] = cnt
			cnt = cnt + 1
		}
	}
	for i := range n.Dsts {
		dsti := n.Dsts[i]
		if dsti == nil {
			break
		}
		dsti.RdyCnt = 0

		if dsti.Ack != nil {
			if pool {
				dsti.Ack = make(chan struct{}, ChannelSize)
			}
			n.cases = append(n.cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(dsti.Ack)})
			n.caseToEdgeDir[cnt] = edgeDir{dsti, false}
			n.edgeToCase[dsti] = cnt
			cnt = cnt + 1
		}
	}
}

// makeNodeForPool returns a new Node with copies of source and destination Edge's.
// Both source channels and the destination data channel get shared.
// The destination ack channel is unique.
func makeNodeForPool(
	name string,
	srcs, dsts []Edge,
	ready NodeRdy,
	fire NodeFire,
	recurse bool) Node {
	var srcsp, dstsp []*Edge
	for i := 0; i < len(srcs); i++ {
		srcsp = append(srcsp, &srcs[i])
	}
	for i := 0; i < len(dsts); i++ {
		dstsp = append(dstsp, &dsts[i])
	}
	return makeNode(name, srcsp, dstsp, ready, fire, true, recurse)
}

// MakeNode returns a new Node with slices of input and output Edge's and functions for testing readiness then firing.
func MakeNode(
	name string,
	srcs, dsts []*Edge,
	ready NodeRdy,
	fire NodeFire) Node {
	return makeNode(name, srcs, dsts, ready, fire, false, false)
}

func prefixTracef(n *Node) (format string) {
	var newFmt string
	if TraceIndent {
		for i := int64(0); i < n.ID; i++ {
			newFmt += "\t"
		}
	}
	if TraceStyle == New {
		if n.Cnt == -1 {
			newFmt += "*:"
		} else {
			newFmt += fmt.Sprintf("%d:", n.Cnt)
		}
		if TraceSeconds {
			t := TimeSinceStart()
			if t >= 0.0 {
				newFmt += fmt.Sprintf("%.4f:", TimeSinceStart())
			} else {
				newFmt += "*:"
			}
		}
		newFmt += " "
	}
	newFmt += n.Name
	if TraceStyle == New {
		newFmt += fmt.Sprintf("_%d", n.ID)
	} else {
		newFmt += fmt.Sprintf("(%d", n.ID)
	}

	if TraceStyle == Old && TraceFireCnt {
		if n.Cnt >= 0 {
			newFmt += fmt.Sprintf(":%d", n.Cnt)
		} else {
			newFmt += ":*"
		}
	}

	if TraceStyle == Old && (TraceSeconds || TraceLevel >= VVVV) {
		t := TimeSinceStart()
		if t >= 0.0 {
			newFmt += fmt.Sprintf(":%.4f", TimeSinceStart())
		} else {
			newFmt += ":*"
		}
	}

	if TraceStyle == Old && TracePointer {
		newFmt += fmt.Sprintf(":%p", n)
	}

	if TraceStyle == Old {
		newFmt += ")\t"
	}
	return newFmt
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (n *Node) Tracef(format string, v ...interface{}) {
	if TraceLevel < V {
		return
	}
	newFmt := prefixTracef(n) + " "
	newFmt += format
	StdoutLog.Printf(newFmt, v...)
}

// LogError for logging of error messages.  Uses atomic log mechanism.
func (n *Node) LogError(format string, v ...interface{}) {
	// _,nm,ln,_ := runtime.Caller(1)
	newFmt := prefixTracef(n)
	newFmt += " ERROR:  "
	newFmt += format
	// newFmt += fmt.Sprintf(" -- %s:%d ", nm, ln)
	StderrLog.Printf(newFmt, v...)
}

// Panicf for quitting with formatted panic message.
func (n *Node) Panicf(format string, v ...interface{}) {
	newFmt := prefixTracef(n)
	newFmt += " ERROR:  "
	newFmt += format
	panic(fmt.Sprintf(newFmt, v...))
}

// traceValRdySrc lists Node input values
func (n *Node) traceValRdySrc(valOnly bool) string {
	newFmt := prefixTracef(n)
	if !valOnly {
		if TraceStyle == Old {
			newFmt += "<<"
		} else {
			newFmt += " <"
		}
	} else if TraceStyle == New {
		newFmt += " ("
	}
	for i := range n.Srcs {
		if i != 0 {
			newFmt += ","
		}
		srci := n.Srcs[i]
		if TracePorts && n.srcNames != nil {
			newFmt += "." + n.srcNames[i] + "("
		}
		if srci == nil {
			newFmt += "*"
		} else {
			newFmt += fmt.Sprintf("%s=", srci.Name)
			if srci.RdyZero() {
				if IsSlice(srci.Val) {
					newFmt += StringSlice(srci.Val)
				} else {
					if srci.Val == nil {
						newFmt += "<nil>"
					} else if v, ok := srci.Val.(error); ok && v.Error() == "EOF" {
						newFmt += "EOF"
					} else {
						newFmt += fmt.Sprintf("%s", String(srci.Val))
					}
				}
			} else {
				newFmt += "_"
			}
		}
		if TracePorts && n.srcNames != nil {
			newFmt += ")"
		}
		if srci.blocked == ackBlock {
			newFmt += "(α)"
		}
	}
	if TraceStyle == New && valOnly {
		newFmt += ")"
	} else if TraceStyle == Old {
		newFmt += ";"
	} else {
		newFmt += "><"
	}
	return newFmt
}

// traceValRdyDst lists Node output values or readiness.
func (n *Node) traceValRdyDst(valOnly bool) string {

	var newFmt string
	if TraceStyle == New && valOnly {
		newFmt += "("
	}
	for i := range n.Dsts {
		if i != 0 {
			newFmt += ","
		}
		dsti := n.Dsts[i]
		if TracePorts && n.dstNames != nil {
			newFmt += "." + n.dstNames[i] + "("
		}
		if dsti == nil {
			newFmt += "*"
		} else {
			dstiv := dsti.Val
			if _, ok := dstiv.(ackWrap); ok {
				dstiv = dstiv.(ackWrap).datum // remove wrapper for tracing
			}
			if valOnly {
				newFmt += fmt.Sprintf("%s=", dsti.Name)
				if !dsti.Flow {
					newFmt += "_"
				} else {
					if dstiv != nil {
						s := String(dstiv)
						if v, ok := dstiv.(error); ok && v.Error() == "EOF" {
							newFmt += "EOF"
						} else if !IsSlice(dstiv) {
							newFmt += fmt.Sprintf("%s", s)
						} else {
							newFmt += s
						}

					} else {
						newFmt += "<nil>"
					}
				}

			} else {
				newFmt += fmt.Sprintf("%s=k%v", dsti.Name, dsti.RdyCnt)
				if dsti.blocked == dataBlock {
					newFmt += "(δ)"
				}
			}
		}
		if TracePorts && n.dstNames != nil {
			newFmt += ")"
		}
	}
	if !valOnly {
		if TraceStyle == Old {
			newFmt += ">>"
		} else {
			newFmt += ">"
		}
	}
	if TraceStyle == New && valOnly {
		newFmt += ")"
	}
	if summarizing {
		newFmt += "\t// "
		if n.flag&flagSelecting == flagSelecting {
			newFmt += "!select,"
		}
		newFmt += "cases["
		for i := range n.cases {
			if i != 0 {
				newFmt += " "
			}
			if n.cases[i].Chan == reflect.ValueOf(nil) {
				newFmt += fmt.Sprintf("nil")
			} else {
				newFmt += fmt.Sprintf("%p", n.cases[i].Chan.Interface())
			}
		}
		newFmt += "]"
	}
	return newFmt
}

// TraceValRdy lists Node input values and output readiness when TraceLevel >= VVV.
func (n *Node) TraceValRdy() {
	if TraceLevel >= VVV {
		n.traceValRdy(false)
	}
}

// traceValRdy lists Node input values and output values or readiness.
func (n *Node) traceValRdy(valOnly bool) {

	newFmt := n.traceValRdySrc(valOnly)
	newFmt += n.traceValRdyDst(valOnly)
	newFmt += "\n"
	StdoutLog.Printf(newFmt)
}

// traceValRdyErr lists Node input values and output readiness to stderr.
func (n *Node) traceValRdyErr() {

	newFmt := n.traceValRdySrc(false)
	newFmt += n.traceValRdyDst(false)
	newFmt += "\n"
	StderrLog.Printf(newFmt)
}

// TraceVals lists input and output values for a Node.
func (n *Node) TraceVals() {
	if TraceLevel != Q {
		n.traceValRdy(true)
	}
}

// incrFireCnt increments execution count of Node.
func (n *Node) incrFireCnt() {
	if GlobalStats {
		c := atomic.AddInt64(&globalFireCnt, 1)
		n.Cnt = c - 1
	} else {
		n.Cnt = n.Cnt + 1
	}
}

// DefaultRdyFunc tests for everything ready.
func (n *Node) DefaultRdyFunc() bool {
	for i := range n.Srcs {
		if !n.Srcs[i].SrcRdy(n) {
			return false
		}
	}
	for i := range n.Dsts {
		if !n.Dsts[i].DstRdy(n) {
			return false
		}
	}
	return true
}

// restoreDataChannels restores data channels for next use
func (n *Node) restoreDataChannels() {
	j := 0
	for i := range n.Srcs {
		if n.Srcs[i].Data != nil {
			if n.Srcs[i].RdyCnt > 0 {
				n.cases[j].Chan = n.dataBackup[j]
			}
			j++
		}
	}

	if false {
		chanstr := "cases["
		for i := range n.cases {
			if i != 0 {
				chanstr += " "
			}
			if n.cases[i].Chan == reflect.ValueOf(nil) {
				chanstr += fmt.Sprintf("nil")
			} else {
				chanstr += fmt.Sprintf("%p", n.cases[i].Chan.Interface())
			}
		}
		chanstr += "]"

		n.Tracef("DATA CHANNEL %p RESTORED %s\n", n.cases, chanstr)
	}
}

// RdyAll tests readiness of Node to execute.
func (n *Node) RdyAll() (rdy bool) {

	if n.RdyFunc == nil {
		if !n.DefaultRdyFunc() {
			rdy = false
			return
		}
	} else {
		if !n.RdyFunc(n) {
			rdy = false
			return
		}
	}

	rdy = true
	return
}

// Fire executes Node using function pointer.
func (n *Node) Fire() error {
	var err error
	n.incrFireCnt()
	var newFmt string
	if TraceLevel > Q {
		newFmt = n.traceValRdySrc(true)
	}
	if n.FireFunc != nil {
		err = n.FireFunc(n)
	} else {
		/* Generic PASS */
		var v interface{}
		for i := range n.Srcs {
			v = n.Srcs[i].SrcGet()
			if len(n.Dsts) > i {
				n.Dsts[i].DstPut(v)
			}
		}
		for i := len(n.Srcs); i < len(n.Dsts); i++ {
			n.Dsts[i].DstPut(v)

		}
	}
	if TraceLevel > Q {
		// newFmt += "\t"
		newFmt += n.traceValRdyDst(true)
		if n.Aux != nil {
			var s = ""
			if IsStruct(n.Aux) {
				s = String(n.Aux)
				if s != "{}" {
					// newFmt += "\t// " + s
					newFmt += " // " + s
				}
			} else {
				if v, ok := n.Aux.(fmt.Stringer); ok {
					s = v.String()
					if s != "&{}" {
						newFmt += " // " + v.String()
					}
				}
			}
		}
		newFmt += "\n"
		StdoutLog.Printf(newFmt)
	}
	return err
}

// SendAll writes all data and acks after new result is computed.
func (n *Node) SendAll() bool {
	sent := false
	for i := range n.Srcs {
		sent = n.Srcs[i].SendAck(n) || sent
	}
	for i := range n.Dsts {
		sent = n.Dsts[i].SendData(n) || sent
	}
	return sent
}

// RecvOne reads one data or ack and decrements RdyCnt.
func (n *Node) RecvOne() (recvOK bool) {
	if TraceLevel >= VVV {
		n.traceValRdy(false)
	}
	if len(n.cases) == 0 {
		return false
	}
	n.flag = n.flag | flagSelecting
	i, recv, recvOK := reflect.Select(n.cases)
	n.flag = n.flag ^ flagSelecting
	if !recvOK {
		n.LogError("receive from select not ok for case %d", i)
		return false
	}
	if n.caseToEdgeDir[i].srcFlag {
		srci := n.caseToEdgeDir[i].edge
		srci.Val = recv.Interface()
		n.RemoveInputCase(srci)
		srci.srcReadHandle(n, true)
	} else {
		dsti := n.caseToEdgeDir[i].edge
		dsti.dstReadHandle(n, true)
	}
	return recvOK
}

// DefaultRunFunc is the default run func.
func (n *Node) DefaultRunFunc() error {
	var err error

	for {
		for n.RdyAll() {
			if TraceLevel >= VVV {
				n.traceValRdy(false)
			}
			err = n.Fire()
			sent := n.SendAll()
			if err != nil {
				return err
			}
			if !sent {
				break
			} // wait for external event

			// just moved inside loop so it doesn't happen
			n.restoreDataChannels()
		}
		if !n.RecvOne() { // bad receiving shuts down go-routine
			break
		}
	}
	return nil
}

// Run is an event loop that runs forever for each Node.
func (n *Node) Run() error {
	if n.RunFunc != nil {
		return n.RunFunc(n)
	}

	err := n.DefaultRunFunc()
	return err
}

// MakeNodes returns a slice of Node.
func MakeNodes(sz int) []Node {
	n := make([]Node, sz)
	return n
}

// extendChannelCaps extends the channel capacity to support arbitrated fan-in.
func extendChannelCaps(nodes []*Node) {
	// for all the nodes in the slice
	for _, n := range nodes {
		// for all the destination edges
		for j := range n.Dsts {
			dstj := n.Dsts[j]
			if dstj == nil {
				break
			}

			// if that edge has more than one upstream node and isn't a pool node
			if dstj.SrcCnt() > 1 && !n.IsPool() {

				// for all the data channels on that node shared with downstream nodes
				for k := range *dstj.Data {

					// if the capacity of that channel is less than the the number of upstream nodes
					if cap((*dstj.Data)[k]) < dstj.SrcCnt() {

						// create and plugin a new channel with greater capacity
						StdoutLog.Printf("Multiple upstream nodes on %s (len(*dstj.Data)=%d vs dstj.SrcCnt()=%d)\n", dstj.Name, len(*dstj.Data), dstj.DstCnt())
						c := make(chan interface{}, dstj.SrcCnt())
						(*dstj.Data)[k] = c

					}

				}
			}
			if false {
				if len(*dstj.Data) != dstj.DstCnt() {
					StdoutLog.Printf("Multiple downstream nodes on %s (len(*dstj.Data)=%d vs dstj.DstCnt()=%d) -- %v\n", dstj.Name, len(*dstj.Data), dstj.SrcCnt(), dstj.edgeNodes)
				}
			}

		}
	}
}

// clearUpstreamAcks increments RdyCnt upstream for every initialized downstream Edge
// (Node input edge) to reflect the fact that flow is initialized here.
func clearUpstreamAcks(nodes []*Node) {
	for _, n := range nodes {
		for j := range n.Srcs {
			if n.Srcs[j] == nil {
				break
			}
			if n.Srcs[j].Val != nil {
				n.RemoveInputCase(n.Srcs[j])
			}
		}
		for j := range n.Dsts {
			if n.Dsts[j] == nil {
				break
			}
			if n.Dsts[j].Val != nil {

				ecopies := n.Dsts[j].allEdgesPlus()
				// count up all downstream copies of this edge that have Val set
				for _, es := range ecopies {
					if !es.srcFlag && es.edge.Val != nil {
						n.Dsts[j].RdyCnt++
					}
				}
			}
		}
	}
}

// RunAll calls Run for each Node, and times out after RunTime.
func RunAll(nodes []Node) {
	// build slice of pointers
	pnodes := make([]*Node, len(nodes))
	for i := range nodes {
		pnodes[i] = &nodes[i]
	}

	runAll(pnodes)
}

// RunGraph calls Run for each *Node, and times out after RunTime.
func RunGraph(nodes []*Node) {
	runAll(nodes)
}

// runAll calls Run for each Node, and times out after RunTime.
func runAll(nodes []*Node) {

	extendChannelCaps(nodes)

	// builds node internals after edges attached
	for _, v := range nodes {
		v.Init()
	}

	if GmlOutput || DotOutput {
		if GmlOutput {
			OutputGml(nodes)
		} else {
			OutputDot(nodes)
		}
		TraceLevel = QQ
		return
	}

	clearUpstreamAcks(nodes)

	if TraceLevel >= VVVV {
		summarizing = true
		for _, n := range nodes {
			n.TraceValRdy()
		}
		StdoutLog.Printf("<<<<<<<<<<<<<<<<<<<<>>>>>>>>>>>>>>>>>>>>\n")
		summarizing = false
	}
	StartTime = time.Now()
	var wg sync.WaitGroup
	wg.Add(len(nodes))
	for i := range nodes {
		node := nodes[i]
		go func() {
			defer func() {
				wg.Done()
			}()
			node.Run()
		}()
	}

	timeout := RunTime
	if timeout > 0 {
		time.Sleep(timeout)
		if TraceLevel > QQ {
			defer StdoutLog.Printf("\n")
		}
	} else {
		wg.Wait()
	}

	if TraceLevel >= VVV {
		time.Sleep(time.Second)
		if TraceLevel >= VVVV {
			summarizing = true
		}
		StdoutLog.Printf("<<<<<<<<<<<<<<<<<<<<>>>>>>>>>>>>>>>>>>>>\n")
		for i := 0; i < len(nodes); i++ {
			nodes[i].traceValRdy(false)
		}
		/*
			for i := 0; i < len(nodes); i++ {
				nodes[i].showCases("")
			}
		*/

		summarizing = false
	}

}

// AckWrap bundles a Node pointer, and an ack channel with an empty interface, in order to
// pass information about an upstream node downstream.  Used for acking back in a Pool.
func (n *Node) AckWrap(d interface{}, ack chan struct{}) interface{} {
	return ackWrap{n, d, ack}
}

// Recursed returns true if a Node from the same Pool is upstream of this Node.
func (n *Node) Recursed() bool { return n.flag&flagRecursed == flagRecursed }

// IsPool returns true if Node is part of a Pool.
func (n *Node) IsPool() bool { return n.flag&flagPool == flagPool }

// RemoveInputCase removes the input of a Node from the select switch.
// It is restored after RdyAll.
func (n *Node) RemoveInputCase(e *Edge) {
	if !e.IsConst() {
		n.cases[n.edgeToCase[e]].Chan = reflect.ValueOf(nil) // don't read this again until after RdyAll
	}
}

// SetDotAttr set the attribute string used for outputting this node in dot format
func (n *Node) SetDotAttr(attr string) {
	n.dotAttr = attr
}

// DotAttr returns the attribute string used for outputting this node in dot format
func (n *Node) DotAttr() string {
	return n.dotAttr
}

// OutputDot outputs .dot graphviz format
func OutputDot(nodes []*Node) {

	fmt.Printf("digraph G {\n")
	fmt.Printf("graph [ ordered=\"in\" ordered=\"out\" ]\n")

	for _, iv := range nodes {
		fmt.Printf("\n// %s\n", iv.Name)
		fmt.Printf("%sɸ%d [ label=%s %s]\n", iv.Name, iv.ID, iv.Name, iv.DotAttr())
		k := 0
		for _, jv := range iv.Dsts {
			if jv == nil {
				break
			}
			for _, kv := range *jv.edgeNodes {
				if kv.srcFlag {
					continue
				}
				attr := ""
				l := len(*jv.dotAttrs)
				if l > 0 {
					attr = " " + (*jv.dotAttrs)[k%l]
				}
				fmt.Printf("%sɸ%d", iv.Name, iv.ID)
				fmt.Printf(" -> %sɸ%d", kv.node.Name, kv.node.ID)
				onm := jv.linkName()
				if onm != "" {
					onm = "/" + onm
				}
				fmt.Printf(" [ label=\"%s%s\"%s ]\n", " "+jv.Name, onm, attr)
				if attr != "" {
					k++
				}
			}
		}
	}

	fmt.Printf("}\n")

}

// OutputGml outputs .gml graph modeling language format
func OutputGml(nodes []*Node) {

	fmt.Printf("graph\n[\n")

	for _, iv := range nodes {
		fmt.Printf("  node\n  [\n   id %s_%d\n  ]\n", iv.Name, iv.ID)
	}

	for _, iv := range nodes {
		for _, jv := range iv.Dsts {
			for _, kv := range *jv.edgeNodes {
				if !kv.srcFlag {
					fmt.Printf("  edge\n  [\n   source %s_%d\n", iv.Name, iv.ID)
					fmt.Printf("   target %s_%d\n", kv.node.Name, kv.node.ID)
					fmt.Printf("   label \"%s", jv.Name)
					fmt.Printf("\"\n  ]\n")
				}
			}
		}
	}

	fmt.Printf("]\n")

}

// SrcCnt returns the number of source edges.
func (n *Node) SrcCnt() int {
	return len(n.Srcs)
}

// DstCnt returns the number of destination edges.
func (n *Node) DstCnt() int {
	return len(n.Dsts)
}

// FindSrc returns incoming edge by name
func (n *Node) FindSrc(name string) (*Edge, bool) {
	i, ok := n.FindSrcIndex(name)
	if !ok {
		return nil, false
	}
	return n.Srcs[i], true
}

// FindSrcIndex returns index of incoming edge by name
func (n *Node) FindSrcIndex(name string) (int, bool) {
	if n.srcIndexByName == nil {
		return -1, false
	}
	i, ok := n.srcIndexByName[name]
	return i, ok
}

// FindDst returns outgoing edge by name
func (n *Node) FindDst(name string) (*Edge, bool) {
	i, ok := n.FindDstIndex(name)
	if !ok {
		return nil, false
	}
	return n.Dsts[i], true
}

// FindDstIndex returns index of outgoing edge by name
func (n *Node) FindDstIndex(name string) (int, bool) {
	if n.dstIndexByName == nil {
		return -1, false
	}
	i, ok := n.dstIndexByName[name]
	return i, ok
}

// SetSrcNames names the incoming edges
func (n *Node) SetSrcNames(name ...string) {
	n.srcNames = name
	l := len(n.Srcs)
	if n.srcIndexByName == nil {
		n.srcIndexByName = make(map[string]int)
	}
	for i, v := range name {
		if i >= l {
			n.Srcs = append(n.Srcs, nil)
		}
		n.srcIndexByName[v] = i
	}
}

// SetDstNames names the outgoing edges
func (n *Node) SetDstNames(name ...string) {
	n.dstNames = name
	l := len(n.Dsts)
	if n.dstIndexByName == nil {
		n.dstIndexByName = make(map[string]int)
	}
	for i, v := range name {
		if i >= l {
			n.Dsts = append(n.Dsts, nil)
		}
		n.dstIndexByName[v] = i
	}
}

// SrcNames returns the names of the incoming edges
func (n *Node) SrcNames() []string {
	return n.srcNames
}

// DstNames returns the names of the outgoing edges
func (n *Node) DstNames() []string {
	return n.dstNames
}

// SrcByName returns the incoming edge by name
func (n *Node) SrcByName(name string) *Edge {
	return n.Srcs[n.srcIndexByName[name]]
}

// DstByName returns the outgoing edge by name
func (n *Node) DstByName(name string) *Edge {
	return n.Dsts[n.dstIndexByName[name]]
}

// SrcAppend appends an incoming edge
func (n *Node) SrcAppend(e *Edge) {
	n.Srcs = append(n.Srcs, e)
}

// DstAppend appends an outgoing edge
func (n *Node) DstAppend(e *Edge) {
	n.Dsts = append(n.Dsts, e)
}

// Src gets the edge for a source port
func (n *Node) Src(i int) *Edge {
	return n.Srcs[i]
}

// Dst gets the edge for a destination port
func (n *Node) Dst(i int) *Edge {
	return n.Dsts[i]
}

// SrcSet sets the edge for a source port
func (n *Node) SrcSet(i int, e *Edge) {
	n.Srcs[i] = e
	e.srcRegister(n)
}

// DstSet sets the edge for a destination port
func (n *Node) DstSet(i int, e *Edge) {

	// Setup new ack chan when not the first use of this Edge
	// Used to steer acks back to where the data came from using AckWrap
	if e.DstCnt() > 0 && e.SrcCnt() > 0 {
		e.Ack = make(chan struct{}, ChannelSize)
	}

	n.Dsts[i] = e
	e.dstRegister(n)
}

// SetSrcNum sets the number of source ports
func (n *Node) SetSrcNum(num int) {
	n.Srcs = make([]*Edge, num)
}

// SetDstNum sets the number of result ports
func (n *Node) SetDstNum(num int) {
	n.Dsts = make([]*Edge, num)
}

// Link links an internal stream to an external stream
func (n *Node) Link(in, ex *Edge) {

	for _, v := range in.allEdgesPlus() {
		// need to get at associated *Node and srcFlag
		vi := v.edge
		vi.Data = ex.Data
		vi.Ack = ex.Ack
		vi.edgeNodes = ex.edgeNodes
		vi.srcCnt = ex.srcCnt
		vi.dstCnt = ex.dstCnt
		if v.srcFlag {
			if !DotOutput {
				(*vi.srcCnt)--
			}
			vi.dstRegister(v.node)
		} else {
			if !DotOutput {
				(*vi.dstCnt)--
			}
			vi.srcRegister(v.node)
		}
	}
}

// String returns a string representation of the node
func (n *Node) String() string {
	srcs := ""
	for i := 0; i < n.SrcCnt(); i++ {
		if n.srcNames != nil {
			srcs += fmt.Sprintf(".%s(", n.srcNames[i])
		}
		if i != 0 {
			srcs += ","
		}
		if n.Srcs[i] == nil {
			srcs += "nil"
		} else {
			onm := n.Srcs[i].linkName()
			if onm != "" {
				onm = "/" + onm
			}
			srcs += n.Srcs[i].Name + onm
		}
		if n.srcNames != nil {
			srcs += ")"
		}
	}
	dsts := ""
	for i := 0; i < n.DstCnt(); i++ {
		if n.dstNames != nil {
			dsts += fmt.Sprintf(".%s(", n.dstNames[i])
		}
		if i != 0 {
			dsts += ","
		}
		if n.Dsts[i] == nil {
			dsts += "nil"
		} else {
			onm := n.Dsts[i].linkName()
			if onm != "" {
				onm = "/" + onm
			}
			dsts += n.Dsts[i].Name + onm
		}
		if n.dstNames != nil {
			dsts += ")"
		}
	}
	s := fmt.Sprintf("%s(%s)(%s)", n.Name, srcs, dsts)
	return s
}

// showCases prints out the current case for a node
func (n *Node) showCases(prefix string) {
	for i, c := range n.cases {
		n.Tracef("%s -- CASE %d (%s):  %+v\n", prefix, i, n.caseToEdgeDir[i].edge.Name, c.Chan)
	}
}

// isSrc checks if this is a source edge for this node
func (n *Node) isSrc(e *Edge) bool {
	for i := 0; i < len(n.Srcs); i++ {
		if n.Srcs[i] == e {
			return true
		}
	}
	return false
}
