ALTER TYPE product_status RENAME TO product_status_legacy;
CREATE TYPE product_status AS ENUM ('draft', 'published', 'archived', 'suspended');

ALTER TABLE products ALTER COLUMN status DROP DEFAULT;
ALTER TABLE products
    ALTER COLUMN status TYPE product_status USING status::text::product_status;
ALTER TABLE products ALTER COLUMN status SET DEFAULT 'draft';
DROP TYPE product_status_legacy;

INSERT INTO audit_log (id, actor_id, action, resource_type, resource_id, metadata, created_at)
SELECT
    md5('product.moderation.migrated:' || p.id::text)::uuid,
    COALESCE(
        (
            SELECT ae.actor_id
            FROM audit_log ae
            WHERE ae.resource_type = 'product'
              AND ae.resource_id = p.id
              AND ae.action LIKE 'product.moderated.%'
            ORDER BY ae.created_at DESC, ae.id DESC
            LIMIT 1
        ),
        s.seller_id
    ),
    'product.moderation.migrated',
    'product',
    p.id,
    jsonb_build_object(
        'legacyProductStatus', p.status::text,
        'legacyModerationStatus', p.moderation_status::text,
        'legacyModerationNote', p.moderation_note
    ),
    p.updated_at
FROM products p
JOIN stores s ON s.id = p.store_id
WHERE btrim(p.moderation_note) <> ''
ON CONFLICT (id) DO NOTHING;

UPDATE products
SET status = 'draft'
WHERE status = 'published'
  AND moderation_status = 'rejected';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM products
        WHERE status::text NOT IN ('draft', 'published', 'archived')
    ) THEN
        RAISE EXCEPTION 'unmapped product state remains';
    END IF;
END
$$;

ALTER TABLE products
    DROP COLUMN moderation_status,
    DROP COLUMN moderation_note;
