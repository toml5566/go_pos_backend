-- name: CreateOrderItem :one
INSERT INTO orders (id, shop_name, order_id, order_day, product_name, product_price, amount, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateOrderItem :one
UPDATE orders
SET amount = $3, status = $4
WHERE shop_name = $1 AND id = $2
RETURNING *;

-- name: DeleteOrderItem :exec
DELETE FROM orders
WHERE shop_name = $1 AND id = $2;

-- name: GetOrdersByDay :many
SELECT * FROM orders
WHERE shop_name = $1 AND order_day = $2;

-- name: GetOrdersByOrderID :many
SELECT * FROM orders
WHERE shop_name = $1 AND order_id = $2;