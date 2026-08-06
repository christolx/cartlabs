INSERT INTO platform_metadata (key, value, updated_at)
VALUES (
    'demo_seed',
    '{"version": 2, "accounts": ["buyer", "seller", "admin"]}'::jsonb,
    now()
)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = EXCLUDED.updated_at;

INSERT INTO users (id, email, password_hash, display_name, role, status)
VALUES
    ('01989f00-0000-7000-8000-000000000001', 'buyer@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Demo Buyer', 'buyer', 'active'),
    ('01989f00-0000-7000-8000-000000000002', 'seller@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Demo Seller', 'seller', 'active'),
    ('01989f00-0000-7000-8000-000000000003', 'admin@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Demo Admin', 'admin', 'active'),
    ('01989f00-0000-7000-8000-000000000004', 'merchant@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Seed Merchant', 'seller', 'active')
ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    display_name = EXCLUDED.display_name,
    role = EXCLUDED.role,
    status = EXCLUDED.status,
    updated_at = now();

INSERT INTO categories (id, name, slug)
VALUES
    ('01989f00-0000-7000-8000-000000000101', 'Apparel', 'apparel'),
    ('01989f00-0000-7000-8000-000000000102', 'Home & Living', 'home-living'),
    ('01989f00-0000-7000-8000-000000000103', 'Electronics', 'electronics')
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name;

INSERT INTO stores (id, seller_id, name, slug, description, status)
VALUES (
    '01989f00-0000-7000-8000-000000000201',
    '01989f00-0000-7000-8000-000000000004',
    'Nusantara Goods',
    'nusantara-goods',
    'Durable everyday goods from independent Indonesian makers.',
    'approved'
)
ON CONFLICT (seller_id) DO UPDATE SET
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    updated_at = now();

INSERT INTO products (id, store_id, category_id, name, slug, description, status, moderation_status)
VALUES (
    '01989f00-0000-7000-8000-000000000301',
    '01989f00-0000-7000-8000-000000000201',
    '01989f00-0000-7000-8000-000000000102',
    'Handwoven Market Basket',
    'handwoven-market-basket',
    'Structured natural-fiber basket woven for markets, picnics, and daily storage.',
    'published',
    'approved'
)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    moderation_status = EXCLUDED.moderation_status,
    updated_at = now();

INSERT INTO product_variants (id, product_id, sku, name, attributes, price_minor, currency, stock)
VALUES (
    '01989f00-0000-7000-8000-000000000401',
    '01989f00-0000-7000-8000-000000000301',
    'NUSA-BASKET-NAT',
    'Natural',
    '{"color": "Natural"}'::jsonb,
    249000,
    'IDR',
    18
)
ON CONFLICT (sku) DO UPDATE SET
    name = EXCLUDED.name,
    attributes = EXCLUDED.attributes,
    price_minor = EXCLUDED.price_minor,
    currency = EXCLUDED.currency,
    stock = EXCLUDED.stock,
    active = true,
    updated_at = now();

INSERT INTO product_images (id, product_id, url, alt_text, position)
VALUES (
    '01989f00-0000-7000-8000-000000000501',
    '01989f00-0000-7000-8000-000000000301',
    '/images/shared-product.webp',
    'Natural handwoven market basket',
    0
)
ON CONFLICT (product_id, position) DO UPDATE SET
    url = EXCLUDED.url,
    alt_text = EXCLUDED.alt_text;
