package service

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

			progressModel "mangahub-backend/internal/modules/progress/model"
)

var ErrNotFound = errors.New("reading progress not found")

type Repo interface {
	Upsert(ctx context.Context, p *progressModel.ReadingProgress) (*progressModel.ReadingProgress, error)
	Get(ctx context.Context, userID, mangaID primitive.ObjectID) (*progressModel.ReadingProgress, error)
	List(ctx context.Context, userID primitive.ObjectID, status string, page, limit int) ([]*progressModel.ReadingProgress, int64, error)
	Delete(ctx context.Context, userID, mangaID primitive.ObjectID) error
	Stats(ctx context.Context, userID primitive.ObjectID) (map[string]int, error)
}

type mongoRepo struct {
	coll *mongo.Collection
}

func NewMongoRepo(db *mongo.Database) Repo {
	return &mongoRepo{coll: db.Collection("reading_progress")}
}

func (r *mongoRepo) Upsert(ctx context.Context, p *progressModel.ReadingProgress) (*progressModel.ReadingProgress, error) {
	p.LastReadAt = time.Now().UTC()

	set := bson.M{
		"status":          p.Status,
		"current_chapter": p.CurrentChapter,
		"last_read_at":    p.LastReadAt,
	}
	if p.Rating > 0 {
		set["rating"] = p.Rating
	}

	after := options.After
	upsert := true
	var out progressModel.ReadingProgress
	err := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"user_id": p.UserID, "manga_id": p.MangaID},
		bson.M{
			"$set":         set,
			"$setOnInsert": bson.M{"user_id": p.UserID, "manga_id": p.MangaID},
		},
		&options.FindOneAndUpdateOptions{ReturnDocument: &after, Upsert: &upsert},
	).Decode(&out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *mongoRepo) Get(ctx context.Context, userID, mangaID primitive.ObjectID) (*progressModel.ReadingProgress, error) {
	var p progressModel.ReadingProgress
	err := r.coll.FindOne(ctx, bson.M{"user_id": userID, "manga_id": mangaID}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *mongoRepo) List(ctx context.Context, userID primitive.ObjectID, status string, page, limit int) ([]*progressModel.ReadingProgress, int64, error) {
	filter := bson.M{"user_id": userID}
	if status != "" {
		filter["status"] = status
	}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	cur, err := r.coll.Find(ctx, filter, options.Find().
		SetSort(bson.D{{Key: "last_read_at", Value: -1}}).
		SetSkip(int64((page-1)*limit)).
		SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)
	out := []*progressModel.ReadingProgress{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *mongoRepo) Delete(ctx context.Context, userID, mangaID primitive.ObjectID) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"user_id": userID, "manga_id": mangaID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *mongoRepo) Stats(ctx context.Context, userID primitive.ObjectID) (map[string]int, error) {
	pipe := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"user_id": userID}}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":   "$status",
			"count": bson.M{"$sum": 1},
		}}},
	}
	cur, err := r.coll.Aggregate(ctx, pipe)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := map[string]int{}
	for cur.Next(ctx) {
		var row struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		out[row.ID] = row.Count
	}
	return out, nil
}
