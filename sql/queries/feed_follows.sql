-- name: CreateFeedFollow :one
WITH inserted_feed_follows AS 
	(INSERT INTO feed_follows (id, user_id, feed_id, created_at, updated_at)
	VALUES (
		$1,
		$2,
		$3,
		$4,
		$5
		)
	RETURNING *)
SELECT 	inserted_feed_follows.*,
		users.name AS user_name,
		feeds.name AS feed_name
FROM inserted_feed_follows 
INNER JOIN users ON inserted_feed_follows.user_id = users.id
INNER JOIN feeds ON inserted_feed_follows.feed_id = feeds.id;

--It should return all the feed follows for a given user, and include the names of the feeds and user in the result.
-- name: GetFeedFollowsForUser :many
SELECT 	feed_follows.*,
		feeds.name AS feed_name,
		users.name AS user_name
FROM feed_follows
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
INNER JOIN users ON feed_follows.user_id = users.id 
WHERE users.name = $1;

-- name: DeleteFeedFollowsByUserAndFeed :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;

