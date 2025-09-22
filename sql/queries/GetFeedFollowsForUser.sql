-- name: GetFeedFollowsForUser :many
select feed_id,users.name as username,feeds.name as feedname
from feed_follows
join users
on users.id = feed_follows.user_id
join feeds
on feeds.id = feed_follows.feed_id
where feed_follows.user_id = $1;