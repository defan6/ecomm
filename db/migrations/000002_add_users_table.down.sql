ALTER TABLE "orders"
    DROP CONSTRAINT IF EXISTS "user_id_fk";

ALTER TABLE "orders"
    DROP COLUMN IF EXISTS "user_id";

DROP TABLE IF EXISTS "users";