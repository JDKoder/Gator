-- name: CreateFeed :one
INSERT INTO feeds (id, name, url, user_id, created_at, updated_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
	$5,
	$6
)
RETURNING *;

-- name: GetUserFeeds :many
SELECT F.*, U.name AS UserName FROM feeds F
INNER JOIN users U ON F.user_id = U.id;

-- name: GetFeedByURL :one
SELECT * FROM feeds
WHERE url = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds 
SET last_fetched_at=$1, updated_at=$1
WHERE id = $2; 

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY 
	last_fetched_at ASC NULLS FIRST,
	updated_at ASC
LIMIT(1);




