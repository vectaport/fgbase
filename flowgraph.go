/*
Package flowgraph layers a ready-send flow mechanism on top of goroutines.
*/

package flowgraph

import (
	"log"
	"os"
)

var nodeID int64
var globalExecCnt int64

// Log for tracing flowgraph execution
var StdoutLog = log.New(os.Stdout, "", 0)

// Enable debug tracing
var Debug = false

// Indent trace by node id
var Indent = false

// Use global execution count
var GlobalExecCnt = false

