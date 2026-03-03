CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    dialogue_id INT NOT NULL REFERENCES dialogues(id),
    sender_id INT NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);
