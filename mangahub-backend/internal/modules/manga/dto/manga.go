package dto

type CreateMangaInput struct {
	Title       string   `json:"title"        binding:"required,min=1,max=300"`
	AltTitles   []string `json:"alt_titles"`
	ArtistIDs   []string `json:"artist_ids"`
	AuthorIDs   []string `json:"author_ids"`
	Description string   `json:"description"`
	Status      string   `json:"status"       binding:"required,oneof=ongoing completed hiatus"`
	Genres      []string `json:"genres"       binding:"required,min=1,dive,min=1"`
	Tags        []string `json:"tags"`
	Chapters    int      `json:"chapters"     binding:"gte=0"`
	Rating      float64  `json:"rating"       binding:"gte=1,lte=10"`
	CoverURL    string   `json:"cover_url"`
	Popularity  int      `json:"popularity"   binding:"gte=0"`
}

type UpdateMangaInput struct {
	Title       *string   `json:"title"        binding:"omitempty,min=1,max=300"`
	AltTitles   *[]string `json:"alt_titles"`
	ArtistIDs   *[]string `json:"artist_ids"`
	AuthorIDs   *[]string `json:"author_ids"`
	Description *string   `json:"description"`
	Status      *string   `json:"status"       binding:"omitempty,oneof=ongoing completed hiatus"`
	Genres      *[]string `json:"genres"       binding:"omitempty,min=1"`
	Tags        *[]string `json:"tags"`
	Chapters    *int      `json:"chapters"     binding:"omitempty,gte=0"`
	Rating      *float64  `json:"rating"       binding:"omitempty,gte=1,lte=10"`
	CoverURL    *string   `json:"cover_url"`
	Popularity  *int      `json:"popularity"   binding:"omitempty,gte=0"`
}

type ListMangaQuery struct {
	Page  int    `form:"page,default=1"   binding:"gte=1"`
	Limit int    `form:"limit,default=20" binding:"gte=1,lte=100"`
	Genre string `form:"genre"`
	Tags  string `form:"tags"`
	Q     string `form:"q"`
	Sort  string `form:"sort"`
}
