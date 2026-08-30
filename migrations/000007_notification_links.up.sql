ALTER TABLE notifications
    ADD COLUMN href text NOT NULL DEFAULT ''
        CHECK (href = '' OR href ~ '^/');

UPDATE notifications n
SET href = CASE
    WHEN (SELECT u.role FROM users u WHERE u.id = n.user_id) = 'seller'
        THEN '/seller/orders'
    WHEN (SELECT u.role FROM users u WHERE u.id = n.user_id) = 'buyer'
        AND (SELECT e.payload->>'purchaseId' FROM outbox_events e WHERE e.id = n.source_event_id) IS NOT NULL
        THEN '/purchases/' || (
            SELECT e.payload->>'purchaseId'
            FROM outbox_events e
            WHERE e.id = n.source_event_id
        )
    ELSE ''
END
WHERE n.kind LIKE 'purchase.%' OR n.kind LIKE 'order.%';
