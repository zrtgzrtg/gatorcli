-- name: GetUserById :one
Select *
from users
where id = $1;