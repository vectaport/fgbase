/*
Package flowgraph layers a ready-send flow mechanism on top of goroutines.
*/

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
	Q = iota
	V
	VV
	VVV
)

// Enable tracing, writes to StdoutLog if TraceLevel>Q.
var TraceLevel = Q

// Indent trace by node id tabs.
var TraceIndent = false


// Unique node id.
var NodeID int64

// Global count of number of Node firings.
var globalFireCnt int64

