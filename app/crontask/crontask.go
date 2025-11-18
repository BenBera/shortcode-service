package crontask

import (
	"context"
	"database/sql"

	"github.com/BenBera/shortcode-service/app/controllers"

	"github.com/go-redis/redis"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

type Crontask struct {
	RabbitMQConn *amqp.Connection
	DB           *sql.DB
	DBSlave      *sql.DB
	RedisConn    *redis.Client
	Tracer       trace.Tracer

	Controller *controllers.Controller
}

func (cron *Crontask) SetupJobs(ctx context.Context) {

	// Start a new root span representing the entire request.
	ctx, span := cron.Tracer.Start(ctx, "SetupJobs")
	defer span.End()

	//go cron.GetMarketsData(ctx)
	//go cron.SendDashboardReports(ctx)
	select {}
}
