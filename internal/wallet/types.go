package wallet


type PaginationQuery struct {
	page int32     `form:"page,default=1"`
	pageSize int32 `form:"page_size,default=20"`
}