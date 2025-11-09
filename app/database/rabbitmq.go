package database

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
)

func GetRabbitMQConnection() *amqp.Connection {

	host := os.Getenv("rabbitmq_host")
	user := os.Getenv("rabbitmq_user")
	pass := os.Getenv("rabbitmq_pass")
	port := os.Getenv("rabbitmq_port")
	vhost := os.Getenv("rabbitmq_vhost")

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, pass, host, port, vhost)

	conn, err := amqp.Dial(uri)

	if err != nil {

		log.Printf("got error connecting to rabbitMQ %s with %s", err.Error(), uri)
		return nil
	}

	return conn
}
