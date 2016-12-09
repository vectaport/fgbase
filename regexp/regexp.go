// Package regexp builds regular expressions out of scalable flowgraphs
package regexp

import (
//	"github.com/vectaport/flowgraph"
)

type Mode int
const (
	Fail Mode = iota
	Live
)

var Modes = map[Mode]string {
	Fail: "Fail",
	Live: "Live",
}

func (m Mode) String() string {
	return Modes[m]
}

type Search struct {
	Curr string
	Orig string
	State Mode
}
