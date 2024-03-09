package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	db "github.com/toml5566/go_pos_backend/internal/database"
	"github.com/toml5566/go_pos_backend/token"
	"github.com/toml5566/go_pos_backend/utils"
)

type Server struct {
	config     utils.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// create a new HTTP server and setup routing
func NewServer(config utils.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenSecretKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.RedirectTrailingSlash = false // it will redirect /users => /users/ if set to true

	// add routes to router
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	router.GET("/:shop_name/menus", server.getAllMenuItems)

	router.POST("/:shop_name/order", server.createOrders)
	router.GET("/:shop_name/order/:order_id", server.getOrdersByOrderID)

	// protected routes
	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))
	authRoutes.GET("/users/:username", server.getUser)

	authRoutes.GET("/users/:username/products", server.getAllProducts)
	authRoutes.POST("/users/:username/products", server.createProduct)
	authRoutes.PATCH("/users/:username/products/:productid", server.updateProduct)
	authRoutes.DELETE("/users/:username/products/:productid", server.deleteProduct)

	authRoutes.POST("/users/:username/menus", server.addMenuItem)
	authRoutes.PATCH("/users/:username/menus/:menu_item_id", server.updateMenuItem)
	authRoutes.DELETE("/users/:username/menus/:menu_item_id", server.deleteMenuItem)

	authRoutes.PATCH("/users/:username/orders/:order_id", server.updateOrderItem)
	authRoutes.GET("/users/:username/orders", server.getOrdersByDay)
	authRoutes.DELETE("/users/:username/orders/:order_id", server.deleteOrderItem)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func textResponse(s string) gin.H {
	return gin.H{"respond": s}
}
