-- name: AddMenuItem :one
INSERT INTO menus (id, user_id, shop_name, product_id, product_name, product_price, catalog, description)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateMenuItem :one
UPDATE menus
SET product_name = $3, product_price = $4, catalog = $5, description = $6
WHERE user_id = $1 AND id = $2
RETURNING *;

-- name: DeleteMenuItem :exec
DELETE FROM menus
WHERE user_id = $1 AND id = $2;

-- name: GetAllMenuItems :many
SELECT * FROM menus 
WHERE shop_name = $1;

