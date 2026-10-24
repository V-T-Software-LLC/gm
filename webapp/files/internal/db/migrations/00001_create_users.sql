-- +goose Up
create table users (
    id bigint generated always as identity primary key,
    first_name text not null,
    last_name text not null,
    email text not null unique,
    password_hash text not null,
    updated_at timestamptz not null default now(),
    created_at timestamptz not null default now()
);

-- +goose Down
drop table users;
