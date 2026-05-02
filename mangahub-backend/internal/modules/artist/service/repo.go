package service

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

		artistModel "mangahub-backend/internal/modules/artist/model"
	)

var ErrNotFound = errors.New("artist not found")

type Repo interface {
	Create(ctx context.Context, a *artistModel.Artist) (primitive.ObjectID, error)
	Get(ctx context.Context, id primitive.ObjectID) (*artistModel.Artist, error)
	Update(ctx context.Context, id primitive.ObjectID, set bson.M) (*artistModel.Artist, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, q string, page, limit int) ([]*artistModel.Artist, int64, error)
}

type mongoRepo struct {
	coll *mongo.Collection
}

func NewMongoRepo(db *mongo.Database) Repo {
	return &mongoRepo{coll: db.Collection("artists")}
}

func (r *mongoRepo) Create(ctx context.Context, a *artistModel.Artist) (primitive.ObjectID, error) {
	now := time.Now().UTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	res, err := r.coll.InsertOne(ctx, a)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func (r *mongoRepo) Get(ctx context.Context, id primitive.ObjectID) (*artistModel.Artist, error) {
	var a artistModel.Artist
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&a)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *mongoRepo) Update(ctx context.Context, id primitive.ObjectID, set bson.M) (*artistModel.Artist, error) {
	set["updated_at"] = time.Now().UTC()
	var a artistModel.Artist
	after := options.After
	err := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": set},
		&options.FindOneAndUpdateOptions{ReturnDocument: &after},
	).Decode(&a)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *mongoRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *mongoRepo) List(ctx context.Context, q string, page, limit int) ([]*artistModel.Artist, int64, error) {
	filter := bson.M{}
	if q != "" {
		filter["$text"] = bson.M{"$search": q}
	}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	cur, err := r.coll.Find(ctx, filter, options.Find().
		SetSort(bson.D{{Key: "name", Value: 1}}).
		SetSkip(int64((page-1)*limit)).
		SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)
	out := []*artistModel.Artist{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
