package flowgraph

import ()

// Connector interface
type Connector interface {
	Name() string
	Value() interface{}
	Source(i int) Pipe
	Destination(i int) Pipe
}

// implementation of Connector
type conn struct {
	edge *Edge
}

// Name returns the connector name
func (c conn) Name() string {
	return c.edge.Name
}

// Value returns the connector's current value
func (c conn) Value() interface{} {
	return c.edge.Val
}

// Source returns the nth downstream pipe for this connector
func (c conn) Source(n int) Pipe {
	return pipe{c.edge.SrcNode(n)}
}

// Destination returns the nth upstream pipe for this connector
func (c conn) Destination(n int) Pipe {
	return pipe{c.edge.DstNode(n)}
}
