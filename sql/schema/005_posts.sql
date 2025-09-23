-- +goose Up
CREATE TABLE posts(
    id uuid primary key,
    created_at timestamp,
    updated_at timestamp,
    title text not null,
    url text unique not null,
    description text not null,
    published_at timestamp not null,
    feed_id uuid not null,
    foreign key (feed_id) references feeds(id)
);

-- +goose Down
DROP TABLE posts;