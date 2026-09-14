-- name: CreateTrack :one
INSERT INTO tracks
(title, duration, track, disc, composer, lyrics, release_id, audio_file_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetTrackById :one
SELECT t.id, t.title, t.duration, t.track, t.disc, t.composer, t.lyrics,
       r.id AS release_id, r.title AS release_title, r.name AS release_name,
       af.id AS audio_file_id, af.name AS audio_file_name, af.codec, af.bit_rate, af.bit_depth,
       af.channels, af.sample_rate, af.size
FROM tracks AS t
JOIN releases AS r ON t.release_id = r.id
JOIN audio_files AS af ON t.audio_file_id = af.id
WHERE t.id = ?;

-- name: DeleteTrack :exec
DELETE FROM tracks
WHERE id = ?;

-- name: UpdateTrack :one
UPDATE tracks
SET title = ?, duration = ?, track = ?, disc = ?, composer = ?, lyrics = ?
WHERE id = ?
RETURNING *;
