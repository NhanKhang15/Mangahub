package dto

type UpsertProgressInput struct {
	Status         string  `json:"status"          binding:"required,oneof=reading completed plan_to_read dropped"`
	CurrentChapter int     `json:"current_chapter" binding:"gte=0"`
	Rating         float64 `json:"rating"          binding:"omitempty,gte=1,lte=10"`
}

type ListProgressQuery struct {
	Page   int    `form:"page,default=1"   binding:"gte=1"`
	Limit  int    `form:"limit,default=20" binding:"gte=1,lte=100"`
	Status string `form:"status"           binding:"omitempty,oneof=reading completed plan_to_read dropped"`
}
