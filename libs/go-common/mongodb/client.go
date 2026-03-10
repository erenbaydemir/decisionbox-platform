package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Config struct {
	URI                    string
	Database               string
	MaxPoolSize            uint64
	MinPoolSize            uint64
	MaxConnIdleTime        time.Duration
	ConnectTimeout         time.Duration
	ServerSelectionTimeout time.Duration
}

func DefaultConfig() Config {
	return Config{
		MaxPoolSize:            100,
		MinPoolSize:            10,
		MaxConnIdleTime:        60 * time.Second,
		ConnectTimeout:         10 * time.Second,
		ServerSelectionTimeout: 10 * time.Second,
	}
}

type Client struct {
	client   *mongo.Client
	database string
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	opts := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime).
		SetConnectTimeout(cfg.ConnectTimeout).
		SetServerSelectionTimeout(cfg.ServerSelectionTimeout)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return &Client{client: client, database: cfg.Database}, nil
}

func (c *Client) Database() *mongo.Database {
	return c.client.Database(c.database)
}

func (c *Client) Collection(name string) *mongo.Collection {
	return c.Database().Collection(name)
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, readpref.Primary())
}

func (c *Client) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Raw returns the underlying mongo.Client for advanced use cases.
func (c *Client) Raw() *mongo.Client {
	return c.client
}
