package dto

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	UserID      uint    `json:"user_id" binding:"required" example:"1"`
	TotalPrice  float64 `json:"total_price" binding:"required" example:"99.99"`
	Description string  `json:"description" binding:"omitempty" example:"Order for John Doe"`
}

// UpdateOrderRequest represents the request to update order information
type UpdateOrderRequest struct {
	TotalPrice  float64 `json:"total_price" binding:"omitempty" example:"99.99"`
	Description string  `json:"description" binding:"omitempty" example:"Order for John Doe"`
	Status      int8    `json:"status" binding:"omitempty,oneof=0 1" example:"1"`
}
