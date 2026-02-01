CREATE TABLE "users"
(
    "id"       SERIAL PRIMARY KEY,
    "name"     VARCHAR(255) NOT NULL,
    "email"    VARCHAR(255) NOT NULL,
    "password" VARCHAR(255) NOT NULL,
    "is_admin" BOOLEAN      NOT NULL DEFAULT FALSE
);

ALTER TABLE "orders"
    ADD COLUMN "user_id" INT NOT NULL,
    ADD CONSTRAINT "user_id_fk" FOREIGN KEY ("user_id") REFERENCES "users" ("id");