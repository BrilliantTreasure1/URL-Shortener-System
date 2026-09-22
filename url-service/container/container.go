package container

import (
	"log"

	"url-shortener/config"
	"url-shortener/database"

	"github.com/redis/go-redis/v9"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Container struct {
	User   *UserContainer
	Link   *LinkContainer
	Report *ReportContainer
	Redis  *redis.Client
	MQ     *amqp.Connection
	Tracer *sdktrace.TracerProvider

	shutdownTelemetry func()
}

func NewContainer() (*Container, error) {

	db, err := config.NewDatabase()
	if err != nil {
		return nil, err
	}

	if err := database.Migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	redisClient, err := config.NewRedisClient()
	if err != nil {
		log.Printf("warn: redis unavailable, continuing without cache: %v", err)
		redisClient = nil
	}

	rabbitMQConnection, err := config.NewRabbitMQConnection()
	if err != nil {
		log.Printf("warn: rabbitmq unavailable, continuing without queue: %v", err)
		rabbitMQConnection = nil
	}

	var traceProvider *sdktrace.TracerProvider
	var telemetryShutdown func()

	provider, shutdown, err := config.NewTraceProvider()
	if err != nil {
		log.Printf("warn: tracing unavailable, continuing without traces: %v", err)
	} else {
		traceProvider = provider
		telemetryShutdown = shutdown
		otel.SetTracerProvider(provider)
	}

	userContainer, err := NewUserContainer(db)
	if err != nil {
		return nil, err
	}

	linkContainer, err := NewLinkContainer(db, redisClient, rabbitMQConnection, config.NewCacheTTL())
	if err != nil {
		return nil, err
	}

	reportContainer, err := NewReportContainer(db)
	if err != nil {
		return nil, err
	}

	return &Container{
		User:              userContainer,
		Link:              linkContainer,
		Report:            reportContainer,
		Redis:             redisClient,
		MQ:                rabbitMQConnection,
		Tracer:            traceProvider,
		shutdownTelemetry: telemetryShutdown,
	}, nil
}

func (c *Container) Close() {
	if c.MQ != nil {
		c.MQ.Close()
	}

	if c.Redis != nil {
		c.Redis.Close()
	}

	if c.shutdownTelemetry != nil {
		c.shutdownTelemetry()
	}
}