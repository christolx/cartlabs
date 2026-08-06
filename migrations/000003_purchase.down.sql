DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS payment_events;
DROP TABLE IF EXISTS inventory_reservations;
DROP TABLE IF EXISTS purchase_items;
DROP TABLE IF EXISTS seller_orders;
DROP TABLE IF EXISTS purchases;
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS carts;

DROP TYPE IF EXISTS reservation_status;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS seller_order_status;
DROP TYPE IF EXISTS purchase_status;
