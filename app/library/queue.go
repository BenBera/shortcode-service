package library

import (
	"bitbucket.org/maybets/shortcode-service/app/constants"
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"log"
)

func Publish(ctx context.Context, conn *amqp.Connection, name string, payload interface{}, priority uint8) error {

	//js, _ := json.MarshalIndent(payload,"","\t")
	//log.Printf("publish to %s | %s",name,string(js) )

	queue := name
	exchange := name
	key := name
	exchangeType := "direct"
	ch, err := conn.Channel()
	if err != nil {

		log.Printf(" got error opening rabbitMQ channel %s ", err.Error())
		return err
	}

	defer ch.Close()

	err = ch.ExchangeDeclare(
		queue,        // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {

		log.Printf(" got error Failed to declare a queue %s error %s ", name, err.Error())
		return err
	}
	message, err := json.Marshal(payload)

	if err != nil {

		log.Printf(" got error decoding payload to string %s ", err.Error())
		return err
	}

	headers := InjectAMQPHeaders(ctx)

	msg := amqp.Publishing{
		Headers: headers,
		ContentType: "text/plain",
		Body:        message,
		Priority:    priority,
	}

	err = ch.PublishWithContext(
		ctx,
		exchange, // exchange
		key,      // routing key
		false,    // mandatory
		false,    // immediate
		msg)

	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "got error publishing message",
				constants.DATA: message,
			}).
			Error(err.Error())

		return err
	}

	return nil
}
