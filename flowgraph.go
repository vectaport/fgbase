// Package flowgraph layers a ready-send flow mechanism on top of goroutines.
// https://github.com/vectaport/flowgraph/wiki
package flowgraph

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
)

/*=====================================================================*/

type Getter interface {
	Get(p Pipe) (interface{}, error)
}

type Putter interface {
	Put(p Pipe, v interface{}) error
}

type Transformer interface {
	Transform(p Pipe, c ...interface{}) ([]interface{}, error)
}

/*=====================================================================*/

// Flowgraph interface
type Flowgraph interface {
	Name() string

	FindPipe(name string) Pipe
	FindConnector(name string) Connector

	Pipe(i int) Pipe
	Connector(i int) Connector

	NumPipe() int
	NumConnector() int

	InsertIncoming(name string, getter Getter)
	InsertOutgoing(name string, putter Putter)

	InsertConst(name string, v interface{})
	InsertArray(name string, arr []interface{})
	InsertSink(name string)

	InsertAllOf(name string, transformer Transformer)

	RunAll()
}

// implementation of Flowgraph
type graph struct {
	name  string
	nodes []Node
	edges []Edge
}

// New returns a named flowgraph
func New(nm string) Flowgraph {
	return &graph{nm, nil, nil}
}

// Name returns the name of a Flowgraph
func (fg *graph) Name() string {
	return fg.Name()
}

// Pipe returns the nth pipe in a Flowgraph
func (fg *graph) Pipe(n int) Pipe {
	return pipe{&fg.nodes[n]}
}

// Connector returns the nth connector in a Flowgraph
func (fg *graph) Connector(n int) Connector {
	return conn{&fg.edges[n]}
}

// FindPipe finds a Pipe by name
func (fg *graph) FindPipe(name string) Pipe {
	// simple search for now
	for i, v := range fg.nodes {
		if fg.nodes[i].Name == name {
			return pipe{&v}
		}
	}
	return nil
}

// FindConnector finds a Connector by name
func (fg *graph) FindConnector(name string) Connector {
	// simple search for now
	for i, v := range fg.edges {
		if fg.edges[i].Name == name {
			return conn{&v}
		}
	}
	return nil
}

// NumPipe returns the number of pipes in this graph
func (fg *graph) NumPipe() int {
	return len(fg.nodes)
}

// NumConnector returns the number of pipes in this graph
func (fg *graph) NumConnector() int {
	return len(fg.edges)
}

// InsertIncoming adds a single input source to a flowgraph that uses a Getter
func (fg *graph) InsertIncoming(name string, getter Getter) {
	e := makeEdge(fmt.Sprintf("e%d", len(fg.edges)), nil)
	fg.edges = append(fg.edges, e)
	node := FuncIncoming(e, getter)
	fg.nodes = append(fg.nodes, node)
	node.Owner = pipe{&node}
}

// InsertOutgoing adds a single output source to a flowgraph that uses a Putter
func (fg *graph) InsertOutgoing(name string, putter Putter) {
	node := FuncOutgoing(fg.edges[len(fg.edges)-1], putter)
	fg.nodes = append(fg.nodes, node)
	fg.nodes[len(fg.nodes)-1].Owner = pipe{&node}
}

// InsertAllOf adds a transform that waits for all inputs before producing outputs
func (fg *graph) InsertAllOf(name string, transformer Transformer) {
	node := FuncAllOf([]Edge{fg.edges[0]}, []Edge{fg.edges[len(fg.edges)-1]},
		name, transformer)
	fg.nodes = append(fg.nodes, node)
	node.Owner = pipe{&node}
}

// InsertConst adds an input constant as an incoming source.
func (fg *graph) InsertConst(name string, v interface{}) {
	e := makeEdge(fmt.Sprintf("e%d", len(fg.edges)), nil)
	fg.edges = append(fg.edges, e)
	node := FuncConst(e, v)
	fg.nodes = append(fg.nodes, node)
	node.Owner = pipe{&node}
}

// InsertArray adds an array as an incoming source.
func (fg *graph) InsertArray(name string, arr []interface{}) {
	e := makeEdge(fmt.Sprintf("e%d", len(fg.edges)), nil)
	fg.edges = append(fg.edges, e)
	node := FuncArray(e, arr)
	fg.nodes = append(fg.nodes, node)
	node.Owner = pipe{&node}
}

// InsertSink adds a output sink on the latest edge
func (fg *graph) InsertSink(name string) {
	i := len(fg.edges) - 1
	node := FuncSink(fg.edges[i])
	fg.nodes = append(fg.nodes, node)
	node.Owner = pipe{&node}
}

// RunAll runs the flowgraph
func (fg *graph) RunAll() {
	RunAll(fg.nodes)
}

/*=====================================================================*/

func check(e error) {
	if e != nil {
		StderrLog.Printf("ERROR:  %v\n", e)
		os.Exit(1)
	}
}

// Log for tracing flowgraph execution.
var StdoutLog = log.New(os.Stdout, "", 0)

// Log for collecting error messages.
var StderrLog = log.New(os.Stderr, "", 0)

// Compile global flowgraph stats.
var GlobalStats = false

// Trace level constants.
type TraceLevelType int

const (
	QQ   TraceLevelType = iota // ultra-quiet for minimal stats
	Q                          // quiet, default
	V                          // trace Node execution
	VV                         // trace channel IO
	VVV                        // trace state before select
	VVVV                       // full-length array dumps
)

// Map from string to enum for trace flag checking.
var TraceLevels = map[string]TraceLevelType{
	"QQ":   QQ,
	"Q":    Q,
	"V":    V,
	"VV":   VV,
	"VVV":  VVV,
	"VVVV": VVVV,
}

// String method for TraceLevelType
func (t TraceLevelType) String() string {
	return []string{
		"QQ",
		"Q",
		"V",
		"VV",
		"VVV",
		"VVVV",
	}[t]
}

// Enable tracing, writes to StdoutLog if TraceLevel>Q.
var TraceLevel = Q

// Indent trace by Node id tabs.
var TraceIndent = false

// Trace number of node executions.
var TraceFireCnt = true

// Trace elapsed seconds.
var TraceSeconds = false

// Trace types in full detail, including common types.
var TraceTypes = false

// Trace Node pointer.
var TracePointer = false

// Graphviz .dot output
var DotOutput = false

// GML output
var GmlOutput = false

// Unique Node id.
var NodeID int64

// Unique Edged id.
var EdgeID int64

// Duration to run this flowgraph.
var RunTime time.Duration = -1

// Global count of number of Node executions.
var globalFireCnt int64

// Buffer size for every channel.
var ChannelSize = 1

// node channel wrapper
type nodeWrap struct {
	node  *Node
	datum interface{}
	ack2  chan struct{}
}

// MakeGraph returns a slice of Edge and a slice of Node.
func MakeGraph(sze, szn int) ([]Edge, []Node) {
	return MakeEdges(sze), MakeNodes(szn)
}

// ConfigByFlag initializes a standard set of command line arguments for flowgraph utilities,
// while at the same time parsing all other flags.  Use the defaults argument to override
// default settings for ncore, chanz, sec, trace, trsec, trtyp, dot, and gml.  Use -help to see the standard set.
func ConfigByFlag(defaults map[string]interface{}) {

	var ncoreDef interface{} = runtime.NumCPU() - 1
	var secDef interface{} = 1
	var traceDef interface{} = "V"
	var chanszDef interface{} = 1
	var trsecDef interface{} = false
	var trtypDef interface{} = false
	var dotDef interface{} = false
	var gmlDef interface{} = false

	if defaults != nil {
		if defaults["ncore"] != nil {
			ncoreDef = defaults["ncore"]
		}
		if defaults["sec"] != nil {
			secDef = defaults["sec"]
		}
		if defaults["trace"] != nil {
			traceDef = defaults["trace"]
		}
		if defaults["chansz"] != nil {
			chanszDef = defaults["chansz"]
		}
		if defaults["trsec"] != nil {
			trsecDef = defaults["trsec"]
		}
		if defaults["trtyp"] != nil {
			trtypDef = defaults["trtyp"]
		}
		if defaults["dot"] != nil {
			dotDef = defaults["dot"]
		}
		if defaults["gml"] != nil {
			gmlDef = defaults["gml"]
		}
	}

	ncorePtr := flag.Int("ncore", ncoreDef.(int), "# cores to use, max "+strconv.Itoa(runtime.NumCPU()))
	secPtr := flag.Int("sec", secDef.(int), "seconds to run")
	tracePtr := flag.String("trace", traceDef.(string), "trace level, QQ|Q|V|VV|VVV|VVVV")
	chanszPtr := flag.Int("chansz", chanszDef.(int), "channel size")
	trsecPtr := flag.Bool("trsec", trsecDef.(bool), "trace seconds")
	trtypPtr := flag.Bool("trtyp", trtypDef.(bool), "trace types")
	dotPtr := flag.Bool("dot", dotDef.(bool), "graphviz output")
	gmlPtr := flag.Bool("gml", gmlDef.(bool), "GML output")

	flag.Parse()

	runtime.GOMAXPROCS(*ncorePtr)
	RunTime = time.Duration(*secPtr) * time.Second
	TraceLevel = TraceLevels[*tracePtr]
	ChannelSize = *chanszPtr
	TraceSeconds = *trsecPtr
	TraceTypes = *trtypPtr
	DotOutput = *dotPtr
	GmlOutput = *gmlPtr
}

// When the flowgraph started running.
var StartTime time.Time

// TimeSinceStart returns time since start of running flowgraph.
func TimeSinceStart() float64 {
	if IsZero(StartTime) {
		return -1
	}
	return time.Since(StartTime).Seconds()
}

// StringsToMap converts []string into map[string]int.
func StringsToMap(strings []string) map[string]int {
	m := make(map[string]int)
	for i := range strings {
		m[strings[i]] = i
	}
	return m
}
