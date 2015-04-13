/*
Package flowgraph layers a ready-send flow mechanism on top of goroutines.
*/

package flowgraph

import (
	"log"
	"os"
)

// Unique node id.
var nodeID int64

// Global count of number of Node's executed.
var globalExecCnt int64

// Log for tracing flowgraph execution.
var StdoutLog = log.New(os.Stdout, "", 0)

// Log for collecting error messages.
var StderrLog = log.New(os.Stderr, "", 0)

// Use global execution count.
var GlobalStats = false

// Trace level constants
const (
	Q = iota
	V
	VV
	VVV
)

// Enable execution tracing, writes to StdoutLog if TraceLevel>Q
var TraceLevel = Q

// Indent trace by node id
var TraceIndent = false


