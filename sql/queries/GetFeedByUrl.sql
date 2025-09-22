-- name: GetFeedByUrl :one
Select *
from feeds
where feeds.url = $1;