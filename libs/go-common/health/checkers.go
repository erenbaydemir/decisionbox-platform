package health

import "context"

// MongoChecker checks MongoDB connectivity.
type MongoChecker struct {
	Pinger interface {
		Ping(ctx context.Context) error
	}
}

func (c MongoChecker) Name() string                        { return "mongodb" }
func (c MongoChecker) Check(ctx context.Context) error     { return c.Pinger.Ping(ctx) }

// RedisChecker checks Redis connectivity.
type RedisChecker struct {
	Pinger interface {
		Ping(ctx context.Context) error
	}
}

func (c RedisChecker) Name() string                        { return "redis" }
func (c RedisChecker) Check(ctx context.Context) error     { return c.Pinger.Ping(ctx) }

// RabbitMQChecker checks RabbitMQ connectivity.
type RabbitMQChecker struct {
	HealthChecker interface {
		HealthCheck() error
	}
}

func (c RabbitMQChecker) Name() string                     { return "rabbitmq" }
func (c RabbitMQChecker) Check(ctx context.Context) error  { return c.HealthChecker.HealthCheck() }
