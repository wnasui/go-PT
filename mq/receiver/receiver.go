package receiver

import (
	"12305/model"
	"12305/repository"
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ReceiverStruct struct {
	conn      *amqp.Connection
	orderRepo repository.OrderRepoInterface
}

type Receiver interface {
	ReceiveOrder(ctx context.Context)
	StartOrderConsumer(ctx context.Context) error
}

func NewReceiver(conn *amqp.Connection, orderRepo repository.OrderRepoInterface) *ReceiverStruct {
	return &ReceiverStruct{
		conn:      conn,
		orderRepo: orderRepo,
	}
}

func (r *ReceiverStruct) ReceiveOrder(ctx context.Context) model.Order {
	ch, err := r.conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
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
		log.Fatalf("Failed to declare a queue: %v", err)
	}
	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to publish a message: %v", err)
	}
	Order := model.Order{}
	for d := range msgs {
		err := json.Unmarshal(d.Body, &Order)
		if err != nil {
			log.Fatalf("Failed to unmarshal body: %v", err)
		}
	}
	return Order
}

// StartOrderConsumer 启动订单消息消费者
func (r *ReceiverStruct) StartOrderConsumer(ctx context.Context) error {
	ch, err := r.conn.Channel()
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

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Println("开始监听订单消息队列...")

	for {
		select {
		case <-ctx.Done():
			log.Println("订单消费者已停止")
			return nil
		case d := <-msgs:
			var order model.Order
			if err := json.Unmarshal(d.Body, &order); err != nil {
				log.Printf("解析订单消息失败: %v", err)
				continue
			}

			// 将消息存储到数据库
			if err := r.orderRepo.ProcessOrderFromMQ(ctx, &order); err != nil {
				log.Printf("处理订单消息失败: %v", err)
				continue
			}

			log.Printf("成功处理订单: %s", order.OrderId)
		}
	}
}
