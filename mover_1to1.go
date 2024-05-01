package kafkamover

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type (
	Reader interface {
		FetchMessage(ctx context.Context) (kafka.Message, error)
		CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	}

	Writer interface {
		WriteMessages(msgs ...kafka.Message) (int, error)
	}

	// Move1To1 is a struct that represents a 1 to 1 message mover.
	// It reads messages from a [reader] and moves them to a [destination].
	Move1To1 struct {
		// reader is the source where messages will be read from
		reader Reader
		// destination is the destination where messages will be moved to
		destination Writer
		// beforeMoveToDestHandler is a function that will be called before moving the message to the destination.
		// It can be used to modify the message before moving it to the destination.
		beforeMoveToDestHandler func(msg kafka.Message) (kafka.Message, error)
	}

	Option[T any] func(*T) error
)

var (
	// defaultBeforeMoveHandler is a default implementation of the beforeMoveToDestHandler function.
	defaultBeforeMoveHandler = func(msg kafka.Message) (kafka.Message, error) {
		return msg, nil
	}
)

// New1To1Mover creates a new 1 to 1 mover.
// It returns an error if any of the options passed to the mover is invalid.
func New1To1Mover(reader Reader, dest Writer, opts ...Option[Move1To1]) (*Move1To1, error) {
	m := &Move1To1{
		reader:                  reader,
		destination:             dest,
		beforeMoveToDestHandler: defaultBeforeMoveHandler,
	}

	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// Move reads messages from the reader and moves them to the destination.
// It returns an error if any of the operations fail.
func (m *Move1To1) Move(ctx context.Context) error {
	for {
		msg, err := m.reader.FetchMessage(ctx)
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
		if err = m.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit msg: %w", err)
		}
	}
}

// Move1To1WithBeforeMoveHandler is an option that sets the beforeMoveToDestHandler function.
func Move1To1WithBeforeMoveHandler(fn func(msg kafka.Message) (kafka.Message, error)) Option[Move1To1] {
	return func(m *Move1To1) error {
		m.beforeMoveToDestHandler = fn
		return nil
	}
}
