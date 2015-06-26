package weblab

import (
	"github.com/shopify/sarama"
	"github.com/vectaport/flowgraph"
)

func kconsFire (n *flowgraph.Node) {

	x := n.Dsts[0]
	partitionConsumer := x.Aux.(sarama.PartitionConsumer)
	x.Val = <- partitionConsumer.Messages() 

}

// FuncKcons wraps a Kafka consumer.
func FuncKcons(x flowgraph.Edge, topic string) flowgraph.Node {
	node := flowgraph.MakeNode("kcons", nil, []*flowgraph.Edge{&x}, nil, kconsFire)

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, sarama.NewConfig())
	if err != nil {
		panic(err)
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		node.Tracef("%v\n", err)
	}

	node.RunFunc = func (n *flowgraph.Node) {
		defer func() {
			partitionConsumer.AsyncClose()
			if err := consumer.Close(); err != nil {
				panic(err)
			}

		}()
		n.DefaultRunFunc()
	}

	x.Aux = partitionConsumer

	return node
}

