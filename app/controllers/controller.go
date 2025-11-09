package controllers

import (
	"database/sql"

	"github.com/BenBera/shortcode-service/app/grpc/betting"
	"github.com/BenBera/shortcode-service/app/grpc/fixture"
	"github.com/BenBera/shortcode-service/app/grpc/identity"
	"github.com/BenBera/shortcode-service/app/grpc/jackpot"
	"github.com/BenBera/shortcode-service/app/grpc/wallet"
	"github.com/go-redis/redis"
	amqp "github.com/rabbitmq/amqp091-go"
	trace "go.opentelemetry.io/otel/trace"
)

type Controller struct {
	RabbitMQConn          *amqp.Connection
	DB                    *sql.DB
	DBSlave               *sql.DB
	RedisConn             *redis.Client
	IdentityServiceClient identity.IdentityClient
	FixtureServiceClient  fixture.FixtureClient
	WalletServiceClient   wallet.WalletClient
	BettingServiceClient  betting.BettingClient
	JackpotServiceClient  jackpot.JackpotClient
	Tracer                trace.Tracer
}
