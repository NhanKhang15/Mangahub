package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	Mongo *mongo.Client
	DB    *mongo.Database
}

func Connect(ctx context.Context, uri, dbName string) (*Client, error) {
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cl, err := mongo.Connect(cctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	if err := cl.Ping(cctx, nil); err != nil {
		return nil, fmt.Errorf("mongo ping: %w", err)
	}
	return &Client{Mongo: cl, DB: cl.Database(dbName)}, nil
}

func (c *Client) Close(ctx context.Context) error {
	return c.Mongo.Disconnect(ctx)
}

// EnsureIndexes creates the collections (implicitly, via index creation) and
// installs all indexes described in BACKEND_PLAN.md §3.
func (c *Client) EnsureIndexes(ctx context.Context) error {
	specs := []struct {
		coll    string
		indexes []mongo.IndexModel
	}{
		{
			coll: "manga",
			indexes: []mongo.IndexModel{
				{Keys: bson.D{{Key: "title", Value: "text"}, {Key: "description", Value: "text"}}},
				{Keys: bson.D{{Key: "genres", Value: 1}}},
				{Keys: bson.D{{Key: "tags", Value: 1}}},
				{Keys: bson.D{{Key: "popularity", Value: -1}}},
				{
					Keys:    bson.D{{Key: "external_ids.mangadex", Value: 1}},
					Options: options.Index().SetUnique(true).SetSparse(true),
				},
			},
		},
		{
			coll: "artists",
			indexes: []mongo.IndexModel{
				{Keys: bson.D{{Key: "name", Value: "text"}}},
				{Keys: bson.D{{Key: "role", Value: 1}}},
			},
		},
		{
			coll: "users",
			indexes: []mongo.IndexModel{
				{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
			},
		},
		{
			coll: "reading_progress",
			indexes: []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "manga_id", Value: 1}},
					Options: options.Index().SetUnique(true),
				},
			},
		},
		{
			coll: "subscriptions",
			indexes: []mongo.IndexModel{
				{Keys: bson.D{{Key: "user_id", Value: 1}}},
				{Keys: bson.D{{Key: "room", Value: 1}}},
			},
		},
		{
			coll: "notifications",
			indexes: []mongo.IndexModel{
				{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "sent_at", Value: -1}}},
			},
		},
	}

	for _, s := range specs {
		if _, err := c.DB.Collection(s.coll).Indexes().CreateMany(ctx, s.indexes); err != nil {
			return fmt.Errorf("ensure indexes on %s: %w", s.coll, err)
		}
	}
	return nil
}
