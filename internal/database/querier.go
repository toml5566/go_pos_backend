// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package database

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	AddMenuItem(ctx context.Context, arg AddMenuItemParams) (Menu, error)
	CreateOrderItem(ctx context.Context, arg CreateOrderItemParams) (Order, error)
	CreateProduct(ctx context.Context, arg CreateProductParams) (Product, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	DeleteMenuItem(ctx context.Context, arg DeleteMenuItemParams) error
	DeleteOrderItem(ctx context.Context, arg DeleteOrderItemParams) error
	DeleteProduct(ctx context.Context, arg DeleteProductParams) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetAllMenuItems(ctx context.Context, shopName string) ([]Menu, error)
	GetAllProducts(ctx context.Context, userID uuid.UUID) ([]Product, error)
	GetOrdersByDay(ctx context.Context, arg GetOrdersByDayParams) ([]Order, error)
	GetOrdersByOrderID(ctx context.Context, arg GetOrdersByOrderIDParams) ([]Order, error)
	GetProduct(ctx context.Context, arg GetProductParams) (Product, error)
	GetProductsByName(ctx context.Context, arg GetProductsByNameParams) ([]Product, error)
	GetUser(ctx context.Context, username string) (User, error)
	UpdateMenuItem(ctx context.Context, arg UpdateMenuItemParams) (Menu, error)
	UpdateOrderItem(ctx context.Context, arg UpdateOrderItemParams) (Order, error)
	UpdateProduct(ctx context.Context, arg UpdateProductParams) (Product, error)
}

var _ Querier = (*Queries)(nil)
