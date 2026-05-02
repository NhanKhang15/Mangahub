package service

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	mangaModel "mangahub-backend/internal/modules/manga/model"
		)

var (
	ErrNotFound = errors.New("manga not found")
)

// UpsertAction tells callers whether UpsertByExternalIDs created a new
// document or updated an existing one.
type UpsertAction string

const (
	UpsertInserted UpsertAction = "insert"
	UpsertUpdated  UpsertAction = "update"
)

type Repo interface {
	Create(ctx context.Context, m *mangaModel.Manga) (primitive.ObjectID, error)
	Get(ctx context.Context, id primitive.ObjectID) (*mangaModel.Manga, error)
	Update(ctx context.Context, id primitive.ObjectID, set bson.M) (*mangaModel.Manga, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, q mangaModel.MangaListQuery) ([]*mangaModel.Manga, int64, error)
	ListByArtist(ctx context.Context, artistID primitive.ObjectID, page, limit int) ([]*mangaModel.Manga, int64, error)
	Popular(ctx context.Context, limit int) ([]*mangaModel.Manga, error)
	Trending(ctx context.Context, limit int) ([]*mangaModel.Manga, error)
	UpsertByExternalIDs(ctx context.Context, m *mangaModel.Manga) (UpsertAction, *mangaModel.Manga, error)
}

type mongoRepo struct {
	coll *mongo.Collection
}

func NewMongoRepo(db *mongo.Database) Repo {
	return &mongoRepo{coll: db.Collection("manga")}
}

func (r *mongoRepo) Create(ctx context.Context, m *mangaModel.Manga) (primitive.ObjectID, error) {
	now := time.Now().UTC()
	m.CreatedAt = now
	m.UpdatedAt = now
	res, err := r.coll.InsertOne(ctx, m)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func (r *mongoRepo) Get(ctx context.Context, id primitive.ObjectID) (*mangaModel.Manga, error) {
	var m mangaModel.Manga
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&m)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *mongoRepo) Update(ctx context.Context, id primitive.ObjectID, set bson.M) (*mangaModel.Manga, error) {
	set["updated_at"] = time.Now().UTC()
	var m mangaModel.Manga
	after := options.After
	err := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": set},
		&options.FindOneAndUpdateOptions{ReturnDocument: &after},
	).Decode(&m)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
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

func (r *mongoRepo) List(ctx context.Context, q mangaModel.MangaListQuery) ([]*mangaModel.Manga, int64, error) {
	filter := bson.M{}
	if q.Genre != "" {
		filter["genres"] = q.Genre
	}
	if len(q.Tags) > 0 {
		filter["tags"] = bson.M{"$all": q.Tags}
	}
	if q.Q != "" {
		filter["$text"] = bson.M{"$search": q.Q}
	}

	sort := sortSpec(q.Sort)
	skip := int64((q.Page - 1) * q.Limit)
	limit := int64(q.Limit)

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cur, err := r.coll.Find(ctx, filter, options.Find().
		SetSort(sort).
		SetSkip(skip).
		SetLimit(limit),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	out := make([]*mangaModel.Manga, 0, q.Limit)
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *mongoRepo) ListByArtist(ctx context.Context, artistID primitive.ObjectID, page, limit int) ([]*mangaModel.Manga, int64, error) {
	filter := bson.M{"artist_ids": artistID}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	cur, err := r.coll.Find(ctx, filter, options.Find().
		SetSort(bson.D{{Key: "updated_at", Value: -1}}).
		SetSkip(int64((page-1)*limit)).
		SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)
	out := []*mangaModel.Manga{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *mongoRepo) Popular(ctx context.Context, limit int) ([]*mangaModel.Manga, error) {
	cur, err := r.coll.Find(ctx, bson.M{}, options.Find().
		SetSort(bson.D{{Key: "popularity", Value: -1}}).
		SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []*mangaModel.Manga{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *mongoRepo) Trending(ctx context.Context, limit int) ([]*mangaModel.Manga, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -30)
	cur, err := r.coll.Find(ctx,
		bson.M{"updated_at": bson.M{"$gte": cutoff}},
		options.Find().
			SetSort(bson.D{{Key: "popularity", Value: -1}, {Key: "updated_at", Value: -1}}).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []*mangaModel.Manga{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpsertByExternalIDs finds an existing manga whose external_ids.<source>
// matches any entry in m.ExternalIDs, then updates it in place; otherwise it
// inserts m as a new document. Returns whether the call inserted or updated.
func (r *mongoRepo) UpsertByExternalIDs(ctx context.Context, m *mangaModel.Manga) (UpsertAction, *mangaModel.Manga, error) {
	if len(m.ExternalIDs) == 0 {
		return "", nil, errors.New("manga has no external_ids")
	}
	or := bson.A{}
	for source, id := range m.ExternalIDs {
		if source == "" || id == "" {
			continue
		}
		or = append(or, bson.M{"external_ids." + source: id})
	}
	if len(or) == 0 {
		return "", nil, errors.New("manga has no external_ids")
	}
	filter := bson.M{"$or": or}

	now := time.Now().UTC()
	set := bson.M{
		"title":        m.Title,
		"alt_titles":   m.AltTitles,
		"external_ids": m.ExternalIDs,
		"description":  m.Description,
		"status":       m.Status,
		"genres":       m.Genres,
		"tags":         m.Tags,
		"chapters":     m.Chapters,
		"rating":       m.Rating,
		"cover_url":    m.CoverURL,
		"popularity":   m.Popularity,
		"updated_at":   now,
	}
	setOnInsert := bson.M{"created_at": now}
	after := options.After
	upsert := true

	var out mangaModel.Manga
	err := r.coll.FindOneAndUpdate(
		ctx, filter,
		bson.M{"$set": set, "$setOnInsert": setOnInsert},
		&options.FindOneAndUpdateOptions{
			ReturnDocument: &after,
			Upsert:         &upsert,
		},
	).Decode(&out)
	if err != nil {
		return "", nil, err
	}

	action := UpsertUpdated
	if out.CreatedAt.Equal(now) {
		action = UpsertInserted
	}
	return action, &out, nil
}

func sortSpec(s string) bson.D {
	switch s {
	case "rating_desc":
		return bson.D{{Key: "rating", Value: -1}}
	case "popularity_desc":
		return bson.D{{Key: "popularity", Value: -1}}
	case "title_asc":
		return bson.D{{Key: "title", Value: 1}}
	case "updated_at_desc", "":
		return bson.D{{Key: "updated_at", Value: -1}}
	default:
		return bson.D{{Key: "updated_at", Value: -1}}
	}
}
