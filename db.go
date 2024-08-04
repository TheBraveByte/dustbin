package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrDBConnectionFailed = errors.New("cannot connect to the database")
	ErrPingFailed         = errors.New("cannot ping the database")
)

// SetConnection function is used to set a connection to the MongoDB database
func SetConnection(uri string) (*mongo.Client, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 100*time.Second)

	defer cancelCtx()

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions))
	if err != nil {
		return nil, fmt.Errorf("%w:%q", ErrDBConnectionFailed, err.Error())
	}
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("%w:%q", ErrPingFailed, err.Error())
	}

	return client, nil
}

// OpenConnection function is used to open a connection to the MongoDB database
func OpenConnection(uri string) *mongo.Client {
	connectCount := 0

	slog.Info("........... Setting Connection to MongoDB ...........")

	for {

		client, err := SetConnection(uri)

		if err != nil {
			slog.Info(".......... MongoDB not ready for connection ..........")
			connectCount++
		} else {
			slog.Info(".......... MongoDB client Connected  ..........")
			return client
		}

		if connectCount >= 5 {
			slog.Info(err.Error())
			return nil

		}
		slog.Info(".......... MongoDB client trying to reconnect ..........")
		time.Sleep(2 * time.Second)
		continue
	}
}
