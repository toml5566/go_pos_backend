package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/toml5566/go_pos_backend/internal/database"
	"github.com/toml5566/go_pos_backend/token"
	"github.com/toml5566/go_pos_backend/utils"
)

type createOrderItemRequest struct {
	ShopName     string  `json:"shop_name" binding:"required"`
	OrderDay     string  `json:"order_day" binding:"required"`
	ProductName  string  `json:"product_name" binding:"required"`
	ProductPrice float64 `json:"product_price" binding:"required"`
	Amount       int32   `json:"amount" binding:"required"`
	Status       string  `json:"status" binding:"required"`
}

type createOrderRequest struct {
	OrderID uuid.UUID                `json:"order_id" binding:"required"`
	Orders  []createOrderItemRequest `json:"orders" binding:"required"`
}

func (server *Server) createOrders(ctx *gin.Context) {
	var orderReq createOrderRequest
	var orders []db.Order

	if err := ctx.ShouldBindJSON(&orderReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	for _, req := range orderReq.Orders {
		arg := db.CreateOrderItemParams{
			ID:           uuid.New(),
			ShopName:     req.ShopName,
			OrderID:      orderReq.OrderID,
			OrderDay:     req.OrderDay,
			ProductName:  req.ProductName,
			ProductPrice: utils.FormottedDecimalToString(req.ProductPrice),
			Amount:       req.Amount,
			Status:       req.Status,
		}

		orderItem, err := server.store.CreateOrderItem(ctx, arg)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		orders = append(orders, orderItem)
	}

	ctx.JSON(http.StatusOK, orders)
}

type updateOrderItemRequest struct {
	ID       uuid.UUID `json:"id" binding:"required"`
	ShopName string    `json:"shop_name" binding:"required"`
	Amount   int32     `json:"amount" binding:"required"`
	Status   string    `json:"status" binding:"required"`
}

func (server *Server) updateOrderItem(ctx *gin.Context) {
	var req updateOrderItemRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Username != req.ShopName {
		err := errors.New("unauthorizated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.UpdateOrderItemParams{
		ShopName: req.ShopName,
		ID:       req.ID,
		Amount:   req.Amount,
		Status:   req.Status,
	}

	updatedOrder, err := server.store.UpdateOrderItem(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, updatedOrder)
}

type deleteOrderItemRequest struct {
	ShopName string    `json:"shop_name" binding:"required"`
	ID       uuid.UUID `json:"id" binding:"required"`
}

func (server *Server) deleteOrderItem(ctx *gin.Context) {
	var req deleteOrderItemRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Username != req.ShopName {
		err := errors.New("unauthorizated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.DeleteOrderItemParams(req)

	err := server.store.DeleteOrderItem(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, textResponse("delete successfully"))
}

type getOrdersByDayRequest struct {
	ShopName string `json:"shop_name" binding:"required"`
	OrderDay string `json:"order_day" binding:"required"`
}

func (server *Server) getOrdersByDay(ctx *gin.Context) {
	var req getOrdersByDayRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Username != req.ShopName {
		err := errors.New("unauthorizated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.GetOrdersByDayParams(req)

	orders, err := server.store.GetOrdersByDay(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

type getOrdersByOrderIDRequest struct {
	ShopName string    `json:"shop_name" binding:"required"`
	OrderID  uuid.UUID `json:"order_id" binding:"required"`
}

func (server *Server) getOrdersByOrderID(ctx *gin.Context) {
	var req getOrdersByOrderIDRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetOrdersByOrderIDParams(req)

	orders, err := server.store.GetOrdersByOrderID(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, orders)
}
