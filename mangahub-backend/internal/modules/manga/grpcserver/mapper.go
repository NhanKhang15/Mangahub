package grpcserver

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"

	mangaModel "mangahub-backend/internal/modules/manga/model"
	catalogpb "mangahub-backend/proto/catalogpb"
)

func MangaToProto(m *mangaModel.Manga) *catalogpb.Manga {
	if m == nil {
		return nil
	}
	out := &catalogpb.Manga{
		Id:          m.ID.Hex(),
		ExternalIds: m.ExternalIDs,
		Title:       m.Title,
		AltTitles:   m.AltTitles,
		ArtistIds:   objectIDsToHex(m.ArtistIDs),
		AuthorIds:   objectIDsToHex(m.AuthorIDs),
		Description: m.Description,
		Status:      m.Status,
		Genres:      m.Genres,
		Tags:        m.Tags,
		Chapters:    int32(m.Chapters),
		Rating:      m.Rating,
		CoverUrl:    m.CoverURL,
		Popularity:  int32(m.Popularity),
	}
	if !m.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(m.CreatedAt)
	}
	if !m.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(m.UpdatedAt)
	}
	return out
}

func MangaFromProto(p *catalogpb.Manga) (*mangaModel.Manga, error) {
	if p == nil {
		return &mangaModel.Manga{}, nil
	}
	artistIDs, err := hexesToObjectIDs(p.GetArtistIds())
	if err != nil {
		return nil, err
	}
	authorIDs, err := hexesToObjectIDs(p.GetAuthorIds())
	if err != nil {
		return nil, err
	}
	m := &mangaModel.Manga{
		ExternalIDs: p.GetExternalIds(),
		Title:       p.GetTitle(),
		AltTitles:   p.GetAltTitles(),
		ArtistIDs:   artistIDs,
		AuthorIDs:   authorIDs,
		Description: p.GetDescription(),
		Status:      p.GetStatus(),
		Genres:      p.GetGenres(),
		Tags:        p.GetTags(),
		Chapters:    int(p.GetChapters()),
		Rating:      p.GetRating(),
		CoverURL:    p.GetCoverUrl(),
		Popularity:  int(p.GetPopularity()),
	}
	if id := p.GetId(); id != "" {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		m.ID = oid
	}
	if ts := p.GetCreatedAt(); ts != nil {
		m.CreatedAt = ts.AsTime()
	}
	if ts := p.GetUpdatedAt(); ts != nil {
		m.UpdatedAt = ts.AsTime()
	}
	return m, nil
}

// PatchToBSON converts a MangaPatch into a bson.M usable by Repo.Update,
// honouring the *_set presence flags. Returns nil when no fields are set.
func PatchToBSON(p *catalogpb.MangaPatch) (bson.M, error) {
	if p == nil {
		return nil, nil
	}
	set := bson.M{}
	if p.GetTitleSet() {
		set["title"] = p.GetTitle()
	}
	if p.GetAltTitlesSet() {
		set["alt_titles"] = p.GetAltTitles()
	}
	if p.GetArtistIdsSet() {
		ids, err := hexesToObjectIDs(p.GetArtistIds())
		if err != nil {
			return nil, err
		}
		set["artist_ids"] = ids
	}
	if p.GetAuthorIdsSet() {
		ids, err := hexesToObjectIDs(p.GetAuthorIds())
		if err != nil {
			return nil, err
		}
		set["author_ids"] = ids
	}
	if p.GetDescriptionSet() {
		set["description"] = p.GetDescription()
	}
	if p.GetStatusSet() {
		set["status"] = p.GetStatus()
	}
	if p.GetGenresSet() {
		set["genres"] = p.GetGenres()
	}
	if p.GetTagsSet() {
		set["tags"] = p.GetTags()
	}
	if p.GetChaptersSet() {
		set["chapters"] = int(p.GetChapters())
	}
	if p.GetRatingSet() {
		set["rating"] = p.GetRating()
	}
	if p.GetCoverUrlSet() {
		set["cover_url"] = p.GetCoverUrl()
	}
	if p.GetPopularitySet() {
		set["popularity"] = int(p.GetPopularity())
	}
	if len(set) == 0 {
		return nil, nil
	}
	return set, nil
}

func MangasToProto(items []*mangaModel.Manga) []*catalogpb.Manga {
	out := make([]*catalogpb.Manga, 0, len(items))
	for _, m := range items {
		out = append(out, MangaToProto(m))
	}
	return out
}

func objectIDsToHex(ids []primitive.ObjectID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.Hex())
	}
	return out
}

func hexesToObjectIDs(hexes []string) ([]primitive.ObjectID, error) {
	if len(hexes) == 0 {
		return nil, nil
	}
	out := make([]primitive.ObjectID, 0, len(hexes))
	for _, h := range hexes {
		oid, err := primitive.ObjectIDFromHex(h)
		if err != nil {
			return nil, err
		}
		out = append(out, oid)
	}
	return out, nil
}

// ValidateHex is a small re-export of primitive.ObjectIDFromHex for callers
// that only need to confirm a string is a well-formed object id without
// pulling the mongo driver into their imports.
func ValidateHex(s string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(s)
}
