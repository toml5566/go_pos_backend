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

type usersProductsUri struct {
	Username string `uri:"username" binding:"required,alphanum,min=1"`
}

type createProductRequest struct {
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	Username    string    `json:"username" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Price       float64   `json:"price" binding:"required,min=0"`
	Description string    `json:"description" binding:"required"`
}

func (server *Server) createProduct(ctx *gin.Context) {
	var req createProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload) // type assertion to convert to token.Payload type
	if req.Username != authPayload.Username {
		err := errors.New("unauthorized user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.CreateProductParams{
		ID:          uuid.New(),
		UserID:      req.UserID,
		Name:        req.Name,
		Price:       utils.FormottedDecimalToString(req.Price),
		Description: req.Description,
	}

	product, err := server.store.CreateProduct(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, product)
}

type getAllProductsRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

func (server *Server) getAllProducts(ctx *gin.Context) {
	var uri usersProductsUri
	var req getAllProductsRequest

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// only allow logined users to check their own products
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload) // type assertion to convert to token.Payload type
	if uri.Username != authPayload.Username {
		err := errors.New("unauthorized user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	products, err := server.store.GetAllProducts(ctx, req.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, products)
}

type updateProductRequest struct {
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	ID          uuid.UUID `json:"id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Price       float64   `json:"price" binding:"required"`
	Description string    `json:"description"`
}

func (server *Server) updateProduct(ctx *gin.Context) {
	var req updateProductRequest
	var uri usersProductsUri

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if uri.Username != authPayload.Username {
		err := errors.New("unauthorized user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.UpdateProductParams{
		UserID:      req.UserID,
		ID:          req.ID,
		Name:        req.Name,
		Price:       utils.FormottedDecimalToString(req.Price),
		Description: req.Description,
	}

	updatedProduct, err := server.store.UpdateProduct(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, updatedProduct)
}

type deleteProductRequest struct {
	UserID    uuid.UUID `json:"user_id" binding:"required"`
	Username  string    `json:"username" binding:"required,alphanum,min=1"`
	ProductID uuid.UUID `json:"product_id" binding:"required"`
}

func (server *Server) deleteProduct(ctx *gin.Context) {
	var req deleteProductRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if req.Username != authPayload.Username {
		err := errors.New("unauthorized user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.DeleteProductParams{
		UserID: req.UserID,
		ID:     req.ProductID,
	}

	err := server.store.DeleteProduct(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, "delete successfully")
}
