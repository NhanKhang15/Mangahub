package service

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	prefsModel "mangahub-backend/internal/modules/prefs/model"
)

var ErrNotFound = errors.New("preferences not found")

// PreferencesPatch carries the optional fields supported by UpdatePreferences.
// Pointer == nil means "leave unchanged".
type PreferencesPatch struct {
	FavoriteGenres *[]string
	Language       *string
	NSFW           *bool
}

type Repo interface {
	GetPreferences(ctx context.Context, userID primitive.ObjectID) (*prefsModel.Preferences, error)
	UpdatePreferences(ctx context.Context, userID primitive.ObjectID, patch PreferencesPatch) (*prefsModel.Preferences, error)
	Subscribe(ctx context.Context, userID primitive.ObjectID, room string) (*prefsModel.Subscription, error)
	Unsubscribe(ctx context.Context, userID primitive.ObjectID, room string) error
	ListSubscriptions(ctx context.Context, userID primitive.ObjectID) ([]*prefsModel.Subscription, error)
}

type mongoRepo struct {
	prefs *mongo.Collection
	subs  *mongo.Collection
}

func NewMongoRepo(db *mongo.Database) Repo {
	return &mongoRepo{
		prefs: db.Collection("user_preferences"),
		subs:  db.Collection("subscriptions"),
	}
}

func (r *mongoRepo) GetPreferences(ctx context.Context, userID primitive.ObjectID) (*prefsModel.Preferences, error) {
	var p prefsModel.Preferences
	err := r.prefs.FindOne(ctx, bson.M{"_id": userID}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		// Treat absence as defaults rather than 404 — every authenticated
		// user is allowed to read their preferences and the row is created
		// lazily on first update.
		return &prefsModel.Preferences{UserID: userID}, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *mongoRepo) UpdatePreferences(ctx context.Context, userID primitive.ObjectID, patch PreferencesPatch) (*prefsModel.Preferences, error) {
	now := time.Now().UTC()
	set := bson.M{"updated_at": now}
	if patch.FavoriteGenres != nil {
		set["favorite_genres"] = *patch.FavoriteGenres
	}
	if patch.Language != nil {
		set["language"] = *patch.Language
	}
	if patch.NSFW != nil {
		set["nsfw"] = *patch.NSFW
	}

	after := options.After
	upsert := true
	var out prefsModel.Preferences
	err := r.prefs.FindOneAndUpdate(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": set, "$setOnInsert": bson.M{"_id": userID}},
		&options.FindOneAndUpdateOptions{ReturnDocument: &after, Upsert: &upsert},
	).Decode(&out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *mongoRepo) Subscribe(ctx context.Context, userID primitive.ObjectID, room string) (*prefsModel.Subscription, error) {
	now := time.Now().UTC()
	after := options.After
	upsert := true
	var out prefsModel.Subscription
	err := r.subs.FindOneAndUpdate(
		ctx,
		bson.M{"user_id": userID, "room": room},
		bson.M{
			"$setOnInsert": bson.M{
				"user_id":    userID,
				"room":       room,
				"created_at": now,
			},
		},
		&options.FindOneAndUpdateOptions{ReturnDocument: &after, Upsert: &upsert},
	).Decode(&out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *mongoRepo) Unsubscribe(ctx context.Context, userID primitive.ObjectID, room string) error {
	res, err := r.subs.DeleteOne(ctx, bson.M{"user_id": userID, "room": room})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *mongoRepo) ListSubscriptions(ctx context.Context, userID primitive.ObjectID) ([]*prefsModel.Subscription, error) {
	cur, err := r.subs.Find(ctx, bson.M{"user_id": userID}, options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []*prefsModel.Subscription{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
