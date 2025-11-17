-- +goose Up
CREATE TABLE posts (
	id UUID PRIMARY KEY,
	title text NOT NULL,
	url text UNIQUE NOT NULL,
	description text,
	published_at TIMESTAMP,
	feed_id UUID NOT NULL REFERENCES feeds(id),
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE posts;
