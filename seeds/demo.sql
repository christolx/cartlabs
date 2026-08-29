INSERT INTO platform_metadata (key, value, updated_at)
VALUES (
    'demo_seed',
    '{"version": 14, "accounts": ["buyer", "buyer2", "buyer3", "seller", "admin"], "catalogProducts": 44, "story": "marketplace-natural-reviewed-catalog"}'::jsonb,
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
    ,('01989f00-0000-7000-8000-000000000014', 'pending@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Pending Merchant', 'seller', 'active')
    ,('01989f00-0000-7000-8000-000000000015', 'rejected@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Rejected Merchant', 'seller', 'active')
    ,('01989f00-0000-7000-8000-000000000016', 'suspended@demo.cartlabs.local', '$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g', 'Suspended Example', 'buyer', 'suspended')
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
    ('01989f00-0000-7000-8000-000000000200', '01989f00-0000-7000-8000-000000000002', 'Demo Seller Store', 'demo-seller-store', 'Official demo seller store for testing products and orders.', 'approved'),
    ('01989f00-0000-7000-8000-000000000201', '01989f00-0000-7000-8000-000000000004', 'Nusantara Goods', 'nusantara-goods', 'Durable everyday goods from independent Indonesian makers.', 'approved'),
    ('01989f00-0000-7000-8000-000000000202', '01989f00-0000-7000-8000-000000000006', 'Java Loom Studio', 'java-loom-studio', 'Small-run textiles and useful woven pieces.', 'approved'),
    ('01989f00-0000-7000-8000-000000000203', '01989f00-0000-7000-8000-000000000008', 'Studio Minimal', 'studio-minimal', 'Exact lighting, clocks, and compact furniture for focused rooms.', 'approved'),
    ('01989f00-0000-7000-8000-000000000204', '01989f00-0000-7000-8000-000000000009', 'Tenfold Supply', 'tenfold-supply', 'Travel equipment designed for daily movement and short trips.', 'approved'),
    ('01989f00-0000-7000-8000-000000000205', '01989f00-0000-7000-8000-000000000010', 'Brew Studio', 'brew-studio', 'Kitchen and table tools selected for repeatable daily use.', 'approved'),
    ('01989f00-0000-7000-8000-000000000206', '01989f00-0000-7000-8000-000000000011', 'Object Works', 'object-works', 'Quiet desk technology with durable materials and clear controls.', 'approved'),
    ('01989f00-0000-7000-8000-000000000207', '01989f00-0000-7000-8000-000000000012', 'Carry Collective', 'carry-collective', 'Practical apparel and carry pieces for work and weekends.', 'approved'),
    ('01989f00-0000-7000-8000-000000000208', '01989f00-0000-7000-8000-000000000013', 'Signal House', 'signal-house', 'Compact personal electronics with restrained industrial design.', 'approved')
    ,('01989f00-0000-7000-8000-000000000209', '01989f00-0000-7000-8000-000000000014', 'Awaiting Workshop', 'awaiting-workshop', 'Application awaiting marketplace verification.', 'pending')
    ,('01989f00-0000-7000-8000-000000000210', '01989f00-0000-7000-8000-000000000015', 'Incomplete Traders', 'incomplete-traders', 'Example rejected store application.', 'rejected')
ON CONFLICT (seller_id) DO UPDATE SET
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    moderation_note = CASE WHEN EXCLUDED.status = 'rejected' THEN 'Business registration could not be verified.' ELSE '' END,
    updated_at = now();

UPDATE stores
SET moderation_note = 'Business registration could not be verified.'
WHERE id = '01989f00-0000-7000-8000-000000000210';

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
    ,('01989f00-0000-7000-8000-000000000341', '01989f00-0000-7000-8000-000000000200', '01989f00-0000-7000-8000-000000000102', 'Demo Desk Caddy', 'demo-desk-caddy', 'Modular desk caddy used throughout the demo order lifecycle.', 'published')
    ,('01989f00-0000-7000-8000-000000000342', '01989f00-0000-7000-8000-000000000200', '01989f00-0000-7000-8000-000000000104', 'Demo Travel Pouch', 'demo-travel-pouch', 'Compact pouch prepared as a draft listing.', 'draft')
    ,('01989f00-0000-7000-8000-000000000343', '01989f00-0000-7000-8000-000000000200', '01989f00-0000-7000-8000-000000000102', 'Demo Archive Tray', 'demo-archive-tray', 'Retired tray retained for seller history.', 'archived')
    ,('01989f00-0000-7000-8000-000000000344', '01989f00-0000-7000-8000-000000000200', '01989f00-0000-7000-8000-000000000103', 'Demo Safety Charger', 'demo-safety-charger', 'Charging accessory suspended by marketplace moderation.', 'suspended')
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
    ,('01989f00-0000-7000-8000-000000000441', '01989f00-0000-7000-8000-000000000341', 'DEMO-CADDY-NAVY', 'Navy', '{"color": "Navy", "modules": "3"}'::jsonb, 450000, 'IDR', 30)
    ,('01989f00-0000-7000-8000-000000000442', '01989f00-0000-7000-8000-000000000341', 'DEMO-CADDY-STONE', 'Stone', '{"color": "Stone", "modules": "3"}'::jsonb, 450000, 'IDR', 2)
    ,('01989f00-0000-7000-8000-000000000443', '01989f00-0000-7000-8000-000000000342', 'DEMO-POUCH-OLIVE', 'Olive', '{"color": "Olive"}'::jsonb, 275000, 'IDR', 0)
    ,('01989f00-0000-7000-8000-000000000444', '01989f00-0000-7000-8000-000000000344', 'DEMO-CHARGER-BLACK', 'Black', '{"color": "Black", "power": "30 W"}'::jsonb, 525000, 'IDR', 7)
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
    ,('01989f00-0000-7000-8000-000000000541', '01989f00-0000-7000-8000-000000000341', '/images/catalog-v2/modular-desk-organizer.webp', 'Navy modular demo desk caddy', 0)
    ,('01989f00-0000-7000-8000-000000000542', '01989f00-0000-7000-8000-000000000342', '/images/catalog-v2/crossbody-pouch.webp', 'Olive demo travel pouch', 0)
    ,('01989f00-0000-7000-8000-000000000543', '01989f00-0000-7000-8000-000000000343', '/images/catalog-v2/studio-tray.webp', 'Archived demo tray', 0)
    ,('01989f00-0000-7000-8000-000000000544', '01989f00-0000-7000-8000-000000000344', '/images/catalog-v2/wireless-charging-stand.webp', 'Suspended black demo charger', 0)
ON CONFLICT (product_id, position) DO UPDATE SET
    url = EXCLUDED.url,
    alt_text = EXCLUDED.alt_text;

-- Buyer cart: two available products from two approved stores.
INSERT INTO carts (id, buyer_id, currency, created_at, updated_at)
VALUES ('01989f00-0000-7000-8000-000000000600', '01989f00-0000-7000-8000-000000000001', 'IDR', now() - interval '2 days', now() - interval '20 minutes')
ON CONFLICT (buyer_id) DO UPDATE SET currency = EXCLUDED.currency, updated_at = EXCLUDED.updated_at;

INSERT INTO cart_items (cart_id, variant_id, quantity, created_at, updated_at)
VALUES
    ('01989f00-0000-7000-8000-000000000600', '01989f00-0000-7000-8000-000000000441', 1, now() - interval '30 minutes', now() - interval '20 minutes'),
    ('01989f00-0000-7000-8000-000000000600', '01989f00-0000-7000-8000-000000000403', 2, now() - interval '25 minutes', now() - interval '20 minutes')
ON CONFLICT (cart_id, variant_id) DO UPDATE SET quantity = EXCLUDED.quantity, updated_at = EXCLUDED.updated_at;

-- Reset only owned lifecycle fixtures so revisions remain deterministic on existing demo databases.
DELETE FROM reviews
WHERE purchase_item_id IN (
    SELECT pi.id
    FROM purchase_items pi
    JOIN seller_orders so ON so.id = pi.seller_order_id
    WHERE so.purchase_id BETWEEN '01989f00-0000-7000-8000-000000000601' AND '01989f00-0000-7000-8000-000000000610'
);
DELETE FROM inventory_reservations
WHERE purchase_id BETWEEN '01989f00-0000-7000-8000-000000000601' AND '01989f00-0000-7000-8000-000000000610';
DELETE FROM purchase_items
WHERE seller_order_id IN (
    SELECT id FROM seller_orders
    WHERE purchase_id BETWEEN '01989f00-0000-7000-8000-000000000601' AND '01989f00-0000-7000-8000-000000000610'
);
DELETE FROM seller_orders
WHERE purchase_id BETWEEN '01989f00-0000-7000-8000-000000000601' AND '01989f00-0000-7000-8000-000000000610';
DELETE FROM purchases
WHERE id BETWEEN '01989f00-0000-7000-8000-000000000601' AND '01989f00-0000-7000-8000-000000000610';

-- Buyer history: completed orders only. Fresh checkout states come from interactive use,
-- while deterministic seed orders remain coherent regardless of seed age.
INSERT INTO purchases (id, reference, buyer_id, idempotency_key, status, payment_status, payment_intent_id, currency, subtotal_minor, total_minor, reservation_expires_at, created_at, updated_at)
VALUES
    ('01989f00-0000-7000-8000-000000000601', 'CL-DEMO00000001', '01989f00-0000-7000-8000-000000000001', 'demo-completed-order-one', 'paid', 'succeeded', 'pi_demo_completed_one', 'IDR', 450000, 450000, now() - interval '9 days', now() - interval '10 days', now() - interval '6 days'),
    ('01989f00-0000-7000-8000-000000000602', 'CL-DEMO00000002', '01989f00-0000-7000-8000-000000000001', 'demo-completed-order-two', 'paid', 'succeeded', 'pi_demo_completed_two', 'IDR', 249000, 249000, now() - interval '11 days', now() - interval '12 days', now() - interval '8 days'),
    ('01989f00-0000-7000-8000-000000000603', 'CL-DEMO00000003', '01989f00-0000-7000-8000-000000000001', 'demo-completed-order-three', 'paid', 'succeeded', 'pi_demo_completed_three', 'IDR', 318000, 318000, now() - interval '13 days', now() - interval '14 days', now() - interval '10 days'),
    ('01989f00-0000-7000-8000-000000000604', 'CL-DEMO00000004', '01989f00-0000-7000-8000-000000000001', 'demo-completed-order-four', 'paid', 'succeeded', 'pi_demo_completed_four', 'IDR', 1099000, 1099000, now() - interval '3 days', now() - interval '4 days', now() - interval '1 day'),
    ('01989f00-0000-7000-8000-000000000605', 'CL-DEMO00000005', '01989f00-0000-7000-8000-000000000001', 'demo-delivered-order', 'paid', 'succeeded', 'pi_demo_delivered', 'IDR', 1230000, 1230000, now() - interval '17 days', now() - interval '18 days', now() - interval '12 days'),
    ('01989f00-0000-7000-8000-000000000606', 'CL-DEMO00000006', '01989f00-0000-7000-8000-000000000001', 'demo-completed-order-six', 'paid', 'succeeded', 'pi_demo_completed_six', 'IDR', 450000, 450000, now() - interval '19 days', now() - interval '20 days', now() - interval '16 days'),
    ('01989f00-0000-7000-8000-000000000607', 'CL-DEMO00000007', '01989f00-0000-7000-8000-000000000001', 'demo-completed-order-seven', 'paid', 'succeeded', 'pi_demo_completed_seven', 'IDR', 450000, 450000, now() - interval '21 days', now() - interval '22 days', now() - interval '18 days'),
    ('01989f00-0000-7000-8000-000000000608', 'CL-DEMO00000008', '01989f00-0000-7000-8000-000000000005', 'demo-review-four', 'paid', 'succeeded', 'pi_demo_review_four', 'IDR', 249000, 249000, now() - interval '25 days', now() - interval '26 days', now() - interval '23 days'),
    ('01989f00-0000-7000-8000-000000000609', 'CL-DEMO00000009', '01989f00-0000-7000-8000-000000000007', 'demo-review-three', 'paid', 'succeeded', 'pi_demo_review_three', 'IDR', 159000, 159000, now() - interval '29 days', now() - interval '30 days', now() - interval '27 days'),
    ('01989f00-0000-7000-8000-000000000610', 'CL-DEMO00000010', '01989f00-0000-7000-8000-000000000001', 'demo-delivered-carry', 'paid', 'succeeded', 'pi_demo_delivered_carry', 'IDR', 979000, 979000, now() - interval '32 days', now() - interval '33 days', now() - interval '30 days');

INSERT INTO seller_orders (id, purchase_id, store_id, status, currency, subtotal_minor, processing_at, shipped_at, delivered_at, cancelled_at, cancellation_reason, created_at, updated_at)
VALUES
    ('01989f00-0000-7000-8000-000000000611', '01989f00-0000-7000-8000-000000000601', '01989f00-0000-7000-8000-000000000200', 'delivered', 'IDR', 450000, now() - interval '9 days', now() - interval '8 days', now() - interval '6 days', NULL, '', now() - interval '10 days', now() - interval '6 days'),
    ('01989f00-0000-7000-8000-000000000612', '01989f00-0000-7000-8000-000000000602', '01989f00-0000-7000-8000-000000000201', 'delivered', 'IDR', 249000, now() - interval '11 days', now() - interval '10 days', now() - interval '8 days', NULL, '', now() - interval '12 days', now() - interval '8 days'),
    ('01989f00-0000-7000-8000-000000000613', '01989f00-0000-7000-8000-000000000603', '01989f00-0000-7000-8000-000000000202', 'delivered', 'IDR', 318000, now() - interval '13 days', now() - interval '12 days', now() - interval '10 days', NULL, '', now() - interval '14 days', now() - interval '10 days'),
    ('01989f00-0000-7000-8000-000000000614', '01989f00-0000-7000-8000-000000000604', '01989f00-0000-7000-8000-000000000204', 'delivered', 'IDR', 1099000, now() - interval '3 days', now() - interval '2 days', now() - interval '1 day', NULL, '', now() - interval '4 days', now() - interval '1 day'),
    ('01989f00-0000-7000-8000-000000000615', '01989f00-0000-7000-8000-000000000605', '01989f00-0000-7000-8000-000000000200', 'delivered', 'IDR', 450000, now() - interval '16 days', now() - interval '14 days', now() - interval '12 days', NULL, '', now() - interval '18 days', now() - interval '12 days'),
    ('01989f00-0000-7000-8000-000000000616', '01989f00-0000-7000-8000-000000000606', '01989f00-0000-7000-8000-000000000200', 'delivered', 'IDR', 450000, now() - interval '19 days', now() - interval '18 days', now() - interval '16 days', NULL, '', now() - interval '20 days', now() - interval '16 days'),
    ('01989f00-0000-7000-8000-000000000617', '01989f00-0000-7000-8000-000000000607', '01989f00-0000-7000-8000-000000000200', 'delivered', 'IDR', 450000, now() - interval '21 days', now() - interval '20 days', now() - interval '18 days', NULL, '', now() - interval '22 days', now() - interval '18 days'),
    ('01989f00-0000-7000-8000-000000000618', '01989f00-0000-7000-8000-000000000608', '01989f00-0000-7000-8000-000000000201', 'delivered', 'IDR', 249000, now() - interval '25 days', now() - interval '24 days', now() - interval '23 days', NULL, '', now() - interval '26 days', now() - interval '23 days'),
    ('01989f00-0000-7000-8000-000000000619', '01989f00-0000-7000-8000-000000000609', '01989f00-0000-7000-8000-000000000202', 'delivered', 'IDR', 159000, now() - interval '29 days', now() - interval '28 days', now() - interval '27 days', NULL, '', now() - interval '30 days', now() - interval '27 days'),
    ('01989f00-0000-7000-8000-000000000620', '01989f00-0000-7000-8000-000000000605', '01989f00-0000-7000-8000-000000000205', 'delivered', 'IDR', 780000, now() - interval '16 days', now() - interval '14 days', now() - interval '12 days', NULL, '', now() - interval '18 days', now() - interval '12 days'),
    ('01989f00-0000-7000-8000-000000000680', '01989f00-0000-7000-8000-000000000610', '01989f00-0000-7000-8000-000000000207', 'delivered', 'IDR', 979000, now() - interval '32 days', now() - interval '31 days', now() - interval '30 days', NULL, '', now() - interval '33 days', now() - interval '30 days');

INSERT INTO purchase_items (id, seller_order_id, product_id, variant_id, product_name, variant_name, sku, image_url, quantity, unit_price_minor, line_total_minor, currency, created_at)
VALUES
    ('01989f00-0000-7000-8000-000000000621', '01989f00-0000-7000-8000-000000000611', '01989f00-0000-7000-8000-000000000341', '01989f00-0000-7000-8000-000000000441', 'Demo Desk Caddy', 'Navy', 'DEMO-CADDY-NAVY', '/images/catalog-v2/modular-desk-organizer.webp', 1, 450000, 450000, 'IDR', now() - interval '10 days'),
    ('01989f00-0000-7000-8000-000000000622', '01989f00-0000-7000-8000-000000000612', '01989f00-0000-7000-8000-000000000301', '01989f00-0000-7000-8000-000000000401', 'Handwoven Market Basket', 'Natural', 'NUSA-BASKET-NAT', '/images/catalog-v2/loom-carry-basket.webp', 1, 249000, 249000, 'IDR', now() - interval '12 days'),
    ('01989f00-0000-7000-8000-000000000623', '01989f00-0000-7000-8000-000000000613', '01989f00-0000-7000-8000-000000000302', '01989f00-0000-7000-8000-000000000402', 'Indigo Utility Tray', 'Indigo', 'JLS-TRAY-INDIGO', '/images/catalog-v2/studio-tray.webp', 2, 159000, 318000, 'IDR', now() - interval '14 days'),
    ('01989f00-0000-7000-8000-000000000624', '01989f00-0000-7000-8000-000000000614', '01989f00-0000-7000-8000-000000000316', '01989f00-0000-7000-8000-000000000416', 'Roll-Top Backpack', 'Navy', 'TEN-BACKPACK-NAVY', '/images/catalog-v2/roll-top-backpack.webp', 1, 1099000, 1099000, 'IDR', now() - interval '4 days'),
    ('01989f00-0000-7000-8000-000000000625', '01989f00-0000-7000-8000-000000000615', '01989f00-0000-7000-8000-000000000341', '01989f00-0000-7000-8000-000000000441', 'Demo Desk Caddy', 'Navy', 'DEMO-CADDY-NAVY', '/images/catalog-v2/modular-desk-organizer.webp', 1, 450000, 450000, 'IDR', now() - interval '18 days'),
    ('01989f00-0000-7000-8000-000000000626', '01989f00-0000-7000-8000-000000000620', '01989f00-0000-7000-8000-000000000321', '01989f00-0000-7000-8000-000000000421', 'Pour Over Set', 'Three piece', 'BREW-POUR-SET', '/images/catalog-v2/pour-over-set.webp', 1, 780000, 780000, 'IDR', now() - interval '18 days'),
    ('01989f00-0000-7000-8000-000000000627', '01989f00-0000-7000-8000-000000000616', '01989f00-0000-7000-8000-000000000341', '01989f00-0000-7000-8000-000000000441', 'Demo Desk Caddy', 'Navy', 'DEMO-CADDY-NAVY', '/images/catalog-v2/modular-desk-organizer.webp', 1, 450000, 450000, 'IDR', now() - interval '20 days'),
    ('01989f00-0000-7000-8000-000000000628', '01989f00-0000-7000-8000-000000000617', '01989f00-0000-7000-8000-000000000341', '01989f00-0000-7000-8000-000000000441', 'Demo Desk Caddy', 'Navy', 'DEMO-CADDY-NAVY', '/images/catalog-v2/modular-desk-organizer.webp', 1, 450000, 450000, 'IDR', now() - interval '22 days'),
    ('01989f00-0000-7000-8000-000000000629', '01989f00-0000-7000-8000-000000000618', '01989f00-0000-7000-8000-000000000301', '01989f00-0000-7000-8000-000000000401', 'Handwoven Market Basket', 'Natural', 'NUSA-BASKET-NAT', '/images/catalog-v2/loom-carry-basket.webp', 1, 249000, 249000, 'IDR', now() - interval '26 days'),
    ('01989f00-0000-7000-8000-000000000630', '01989f00-0000-7000-8000-000000000619', '01989f00-0000-7000-8000-000000000302', '01989f00-0000-7000-8000-000000000402', 'Indigo Utility Tray', 'Indigo', 'JLS-TRAY-INDIGO', '/images/catalog-v2/studio-tray.webp', 1, 159000, 159000, 'IDR', now() - interval '30 days'),
    ('01989f00-0000-7000-8000-000000000681', '01989f00-0000-7000-8000-000000000680', '01989f00-0000-7000-8000-000000000331', '01989f00-0000-7000-8000-000000000431', 'Canvas Tote', 'Off-white navy', 'CARRY-TOTE-OFFWHITE', '/images/catalog-v2/canvas-tote.webp', 1, 680000, 680000, 'IDR', now() - interval '33 days'),
    ('01989f00-0000-7000-8000-000000000682', '01989f00-0000-7000-8000-000000000680', '01989f00-0000-7000-8000-000000000333', '01989f00-0000-7000-8000-000000000433', 'Knit Beanie', 'One size', 'CARRY-BEANIE-CHARCOAL', '/images/catalog-v2/knit-beanie.webp', 1, 299000, 299000, 'IDR', now() - interval '33 days');

INSERT INTO inventory_reservations (id, purchase_id, variant_id, quantity, status, expires_at, resolved_at, created_at)
VALUES
    ('01989f00-0000-7000-8000-000000000631', '01989f00-0000-7000-8000-000000000601', '01989f00-0000-7000-8000-000000000441', 1, 'converted', now() - interval '9 days', now() - interval '9 days', now() - interval '10 days'),
    ('01989f00-0000-7000-8000-000000000632', '01989f00-0000-7000-8000-000000000602', '01989f00-0000-7000-8000-000000000401', 1, 'converted', now() - interval '11 days', now() - interval '11 days', now() - interval '12 days'),
    ('01989f00-0000-7000-8000-000000000633', '01989f00-0000-7000-8000-000000000603', '01989f00-0000-7000-8000-000000000402', 2, 'converted', now() - interval '13 days', now() - interval '13 days', now() - interval '14 days'),
    ('01989f00-0000-7000-8000-000000000634', '01989f00-0000-7000-8000-000000000604', '01989f00-0000-7000-8000-000000000416', 1, 'converted', now() - interval '3 days', now() - interval '3 days', now() - interval '4 days'),
    ('01989f00-0000-7000-8000-000000000635', '01989f00-0000-7000-8000-000000000605', '01989f00-0000-7000-8000-000000000441', 1, 'converted', now() - interval '17 days', now() - interval '17 days', now() - interval '18 days'),
    ('01989f00-0000-7000-8000-000000000636', '01989f00-0000-7000-8000-000000000605', '01989f00-0000-7000-8000-000000000421', 1, 'converted', now() - interval '17 days', now() - interval '17 days', now() - interval '18 days'),
    ('01989f00-0000-7000-8000-000000000637', '01989f00-0000-7000-8000-000000000606', '01989f00-0000-7000-8000-000000000441', 1, 'converted', now() - interval '19 days', now() - interval '19 days', now() - interval '20 days'),
    ('01989f00-0000-7000-8000-000000000638', '01989f00-0000-7000-8000-000000000607', '01989f00-0000-7000-8000-000000000441', 1, 'converted', now() - interval '21 days', now() - interval '21 days', now() - interval '22 days'),
    ('01989f00-0000-7000-8000-000000000639', '01989f00-0000-7000-8000-000000000608', '01989f00-0000-7000-8000-000000000401', 1, 'converted', now() - interval '25 days', now() - interval '25 days', now() - interval '26 days'),
    ('01989f00-0000-7000-8000-000000000640', '01989f00-0000-7000-8000-000000000609', '01989f00-0000-7000-8000-000000000402', 1, 'converted', now() - interval '29 days', now() - interval '29 days', now() - interval '30 days'),
    ('01989f00-0000-7000-8000-000000000683', '01989f00-0000-7000-8000-000000000610', '01989f00-0000-7000-8000-000000000431', 1, 'converted', now() - interval '32 days', now() - interval '32 days', now() - interval '33 days'),
    ('01989f00-0000-7000-8000-000000000684', '01989f00-0000-7000-8000-000000000610', '01989f00-0000-7000-8000-000000000433', 1, 'converted', now() - interval '32 days', now() - interval '32 days', now() - interval '33 days');

INSERT INTO reviews (id, buyer_id, purchase_item_id, product_id, rating, title, body, created_at, updated_at)
VALUES
    ('01989f00-0000-7000-8000-000000000641', '01989f00-0000-7000-8000-000000000001', '01989f00-0000-7000-8000-000000000625', '01989f00-0000-7000-8000-000000000341', 5, 'Keeps everything in reach', 'Solid modules and clean finish. Fits the demo workspace perfectly.', now() - interval '11 days', now() - interval '11 days')
    ,('01989f00-0000-7000-8000-000000000642', '01989f00-0000-7000-8000-000000000005', '01989f00-0000-7000-8000-000000000629', '01989f00-0000-7000-8000-000000000301', 4, 'Strong market-day carry', 'Handles sit comfortably and the basket keeps its shape under a full load.', now() - interval '22 days', now() - interval '22 days')
    ,('01989f00-0000-7000-8000-000000000643', '01989f00-0000-7000-8000-000000000007', '01989f00-0000-7000-8000-000000000630', '01989f00-0000-7000-8000-000000000302', 3, 'Useful entryway tray', 'Color is rich and the low edge works well for keys, though the base can slide.', now() - interval '26 days', now() - interval '26 days');

-- Catalog proof: every published product has three delivered sales and three verified
-- reviews. Separate orders and buyers keep marketplace activity believable.
DELETE FROM reviews
WHERE id BETWEEN '01989f00-0000-7000-8000-000000005001' AND '01989f00-0000-7000-8000-000000005999';
DELETE FROM inventory_reservations
WHERE id BETWEEN '01989f00-0000-7000-8000-000000004001' AND '01989f00-0000-7000-8000-000000004999';
DELETE FROM purchase_items
WHERE id BETWEEN '01989f00-0000-7000-8000-000000003001' AND '01989f00-0000-7000-8000-000000003999';
DELETE FROM seller_orders
WHERE id BETWEEN '01989f00-0000-7000-8000-000000002001' AND '01989f00-0000-7000-8000-000000002999';
DELETE FROM purchases
WHERE id BETWEEN '01989f00-0000-7000-8000-000000001001' AND '01989f00-0000-7000-8000-000000001999';

WITH catalog_sales AS (
    SELECT p.*, v.id AS variant_id, v.name AS variant_name, v.sku,
        v.price_minor, v.currency, COALESCE(i.url, '') AS image_url,
        row_number() OVER (ORDER BY p.id, review_copy) AS sale_number
    FROM products p
    CROSS JOIN generate_series(1, 3) AS review_series(review_copy)
    JOIN LATERAL (
        SELECT pv.* FROM product_variants pv
        WHERE pv.product_id = p.id ORDER BY pv.id LIMIT 1
    ) v ON true
    LEFT JOIN LATERAL (
        SELECT pi.url FROM product_images pi
        WHERE pi.product_id = p.id ORDER BY pi.position LIMIT 1
    ) i ON true
    WHERE p.status = 'published'
)
INSERT INTO purchases (id, reference, buyer_id, idempotency_key, status, payment_status, payment_intent_id, currency, subtotal_minor, total_minor, reservation_expires_at, created_at, updated_at)
SELECT ('01989f00-0000-7000-8000-' || lpad((1000 + sale_number)::text, 12, '0'))::uuid,
    'CL-SEED' || lpad(sale_number::text, 8, '0'),
    CASE sale_number % 3
        WHEN 1 THEN '01989f00-0000-7000-8000-000000000001'::uuid
        WHEN 2 THEN '01989f00-0000-7000-8000-000000000005'::uuid
        ELSE '01989f00-0000-7000-8000-000000000007'::uuid
    END,
    'catalog-sale-' || sale_number, 'paid', 'succeeded',
    'pi_catalog_sale_' || sale_number, currency,
    price_minor * (1 + sale_number % 2), price_minor * (1 + sale_number % 2),
    now() - make_interval(days => 45 + sale_number::int),
    now() - make_interval(days => 46 + sale_number::int),
    now() - make_interval(days => 42 + sale_number::int)
FROM catalog_sales;

WITH catalog_sales AS (
    SELECT p.store_id, v.price_minor, v.currency,
        row_number() OVER (ORDER BY p.id, review_copy) AS sale_number
    FROM products p
    CROSS JOIN generate_series(1, 3) AS review_series(review_copy)
    JOIN LATERAL (SELECT pv.* FROM product_variants pv WHERE pv.product_id = p.id ORDER BY pv.id LIMIT 1) v ON true
    WHERE p.status = 'published'
)
INSERT INTO seller_orders (id, purchase_id, store_id, status, currency, subtotal_minor, processing_at, shipped_at, delivered_at, cancellation_reason, created_at, updated_at)
SELECT ('01989f00-0000-7000-8000-' || lpad((2000 + sale_number)::text, 12, '0'))::uuid,
    ('01989f00-0000-7000-8000-' || lpad((1000 + sale_number)::text, 12, '0'))::uuid,
    store_id, 'delivered', currency, price_minor * (1 + sale_number % 2),
    now() - make_interval(days => 45 + sale_number::int),
    now() - make_interval(days => 44 + sale_number::int),
    now() - make_interval(days => 42 + sale_number::int), '',
    now() - make_interval(days => 46 + sale_number::int),
    now() - make_interval(days => 42 + sale_number::int)
FROM catalog_sales;

WITH catalog_sales AS (
    SELECT p.id AS product_id, p.name AS product_name, v.id AS variant_id,
        v.name AS variant_name, v.sku, v.price_minor, v.currency,
        COALESCE(i.url, '') AS image_url,
        row_number() OVER (ORDER BY p.id, review_copy) AS sale_number
    FROM products p
    CROSS JOIN generate_series(1, 3) AS review_series(review_copy)
    JOIN LATERAL (SELECT pv.* FROM product_variants pv WHERE pv.product_id = p.id ORDER BY pv.id LIMIT 1) v ON true
    LEFT JOIN LATERAL (SELECT pi.url FROM product_images pi WHERE pi.product_id = p.id ORDER BY pi.position LIMIT 1) i ON true
    WHERE p.status = 'published'
)
INSERT INTO purchase_items (id, seller_order_id, product_id, variant_id, product_name, variant_name, sku, image_url, quantity, unit_price_minor, line_total_minor, currency, created_at)
SELECT ('01989f00-0000-7000-8000-' || lpad((3000 + sale_number)::text, 12, '0'))::uuid,
    ('01989f00-0000-7000-8000-' || lpad((2000 + sale_number)::text, 12, '0'))::uuid,
    product_id, variant_id, product_name, variant_name, sku, image_url,
    1 + sale_number % 2, price_minor, price_minor * (1 + sale_number % 2), currency,
    now() - make_interval(days => 46 + sale_number::int)
FROM catalog_sales;

WITH catalog_sales AS (
    SELECT v.id AS variant_id, row_number() OVER (ORDER BY p.id, review_copy) AS sale_number
    FROM products p
    CROSS JOIN generate_series(1, 3) AS review_series(review_copy)
    JOIN LATERAL (SELECT pv.* FROM product_variants pv WHERE pv.product_id = p.id ORDER BY pv.id LIMIT 1) v ON true
    WHERE p.status = 'published'
)
INSERT INTO inventory_reservations (id, purchase_id, variant_id, quantity, status, expires_at, resolved_at, created_at)
SELECT ('01989f00-0000-7000-8000-' || lpad((4000 + sale_number)::text, 12, '0'))::uuid,
    ('01989f00-0000-7000-8000-' || lpad((1000 + sale_number)::text, 12, '0'))::uuid,
    variant_id, 1 + sale_number % 2, 'converted',
    now() - make_interval(days => 45 + sale_number::int),
    now() - make_interval(days => 45 + sale_number::int),
    now() - make_interval(days => 46 + sale_number::int)
FROM catalog_sales;

WITH catalog_sales AS (
    SELECT p.id AS product_id, p.name AS product_name, review_copy,
        row_number() OVER (ORDER BY p.id, review_copy) AS sale_number
    FROM products p
    CROSS JOIN generate_series(1, 3) AS review_series(review_copy)
    WHERE p.status = 'published'
      AND EXISTS (SELECT 1 FROM product_variants v WHERE v.product_id = p.id)
)
INSERT INTO reviews (id, buyer_id, purchase_item_id, product_id, rating, title, body, created_at, updated_at)
SELECT ('01989f00-0000-7000-8000-' || lpad((5000 + sale_number)::text, 12, '0'))::uuid,
    CASE sale_number % 3
        WHEN 1 THEN '01989f00-0000-7000-8000-000000000001'::uuid
        WHEN 2 THEN '01989f00-0000-7000-8000-000000000005'::uuid
        ELSE '01989f00-0000-7000-8000-000000000007'::uuid
    END,
    ('01989f00-0000-7000-8000-' || lpad((3000 + sale_number)::text, 12, '0'))::uuid,
    product_id, 4 + sale_number % 2,
    CASE review_copy
        WHEN 1 THEN product_name || ' feels well made'
        WHEN 2 THEN 'Happy with the ' || product_name
        ELSE product_name || ' works well day to day'
    END,
    CASE sale_number % 16
        WHEN 0 THEN product_name || ' has a solid finish and feels durable in regular use. It arrived in good condition.'
        WHEN 1 THEN 'The ' || product_name || ' matches the listing photos and dimensions. Quality is good for the price.'
        WHEN 2 THEN product_name || ' is simple, practical, and easy to use. The packaging could be better, but the item was fine.'
        WHEN 3 THEN 'I have used the ' || product_name || ' for a few weeks now. It still looks good and works as expected.'
        WHEN 4 THEN 'The material and finish on the ' || product_name || ' are good. Color is slightly darker than it looked on my screen.'
        WHEN 5 THEN product_name || ' arrived on time and was packed securely. No issues with it so far.'
        WHEN 6 THEN 'Build quality on the ' || product_name || ' feels dependable. I would have liked one more color option.'
        WHEN 7 THEN product_name || ' fits the space well and looks like the photos. Setup was straightforward.'
        WHEN 8 THEN 'I bought the ' || product_name || ' for everyday use. It does the job well without taking up too much room.'
        WHEN 9 THEN 'The ' || product_name || ' feels nicer in person than I expected. Clean finish and no visible defects.'
        WHEN 10 THEN product_name || ' has been easy to maintain and comfortable to use. Delivery took a day longer than expected.'
        WHEN 11 THEN 'Happy with the size and quality of the ' || product_name || '. It has held up well so far.'
        WHEN 12 THEN 'The ' || product_name || ' is well proportioned and feels sturdy. Instructions were brief but sufficient.'
        WHEN 13 THEN product_name || ' works well for what I needed. The design is understated and the finish is consistent.'
        WHEN 14 THEN 'After regular use, the ' || product_name || ' still feels solid. Price seems fair for the quality.'
        ELSE 'No major complaints about the ' || product_name || '. It arrived as described and has worked reliably.'
    END,
    now() - make_interval(days => 40 + sale_number::int),
    now() - make_interval(days => 40 + sale_number::int)
FROM catalog_sales;

DELETE FROM notifications
WHERE id BETWEEN '01989f00-0000-7000-8000-000000000651' AND '01989f00-0000-7000-8000-000000000654';

INSERT INTO notifications (id, user_id, source_event_id, kind, title, body, read_at, created_at)
VALUES
    ('01989f00-0000-7000-8000-000000000651', '01989f00-0000-7000-8000-000000000001', '01989f00-0000-7000-8000-000000000604', 'order.delivered', 'Order delivered', 'CL-DEMO00000004 was delivered.', now() - interval '1 day', now() - interval '1 day'),
    ('01989f00-0000-7000-8000-000000000652', '01989f00-0000-7000-8000-000000000001', '01989f00-0000-7000-8000-000000000605', 'order.delivered', 'Order delivered', 'CL-DEMO00000005 was delivered and can now be reviewed.', NULL, now() - interval '12 days'),
    ('01989f00-0000-7000-8000-000000000653', '01989f00-0000-7000-8000-000000000006', '01989f00-0000-7000-8000-000000000603', 'order.delivered', 'Order delivered', 'CL-DEMO00000003 was delivered.', NULL, now() - interval '10 days'),
    ('01989f00-0000-7000-8000-000000000654', '01989f00-0000-7000-8000-000000000002', '01989f00-0000-7000-8000-000000000605', 'order.delivered', 'Order delivered', 'CL-DEMO00000005 was delivered and can now be reviewed.', now() - interval '10 days', now() - interval '12 days')
ON CONFLICT (user_id, source_event_id) DO UPDATE SET kind = EXCLUDED.kind, title = EXCLUDED.title, body = EXCLUDED.body, read_at = EXCLUDED.read_at, created_at = EXCLUDED.created_at;

INSERT INTO audit_log (id, actor_id, action, resource_type, resource_id, metadata, created_at)
VALUES
    ('01989f00-0000-7000-8000-000000000661', '01989f00-0000-7000-8000-000000000003', 'store.verified.approved', 'store', '01989f00-0000-7000-8000-000000000200', '{"note":"Demo seller identity and business verified."}', now() - interval '40 days'),
    ('01989f00-0000-7000-8000-000000000662', '01989f00-0000-7000-8000-000000000003', 'store.verified.rejected', 'store', '01989f00-0000-7000-8000-000000000210', '{"note":"Business registration could not be verified."}', now() - interval '2 days'),
    ('01989f00-0000-7000-8000-000000000663', '01989f00-0000-7000-8000-000000000003', 'product.status.suspended', 'product', '01989f00-0000-7000-8000-000000000344', '{"reason":"Safety certification is missing."}', now() - interval '1 day'),
    ('01989f00-0000-7000-8000-000000000664', '01989f00-0000-7000-8000-000000000003', 'user.status.suspended', 'user', '01989f00-0000-7000-8000-000000000016', '{"from":"active","to":"suspended","reason":"Repeated policy violations in demo data."}', now() - interval '3 days'),
    ('01989f00-0000-7000-8000-000000000665', '01989f00-0000-7000-8000-000000000002', 'inventory.adjusted', 'product_variant', '01989f00-0000-7000-8000-000000000442', '{"delta":-3,"resultingStock":2,"reason":"Demo cycle count correction."}', now() - interval '7 days'),
    ('01989f00-0000-7000-8000-000000000666', '01989f00-0000-7000-8000-000000000009', 'seller_order.shipped', 'seller_order', '01989f00-0000-7000-8000-000000000614', '{"from":"processing","to":"shipped"}', now() - interval '2 days'),
    ('01989f00-0000-7000-8000-000000000667', '01989f00-0000-7000-8000-000000000002', 'seller_order.delivered', 'seller_order', '01989f00-0000-7000-8000-000000000615', '{"from":"shipped","to":"delivered"}', now() - interval '12 days'),
    ('01989f00-0000-7000-8000-000000000668', '01989f00-0000-7000-8000-000000000001', 'seller_order.delivered', 'seller_order', '01989f00-0000-7000-8000-000000000617', '{"from":"shipped","to":"delivered"}', now() - interval '18 days'),
    ('01989f00-0000-7000-8000-000000000669', '01989f00-0000-7000-8000-000000000001', 'review.created', 'review', '01989f00-0000-7000-8000-000000000641', '{"rating":5,"productId":"01989f00-0000-7000-8000-000000000341","purchaseItemId":"01989f00-0000-7000-8000-000000000625"}', now() - interval '11 days')
ON CONFLICT (id) DO UPDATE SET actor_id = EXCLUDED.actor_id, action = EXCLUDED.action, resource_type = EXCLUDED.resource_type, resource_id = EXCLUDED.resource_id, metadata = EXCLUDED.metadata, created_at = EXCLUDED.created_at;

INSERT INTO inventory_ledger (id, variant_id, actor_id, delta, resulting_stock, reason, created_at)
VALUES ('01989f00-0000-7000-8000-000000000670', '01989f00-0000-7000-8000-000000000442', '01989f00-0000-7000-8000-000000000002', -3, 2, 'Demo cycle count correction.', now() - interval '7 days')
ON CONFLICT (id) DO UPDATE SET delta = EXCLUDED.delta, resulting_stock = EXCLUDED.resulting_stock, reason = EXCLUDED.reason, created_at = EXCLUDED.created_at;
