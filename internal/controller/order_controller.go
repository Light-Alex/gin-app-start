package controller

import (
	"net/http"
	"strconv"

	"gin-app-start/internal/code"
	"gin-app-start/internal/common"
	"gin-app-start/internal/dto"
	"gin-app-start/internal/service"
	"gin-app-start/internal/validation"
	"gin-app-start/pkg/errors"
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
func (oc *OrderController) CreateOrder() common.HandlerFunc {
	return func(c common.Context) {
		var req dto.CreateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				validation.Error(err)).WithError(err),
			)
			return
		}

		sessionData := c.SessionUserInfo()
		user, err := getUserSession(sessionData)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(err),
			)
			return
		}

		if user.UserName != common.ADMIN_NAME && user.UserName != req.Username {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New(user.UserName + " overstepping authority")),
			)
			return
		}

		req.UserId = user.UserId
		order, err := oc.orderService.CreateOrder(c, &req)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.OrderCreateError,
				code.Text(code.OrderCreateError)).WithError(err),
			)
			return
		}

		c.Payload(order)
	}
}

// GetOrderByOrderNumber godoc
//
//	@Summary		Get order by order_number
//	@Description	Get order information by order_number
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			username	    path		string	true	"Username"
//	@Param			order_number	path		string	true	"Order Number"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/api/v1/orders/search?order_number={order_number}&username={username} [get]
func (oc *OrderController) GetOrderByOrderNumber() common.HandlerFunc {
	return func(c common.Context) {
		orderNumber := c.Query("order_number")
		username := c.Query("username")

		sessionData := c.SessionUserInfo()
		user, err := getUserSession(sessionData)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(err),
			)
			return
		}

		if user.UserName != common.ADMIN_NAME && user.UserName != username {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New(user.UserName + " overstepping authority")),
			)
			return
		}

		order, err := oc.orderService.GetOrderByOrderNumber(c, orderNumber)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.OrderGetError,
				code.Text(code.OrderGetError)).WithError(err),
			)
			return
		}

		c.Payload(order)
	}
}

// UpdateOrderByOrderNumber godoc
//
//	@Summary		Update order information
//	@Description	Update order information by order_number
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.UpdateOrderRequest	true	"Order information to update"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		404		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/api/v1/orders [put]
func (oc *OrderController) UpdateOrderByOrderNumber() common.HandlerFunc {
	return func(c common.Context) {
		var req dto.UpdateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				validation.Error(err)).WithError(err),
			)
			return
		}

		sessionData := c.SessionUserInfo()
		user, err := getUserSession(sessionData)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(err),
			)
			return
		}

		if user.UserName != common.ADMIN_NAME && user.UserName != req.Username {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New(user.UserName + " overstepping authority")),
			)
			return
		}

		order, err := oc.orderService.UpdateOrderByOrderNumber(c, &req)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.OrderUpdateError,
				code.Text(code.OrderUpdateError)).WithError(err),
			)
			return
		}
		c.Payload(order)
	}
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
//	@Router			/api/v1/orders [delete]
func (oc *OrderController) DeleteOrderByOrderNumber() common.HandlerFunc {
	return func(c common.Context) {
		var req dto.DeleteOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				validation.Error(err)).WithError(err),
			)
			return
		}

		sessionData := c.SessionUserInfo()
		user, err := getUserSession(sessionData)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(err),
			)
			return
		}

		if user.UserName != common.ADMIN_NAME && user.UserName != req.Username {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New(user.UserName + " overstepping authority")),
			)
			return
		}

		order, err := oc.orderService.GetOrderByOrderNumber(c, req.OrderNumber)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.OrderGetError,
				code.Text(code.OrderGetError)).WithError(err),
			)
			return
		}

		if order == nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.OrderGetError,
				code.Text(code.OrderGetError)).WithError(errors.New("order not found")),
			)
			return
		}

		if order.Username != req.Username {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.OrderDeleteError,
				code.Text(code.OrderDeleteError)).WithError(errors.New("username and order number not match")),
			)
			return
		}

		err = oc.orderService.DeleteOrderByOrderNumber(c, req.OrderNumber)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.OrderDeleteError,
				code.Text(code.OrderDeleteError)).WithError(err),
			)
			return
		}

		c.Payload(req)
	}
}

// ListOrders godoc
//
//	@Summary		List orders
//	@Description	Get paginated list of orders
//	@Tags			orders
//	@Accept			json
//	@Produce		json
//	@Param          username    query       string     false    "Username"
//	@Param			page		query		int	       false	"Page number"		default(1)
//	@Param			page_size	query		int	       false	"Page size"			default(10)
//	@Success		200			{object}	response.Response
//	@Failure		500			{object}	response.Response
//	@Router			/api/v1/orders [get]
func (oc *OrderController) ListOrders() common.HandlerFunc {
	return func(c common.Context) {
		var res dto.ListOrdersResponse

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		username := c.DefaultQuery("username", "")

		sessionData := c.SessionUserInfo()
		user, err := getUserSession(sessionData)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(err),
			)
			return
		}

		if user.UserName != common.ADMIN_NAME && user.UserName != username {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New(user.UserName + " overstepping authority")),
			)
			return
		}

		orders, total, err := oc.orderService.ListOrders(c, username, page, pageSize)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.OrderListError,
				code.Text(code.OrderListError)).WithError(err),
			)
			return
		}
		res.Orders = orders
		res.Total = total
		c.Payload(res)
	}
}
