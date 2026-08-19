INSERT INTO platform_metadata (key, value, updated_at)
VALUES (
    'demo_seed',
    '{"version": 6, "accounts": ["buyer", "buyer2", "buyer3", "seller", "admin"], "catalogProducts": 40}'::jsonb,
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
    ('01989f00-0000-7000-8000-000000000004', 'merchant@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Seed Merchant', 'seller', 'active'),
    ('01989f00-0000-7000-8000-000000000005', 'buyer2@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Second Demo Buyer', 'buyer', 'active'),
    ('01989f00-0000-7000-8000-000000000006', 'merchant2@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Second Seed Merchant', 'seller', 'active'),
    ('01989f00-0000-7000-8000-000000000007', 'buyer3@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Synthetic Traffic Buyer', 'buyer', 'active'),
    ('01989f00-0000-7000-8000-000000000008', 'studio@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Studio Minimal Merchant', 'seller', 'active'),
    ('01989f00-0000-7000-8000-000000000009', 'tenfold@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Tenfold Merchant', 'seller', 'active'),
    ('01989f00-0000-7000-8000-000000000010', 'brew@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Brew Studio Merchant', 'seller', 'active'),
    ('01989f00-0000-7000-8000-000000000011', 'object@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Object Works Merchant', 'seller', 'active'),
    ('01989f00-0000-7000-8000-000000000012', 'carry@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Carry Collective Merchant', 'seller', 'active'),
    ('01989f00-0000-7000-8000-000000000013', 'signal@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Signal House Merchant', 'seller', 'active')
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
    ('01989f00-0000-7000-8000-000000000103', 'Electronics', 'electronics'),
    ('01989f00-0000-7000-8000-000000000104', 'Travel', 'travel')
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name;

INSERT INTO stores (id, seller_id, name, slug, description, status)
VALUES
    ('01989f00-0000-7000-8000-000000000201', '01989f00-0000-7000-8000-000000000004', 'Nusantara Goods', 'nusantara-goods', 'Durable everyday goods from independent Indonesian makers.', 'approved'),
    ('01989f00-0000-7000-8000-000000000202', '01989f00-0000-7000-8000-000000000006', 'Java Loom Studio', 'java-loom-studio', 'Small-run textiles and useful woven pieces.', 'approved'),
    ('01989f00-0000-7000-8000-000000000203', '01989f00-0000-7000-8000-000000000008', 'Studio Minimal', 'studio-minimal', 'Exact lighting, clocks, and compact furniture for focused rooms.', 'approved'),
    ('01989f00-0000-7000-8000-000000000204', '01989f00-0000-7000-8000-000000000009', 'Tenfold Supply', 'tenfold-supply', 'Travel equipment designed for daily movement and short trips.', 'approved'),
    ('01989f00-0000-7000-8000-000000000205', '01989f00-0000-7000-8000-000000000010', 'Brew Studio', 'brew-studio', 'Kitchen and table tools selected for repeatable daily use.', 'approved'),
    ('01989f00-0000-7000-8000-000000000206', '01989f00-0000-7000-8000-000000000011', 'Object Works', 'object-works', 'Quiet desk technology with durable materials and clear controls.', 'approved'),
    ('01989f00-0000-7000-8000-000000000207', '01989f00-0000-7000-8000-000000000012', 'Carry Collective', 'carry-collective', 'Practical apparel and carry pieces for work and weekends.', 'approved'),
    ('01989f00-0000-7000-8000-000000000208', '01989f00-0000-7000-8000-000000000013', 'Signal House', 'signal-house', 'Compact personal electronics with restrained industrial design.', 'approved')
ON CONFLICT (seller_id) DO UPDATE SET
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    updated_at = now();

INSERT INTO products (id, store_id, category_id, name, slug, description, status)
VALUES
    ('01989f00-0000-7000-8000-000000000301', '01989f00-0000-7000-8000-000000000201', '01989f00-0000-7000-8000-000000000102', 'Handwoven Market Basket', 'handwoven-market-basket', 'Structured natural-fiber basket woven for markets, picnics, and daily storage.', 'published'),
    ('01989f00-0000-7000-8000-000000000302', '01989f00-0000-7000-8000-000000000202', '01989f00-0000-7000-8000-000000000102', 'Indigo Utility Tray', 'indigo-utility-tray', 'Low graphite tray for entryways, desks, and shared tables.', 'published'),
    ('01989f00-0000-7000-8000-000000000303', '01989f00-0000-7000-8000-000000000201', '01989f00-0000-7000-8000-000000000104', 'Field Bottle', 'field-bottle', 'Brushed steel bottle with a secure carry loop for workdays and travel.', 'published'),
    ('01989f00-0000-7000-8000-000000000304', '01989f00-0000-7000-8000-000000000201', '01989f00-0000-7000-8000-000000000102', 'Enamel Stock Pot', 'enamel-stock-pot', 'Deep enamel stock pot with fitted lid and balanced side handles.', 'published'),
    ('01989f00-0000-7000-8000-000000000305', '01989f00-0000-7000-8000-000000000201', '01989f00-0000-7000-8000-000000000102', 'Manual Citrus Press', 'manual-citrus-press', 'Lever press built for clean, controlled citrus extraction.', 'published'),
    ('01989f00-0000-7000-8000-000000000306', '01989f00-0000-7000-8000-000000000201', '01989f00-0000-7000-8000-000000000102', 'Pepper Mill', 'pepper-mill', 'Graphite mill with an adjustable ceramic grinding mechanism.', 'published'),
    ('01989f00-0000-7000-8000-000000000307', '01989f00-0000-7000-8000-000000000202', '01989f00-0000-7000-8000-000000000102', 'Grid Throw', 'grid-throw', 'Soft navy throw woven with a restrained windowpane grid.', 'published'),
    ('01989f00-0000-7000-8000-000000000308', '01989f00-0000-7000-8000-000000000202', '01989f00-0000-7000-8000-000000000102', 'Linen Table Runner', 'linen-table-runner', 'Slate linen runner with a dense weave and clean finished edges.', 'published'),
    ('01989f00-0000-7000-8000-000000000309', '01989f00-0000-7000-8000-000000000202', '01989f00-0000-7000-8000-000000000101', 'Woven Scarf', 'woven-scarf', 'Cobalt and gray scarf finished with short woven fringe.', 'published'),
    ('01989f00-0000-7000-8000-000000000310', '01989f00-0000-7000-8000-000000000202', '01989f00-0000-7000-8000-000000000101', 'Work Apron', 'work-apron', 'Heavy slate canvas apron with reinforced pockets and adjustable straps.', 'published'),
    ('01989f00-0000-7000-8000-000000000311', '01989f00-0000-7000-8000-000000000203', '01989f00-0000-7000-8000-000000000103', 'Arc Task Lamp', 'arc-task-lamp', 'Articulated cobalt task lamp with focused downward light.', 'published'),
    ('01989f00-0000-7000-8000-000000000312', '01989f00-0000-7000-8000-000000000203', '01989f00-0000-7000-8000-000000000102', 'Desk Clock', 'desk-clock', 'Compact silver desk clock with a clear analog face.', 'published'),
    ('01989f00-0000-7000-8000-000000000313', '01989f00-0000-7000-8000-000000000203', '01989f00-0000-7000-8000-000000000102', 'Powder-Coated Folding Stool', 'powder-coated-folding-stool', 'Compact folding stool with a stable steel frame and graphite finish.', 'published'),
    ('01989f00-0000-7000-8000-000000000314', '01989f00-0000-7000-8000-000000000203', '01989f00-0000-7000-8000-000000000102', 'Arched Table Mirror', 'arched-table-mirror', 'Freestanding brushed-silver mirror with adjustable side pivots.', 'published'),
    ('01989f00-0000-7000-8000-000000000315', '01989f00-0000-7000-8000-000000000203', '01989f00-0000-7000-8000-000000000103', 'Compact Desk Fan', 'compact-desk-fan', 'Quiet circular desk fan with a weighted graphite base.', 'published'),
    ('01989f00-0000-7000-8000-000000000316', '01989f00-0000-7000-8000-000000000204', '01989f00-0000-7000-8000-000000000104', 'Roll-Top Backpack', 'roll-top-backpack', 'Structured navy backpack with a secure roll-top closure.', 'published'),
    ('01989f00-0000-7000-8000-000000000317', '01989f00-0000-7000-8000-000000000204', '01989f00-0000-7000-8000-000000000104', 'Compact Umbrella', 'compact-umbrella', 'Small charcoal umbrella with a strong folding frame and wrist loop.', 'published'),
    ('01989f00-0000-7000-8000-000000000318', '01989f00-0000-7000-8000-000000000204', '01989f00-0000-7000-8000-000000000104', 'Crossbody Pouch', 'crossbody-pouch', 'Graphite pouch with a simple adjustable shoulder strap.', 'published'),
    ('01989f00-0000-7000-8000-000000000319', '01989f00-0000-7000-8000-000000000204', '01989f00-0000-7000-8000-000000000104', 'Packing Cube Set', 'packing-cube-set', 'Three structured packing cubes sized for organized carry-on storage.', 'published'),
    ('01989f00-0000-7000-8000-000000000320', '01989f00-0000-7000-8000-000000000204', '01989f00-0000-7000-8000-000000000104', 'Weekender Bag', 'weekender-bag', 'Graphite canvas holdall with short handles and a removable shoulder strap.', 'published'),
    ('01989f00-0000-7000-8000-000000000321', '01989f00-0000-7000-8000-000000000205', '01989f00-0000-7000-8000-000000000102', 'Pour Over Set', 'pour-over-set', 'Ceramic dripper, glass carafe, and steel kettle for measured brewing.', 'published'),
    ('01989f00-0000-7000-8000-000000000322', '01989f00-0000-7000-8000-000000000205', '01989f00-0000-7000-8000-000000000102', 'Acacia Serving Board', 'acacia-serving-board', 'Solid acacia board with rounded corners and a compact hanging hole.', 'published'),
    ('01989f00-0000-7000-8000-000000000323', '01989f00-0000-7000-8000-000000000205', '01989f00-0000-7000-8000-000000000102', 'Stoneware Dinner Set', 'stoneware-dinner-set', 'Matte stoneware plate, bowl, and cup for everyday table settings.', 'published'),
    ('01989f00-0000-7000-8000-000000000324', '01989f00-0000-7000-8000-000000000205', '01989f00-0000-7000-8000-000000000104', 'Stainless Lunch Box', 'stainless-lunch-box', 'Brushed steel lunch box with a fitted lid and secure side clips.', 'published'),
    ('01989f00-0000-7000-8000-000000000325', '01989f00-0000-7000-8000-000000000205', '01989f00-0000-7000-8000-000000000102', 'Ceramic Oil Cruet', 'ceramic-oil-cruet', 'Tall off-white cruet with a controlled pour spout and cork stopper.', 'published'),
    ('01989f00-0000-7000-8000-000000000326', '01989f00-0000-7000-8000-000000000206', '01989f00-0000-7000-8000-000000000102', 'Modular Desk Organizer', 'modular-desk-organizer', 'Low aluminum organizer divided for small daily desk tools.', 'published'),
    ('01989f00-0000-7000-8000-000000000327', '01989f00-0000-7000-8000-000000000206', '01989f00-0000-7000-8000-000000000103', 'Wireless Mechanical Keyboard', 'wireless-mechanical-keyboard', 'Compact wireless keyboard with tactile switches and off-white keycaps.', 'published'),
    ('01989f00-0000-7000-8000-000000000328', '01989f00-0000-7000-8000-000000000206', '01989f00-0000-7000-8000-000000000103', 'Aluminum Power Bank', 'aluminum-power-bank', 'Slim brushed-aluminum battery pack for everyday device charging.', 'published'),
    ('01989f00-0000-7000-8000-000000000329', '01989f00-0000-7000-8000-000000000206', '01989f00-0000-7000-8000-000000000103', 'Wireless Charging Stand', 'wireless-charging-stand', 'Angled aluminum charging stand for a clear desk-side view.', 'published'),
    ('01989f00-0000-7000-8000-000000000330', '01989f00-0000-7000-8000-000000000206', '01989f00-0000-7000-8000-000000000103', 'E-Reader', 'e-reader', 'Thin navy reader with a matte display designed for long sessions.', 'published'),
    ('01989f00-0000-7000-8000-000000000331', '01989f00-0000-7000-8000-000000000207', '01989f00-0000-7000-8000-000000000104', 'Canvas Tote', 'canvas-tote', 'Structured off-white tote with reinforced navy handles.', 'published'),
    ('01989f00-0000-7000-8000-000000000332', '01989f00-0000-7000-8000-000000000207', '01989f00-0000-7000-8000-000000000101', 'Waxed Utility Jacket', 'waxed-utility-jacket', 'Olive waxed-cotton jacket with reinforced utility pockets.', 'published'),
    ('01989f00-0000-7000-8000-000000000333', '01989f00-0000-7000-8000-000000000207', '01989f00-0000-7000-8000-000000000101', 'Knit Beanie', 'knit-beanie', 'Dense charcoal rib-knit beanie with a folded cuff.', 'published'),
    ('01989f00-0000-7000-8000-000000000334', '01989f00-0000-7000-8000-000000000207', '01989f00-0000-7000-8000-000000000101', 'Low-Top Canvas Sneakers', 'low-top-canvas-sneakers', 'Off-white canvas sneakers with restrained navy trim.', 'published'),
    ('01989f00-0000-7000-8000-000000000335', '01989f00-0000-7000-8000-000000000207', '01989f00-0000-7000-8000-000000000104', 'Leather Card Wallet', 'leather-card-wallet', 'Slim dark-brown leather wallet with reinforced edge stitching.', 'published'),
    ('01989f00-0000-7000-8000-000000000336', '01989f00-0000-7000-8000-000000000208', '01989f00-0000-7000-8000-000000000102', 'Ceramic Planter', 'ceramic-planter', 'Cobalt ceramic planter with a fitted drainage saucer.', 'published'),
    ('01989f00-0000-7000-8000-000000000337', '01989f00-0000-7000-8000-000000000208', '01989f00-0000-7000-8000-000000000103', 'Portable Radio', 'portable-radio', 'Compact navy analog radio with tactile tuning controls.', 'published'),
    ('01989f00-0000-7000-8000-000000000338', '01989f00-0000-7000-8000-000000000208', '01989f00-0000-7000-8000-000000000103', 'Over-Ear Headphones', 'over-ear-headphones', 'Graphite wireless headphones with padded over-ear cushions.', 'published'),
    ('01989f00-0000-7000-8000-000000000339', '01989f00-0000-7000-8000-000000000208', '01989f00-0000-7000-8000-000000000103', 'Portable Speaker', 'portable-speaker', 'Compact cobalt speaker wrapped in a durable woven grille.', 'published'),
    ('01989f00-0000-7000-8000-000000000340', '01989f00-0000-7000-8000-000000000208', '01989f00-0000-7000-8000-000000000103', 'Compact Digital Camera', 'compact-digital-camera', 'Brushed-silver camera with a retractable optical lens.', 'published')
ON CONFLICT (slug) DO UPDATE SET
    store_id = EXCLUDED.store_id,
    category_id = EXCLUDED.category_id,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    updated_at = now();

INSERT INTO product_variants (id, product_id, sku, name, attributes, price_minor, currency, stock)
VALUES
    ('01989f00-0000-7000-8000-000000000401', '01989f00-0000-7000-8000-000000000301', 'NUSA-BASKET-NAT', 'Natural', '{"color": "Natural"}'::jsonb, 249000, 'IDR', 18),
    ('01989f00-0000-7000-8000-000000000402', '01989f00-0000-7000-8000-000000000302', 'JLS-TRAY-INDIGO', 'Indigo', '{"color": "Graphite"}'::jsonb, 159000, 'IDR', 12),
    ('01989f00-0000-7000-8000-000000000403', '01989f00-0000-7000-8000-000000000303', 'NUSA-BOTTLE-750', '750 ml', '{"capacity": "750 ml", "color": "Steel"}'::jsonb, 329000, 'IDR', 24),
    ('01989f00-0000-7000-8000-000000000404', '01989f00-0000-7000-8000-000000000304', 'NUSA-POT-5L', '5 liter', '{"capacity": "5 liter", "color": "Navy"}'::jsonb, 899000, 'IDR', 8),
    ('01989f00-0000-7000-8000-000000000405', '01989f00-0000-7000-8000-000000000305', 'NUSA-PRESS-STEEL', 'Brushed steel', '{"material": "Stainless steel"}'::jsonb, 575000, 'IDR', 7),
    ('01989f00-0000-7000-8000-000000000406', '01989f00-0000-7000-8000-000000000306', 'NUSA-MILL-GRAPHITE', 'Graphite', '{"color": "Graphite"}'::jsonb, 289000, 'IDR', 0),
    ('01989f00-0000-7000-8000-000000000407', '01989f00-0000-7000-8000-000000000307', 'JLS-THROW-NAVY', 'Navy grid', '{"color": "Navy", "size": "130 x 180 cm"}'::jsonb, 649000, 'IDR', 11),
    ('01989f00-0000-7000-8000-000000000408', '01989f00-0000-7000-8000-000000000308', 'JLS-RUNNER-SLATE', 'Slate', '{"color": "Slate", "size": "40 x 180 cm"}'::jsonb, 279000, 'IDR', 16),
    ('01989f00-0000-7000-8000-000000000409', '01989f00-0000-7000-8000-000000000309', 'JLS-SCARF-COBALT', 'Cobalt gray', '{"color": "Cobalt gray"}'::jsonb, 349000, 'IDR', 6),
    ('01989f00-0000-7000-8000-000000000410', '01989f00-0000-7000-8000-000000000310', 'JLS-APRON-SLATE', 'Standard', '{"color": "Slate"}'::jsonb, 399000, 'IDR', 4),
    ('01989f00-0000-7000-8000-000000000411', '01989f00-0000-7000-8000-000000000311', 'SM-LAMP-COBALT', 'Cobalt', '{"color": "Cobalt"}'::jsonb, 1250000, 'IDR', 9),
    ('01989f00-0000-7000-8000-000000000412', '01989f00-0000-7000-8000-000000000312', 'SM-CLOCK-SILVER', 'Silver', '{"color": "Silver"}'::jsonb, 550000, 'IDR', 0),
    ('01989f00-0000-7000-8000-000000000413', '01989f00-0000-7000-8000-000000000313', 'SM-STOOL-GRAPHITE', 'Graphite', '{"color": "Graphite"}'::jsonb, 475000, 'IDR', 5),
    ('01989f00-0000-7000-8000-000000000414', '01989f00-0000-7000-8000-000000000314', 'SM-MIRROR-SILVER', 'Silver', '{"color": "Silver"}'::jsonb, 625000, 'IDR', 7),
    ('01989f00-0000-7000-8000-000000000415', '01989f00-0000-7000-8000-000000000315', 'SM-FAN-GRAPHITE', 'Graphite', '{"color": "Graphite"}'::jsonb, 399000, 'IDR', 15),
    ('01989f00-0000-7000-8000-000000000416', '01989f00-0000-7000-8000-000000000316', 'TEN-BACKPACK-NAVY', 'Navy', '{"color": "Navy", "capacity": "22 liter"}'::jsonb, 1099000, 'IDR', 10),
    ('01989f00-0000-7000-8000-000000000417', '01989f00-0000-7000-8000-000000000317', 'TEN-UMBRELLA-CHARCOAL', 'Charcoal', '{"color": "Charcoal"}'::jsonb, 299000, 'IDR', 22),
    ('01989f00-0000-7000-8000-000000000418', '01989f00-0000-7000-8000-000000000318', 'TEN-POUCH-GRAPHITE', 'Graphite', '{"color": "Graphite"}'::jsonb, 425000, 'IDR', 14),
    ('01989f00-0000-7000-8000-000000000419', '01989f00-0000-7000-8000-000000000319', 'TEN-CUBES-NAVY', 'Three piece', '{"color": "Navy", "pieces": "3"}'::jsonb, 499000, 'IDR', 0),
    ('01989f00-0000-7000-8000-000000000420', '01989f00-0000-7000-8000-000000000320', 'TEN-WEEKENDER-GRAPHITE', 'Graphite', '{"color": "Graphite", "capacity": "35 liter"}'::jsonb, 1399000, 'IDR', 5),
    ('01989f00-0000-7000-8000-000000000421', '01989f00-0000-7000-8000-000000000321', 'BREW-POUR-SET', 'Three piece', '{"pieces": "3"}'::jsonb, 780000, 'IDR', 13),
    ('01989f00-0000-7000-8000-000000000422', '01989f00-0000-7000-8000-000000000322', 'BREW-BOARD-ACACIA', 'Acacia', '{"material": "Acacia"}'::jsonb, 379000, 'IDR', 8),
    ('01989f00-0000-7000-8000-000000000423', '01989f00-0000-7000-8000-000000000323', 'BREW-DINNER-3PC', 'Three piece', '{"color": "Stone", "pieces": "3"}'::jsonb, 899000, 'IDR', 6),
    ('01989f00-0000-7000-8000-000000000424', '01989f00-0000-7000-8000-000000000324', 'BREW-LUNCH-STEEL', 'Standard', '{"material": "Stainless steel"}'::jsonb, 349000, 'IDR', 19),
    ('01989f00-0000-7000-8000-000000000425', '01989f00-0000-7000-8000-000000000325', 'BREW-CRUET-OFFWHITE', 'Off-white', '{"color": "Off-white", "capacity": "500 ml"}'::jsonb, 319000, 'IDR', 2),
    ('01989f00-0000-7000-8000-000000000426', '01989f00-0000-7000-8000-000000000326', 'OBJ-ORGANIZER-GRAPHITE', 'Graphite', '{"color": "Graphite"}'::jsonb, 420000, 'IDR', 17),
    ('01989f00-0000-7000-8000-000000000427', '01989f00-0000-7000-8000-000000000327', 'OBJ-KEYBOARD-OFFWHITE', 'Off-white', '{"color": "Off-white", "layout": "75 percent"}'::jsonb, 1299000, 'IDR', 7),
    ('01989f00-0000-7000-8000-000000000428', '01989f00-0000-7000-8000-000000000328', 'OBJ-POWERBANK-SILVER', 'Silver', '{"color": "Silver", "capacity": "10000 mAh"}'::jsonb, 499000, 'IDR', 21),
    ('01989f00-0000-7000-8000-000000000429', '01989f00-0000-7000-8000-000000000329', 'OBJ-CHARGER-SILVER', 'Silver', '{"color": "Silver"}'::jsonb, 549000, 'IDR', 12),
    ('01989f00-0000-7000-8000-000000000430', '01989f00-0000-7000-8000-000000000330', 'OBJ-EREADER-NAVY', 'Navy', '{"color": "Navy", "storage": "32 GB"}'::jsonb, 1899000, 'IDR', 3),
    ('01989f00-0000-7000-8000-000000000431', '01989f00-0000-7000-8000-000000000331', 'CARRY-TOTE-OFFWHITE', 'Off-white navy', '{"color": "Off-white navy"}'::jsonb, 680000, 'IDR', 16),
    ('01989f00-0000-7000-8000-000000000432', '01989f00-0000-7000-8000-000000000332', 'CARRY-JACKET-OLIVE-M', 'Medium', '{"color": "Olive", "size": "M"}'::jsonb, 1599000, 'IDR', 4),
    ('01989f00-0000-7000-8000-000000000433', '01989f00-0000-7000-8000-000000000333', 'CARRY-BEANIE-CHARCOAL', 'One size', '{"color": "Charcoal", "size": "One size"}'::jsonb, 299000, 'IDR', 20),
    ('01989f00-0000-7000-8000-000000000434', '01989f00-0000-7000-8000-000000000334', 'CARRY-SNEAKER-41', 'Size 41', '{"color": "Off-white navy", "size": "41"}'::jsonb, 899000, 'IDR', 9),
    ('01989f00-0000-7000-8000-000000000435', '01989f00-0000-7000-8000-000000000335', 'CARRY-WALLET-BROWN', 'Dark brown', '{"color": "Dark brown"}'::jsonb, 389000, 'IDR', 18),
    ('01989f00-0000-7000-8000-000000000436', '01989f00-0000-7000-8000-000000000336', 'SIG-PLANTER-COBALT', 'Cobalt', '{"color": "Cobalt", "diameter": "24 cm"}'::jsonb, 425000, 'IDR', 10),
    ('01989f00-0000-7000-8000-000000000437', '01989f00-0000-7000-8000-000000000337', 'SIG-RADIO-NAVY', 'Navy', '{"color": "Navy"}'::jsonb, 799000, 'IDR', 6),
    ('01989f00-0000-7000-8000-000000000438', '01989f00-0000-7000-8000-000000000338', 'SIG-HEADPHONES-GRAPHITE', 'Graphite', '{"color": "Graphite"}'::jsonb, 1499000, 'IDR', 8),
    ('01989f00-0000-7000-8000-000000000439', '01989f00-0000-7000-8000-000000000339', 'SIG-SPEAKER-COBALT', 'Cobalt', '{"color": "Cobalt"}'::jsonb, 899000, 'IDR', 0),
    ('01989f00-0000-7000-8000-000000000440', '01989f00-0000-7000-8000-000000000340', 'SIG-CAMERA-SILVER', 'Silver', '{"color": "Silver"}'::jsonb, 2499000, 'IDR', 5)
ON CONFLICT (sku) DO UPDATE SET
    product_id = EXCLUDED.product_id,
    name = EXCLUDED.name,
    attributes = EXCLUDED.attributes,
    price_minor = EXCLUDED.price_minor,
    currency = EXCLUDED.currency,
    stock = EXCLUDED.stock,
    active = true,
    updated_at = now();

INSERT INTO product_images (id, product_id, url, alt_text, position)
VALUES
    ('01989f00-0000-7000-8000-000000000501', '01989f00-0000-7000-8000-000000000301', '/images/catalog-v2/loom-carry-basket.webp', 'Natural handwoven market basket on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000502', '01989f00-0000-7000-8000-000000000302', '/images/catalog-v2/studio-tray.webp', 'Low graphite utility tray on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000503', '01989f00-0000-7000-8000-000000000303', '/images/catalog-v2/field-bottle.webp', 'Brushed steel field bottle on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000504', '01989f00-0000-7000-8000-000000000304', '/images/catalog-v2/enamel-stock-pot.webp', 'Navy enamel stock pot on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000505', '01989f00-0000-7000-8000-000000000305', '/images/catalog-v2/manual-citrus-press.webp', 'Brushed metal manual citrus press on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000506', '01989f00-0000-7000-8000-000000000306', '/images/catalog-v2/pepper-mill.webp', 'Graphite pepper mill on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000507', '01989f00-0000-7000-8000-000000000307', '/images/catalog-v2/grid-throw.webp', 'Folded navy grid throw on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000508', '01989f00-0000-7000-8000-000000000308', '/images/catalog-v2/linen-table-runner.webp', 'Folded slate linen table runner on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000509', '01989f00-0000-7000-8000-000000000309', '/images/catalog-v2/woven-scarf.webp', 'Folded cobalt and gray woven scarf on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000510', '01989f00-0000-7000-8000-000000000310', '/images/catalog-v2/work-apron.webp', 'Slate canvas work apron laid flat on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000511', '01989f00-0000-7000-8000-000000000311', '/images/catalog-v2/arc-task-lamp.webp', 'Cobalt articulated task lamp on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000512', '01989f00-0000-7000-8000-000000000312', '/images/catalog-v2/desk-clock.webp', 'Round silver desk clock on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000513', '01989f00-0000-7000-8000-000000000313', '/images/catalog-v2/powder-coated-folding-stool.webp', 'Graphite folding stool on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000514', '01989f00-0000-7000-8000-000000000314', '/images/catalog-v2/arched-table-mirror.webp', 'Brushed-silver arched table mirror on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000515', '01989f00-0000-7000-8000-000000000315', '/images/catalog-v2/compact-desk-fan.webp', 'Compact graphite desk fan on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000516', '01989f00-0000-7000-8000-000000000316', '/images/catalog-v2/roll-top-backpack.webp', 'Navy roll-top backpack on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000517', '01989f00-0000-7000-8000-000000000317', '/images/catalog-v2/compact-umbrella.webp', 'Folded charcoal umbrella on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000518', '01989f00-0000-7000-8000-000000000318', '/images/catalog-v2/crossbody-pouch.webp', 'Graphite crossbody pouch on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000519', '01989f00-0000-7000-8000-000000000319', '/images/catalog-v2/packing-cube-set.webp', 'Three navy packing cubes on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000520', '01989f00-0000-7000-8000-000000000320', '/images/catalog-v2/weekender-bag.webp', 'Graphite canvas weekender bag on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000521', '01989f00-0000-7000-8000-000000000321', '/images/catalog-v2/pour-over-set.webp', 'Ceramic and steel pour-over coffee set on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000522', '01989f00-0000-7000-8000-000000000322', '/images/catalog-v2/acacia-serving-board.webp', 'Acacia serving board on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000523', '01989f00-0000-7000-8000-000000000323', '/images/catalog-v2/stoneware-dinner-set.webp', 'Matte stoneware dinner set on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000524', '01989f00-0000-7000-8000-000000000324', '/images/catalog-v2/stainless-lunch-box.webp', 'Stainless lunch box on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000525', '01989f00-0000-7000-8000-000000000325', '/images/catalog-v2/ceramic-oil-cruet.webp', 'Off-white ceramic oil cruet on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000526', '01989f00-0000-7000-8000-000000000326', '/images/catalog-v2/modular-desk-organizer.webp', 'Graphite modular desk organizer on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000527', '01989f00-0000-7000-8000-000000000327', '/images/catalog-v2/wireless-mechanical-keyboard.webp', 'Off-white wireless mechanical keyboard on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000528', '01989f00-0000-7000-8000-000000000328', '/images/catalog-v2/aluminum-power-bank.webp', 'Slim aluminum power bank on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000529', '01989f00-0000-7000-8000-000000000329', '/images/catalog-v2/wireless-charging-stand.webp', 'Aluminum wireless charging stand on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000530', '01989f00-0000-7000-8000-000000000330', '/images/catalog-v2/e-reader.webp', 'Navy e-reader with blank screen on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000531', '01989f00-0000-7000-8000-000000000331', '/images/catalog-v2/canvas-tote.webp', 'Off-white canvas tote with navy handles on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000532', '01989f00-0000-7000-8000-000000000332', '/images/catalog-v2/waxed-utility-jacket.webp', 'Olive waxed utility jacket laid flat on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000533', '01989f00-0000-7000-8000-000000000333', '/images/catalog-v2/knit-beanie.webp', 'Charcoal knit beanie on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000534', '01989f00-0000-7000-8000-000000000334', '/images/catalog-v2/low-top-canvas-sneakers.webp', 'Off-white low-top canvas sneakers on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000535', '01989f00-0000-7000-8000-000000000335', '/images/catalog-v2/leather-card-wallet.webp', 'Dark-brown leather card wallet on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000536', '01989f00-0000-7000-8000-000000000336', '/images/catalog-v2/ceramic-planter.webp', 'Cobalt ceramic planter on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000537', '01989f00-0000-7000-8000-000000000337', '/images/catalog-v2/portable-radio.webp', 'Navy portable radio on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000538', '01989f00-0000-7000-8000-000000000338', '/images/catalog-v2/over-ear-headphones.webp', 'Graphite over-ear headphones on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000539', '01989f00-0000-7000-8000-000000000339', '/images/catalog-v2/portable-speaker.webp', 'Cobalt portable speaker on a cool gray studio surface', 0),
    ('01989f00-0000-7000-8000-000000000540', '01989f00-0000-7000-8000-000000000340', '/images/catalog-v2/compact-digital-camera.webp', 'Brushed-silver compact digital camera on a cool gray studio surface', 0)
ON CONFLICT (product_id, position) DO UPDATE SET
    url = EXCLUDED.url,
    alt_text = EXCLUDED.alt_text;
