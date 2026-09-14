-- name: CreateCover :one
INSERT INTO covers (name, thumb_name) VALUES (?, ?)
RETURNING id, name, thumb_name;

-- name: GetCoverById :one
SELECT id, name, thumb_name FROM covers
WHERE id = ?;

-- name: ListCovers :many
SELECT id, name, thumb_name FROM covers
LIMIT ? OFFSET ?;

-- name: DeleteCover :exec
DELETE FROM covers WHERE id = ?;
