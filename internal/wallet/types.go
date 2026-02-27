package wallet

type PaginationQuery struct {
	Page     int32 `form:"page,default=1" binding:"min=1"`
	PageSize int32 `form:"page_size,default=20" binding:"min=1,max=100"`
}
