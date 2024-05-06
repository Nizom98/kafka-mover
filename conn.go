package kafkamover

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

func NewSimpleReader(topicName, consumerGroupID string, brokers ...string) (*kafka.Reader, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("no brokers provided")
	}
	if consumerGroupID == "" {
		return nil, fmt.Errorf("no consumer group ID provided")
	}
	if topicName == "" {
		return nil, fmt.Errorf("no topic name provided")
	}

	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: consumerGroupID,
		Topic:   topicName,
	}), nil
}

func NewSimpleWriter(ctx context.Context, addr, topicName string, partition int) (*kafka.Conn, error) {
	if addr == "" {
		return nil, fmt.Errorf("no address provided")
	}
	if topicName == "" {
		return nil, fmt.Errorf("no topic name provided")
	}

	return kafka.DialLeader(
		ctx,
		"tcp",
		addr,
		topicName,
		partition,
	)
}
