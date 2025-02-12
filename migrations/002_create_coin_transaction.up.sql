CREATE TABLE coin_transactions (
   id SERIAL PRIMARY KEY,
   from_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
   to_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
   amount INTEGER NOT NULL CHECK (amount > 0),
   created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
