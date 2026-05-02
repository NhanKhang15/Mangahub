package dto

type LimitQuery struct {
	Limit int `form:"limit,default=10" binding:"gte=1,lte=100"`
}
