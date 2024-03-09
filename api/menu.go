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

type addMenuItemRequest struct {
	UserID       uuid.UUID `json:"user_id" binding:"required"`
	ShopName     string    `json:"shop_name" binding:"required"`
	ProductID    uuid.UUID `json:"product_id" binding:"required"`
	ProductName  string    `json:"product_name" binding:"required"`
	ProductPrice float64   `json:"product_price" binding:"required"`
	Catalog      string    `json:"catalog" binding:"required"`
	Description  string    `json:"description" binding:"required"`
}

func (server *Server) addMenuItem(ctx *gin.Context) {
	var req addMenuItemRequest

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

	arg := db.AddMenuItemParams{
		ID:           uuid.New(),
		UserID:       req.UserID,
		ShopName:     req.ShopName,
		ProductID:    req.ProductID,
		ProductName:  req.ProductName,
		ProductPrice: utils.FormottedDecimalToString(req.ProductPrice),
		Catalog:      req.Catalog,
		Description:  req.Description,
	}

	menuItem, err := server.store.AddMenuItem(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, menuItem)
}

type updateMenuItemRequest struct {
	ID           uuid.UUID `json:"id" binding:"required"`
	UserID       uuid.UUID `json:"user_id" binding:"required"`
	ShopName     string    `json:"shop_name" binding:"required"`
	ProductName  string    `json:"product_name" binding:"required"`
	ProductPrice float64   `json:"product_price" binding:"required"`
	Catalog      string    `json:"catalog" binding:"required"`
	Description  string    `json:"description" binding:"required"`
}

func (server *Server) updateMenuItem(ctx *gin.Context) {
	var req updateMenuItemRequest

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

	arg := db.UpdateMenuItemParams{
		UserID:       req.UserID,
		ID:           req.ID,
		ProductName:  req.ProductName,
		ProductPrice: utils.FormottedDecimalToString(req.ProductPrice),
		Catalog:      req.Catalog,
		Description:  req.Description,
	}

	updatedItem, err := server.store.UpdateMenuItem(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, updatedItem)

}

type deleteMenuItemRequest struct {
	ID       uuid.UUID `json:"id" binding:"required"`
	UserID   uuid.UUID `json:"user_id" binding:"required"`
	ShopName string    `json:"shop_name" binding:"required"`
}

func (server *Server) deleteMenuItem(ctx *gin.Context) {
	var req deleteMenuItemRequest

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

	arg := db.DeleteMenuItemParams{
		UserID: req.UserID,
		ID:     req.ID,
	}

	err := server.store.DeleteMenuItem(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, textResponse("delete successfully"))
}

type getAllMenuItemsUri struct {
	ShopName string `uri:"shop_name" binding:"required"`
}

func (server *Server) getAllMenuItems(ctx *gin.Context) {
	var uri getAllMenuItemsUri

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	menuItems, err := server.store.GetAllMenuItems(ctx, uri.ShopName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, menuItems)
}
