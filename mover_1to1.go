package kafkamover

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type (
	Sourcer interface {
		FetchMessage(ctx context.Context) (kafka.Message, error)
		CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	}

	Writer interface {
		WriteMessages(msgs ...kafka.Message) (int, error)
	}

	Move1To1 struct {
		source      Sourcer
		destination Writer

		beforeMoveToDestHandler func(msg kafka.Message) (kafka.Message, error)
	}
)

var (
	defaultBeforeMoveHandler = func(msg kafka.Message) (kafka.Message, error) {
		return msg, nil
	}
)

func New1To1Mover(source Sourcer, dest Writer) (*Move1To1, error) {
	return &Move1To1{
		source:                  source,
		destination:             dest,
		beforeMoveToDestHandler: defaultBeforeMoveHandler,
	}, nil
}

func (m *Move1To1) Move(ctx context.Context) error {
	for {
		msg, err := m.source.FetchMessage(ctx)
		if err != nil {
			return fmt.Errorf("fetch msg: %w", err)
		}

		if msg, err = m.beforeMoveToDestHandler(msg); err != nil {
			return fmt.Errorf("before move to dest: %w", err)
		}
		msgTopic := msg.Topic
		msg.Topic = ""
		if _, err = m.destination.WriteMessages(msg); err != nil {
			return fmt.Errorf("write msg: %w", err)
		}

		msg.Topic = msgTopic
		if err = m.source.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit msg: %w", err)
		}
	}
}
