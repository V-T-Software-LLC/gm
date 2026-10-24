-- name: CreateUser :one
insert into users (first_name, last_name, email, password_hash)
values ($1, $2, $3, $4)
returning *;

-- name: GetUser :one
select * from users where id = $1;

-- name: GetUserByEmail :one
select * from users where email = $1;
