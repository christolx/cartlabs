ALTER TABLE products
    ADD COLUMN moderation_status moderation_status NOT NULL DEFAULT 'pending',
    ADD COLUMN moderation_note text NOT NULL DEFAULT ''
        CHECK (char_length(moderation_note) <= 500);

UPDATE products p
SET
    status = 'published',
    moderation_status = 'rejected',
    moderation_note = left(COALESCE((
        SELECT ae.metadata->>'reason'
        FROM audit_log ae
        WHERE ae.resource_type = 'product'
          AND ae.resource_id = p.id
          AND ae.action = 'product.status.suspended'
        ORDER BY ae.created_at DESC, ae.id DESC
        LIMIT 1
    ), ''), 500)
WHERE p.status = 'suspended';

UPDATE products
SET moderation_status = 'approved'
WHERE status = 'published';

ALTER TYPE product_status RENAME TO product_status_refactor;
CREATE TYPE product_status AS ENUM ('draft', 'published', 'archived');

ALTER TABLE products ALTER COLUMN status DROP DEFAULT;
ALTER TABLE products
    ALTER COLUMN status TYPE product_status USING status::text::product_status;
ALTER TABLE products ALTER COLUMN status SET DEFAULT 'draft';
DROP TYPE product_status_refactor;
