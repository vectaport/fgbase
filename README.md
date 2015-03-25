# flowgraph
Package of Go flowgraph primitives
----------------------------------

Go offers direct support for concurrent programming with goroutines, channels, and the select statement.  Used together they offer all the building blocks necessary for programming distributed systems across many cores and many Unix boxes.  But so much is possible with goroutines it requires the application or invention of extra concepts to construct scaleable and reliable systems that won't deadlock or be throttled by bottlenecks.

Flowgraphs are a distinct model of concurrent programming that augment channels with ready-send handshake mechanisms to ensure that no data is sent before the receiver is ready.  MPI Version 1 (a framework for distributed supercomputer computation) directly supports flowgraph computation, but doesn't address flow-based computation within a single Unix box.  Go with its goroutines (that are more efficient than threads according to Rob Pike) can now facilitate taking the MPI model down to whatever granularity the concurrent programmer desires.

How to use goroutines, channels, and select to implement the flowgraph model is not completely obvious, and this framework is an attempt to illustrate one possible approach.  

Features of github.com/vectaport/flowgraph:

* flowgraph.Edge augments a channel with a ready-send acknowledge protocol
* flowgraph.Node augments a goroutine with an empty interface data protocol
* an empty interface data protocol allows a small set of primitives to be reused for a wide variety of things
* a way of rendering flowgraphs drawn and simulated in [github.com/vectaport/ipl](http://github.com/vectaport/ipl-1.1) into compiled code




Godoc [documentation](https://godoc.org/github.com/vectaport/flowgraph)
