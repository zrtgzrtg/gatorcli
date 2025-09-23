-- name: GetPostsForUser :many
select *
from posts
where (select user_id from feeds where posts.feed_id=feeds.id) = $1
limit $2;