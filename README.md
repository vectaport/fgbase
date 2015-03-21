# flowgraph
Package of Go flowgraph primitives and test programs
----

Go offers direct support for concurrent programming with goroutines, channels, and the select statement.  Used together they offer all the building blocks necessary for programming distributed systems across many cores and many Unix boxes.  But they do not offer a conceptual model that gives guidance on how to avoid the problems of distributed programming, the avoidance of global bottlenecks to achieve scalability, and the avoidance of deadlock or gridlock to ensure reliability.

Flowgraphs are a distinct model of concurrent programming that augment channels with ready-send handshake mechanisms to ensure that no data is sent before the receiver is ready.  MPI Version 1 (a framework for distributed supercomputer computation) directly supports flowgraph computation, but doesn't address flow-based computation within a single Unix box.  Go with its goroutines (that are more efficient than threads according to Rob Pike) can now facilitate taking the MPI model down to whatever granularity the concurrent programmer desires.

How to use goroutines, channels, and select to implement the flowgraph model is not completely obvious, and this framework is an attempt to illustrate the necessary approach.  



