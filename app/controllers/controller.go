package controllers

import (
	"database/sql"

	"github.com/go-redis/redis"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

type Controller struct {
	RabbitMQConn *amqp.Connection
	DB           *sql.DB
	DBSlave      *sql.DB
	RedisConn    *redis.Client

	Tracer trace.Tracer
}
