package controller

import (
	"gin-app-start/internal/dto"
	"gin-app-start/internal/service"
	"gin-app-start/pkg/logger"
	"gin-app-start/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OrderController struct {
	orderService service.OrderService
}

func NewOrderController(orderService service.OrderService) *OrderController {
	return &OrderController{
		orderService: orderService,
	}
}

// CreateOrder godoc
//
//	@Summary		Create a new order
//	@Description	Create a new order with user_id, total_price, description
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.CreateOrderRequest	true	"Order information"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/api/v1/orders [post]
func (oc *OrderController) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Parameter binding failed", zap.Error(err))
		response.Error(c, 10001, "Parameter binding failed: "+err.Error())
		return
	}

	order, err := oc.orderService.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Create order failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	response.Success(c, order)
}

// GetOrderByOrderNumber godoc
//
//	@Summary		Get order by order_number
//	@Description	Get order information by order_number
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			order_number	path		string	true	"Order Number"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/api/v1/orders/{order_number} [get]
func (oc *OrderController) GetOrderByOrderNumber(c *gin.Context) {
	orderNumber := c.Param("order_number")
	order, err := oc.orderService.GetOrderByOrderNumber(c.Request.Context(), orderNumber)
	if err != nil {
		logger.Error("Get order by order number failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}
	response.Success(c, order)
}

// UpdateOrderByOrderNumber godoc
//
//	@Summary		Update order information
//	@Description	Update order information by order_number
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			order_number		path		string						true	"Order Number"
//	@Param			request	body		dto.UpdateOrderRequest	true	"Order information to update"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		404		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/api/v1/orders/{order_number} [put]
func (oc *OrderController) UpdateOrderByOrderNumber(c *gin.Context) {
	orderNumber := c.Param("order_number")
	var req dto.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Parameter binding failed", zap.Error(err))
		response.Error(c, 10001, "Parameter binding failed: "+err.Error())
		return
	}

	order, err := oc.orderService.UpdateOrderByOrderNumber(c.Request.Context(), orderNumber, &req)
	if err != nil {
		logger.Error("Update order by order number failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}
	response.Success(c, order)
}

// DeleteOrderByOrderNumber godoc
//
//	@Summary		Delete order
//	@Description	Delete order by order_number
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			order_number	path		string	true	"Order Number"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/api/v1/orders/{order_number} [delete]
func (oc *OrderController) DeleteOrderByOrderNumber(c *gin.Context) {
	orderNumber := c.Param("order_number")
	err := oc.orderService.DeleteOrderByOrderNumber(c.Request.Context(), orderNumber)
	if err != nil {
		logger.Error("Delete order by order number failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}
	response.SuccessWithMessage(c, "Deleted successfully", nil)
}

// ListOrders godoc
//
//	@Summary		List orders
//	@Description	Get paginated list of orders
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int	false	"Page number"		default(1)
//	@Param			page_size	query		int	false	"Page size"			default(10)
//	@Success		200			{object}	response.Response
//	@Failure		500			{object}	response.Response
//	@Router			/api/v1/orders [get]
func (oc *OrderController) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	orders, total, err := oc.orderService.ListOrders(c.Request.Context(), page, pageSize)
	if err != nil {
		logger.Error("List orders failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	response.SuccessWithPage(c, orders, total, page, pageSize)
}
