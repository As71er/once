-- name: CreateArtist :one
INSERT INTO artists (name)
VALUES (?) ON CONFLICT(name) DO UPDATE SET name = name
RETURNING *;

-- name: UpdateArtist :one
UPDATE artists
SET name = ?
WHERE id = ?
RETURNING *;

-- name: DeleteArtist :execrows
DELETE FROM artists
WHERE id = ?;

-- name: ListArtists :many
SELECT id, name
FROM artists
ORDER BY name
LIMIT ? OFFSET ?;

-- name: GetArtistByName :one
SELECT * FROM artists
WHERE name = ?;

-- name: GetArtistById :one
SELECT id, name FROM artists
WHERE id = ?;


-------------------------------------------------------------------------


-- name: AddReleaseArtist :one
INSERT INTO release_artists
(release_id, artist_id)
VALUES (?, ?)
RETURNING *;

-- name: GetReleaseArtistsByReleaseId :many
SELECT ar.id as artist_id, ar.name, ra.release_id
FROM release_artists ra
JOIN artists ar ON ra.artist_id = ar.id
WHERE ra.release_id = ?
ORDER BY ar.name;

-- name: RemoveReleaseArtist :execrows
DELETE FROM release_artists
WHERE release_id = ? AND artist_id = ?;


-------------------------------------------------------------------------


-- name: AddTrackArtist :one
INSERT INTO track_artists
(track_id, artist_id, role)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetTrackArtistsByName :one
SELECT * FROM track_artists as ta
JOIN artists AS a ON ta.artist_id = a.id
WHERE a.name = ?;

-- name: UpdateTrackArtistRole :execresult
UPDATE track_artists
SET role = ?
WHERE track_id = ? AND artist_id = ?;

-- name: GetTrackArtistsByTrackId :many
SELECT ar.id, ar.name, ta.role
FROM track_artists ta
JOIN artists AS ar ON ta.artist_id = ar.id
WHERE ta.track_id = ?
ORDER BY
    CASE ta.role
        WHEN 'primary' THEN 1
        WHEN 'featured' THEN 2
    END, ar.name;

-- name: RemoveTrackArtist :execrows
DELETE FROM track_artists
WHERE track_id = ? AND artist_id = ?;

-- name: GetTracksByArtistId :many
SELECT t.id, t.title, r.name AS release_name, ta.role
FROM track_artists ta
JOIN tracks t ON ta.track_id = t.id
JOIN releases r ON t.release_id = r.id
WHERE ta.artist_id = ?
ORDER BY r.date DESC, t.track;

-- name: ListTracksArtistsByRelease :many
SELECT ta.track_id, ar.id AS artist_id, ar.name AS artist_name, ta.role
FROM track_artists AS ta
JOIN artists AS ar ON ta.artist_id = ar.id
JOIN tracks AS tr ON ta.track_id = tr.id
JOIN releases AS re ON tr.release_id = re.id
WHERE re.id = ?
ORDER BY ta.track_id,
    CASE ta.role
        WHEN 'primary'  THEN 1
        WHEN 'featured' THEN 2
    END,
    ar.name;
