CREATE TABLE IF NOT EXISTS urls (
    id serial PRIMARY KEY,
    original_url text UNIQUE NOT NULL,
    code text UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
