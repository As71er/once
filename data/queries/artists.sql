-- name: CreateArtist :one
INSERT INTO artists (name)
VALUES (?) ON CONFLICT(name) DO UPDATE SET name = name
RETURNING *;

-- name: UpdateArtist :exec
UPDATE artists
SET name = ?
WHERE id = ?;

-- name: DeleteArtist :exec
DELETE FROM artists
WHERE id = ?;

-- name: ListArtists :many
SELECT id, name
FROM artists
ORDER BY name;

-- name: GetArtistByName :one
SELECT * FROM artists
WHERE name = ?;

-- name: GetArtistById :one
SELECT id, name FROM artists
WHERE id = ?;


-------------------------------------------------------------------------


-- name: AddArtistRelease :exec
INSERT INTO release_artists
(release_id, artist_id)
VALUES (?, ?);

-- name: GetReleaseArtists :many
SELECT ar.id, ar.name
FROM release_artists ra
JOIN artists ar ON ra.artist_id = ar.id
WHERE ra.release_id = ?
ORDER BY ar.name;

-- name: RemoveReleaseArtist :exec
DELETE FROM release_artists
WHERE release_id = ? AND artist_id = ?;


-------------------------------------------------------------------------


-- name: AddArtistTrack :exec
INSERT INTO track_artists
(track_id, artist_id, role)
VALUES (?, ?, ?);

-- name: UpdateTrackArtistRole :exec
UPDATE track_artists
SET role = ?
WHERE track_id = ? AND artist_id = ?;

-- name: GetTrackArtists :many
SELECT ar.id, ar.name, ta.role
FROM track_artists ta
JOIN artists AS ar ON ta.artist_id = ar.id
WHERE ta.track_id = ?
ORDER BY
    CASE ta.role
        WHEN 'primary' THEN 1
        WHEN 'featured' THEN 2
    END, ar.name;

-- name: RemoveTrackArtist :exec
DELETE FROM track_artists
WHERE track_id = ? AND artist_id = ?;

-- name: GetTracksByArtist :many
SELECT t.id, t.title, r.name AS release_name, ta.role
FROM track_artists ta
JOIN tracks t ON ta.track_id = t.id
JOIN releases r ON t.release_id = r.id
WHERE ta.artist_id = ?
ORDER BY r.date DESC, t.track;
