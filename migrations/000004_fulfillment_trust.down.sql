DROP TABLE IF EXISTS reviews;

ALTER TABLE seller_orders
    DROP COLUMN IF EXISTS cancellation_reason,
    DROP COLUMN IF EXISTS cancelled_at,
    DROP COLUMN IF EXISTS delivered_at,
    DROP COLUMN IF EXISTS shipped_at,
    DROP COLUMN IF EXISTS processing_at;
