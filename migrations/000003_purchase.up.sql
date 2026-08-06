CREATE TYPE purchase_status AS ENUM ('pending_payment', 'paid', 'payment_failed', 'expired');
CREATE TYPE seller_order_status AS ENUM ('pending_payment', 'paid', 'cancelled');
CREATE TYPE payment_status AS ENUM ('pending', 'succeeded', 'failed', 'expired');
CREATE TYPE reservation_status AS ENUM ('active', 'converted', 'released', 'expired');

CREATE TABLE carts (
    id uuid PRIMARY KEY,
    buyer_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    currency char(3) NOT NULL DEFAULT 'IDR' CHECK (currency = 'IDR'),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE cart_items (
    cart_id uuid NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    variant_id uuid NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    quantity integer NOT NULL CHECK (quantity BETWEEN 1 AND 99),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (cart_id, variant_id)
);

CREATE INDEX cart_items_variant_idx ON cart_items (variant_id);

CREATE TABLE purchases (
    id uuid PRIMARY KEY,
    reference text NOT NULL UNIQUE CHECK (reference ~ '^CL-[A-Z0-9]{12}$'),
    buyer_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    idempotency_key text NOT NULL CHECK (char_length(idempotency_key) BETWEEN 8 AND 128),
    status purchase_status NOT NULL DEFAULT 'pending_payment',
    payment_status payment_status NOT NULL DEFAULT 'pending',
    payment_intent_id text UNIQUE,
    currency char(3) NOT NULL DEFAULT 'IDR' CHECK (currency = 'IDR'),
    subtotal_minor bigint NOT NULL CHECK (subtotal_minor >= 0),
    total_minor bigint NOT NULL CHECK (total_minor >= 0),
    reservation_expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT purchases_buyer_idempotency_unique UNIQUE (buyer_id, idempotency_key),
    CONSTRAINT purchases_total_matches_subtotal CHECK (total_minor = subtotal_minor)
);

CREATE INDEX purchases_buyer_idx ON purchases (buyer_id, created_at DESC);
CREATE INDEX purchases_pending_expiry_idx ON purchases (reservation_expires_at)
    WHERE status = 'pending_payment';

CREATE TABLE seller_orders (
    id uuid PRIMARY KEY,
    purchase_id uuid NOT NULL REFERENCES purchases(id) ON DELETE RESTRICT,
    store_id uuid NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
    status seller_order_status NOT NULL DEFAULT 'pending_payment',
    currency char(3) NOT NULL DEFAULT 'IDR' CHECK (currency = 'IDR'),
    subtotal_minor bigint NOT NULL CHECK (subtotal_minor >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT seller_orders_purchase_store_unique UNIQUE (purchase_id, store_id)
);

CREATE INDEX seller_orders_store_idx ON seller_orders (store_id, created_at DESC);

CREATE TABLE purchase_items (
    id uuid PRIMARY KEY,
    seller_order_id uuid NOT NULL REFERENCES seller_orders(id) ON DELETE RESTRICT,
    product_id uuid NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    variant_id uuid NOT NULL REFERENCES product_variants(id) ON DELETE RESTRICT,
    product_name text NOT NULL,
    variant_name text NOT NULL,
    sku text NOT NULL,
    image_url text NOT NULL DEFAULT '',
    quantity integer NOT NULL CHECK (quantity > 0),
    unit_price_minor bigint NOT NULL CHECK (unit_price_minor >= 0),
    line_total_minor bigint NOT NULL CHECK (line_total_minor >= 0),
    currency char(3) NOT NULL DEFAULT 'IDR' CHECK (currency = 'IDR'),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT purchase_items_line_total_check CHECK (line_total_minor = unit_price_minor * quantity),
    CONSTRAINT purchase_items_order_variant_unique UNIQUE (seller_order_id, variant_id)
);

CREATE TABLE inventory_reservations (
    id uuid PRIMARY KEY,
    purchase_id uuid NOT NULL REFERENCES purchases(id) ON DELETE RESTRICT,
    variant_id uuid NOT NULL REFERENCES product_variants(id) ON DELETE RESTRICT,
    quantity integer NOT NULL CHECK (quantity > 0),
    status reservation_status NOT NULL DEFAULT 'active',
    expires_at timestamptz NOT NULL,
    resolved_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT inventory_reservations_purchase_variant_unique UNIQUE (purchase_id, variant_id)
);

CREATE INDEX inventory_reservations_active_expiry_idx ON inventory_reservations (expires_at)
    WHERE status = 'active';

CREATE TABLE payment_events (
    event_id text PRIMARY KEY CHECK (char_length(event_id) BETWEEN 1 AND 128),
    payment_intent_id text NOT NULL,
    event_type text NOT NULL CHECK (event_type IN ('payment.succeeded', 'payment.failed')),
    payload jsonb NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX payment_events_intent_idx ON payment_events (payment_intent_id, received_at DESC);

CREATE TABLE outbox_events (
    id uuid PRIMARY KEY,
    aggregate_type text NOT NULL,
    aggregate_id uuid NOT NULL,
    event_type text NOT NULL,
    payload jsonb NOT NULL,
    occurred_at timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz,
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    last_error text NOT NULL DEFAULT ''
);

CREATE INDEX outbox_events_unpublished_idx ON outbox_events (occurred_at, id)
    WHERE published_at IS NULL;

CREATE TABLE notifications (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_event_id uuid NOT NULL,
    kind text NOT NULL,
    title text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 160),
    body text NOT NULL CHECK (char_length(body) BETWEEN 1 AND 500),
    read_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT notifications_user_event_unique UNIQUE (user_id, source_event_id)
);

CREATE INDEX notifications_user_idx ON notifications (user_id, created_at DESC);
