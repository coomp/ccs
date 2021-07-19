package rocketmq_impl

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

type RocketMQMessageQueueRepository struct {
	producers map[string]rocketmq.Producer
}

func NewRocketMQMessageQueueRepository() (*RocketMQMessageQueueRepository, error) {
	return &RocketMQMessageQueueRepository{}, nil
}

func (r RocketMQMessageQueueRepository) makeProducer(topic string) (rocketmq.Producer, error) {
	// TODO get nameservers by topic

	mqEndpointStr := "118.195.175.6:9876"
	mqEndpoints := strings.SplitN(mqEndpointStr, ",", -1)

	p, _ := rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver(mqEndpoints)),
		producer.WithRetry(2),
	)

	err := p.Start()
	if err != nil {
		fmt.Printf("makeProducer: start producer error: %s", err.Error())
		return nil, err
	}

	return p, nil
}

func (r RocketMQMessageQueueRepository) Send(topic, msg string) error {
	message := &primitive.Message{
		Topic: topic,
		Body:  []byte(msg),
	}

	var producer rocketmq.Producer
	if p, ok := r.producers[topic]; ok {
		producer = p
	} else {
		p, err := r.makeProducer(topic)
		if err != nil {
			log.Println(err)
			return err
		}
		r.producers[topic] = p
		producer = p
	}
	res, err := producer.SendSync(context.Background(), message)
	log.Println(res.String())
	return err
}
