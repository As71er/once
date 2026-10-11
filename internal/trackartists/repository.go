package trackartists

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/sqlc"
)

type Repository interface {
	GetByName(context.Context, string) (TrackArtist, error)
	ListByTrackId(context.Context, int64) (map[int64]TrackArtist, error)
	Add(ctx context.Context, trackID, artistID int64, role string) (TrackArtist, error)
	Remove(ctx context.Context, trackID, artistID int64) error
	Update(ctx context.Context, trackID, artistID int64, role string) error
}

type SqlcRepository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewRepository(db *sql.DB) *SqlcRepository {
	return &SqlcRepository{
		db: db,
		q:  sqlc.New(db),
	}
}

func (sr *SqlcRepository) GetByName(ctx context.Context, name string) (TrackArtist, error) {
	artist, err := sr.q.GetTrackArtistsByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TrackArtist{}, fmt.Errorf("get artist by name %s: %w", name, errors.Join(apperrors.ErrNotFound, err))
		}
		return TrackArtist{}, fmt.Errorf("get artist by name %s: %w", name, err)
	}

	return TrackArtist{
		ArtistID:   artist.ID,
		ArtistName: artist.Name,
		Role:       ArtistRoles(artist.Role),
	}, nil
}

func (sr *SqlcRepository) ListByTrackId(ctx context.Context, id int64) (map[int64]TrackArtist, error) {
	artists, err := sr.q.GetTrackArtistsByTrackId(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make(map[int64]TrackArtist), fmt.Errorf("get artist by track id %d: %w", id, errors.Join(apperrors.ErrNotFound, err))
		}
		return make(map[int64]TrackArtist), fmt.Errorf("get artist by track id %d: %w", id, err)
	}

	trackArtists := make(map[int64]TrackArtist)
	for _, artist := range artists {
		trackArtists[artist.ID] = TrackArtist{
			ArtistID:   artist.ID,
			ArtistName: artist.Name,
			Role:       ArtistRoles(artist.Role),
		}
	}

	return trackArtists, nil
}

func (sr *SqlcRepository) Add(ctx context.Context, trackID, artistID int64, role string) (TrackArtist, error) {
	trackArtist, err := sr.q.AddTrackArtist(ctx, sqlc.AddTrackArtistParams{
		TrackID:  trackID,
		ArtistID: artistID,
		Role:     role,
	})
	if err != nil {
		return TrackArtist{}, fmt.Errorf("add artist id %d to track id %d: %w", artistID, trackID, err)
	}

	artist, err := sr.q.GetArtistById(ctx, trackArtist.ArtistID)
	if err != nil {
		return TrackArtist{}, fmt.Errorf("add artist id %d to track id %d: %w", artistID, trackID, err)
	}

	return TrackArtist{
		TrackID:    trackArtist.TrackID,
		ArtistID:   trackArtist.ArtistID,
		ArtistName: artist.Name,
		Role:       ArtistRoles(trackArtist.Role),
	}, nil
}

func (sr *SqlcRepository) Update(ctx context.Context, trackID, artistID int64, role string) error {
	// For now it makes more sense to just update the role or just delete the artist, it's a single field
	result, err := sr.q.UpdateTrackArtistRole(ctx, sqlc.UpdateTrackArtistRoleParams{
		Role:     role,
		TrackID:  trackID,
		ArtistID: artistID,
	})
	if err != nil {
		return fmt.Errorf("update artist id %d for track id %d: %w", artistID, trackID, err)
	}

	numAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update artist id %d for track id %d: %w", artistID, trackID, err)
	}

	if numAffected < 1 {
		return fmt.Errorf("update artist id %d for track id %d: %w", artistID, trackID, errors.Join(apperrors.ErrNotFound, err))
	}

	return nil
}

func (sr *SqlcRepository) Remove(ctx context.Context, trackID, artistID int64) error {
	rowsDeleted, err := sr.q.RemoveTrackArtist(ctx, sqlc.RemoveTrackArtistParams{
		TrackID:  trackID,
		ArtistID: artistID,
	})
	if err != nil {
		return fmt.Errorf("delete artist id %d from track id %d: %w", artistID, trackID, err)
	}

	if rowsDeleted < 1 {
		return fmt.Errorf("delete artist id %d from track id %d: %w", artistID, trackID, errors.Join(apperrors.ErrNotFound, err))
	}

	return nil
}
