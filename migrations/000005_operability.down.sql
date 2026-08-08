DROP INDEX IF EXISTS outbox_events_dead_letter_idx;
DROP INDEX IF EXISTS outbox_events_dispatch_idx;

ALTER TABLE outbox_events
    DROP COLUMN IF EXISTS dead_lettered_at,
    DROP COLUMN IF EXISTS next_attempt_at;

CREATE INDEX outbox_events_unpublished_idx ON outbox_events (occurred_at, id)
    WHERE published_at IS NULL;
