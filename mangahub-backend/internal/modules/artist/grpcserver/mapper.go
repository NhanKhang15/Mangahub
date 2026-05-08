package grpcserver

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"

	artistModel "mangahub-backend/internal/modules/artist/model"
	artistpb "mangahub-backend/proto/artistpb"
)

func ArtistToProto(a *artistModel.Artist) *artistpb.ArtistEntity {
	if a == nil {
		return nil
	}
	out := &artistpb.ArtistEntity{
		Id:          a.ID.Hex(),
		ExternalIds: a.ExternalIDs,
		Name:        a.Name,
		Role:        a.Role,
		Bio:         a.Bio,
		MangaIds:    objectIDsToHex(a.MangaIDs),
	}
	if !a.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(a.CreatedAt)
	}
	if !a.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(a.UpdatedAt)
	}
	return out
}

func ArtistsToProto(items []*artistModel.Artist) []*artistpb.ArtistEntity {
	out := make([]*artistpb.ArtistEntity, 0, len(items))
	for _, a := range items {
		out = append(out, ArtistToProto(a))
	}
	return out
}

func ArtistFromProto(p *artistpb.ArtistEntity) (*artistModel.Artist, error) {
	if p == nil {
		return &artistModel.Artist{}, nil
	}
	mangaIDs, err := hexesToObjectIDs(p.GetMangaIds())
	if err != nil {
		return nil, err
	}
	a := &artistModel.Artist{
		ExternalIDs: p.GetExternalIds(),
		Name:        p.GetName(),
		Role:        p.GetRole(),
		Bio:         p.GetBio(),
		MangaIDs:    mangaIDs,
	}
	if id := p.GetId(); id != "" {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		a.ID = oid
	}
	return a, nil
}

func PatchToBSON(p *artistpb.ArtistPatch) bson.M {
	if p == nil {
		return nil
	}
	set := bson.M{}
	if p.GetNameSet() {
		set["name"] = p.GetName()
	}
	if p.GetRoleSet() {
		set["role"] = p.GetRole()
	}
	if p.GetBioSet() {
		set["bio"] = p.GetBio()
	}
	if len(set) == 0 {
		return nil
	}
	return set
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
