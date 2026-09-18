package tracks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/releases"
	"github.com/As71er/once/internal/sqlc"
)

type Repository interface {
	Create(ctx context.Context, track Track) (Track, error)
	GetById(ctx context.Context, id int64) (Track, error)
	Delete(ctx context.Context, id int64) error
	DeleteAudioFile(ctx context.Context, id int64) error
	CreateAudioFile(ctx context.Context, audio_file AudioFile) (AudioFile, error)
	GetAudioFile(ctx context.Context, id int64) (AudioFile, error)
	ListTracksRelease(ctx context.Context, id int64) ([]Track, error)
}

type SqlcRepository struct {
	db *sqlc.Queries
}

func NewRepository(db *sqlc.Queries) *SqlcRepository {
	return &SqlcRepository{
		db: db,
	}
}

func (sr *SqlcRepository) Create(ctx context.Context, track Track) (Track, error) {
	trackDb, err := sr.db.CreateTrack(ctx, sqlc.CreateTrackParams{
		Title:       track.Title,
		Duration:    track.Duration,
		Track:       track.Track,
		Disc:        track.Disc,
		Composer:    sql.NullString{String: track.Composer, Valid: true},
		Lyrics:      sql.NullString{String: track.Lyrics, Valid: true},
		ReleaseID:   track.Release.ID,
		AudioFileID: track.AudioFile.ID,
	})
	if err != nil {
		return Track{}, fmt.Errorf("create track %w", errors.Join(apperrors.ErrCreation, err))
	}

	return Track{
		ID:       trackDb.ID,
		Title:    track.Title,
		Duration: track.Duration,
		Track:    track.Track,
		Disc:     track.Disc,
		Composer: track.Composer,
		Lyrics:   track.Lyrics,
		Release: releases.Release{
			ID: track.Release.ID,
		},
		AudioFile: AudioFile{
			ID: trackDb.AudioFileID,
		},
	}, nil
}

func (sr *SqlcRepository) GetById(ctx context.Context, id int64) (Track, error) {
	track, err := sr.db.GetTrackById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Track{}, fmt.Errorf("get track with id %d: %w", id, errors.Join(apperrors.ErrNotFound, err))
		}
		return Track{}, fmt.Errorf("get track with id %d: %w", id, err)
	}

	return Track{
		ID:       track.ID,
		Title:    track.Title,
		Duration: track.Duration,
		Track:    track.Track,
		Disc:     track.Disc,
		Composer: track.Composer.String,
		Release: releases.Release{
			ID:   id,
			Name: track.ReleaseName,
		},
		AudioFile: AudioFile{
			ID:   track.ID,
			Name: track.AudioFileName,
		},
	}, nil
}

func (sr *SqlcRepository) Delete(ctx context.Context, id int64) error {
	if err := sr.db.DeleteTrack(ctx, id); err != nil {
		return fmt.Errorf("delete track with id %d: %w", id, errors.Join(apperrors.ErrDeletion, err))
	}

	return nil
}

func (sr *SqlcRepository) DeleteAudioFile(ctx context.Context, id int64) error {
	if err := sr.db.DeleteAudioFile(ctx, id); err != nil {
		return fmt.Errorf("delete audio file with id %d: %w", id, errors.Join(apperrors.ErrDeletion, err))
	}

	return nil
}

func (sr *SqlcRepository) CreateAudioFile(ctx context.Context, audio_file AudioFile) (AudioFile, error) {
	audioFileDb, err := sr.db.CreateAudioFile(ctx, sqlc.CreateAudioFileParams{
		Name:       audio_file.Name,
		Codec:      audio_file.Codec,
		BitRate:    sql.NullInt64{Int64: audio_file.BitRate, Valid: true},
		BitDepth:   sql.NullInt64{Int64: audio_file.BitDepth, Valid: true},
		Channels:   sql.NullInt64{Int64: audio_file.Channels, Valid: true},
		SampleRate: audio_file.SampleRate,
		Size:       audio_file.Size,
	})
	if err != nil {
		return AudioFile{}, fmt.Errorf("create audio file: %w", errors.Join(apperrors.ErrCreation, err))
	}

	return AudioFile{
		ID:         audioFileDb.ID,
		Name:       audioFileDb.Name,
		Codec:      audio_file.Codec,
		BitRate:    audioFileDb.BitRate.Int64,
		BitDepth:   audioFileDb.BitDepth.Int64,
		Channels:   audioFileDb.Channels.Int64,
		SampleRate: audioFileDb.SampleRate,
		Size:       audioFileDb.Size,
	}, nil
}

func (sr *SqlcRepository) GetAudioFile(ctx context.Context, id int64) (AudioFile, error) {
	audio_file, err := sr.db.GetAudioFileById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AudioFile{}, fmt.Errorf("get audio file with id %d: %w", id, errors.Join(apperrors.ErrNotFound, err))
		}
		return AudioFile{}, fmt.Errorf("get audio file with id %d: %w", id, err)
	}

	return AudioFile{
		ID:         audio_file.ID,
		Name:       audio_file.Name,
		Codec:      audio_file.Codec,
		BitRate:    audio_file.BitRate.Int64,
		BitDepth:   audio_file.BitDepth.Int64,
		Channels:   audio_file.Channels.Int64,
		SampleRate: audio_file.SampleRate,
		Size:       audio_file.Size,
	}, nil
}

func (sr *SqlcRepository) ListTracksRelease(ctx context.Context, id int64) ([]Track, error) {
	tracksDB, err := sr.db.ListTracksRelease(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list tracks for release with id %d: %w", id, err)
	}

	var tracks []Track
	for _, track := range tracksDB {
		tracks = append(tracks, Track{
			ID:       track.ID,
			Title:    track.Title,
			Duration: track.Duration,
			Track:    track.Track,
			Disc:     track.Disc,
			Composer: track.Composer.String,
			Release: releases.Release{
				ID: id,
			},
			AudioFile: AudioFile{
				ID: track.ID,
			},
		})
	}

	return tracks, nil
}
