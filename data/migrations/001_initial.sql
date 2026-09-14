-- +goose Up
CREATE TABLE artists (
    id      INTEGER PRIMARY KEY,
    name    TEXT NOT NULL UNIQUE
);

CREATE INDEX idx_artist_name ON artists (name);


CREATE TABLE covers (
    id      INTEGER PRIMARY KEY,
    name    TEXT NOT NULL,
    thumb_name TEXT NOT NULL
);

-------------------------------------------------------------------------


CREATE TABLE audio_files (
    id          INTEGER PRIMARY KEY,
    name        TEXT NOT NULL,
    codec       TEXT NOT NULL,
    bit_rate    INTEGER,
    bit_depth   INTEGER,
    channels    INTEGER,
    sample_rate INTEGER NOT NULL,
    size        INTEGER NOT NULL
);


-------------------------------------------------------------------------


CREATE TABLE releases (
    id              INTEGER PRIMARY KEY,
    title           TEXT NOT NULL,
    name            TEXT NOT NULL UNIQUE,
    release_type    TEXT NOT NULL,
    date            TEXT,
    total_tracks    INTEGER NOT NULL DEFAULT 1,
    total_discs     INTEGER NOT NULL DEFAULT 1,
    cover_id        INTEGER,

    FOREIGN KEY (cover_id) REFERENCES covers (id) ON DELETE CASCADE,

    CHECK (release_type IN ('album', 'ep', 'compilation', 'single'))
);


-------------------------------------------------------------------------


CREATE TABLE tracks (
    id              INTEGER PRIMARY KEY,
    title           TEXT NOT NULL,
    duration        REAL NOT NULL,
    track           INTEGER NOT NULL DEFAULT 1,
    disc            INTEGER NOT NULL DEFAULT 1,
    composer        TEXT,
    lyrics          TEXT,
    release_id      INTEGER NOT NULL,
    audio_file_id   INTEGER NOT NULL UNIQUE,

    FOREIGN KEY (release_id) REFERENCES releases (id) ON DELETE RESTRICT,
    FOREIGN KEY (audio_file_id) REFERENCES audio_files (id) ON DELETE RESTRICT
);

CREATE INDEX idx_track_release_id ON tracks(release_id);
CREATE INDEX idx_track_audio_file_id ON tracks(audio_file_id);


-------------------------------------------------------------------------


CREATE TABLE release_artists (
    release_id  INTEGER NOT NULL,
    artist_id   INTEGER NOT NULL,

    PRIMARY KEY (release_id, artist_id),
    FOREIGN KEY (release_id) REFERENCES releases (id),
    FOREIGN KEY (artist_id) REFERENCES artists (id)
);


CREATE TABLE track_artists (
    track_id    INTEGER NOT NULL,
    artist_id   INTEGER NOT NULL,
    role        TEXT NOT NULL,

    PRIMARY KEY (track_id, artist_id),
    FOREIGN KEY (track_id) REFERENCES tracks (id),
    FOREIGN KEY (artist_id) REFERENCES artists (id),

    CHECK (role IN ('primary', 'featured'))
);


-- +goose Down
DROP TABLE track_artists;
DROP TABLE release_artists;
DROP TABLE tracks;
DROP TABLE releases;
DROP TABLE audio_files;
DROP TABLE covers;
DROP TABLE artists;
