-- name: CreateProduct :one
INSERT INTO products (id, user_id, name, price, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetProductsByName :many
SELECT * FROM products
WHERE user_id = $1 AND name = $2;

-- name: GetAllProducts :many
SELECT * FROM products
WHERE user_id = $1;

-- name: GetProduct :one
SELECT * FROM products
WHERE user_id = $1 AND id = $2 LIMIT 1;

-- name: UpdateProduct :one
UPDATE products
SET name = $3, price = $4, description = $5
WHERE user_id = $1 AND id = $2
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE user_id = $1 AND id = $2;