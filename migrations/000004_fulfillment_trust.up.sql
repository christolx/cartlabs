ALTER TYPE purchase_status ADD VALUE IF NOT EXISTS 'cancelled';
ALTER TYPE payment_status ADD VALUE IF NOT EXISTS 'cancelled';
ALTER TYPE seller_order_status ADD VALUE IF NOT EXISTS 'processing';
ALTER TYPE seller_order_status ADD VALUE IF NOT EXISTS 'shipped';
ALTER TYPE seller_order_status ADD VALUE IF NOT EXISTS 'delivered';

ALTER TABLE seller_orders
    ADD COLUMN processing_at timestamptz,
    ADD COLUMN shipped_at timestamptz,
    ADD COLUMN delivered_at timestamptz,
    ADD COLUMN cancelled_at timestamptz,
    ADD COLUMN cancellation_reason text NOT NULL DEFAULT '';

CREATE TABLE reviews (
    id uuid PRIMARY KEY,
    buyer_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purchase_item_id uuid NOT NULL UNIQUE REFERENCES purchase_items(id) ON DELETE RESTRICT,
    product_id uuid NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    rating integer NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 120),
    body text NOT NULL CHECK (char_length(body) BETWEEN 1 AND 2000),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX reviews_product_idx ON reviews (product_id, created_at DESC);
CREATE INDEX reviews_buyer_idx ON reviews (buyer_id, created_at DESC);
