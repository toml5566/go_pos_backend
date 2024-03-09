-- +goose Up

CREATE TABLE "users" (
  "id" UUID UNIQUE PRIMARY KEY NOT NULL,
  "username" varchar UNIQUE NOT NULL CHECK (username <> ''),
  "hashed_password" varchar NOT NULL CHECK (hashed_password <> ''),
  "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "products" (
  "id" UUID UNIQUE PRIMARY KEY NOT NULL,
  "user_id" UUID NOT NULL,
  "name" varchar NOT NULL,
  "price" DECIMAL(10, 2) NOT NULL,
  "description" varchar NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "menus" (
  "id" UUID UNIQUE PRIMARY KEY NOT NULL,
  "user_id" UUID NOT NULL,
  "shop_name" varchar NOT NULL,
  "product_id" UUID NOT NULL,
  "product_name" varchar NOT NULL,
  "product_price" DECIMAL(10,2) NOT NULL DEFAULT 0,
  "catalog" varchar NOT NULL,
  "description" varchar NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "orders" (
  "id" UUID UNIQUE PRIMARY KEY NOT NULL,
  "shop_name" varchar NOT NULL,
  "order_id" UUID NOT NULL,
  "order_day" varchar NOT NULL,
  "product_name" varchar NOT NULL,
  "product_price" DECIMAL(10,2) NOT NULL DEFAULT 0,
  "amount" INTEGER NOT NULL,
  "status" varchar NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE INDEX ON "users" ("username");
CREATE INDEX ON "products" ("user_id");
CREATE INDEX ON "menus" ("user_id");
CREATE INDEX ON "orders" ("shop_name");

ALTER TABLE "products" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;
ALTER TABLE "menus" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;
ALTER TABLE "menus" ADD FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE CASCADE;
ALTER TABLE "orders" ADD FOREIGN KEY ("shop_name") REFERENCES "users" ("username") ON DELETE CASCADE;


-- +goose Down
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS menus;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS users;
