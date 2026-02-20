package request

// ListOrdersRequest — query params for listing orders by userId
type ListOrdersRequest struct {
	UserID     string `form:"userId"     binding:"required"`
	MerchantID string `form:"merchantId"`
	Status     string `form:"status"` // optional filter
	Page       int    `form:"page,default=1"`
	PerPage    int    `form:"perPage,default=20"`
}
