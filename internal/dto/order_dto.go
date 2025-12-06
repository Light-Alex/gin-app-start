package dto

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	UserId      uint    `json:"user_id" binding:"omitempty" example:"1"`
	Username    string  `json:"username" binding:"required" example:"John Doe"`
	TotalPrice  float64 `json:"total_price" binding:"required" example:"99.99"`
	Description string  `json:"description" binding:"omitempty" example:"Order for John Doe"`
}

// GetImage represents the request to get image
type GetImage struct {
	Username  string `uri:"username" binding:"required" example:"John Doe"`
	ImageName string `uri:"imageName" binding:"required,gt=1,lte=100" example:"avatar.jpg"`
}

// UpdateOrderRequest represents the request to update order information
type UpdateOrderRequest struct {
	Username    string  `json:"username" binding:"required" example:"John Doe"`
	OrderNumber string  `json:"order_number" binding:"required" example:"123456"`
	TotalPrice  float64 `json:"total_price" binding:"omitempty" example:"99.99"`
	Description string  `json:"description" binding:"omitempty" example:"Order for John Doe"`
	Status      int8    `json:"status" binding:"omitempty,oneof=0 1" example:"1"`
}

// DeleteOrderRequest represents the request to delete an order
type DeleteOrderRequest struct {
	Username    string `json:"username" binding:"required" example:"John Doe"`
	OrderNumber string `json:"order_number" binding:"required" example:"123456"`
}
