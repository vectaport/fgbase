package weblab

import (
	"fmt" 

	"github.com/shopify/sarama"
	"github.com/vectaport/flowgraph"
)

func kprodFire (n *flowgraph.Node) {

	a := n.Srcs[0]
	producer := n.Aux.(sarama.AsyncProducer)
	producer.Input() <- &sarama.ProducerMessage{Topic: "test", Key: nil, Value: sarama.StringEncoder(fmt.Sprintf("%v", a.Val))}

}

// FuncKprod wraps a Kafka producer.
func FuncKprod(a flowgraph.Edge) flowgraph.Node {
	node := flowgraph.MakeNode("kprod", []*flowgraph.Edge{&a}, nil, nil, kprodFire)

	producer, err := sarama.NewAsyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		panic(err)
	}

	node.Aux = producer

	node.RunFunc = func (n *flowgraph.Node) {
		defer func() {
			producer.AsyncClose()
		}()
		n.DefaultRunFunc()
	}

	return node
}

