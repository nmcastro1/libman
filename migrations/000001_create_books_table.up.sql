CREATE TABLE IF NOT EXISTS books (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    authors text[] NOT NULL,
    publisher text NOT NULL,
    year integer NOT NULL,
    language text NOT NULL,
    pages integer NOT NULL,
    version integer NOT NULL DEFAULT 1
);