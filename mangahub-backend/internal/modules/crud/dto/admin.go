package dto

type ImportQuery struct {
	Source string `form:"source" binding:"required,oneof=mangadex myanimelist anilist all"`
	Q      string `form:"q"`
	Page   int    `form:"page,default=1"   binding:"gte=1"`
}

type ImportResult struct {
	Source   string `json:"source"`
	Query    string `json:"query"`
	Fetched  int    `json:"fetched"`
	Inserted int    `json:"inserted"`
	Updated  int    `json:"updated"`
	Skipped  int    `json:"skipped"`
}
