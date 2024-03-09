CREATE TABLE "users" (
  "id" UUID UNIQUE PRIMARY KEY NOT NULL,
  "username" varchar UNIQUE NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "products" (
  "id" UUID UNIQUE PRIMARY KEY NOT NULL,
  "user_id" UUID,
  "name" integer NOT NULL,
  "price" "DECIMAL(10, 2)" NOT NULL,
  "description" varchar NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "menus" (
  "id" UUID UNIQUE PRIMARY KEY NOT NULL,
  "user_id" UUID,
  "name" varchar NOT NULL,
  "description" varchar NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "orders" (
  "id" UUID UNIQUE PRIMARY KEY NOT NULL,
  "user_id" UUID,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "status" varchar NOT NULL,
  "total_price" "DECIMAL(10, 2)" NOT NULL DEFAULT 0
);

CREATE TABLE "menu_products" (
  "id" UUID UNIQUE PRIMARY KEY NOT NULL,
  "menu_id" UUID,
  "product_id" UUID,
  "catalog" varchar,
  "adjusted_price" "DECIMAL(10, 2)" NOT NULL,
  "quantity" integer NOT NULL DEFAULT 0
);

CREATE TABLE "order_products" (
  "order_id" UUID,
  "product_id" UUID
);

CREATE INDEX ON "users" ("username");

CREATE INDEX ON "products" ("user_id");

CREATE INDEX ON "menus" ("user_id");

CREATE INDEX ON "orders" ("user_id");

CREATE INDEX ON "menu_products" ("menu_id");

CREATE INDEX ON "order_products" ("order_id");

ALTER TABLE "products" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "menus" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "orders" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "menu_products" ADD FOREIGN KEY ("menu_id") REFERENCES "menus" ("id");

ALTER TABLE "menu_products" ADD FOREIGN KEY ("product_id") REFERENCES "products" ("id");

ALTER TABLE "order_products" ADD FOREIGN KEY ("order_id") REFERENCES "orders" ("id");

ALTER TABLE "order_products" ADD FOREIGN KEY ("product_id") REFERENCES "menu_products" ("id");
