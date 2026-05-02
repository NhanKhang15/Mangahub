package dto

type CreateArtistInput struct {
	Name string `json:"name" binding:"required,min=1,max=200"`
	Role string `json:"role" binding:"required,oneof=artist author both"`
	Bio  string `json:"bio"`
}

type UpdateArtistInput struct {
	Name *string `json:"name" binding:"omitempty,min=1,max=200"`
	Role *string `json:"role" binding:"omitempty,oneof=artist author both"`
	Bio  *string `json:"bio"`
}

type ListArtistQuery struct {
	Page  int    `form:"page,default=1"   binding:"gte=1"`
	Limit int    `form:"limit,default=20" binding:"gte=1,lte=100"`
	Q     string `form:"q"`
}

type ListArtistMangaQuery struct {
	Page  int `form:"page,default=1"   binding:"gte=1"`
	Limit int `form:"limit,default=20" binding:"gte=1,lte=100"`
}
