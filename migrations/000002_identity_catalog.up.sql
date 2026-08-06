CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TYPE user_role AS ENUM ('buyer', 'seller', 'admin');
CREATE TYPE user_status AS ENUM ('active', 'suspended');
CREATE TYPE moderation_status AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE product_status AS ENUM ('draft', 'published', 'archived');

CREATE TABLE users (
    id uuid PRIMARY KEY,
    email text NOT NULL,
    password_hash text NOT NULL,
    display_name text NOT NULL CHECK (char_length(display_name) BETWEEN 1 AND 100),
    role user_role NOT NULL,
    status user_status NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_email_normalized CHECK (email = lower(email)),
    CONSTRAINT users_email_unique UNIQUE (email)
);

CREATE TABLE refresh_sessions (
    id uuid PRIMARY KEY,
    family_id uuid NOT NULL,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    rotated_at timestamptz,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX refresh_sessions_family_idx ON refresh_sessions (family_id);
CREATE INDEX refresh_sessions_user_idx ON refresh_sessions (user_id);

CREATE TABLE stores (
    id uuid PRIMARY KEY,
    seller_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
    name text NOT NULL CHECK (char_length(name) BETWEEN 2 AND 100),
    slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    description text NOT NULL DEFAULT '' CHECK (char_length(description) <= 1000),
    status moderation_status NOT NULL DEFAULT 'pending',
    moderation_note text NOT NULL DEFAULT '' CHECK (char_length(moderation_note) <= 500),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE categories (
    id uuid PRIMARY KEY,
    name text NOT NULL UNIQUE CHECK (char_length(name) BETWEEN 2 AND 80),
    slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id uuid PRIMARY KEY,
    store_id uuid NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
    category_id uuid NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    name text NOT NULL CHECK (char_length(name) BETWEEN 2 AND 160),
    slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    description text NOT NULL DEFAULT '' CHECK (char_length(description) <= 5000),
    status product_status NOT NULL DEFAULT 'draft',
    moderation_status moderation_status NOT NULL DEFAULT 'pending',
    moderation_note text NOT NULL DEFAULT '' CHECK (char_length(moderation_note) <= 500),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX products_store_idx ON products (store_id);
CREATE INDEX products_category_idx ON products (category_id);
CREATE INDEX products_search_idx ON products USING gin ((name || ' ' || description) gin_trgm_ops);

CREATE TABLE product_variants (
    id uuid PRIMARY KEY,
    product_id uuid NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku text NOT NULL UNIQUE CHECK (char_length(sku) BETWEEN 2 AND 64),
    name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
    attributes jsonb NOT NULL DEFAULT '{}'::jsonb,
    price_minor bigint NOT NULL CHECK (price_minor >= 0),
    currency char(3) NOT NULL DEFAULT 'IDR' CHECK (currency = 'IDR'),
    stock integer NOT NULL DEFAULT 0 CHECK (stock >= 0),
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX product_variants_product_idx ON product_variants (product_id);

CREATE TABLE product_images (
    id uuid PRIMARY KEY,
    product_id uuid NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url text NOT NULL CHECK (url ~ '^https?://' OR url ~ '^/[^/]'),
    alt_text text NOT NULL DEFAULT '' CHECK (char_length(alt_text) <= 160),
    position integer NOT NULL CHECK (position >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT product_images_position_unique UNIQUE (product_id, position)
);

CREATE TABLE inventory_ledger (
    id uuid PRIMARY KEY,
    variant_id uuid NOT NULL REFERENCES product_variants(id) ON DELETE RESTRICT,
    actor_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    delta integer NOT NULL CHECK (delta <> 0),
    resulting_stock integer NOT NULL CHECK (resulting_stock >= 0),
    reason text NOT NULL CHECK (char_length(reason) BETWEEN 1 AND 200),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX inventory_ledger_variant_idx ON inventory_ledger (variant_id, created_at DESC);

CREATE TABLE audit_log (
    id uuid PRIMARY KEY,
    actor_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id uuid NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_log_resource_idx ON audit_log (resource_type, resource_id, created_at DESC);
CREATE INDEX audit_log_actor_idx ON audit_log (actor_id, created_at DESC);
