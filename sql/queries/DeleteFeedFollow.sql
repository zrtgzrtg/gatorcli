-- name: DeleteFeedFollow :many
delete from feed_follows
where user_id = $1 and feed_id = $2
RETURNING *;