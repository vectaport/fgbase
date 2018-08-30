package weblab

import (
	"fmt" 

	"github.com/shopify/sarama"
	"github.com/vectaport/fgbase"
)

func kprodFire (n *fgbase.Node) {

	a := n.Srcs[0]
	producer := n.Aux.(sarama.AsyncProducer)
	producer.Input() <- &sarama.ProducerMessage{Topic: "test", Key: nil, Value: sarama.StringEncoder(fmt.Sprintf("%v", a.SrcGet()))}

}

// FuncKprod wraps a Kafka producer.
func FuncKprod(a fgbase.Edge) fgbase.Node {
	node := fgbase.MakeNode("kprod", []*fgbase.Edge{&a}, nil, nil, kprodFire)

	producer, err := sarama.NewAsyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		panic(err)
	}

	node.Aux = producer

	node.RunFunc = func (n *fgbase.Node) {
		defer func() {
			producer.AsyncClose()
		}()
		n.DefaultRunFunc()
	}

	return node
}

