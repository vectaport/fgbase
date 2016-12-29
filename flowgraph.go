// Package flowgraph layers a ready-send flow mechanism on top of goroutines.
// https://github.com/vectaport/flowgraph/wiki
package flowgraph

import (
	"flag"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
)

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
	QQ TraceLevelType = iota // ultra-quiet for minimal stats
	Q         // quiet, default
	V         // trace Node execution
	VV        // trace channel IO
	VVV       // trace state before select
	VVVV      // full-length array dumps
)

// Map from string to enum for trace flag checking.
var TraceLevels = map[string]TraceLevelType {
	"QQ": QQ,
	"Q": Q,
	"V": V,
	"VV": VV,
	"VVV": VVV,
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

// Unique Node id.
var NodeID int64

// Duration to run this flowgraph.
var RunTime time.Duration = -1

// Global count of number of Node executions.
var globalFireCnt int64

// Buffer size for every channel.
var ChannelSize = 1

// node channel wrapper
type nodeWrap struct {
	node *Node
	datum interface{}
	ack2 chan struct{}
}

// MakeGraph returns a slice of Edge and a slice of Node.
func MakeGraph(sze, szn int) ([]Edge,[]Node) {
	return MakeEdges(sze),MakeNodes(szn)
}

// ConfigByFlag initializes a standard set of command line arguments for flowgraph utilities,
// while at the same time parsing all other flags.  Use the defaults argument to override
// default settings for ncore, sec, trace, trsec, and chansz.  Use -help to see the standard set.
func ConfigByFlag(defaults map[string]interface{}) {

	var ncoreDef interface{} = runtime.NumCPU()-1
	var secDef interface{} = 1
	var traceDef interface{} = "V"
	var chanszDef interface{} = 1
	var trsecDef interface{} = false
	var trtypDef interface{} = false

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
	}

	ncorePtr := flag.Int("ncore", ncoreDef.(int), "# cores to use, max "+strconv.Itoa(runtime.NumCPU()))
	secPtr := flag.Int("sec", secDef.(int), "seconds to run")
	tracePtr := flag.String("trace", traceDef.(string), "trace level, QQ|Q|V|VV|VVV|VVVV")
	chanszPtr := flag.Int("chansz", chanszDef.(int), "channel size")
	trsecPtr := flag.Bool("trsec", trsecDef.(bool), "trace seconds")
	trtypPtr := flag.Bool("trtyp", trtypDef.(bool), "trace types")

	flag.Parse()

	runtime.GOMAXPROCS(*ncorePtr)
	RunTime = time.Duration(*secPtr)*time.Second
	TraceLevel = TraceLevels[*tracePtr]
	ChannelSize = *chanszPtr
	TraceSeconds = *trsecPtr
	TraceTypes = *trtypPtr
}

// When the flowgraph started running.
var StartTime time.Time

// TimeSinceStart returns time since start of running flowgraph.
func TimeSinceStart() float64 {
	if IsZero(StartTime) { return -1 }
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
