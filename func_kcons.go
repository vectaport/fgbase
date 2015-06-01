package flowgraph

import (
	"github.com/shopify/sarama"
)

func kconsFire (n *Node) {

	x := n.Dsts[0]
	partitionConsumer := x.Aux.(sarama.PartitionConsumer)
	x.Val = <- partitionConsumer.Messages() 

}

// FuncKcons wraps a Kafka consumer.
func FuncKcons(x Edge, topic string) Node {
	node := MakeNode("kcons", nil, []*Edge{&x}, nil, kconsFire)

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, sarama.NewConfig())
	if err != nil {
		panic(err)
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		StderrLog.Printf("%v\n", err)
	}

	node.RunFunc = func (n *Node) {
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

