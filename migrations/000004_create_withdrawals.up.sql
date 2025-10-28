CREATE TABLE withdrawals (
    user_id INTEGER NOT NULL REFERENCES users (id),
    order_id INTEGER NOT NULL REFERENCES orders (id),
    sum NUMERIC NOT NULL CHECK(sum >= 0) DEFAULT 0
);
