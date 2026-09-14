-- name: CreateRelease :one
INSERT INTO releases
(title, name, release_type, date, total_tracks, total_discs, cover_id)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetReleaseById :one
SELECT * FROM releases
WHERE id = ?;

-- name: UpdateRelease :one
UPDATE releases
SET title = ?, release_type = ?, date = ?, total_tracks = ?, total_discs = ?
WHERE id = ?
RETURNING *;

-- name: UpdateReleaseCover :exec
UPDATE releases
SET cover_id = ?
WHERE id = ?;

-- name: DeleteRelease :exec
DELETE FROM releases
WHERE id = ?;

-- name: GetReleaseCover :one
SELECT r.id, r.name, c.id, c.name, c.thumb_name FROM releases AS r
JOIN covers AS c ON  r.cover_id = c.id
WHERE r.id = ?;

-- name: GetReleaseDuration :one
SELECT SUM(duration) AS total_duration
FROM tracks
WHERE release_id = ?;

-- name: ListTracksRelease :many
SELECT t.id, t.title, t.duration, t.track, t.disc, t.composer, t.audio_file_id
FROM tracks AS t
WHERE t.release_id = ?;

-- name: ListReleases :many
SELECT id, title, release_type, date, total_tracks, total_discs, cover_id
FROM releases
LIMIT ? OFFSET ?;
