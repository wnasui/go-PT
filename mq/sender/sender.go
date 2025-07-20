package sender

import (
	"12305/model"
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SenderStruct struct {
	conn *amqp.Connection
}

type Sender interface {
	SendOrder(ctx context.Context, body model.Order)
}

func (s *SenderStruct) SendOrder(ctx context.Context, body model.Order) error {
	ch, err := s.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		"order",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        jsonBody,
		},
	)
	if err != nil {
		return err
	}
	return nil
}
