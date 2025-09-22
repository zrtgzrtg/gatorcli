-- name: CreateFeedFollows :one
insert into feed_follows(id,user_id,feed_id)
VALUES (
    $1,
    $2,
    $3
)

RETURNING *,
(select name from users where id = feed_follows.user_id) as username,
(select name from feeds where id = feed_follows.feed_id) as feedname;