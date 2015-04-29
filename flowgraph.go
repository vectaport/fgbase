// Package flowgraph layers a ready-send flow mechanism on top of goroutines.
// https://github.com/vectaport/flowgraph/wiki
package flowgraph

import (
	"log"
	"os"
	"time"
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
	V         // trace Node firing
	VV        // trace channel IO
	VVV       // trace state before select
)

// Enable tracing, writes to StdoutLog if TraceLevel>Q.
var TraceLevel = Q

// Indent trace by Node id tabs.
var TraceIndent = false

// Unique Node id.
var NodeID int64

// Global count of number of Node executions.
var globalWorkCnt int64

// RunAll calls Run for each Node.
func RunAll(n []Node, timeout time.Duration) {
	for i:=0; i<len(n); i++ {
		var node = n[i]
		go node.Run()
	}

	if timeout>0 { time.Sleep(timeout) }
	StdoutLog.Printf("\n")
}

// MakeGraph returns a slice of Edge and a slice of Node.
func MakeGraph(sze, szn int) ([]Edge,[]Node) {
	return MakeEdges(sze),MakeNodes(szn)
}
