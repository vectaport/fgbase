package flowgraph

import (
	"sync"
)

type Pool struct {
	nodes []Node  
	size int
	free int      
	mu sync.Mutex
}

// Nodes returns the Pool's slice of Node.
func (p *Pool) Nodes() []Node { return p.nodes }

// NumFree is the number of available Nodes in the Pool of Node's
func (p *Pool) NumFree() int { return p.free }

// Size is the total size of the Pool of Node's.
func (p *Pool) Size() int { return p.size }

// Mutex returns the mutex for this Pool.
func (p *Pool) Mutex() *sync.Mutex { return &p.mu }

// Free increments the number of free Node's in the Pool.
func (p *Pool) Free(n *Node, incr int) bool { 
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.free+incr < p.size {
		p.free += incr
		p.Trace(n)
		return true
	}
	n.LogError("Unexpected attempt to free node after pool is full.")
	return false
}

// Alloc decrements the number of free Node's in the Pool.
func (p *Pool) Alloc (n *Node, decr int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.free>=decr {
		p.free -= decr
		p.Trace(n)
		return true
	}
	return false
}

// Trace logs the current number of free Pool Node's using "*"
// (or "X" every ten if Pool larger than 128).
func (p *Pool) Trace(n *Node) {
	if TraceLevel>=V {
		delta,c := 1,"*"
		if p.size > 128 {
			delta = 10
			c = "X"
		}
		n.Tracef("\tpool \t%d\t%s\n", p.free, func() string {var s string; for i:=0; i<p.free; i +=delta { s += c }; return s}())
	}
}



// MakePool returns a Pool of Nodes that share both
// data channels and the source ack channel.
func MakePool(
	size int, 
	name string, 
	srcs, dsts []Edge,
	ready NodeRdy, 
	fire NodeFire,
        recurse bool) Pool {

	var p Pool
	p.size = size
	p.nodes = MakeNodes(size)
	p.free = size
	for i:=0; i<size; i++ {
		var srcs2, dsts2 []Edge
		for j := 0; j<len(srcs); j++ { srcs2 = append(srcs2, srcs[j]) }
		for j := 0; j<len(dsts); j++ { dsts2 = append(dsts2, dsts[j]) }
		p.nodes[i] = makeNodeForPool(name, srcs2, dsts2, ready, fire, recurse)
	}
	return p

}

