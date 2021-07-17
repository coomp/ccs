package rocketmq_impl

import (
	"context"
	"fmt"
	"log"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

type RocketMQMessageQueueRepository struct {
	nameServers primitive.NamesrvAddr
	producer    rocketmq.Producer
	topic       string
}

func NewRocketMQMessageQueueRepository(nameServers []string, topic string) (*RocketMQMessageQueueRepository, error) {
	log.Printf("RocketMQMessageQueueRepository: ns = %v, topic = %v", nameServers, topic)
	p, _ := rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver(nameServers)),
		producer.WithRetry(2),
	)

	err := p.Start()
	if err != nil {
		fmt.Printf("RocketMQMessageQueueRepository: start producer error: %s", err.Error())
		return nil, err
	}

	return &RocketMQMessageQueueRepository{
		nameServers: nameServers,
		producer:    p,
		topic:       topic,
	}, nil
}

func (r RocketMQMessageQueueRepository) Send(msg string) error {
	message := &primitive.Message{
		Topic: r.topic,
		Body:  []byte(msg),
	}

	res, err := r.producer.SendSync(context.Background(), message)
	log.Println(res.String())
	return err
}
