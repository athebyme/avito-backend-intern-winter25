DROP INDEX idx_transactions_from_user ON coin_transactions(from_user_id);
DROP INDEX idx_transactions_to_user ON coin_transactions(to_user_id);
DROP INDEX idx_purchases_user_id ON purchases(user_id);
DROP INDEX idx_merch_name ON merch(name);