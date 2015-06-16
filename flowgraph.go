// Package flowgraph layers a ready-send flow mechanism on top of goroutines.
// https://github.com/vectaport/flowgraph/wiki
package flowgraph

import (
	"log"
	"os"
)

// Log for tracing flowgraph execution.
var StdoutLog = log.New(os.Stdout, "", 0)

// Log for collecting error messages.
var StderrLog = log.New(os.Stderr, "", 0)

// Compile global flowgraph stats.
var GlobalStats = false

// Trace level constants.
const (
	Q = iota  // quiet
	V         // trace Node execution
	VV        // trace channel IO
	VVV       // trace state before select
	VVVV      // full-length array dumps
)

// TraceLevels maps from string to enum for flag checking.
var TraceLevels = map[string]int {
	"Q": Q,
	"V": V,
	"VV": VV,
	"VVV": VVV,
	"VVVV": VVVV,
}

// Enable tracing, writes to StdoutLog if TraceLevel>Q.
var TraceLevel = Q

// Indent trace by Node id tabs.
var TraceIndent = false

// Trace number of node executions.
var TraceFireCnt = true

// Trace elapsed seconds.
var TraceSeconds = false

// Trace Node pointer.
var TracePointer = false

// Unique Node id.
var NodeID int64

// Global count of number of Node executions.
var globalFireCnt int64

// node channel wrapper
type nodeWrap struct {
	node *Node
	datum Datum
}

// ChannelSize is the buffer size for every channel.
var ChannelSize = 1

// MakeGraph returns a slice of Edge and a slice of Node.
func MakeGraph(sze, szn int) ([]Edge,[]Node) {
	return MakeEdges(sze),MakeNodes(szn)
}
