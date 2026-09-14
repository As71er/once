-- name: CreateAudioFile :one
INSERT INTO audio_files
(name, codec, bit_rate, bit_depth, channels, sample_rate, size)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetAudioFileById :one
SELECT * FROM audio_files
WHERE id = ?;

-- name: DeleteAudioFile :exec
DELETE FROM audio_files
WHERE id = ?;
