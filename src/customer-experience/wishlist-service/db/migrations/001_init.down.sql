-- 001_init.down.sql
-- Tears down the wishlist_items table.

DROP INDEX IF EXISTS idx_wishlist_items_customer_added;
DROP TABLE IF EXISTS wishlist_items;
