CREATE TABLE IF NOT EXISTS dialogues (
    id SERIAL PRIMARY KEY,
    user1_id INT NOT NULL REFERENCES users(id),
    user2_id INT NOT NULL REFERENCES users(id),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT now(),
    closed_at TIMESTAMP,
    CHECK (user1_id <> user2_id)
);
