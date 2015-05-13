package flowgraph

import (
	"sync"
)

type Pool struct {
	nodes []Node  
	size int32
	free int32      
	reserved int32
	mu sync.Mutex
}

// Nodes returns the Pool's slice of Node.
func (p *Pool) Nodes() []Node { return p.nodes }

// Free is the number of available Nodes in the Pool of Node's
func (p *Pool) Free() int32 { return p.free }

// Size is the total size of the Pool of Node's.
func (p *Pool) Size() int32 { return p.size }

// Reserved is the number of Node's in the Pool kept in reserve for inputs.
func (p *Pool) Reserved() int32 { return p.reserved }

// Mutex returns the mutex for this Pool.
func (p *Pool) Mutex() *sync.Mutex { return &p.mu }

// Increase increments the number of free Node's in the pool.
func (p *Pool) Increase(incr int32) int32 { 
	p.mu.Lock()
	defer p.mu.Unlock()
	p.free += incr
	return p.free
}

func (p *Pool) Decrease(decr int32) int32 { 
	p.mu.Lock()
	defer p.mu.Unlock()
	p.free -= decr
	return p.free
}



// MakePool returns a Pool of Nodes that share both
// data channels and the source ack channel.
func MakePool(
	size, reserve int32, 
	name string, 
	srcs, dsts []Edge,
	ready NodeRdy, 
	fire NodeFire) Pool {

	var p Pool
	p.size = size
	p.nodes = MakeNodes(size)
	p.free = size-reserve
	for i:=int32(0); i<size; i++ {
		var srcs2, dsts2 []Edge
		for j := 0; j<len(srcs); j++ { srcs2 = append(srcs2, srcs[j]) }
		for j := 0; j<len(dsts); j++ { dsts2 = append(dsts2, dsts[j]) }
		p.nodes[i] = MakeNodePool("qsort", srcs2, dsts2, nil, qsortFire)
	}
	return p

}
