// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: orders.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const createOrderItem = `-- name: CreateOrderItem :one
INSERT INTO orders (id, shop_name, order_id, order_day, product_name, product_price, amount, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, shop_name, order_id, order_day, product_name, product_price, amount, status, created_at
`

type CreateOrderItemParams struct {
	ID           uuid.UUID `json:"id"`
	ShopName     string    `json:"shop_name"`
	OrderID      uuid.UUID `json:"order_id"`
	OrderDay     string    `json:"order_day"`
	ProductName  string    `json:"product_name"`
	ProductPrice string    `json:"product_price"`
	Amount       int32     `json:"amount"`
	Status       string    `json:"status"`
}

func (q *Queries) CreateOrderItem(ctx context.Context, arg CreateOrderItemParams) (Order, error) {
	row := q.db.QueryRowContext(ctx, createOrderItem,
		arg.ID,
		arg.ShopName,
		arg.OrderID,
		arg.OrderDay,
		arg.ProductName,
		arg.ProductPrice,
		arg.Amount,
		arg.Status,
	)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.ShopName,
		&i.OrderID,
		&i.OrderDay,
		&i.ProductName,
		&i.ProductPrice,
		&i.Amount,
		&i.Status,
		&i.CreatedAt,
	)
	return i, err
}

const deleteOrderItem = `-- name: DeleteOrderItem :exec
DELETE FROM orders
WHERE shop_name = $1 AND id = $2
`

type DeleteOrderItemParams struct {
	ShopName string    `json:"shop_name"`
	ID       uuid.UUID `json:"id"`
}

func (q *Queries) DeleteOrderItem(ctx context.Context, arg DeleteOrderItemParams) error {
	_, err := q.db.ExecContext(ctx, deleteOrderItem, arg.ShopName, arg.ID)
	return err
}

const getOrdersByDay = `-- name: GetOrdersByDay :many
SELECT id, shop_name, order_id, order_day, product_name, product_price, amount, status, created_at FROM orders
WHERE shop_name = $1 AND order_day = $2
`

type GetOrdersByDayParams struct {
	ShopName string `json:"shop_name"`
	OrderDay string `json:"order_day"`
}

func (q *Queries) GetOrdersByDay(ctx context.Context, arg GetOrdersByDayParams) ([]Order, error) {
	rows, err := q.db.QueryContext(ctx, getOrdersByDay, arg.ShopName, arg.OrderDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Order{}
	for rows.Next() {
		var i Order
		if err := rows.Scan(
			&i.ID,
			&i.ShopName,
			&i.OrderID,
			&i.OrderDay,
			&i.ProductName,
			&i.ProductPrice,
			&i.Amount,
			&i.Status,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getOrdersByOrderID = `-- name: GetOrdersByOrderID :many
SELECT id, shop_name, order_id, order_day, product_name, product_price, amount, status, created_at FROM orders
WHERE shop_name = $1 AND order_id = $2
`

type GetOrdersByOrderIDParams struct {
	ShopName string    `json:"shop_name"`
	OrderID  uuid.UUID `json:"order_id"`
}

func (q *Queries) GetOrdersByOrderID(ctx context.Context, arg GetOrdersByOrderIDParams) ([]Order, error) {
	rows, err := q.db.QueryContext(ctx, getOrdersByOrderID, arg.ShopName, arg.OrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Order{}
	for rows.Next() {
		var i Order
		if err := rows.Scan(
			&i.ID,
			&i.ShopName,
			&i.OrderID,
			&i.OrderDay,
			&i.ProductName,
			&i.ProductPrice,
			&i.Amount,
			&i.Status,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateOrderItem = `-- name: UpdateOrderItem :one
UPDATE orders
SET amount = $3, status = $4
WHERE shop_name = $1 AND id = $2
RETURNING id, shop_name, order_id, order_day, product_name, product_price, amount, status, created_at
`

type UpdateOrderItemParams struct {
	ShopName string    `json:"shop_name"`
	ID       uuid.UUID `json:"id"`
	Amount   int32     `json:"amount"`
	Status   string    `json:"status"`
}

func (q *Queries) UpdateOrderItem(ctx context.Context, arg UpdateOrderItemParams) (Order, error) {
	row := q.db.QueryRowContext(ctx, updateOrderItem,
		arg.ShopName,
		arg.ID,
		arg.Amount,
		arg.Status,
	)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.ShopName,
		&i.OrderID,
		&i.OrderDay,
		&i.ProductName,
		&i.ProductPrice,
		&i.Amount,
		&i.Status,
		&i.CreatedAt,
	)
	return i, err
}
