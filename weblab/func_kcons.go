package weblab

import (
	"github.com/shopify/sarama"
	"github.com/vectaport/fgbase"
)

func kconsFire(n *fgbase.Node) error {

	x := n.Dsts[0]
	partitionConsumer := n.Aux.(sarama.PartitionConsumer)
	x.DstPut(<-partitionConsumer.Messages())
	return nil

}

// FuncKcons wraps a Kafka consumer.
func FuncKcons(x fgbase.Edge, topic string) fgbase.Node {
	node := fgbase.MakeNode("kcons", nil, []*fgbase.Edge{&x}, nil, kconsFire)

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, sarama.NewConfig())
	if err != nil {
		panic(err)
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		node.Tracef("%v\n", err)
	}

	node.RunFunc = func(n *fgbase.Node) error {
		defer func() {
			partitionConsumer.AsyncClose()
			if err := consumer.Close(); err != nil {
				panic(err)
			}

		}()
		n.DefaultRunFunc()
		return nil
	}

	node.Aux = partitionConsumer

	return node
}
