package flowgraph

import (
	"fmt" 

	"github.com/shopify/sarama"
)

func kprodFire (n *Node) {

	a := n.Srcs[0]
	producer := a.Aux.(sarama.AsyncProducer)
	producer.Input() <- &sarama.ProducerMessage{Topic: "test", Key: nil, Value: sarama.StringEncoder(fmt.Sprintf("%v", a.Val))}

}

// FuncKprod wraps a Kafka producer.
func FuncKprod(a Edge) Node {
	node := MakeNode("kprod", []*Edge{&a}, nil, nil, kprodFire)

	producer, err := sarama.NewAsyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		panic(err)
	}

	a.Aux = producer

	node.RunFunc = func (n *Node) {
		defer func() {
			producer.AsyncClose()
		}()
		n.DefaultRunFunc()
	}

	return node
}

