package releaseartists

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/sqlc"
)

type Repository interface {
	Add(ctx context.Context, releaseID, artistID int64) (ReleaseArtist, error)
	ListByReleaseId(context.Context, int64) (map[int64]ArtistRef, error)
	Remove(ctx context.Context, releaseID, artistID int64) error
}

type SqlcRepository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewRepository(db *sql.DB) *SqlcRepository {
	return &SqlcRepository{
		db,
		sqlc.New(db),
	}
}

func (sr *SqlcRepository) Add(ctx context.Context, releaseID, artistID int64) (ReleaseArtist, error) {
	releaseArtist, err := sr.q.AddReleaseArtist(ctx, sqlc.AddReleaseArtistParams{
		ReleaseID: releaseID,
		ArtistID:  artistID,
	})
	if err != nil {
		return ReleaseArtist{}, fmt.Errorf("add artist id %d to release id %d: %w", artistID, releaseID, err)
	}

	return ReleaseArtist{
		ReleaseID: releaseArtist.ReleaseID,
		ArtistID:  releaseArtist.ArtistID,
	}, nil
}

func (sr *SqlcRepository) Remove(ctx context.Context, releaseID, artistID int64) error {
	rowsDeleted, err := sr.q.RemoveReleaseArtist(ctx, sqlc.RemoveReleaseArtistParams{
		ReleaseID: releaseID,
		ArtistID:  artistID,
	})
	if err != nil {
		return fmt.Errorf("delete artist id %d from release id %d: %w", artistID, releaseID, err)
	}

	if rowsDeleted < 1 {
		return fmt.Errorf("delete artist id %d from release id %d: %w", artistID, releaseID, errors.Join(apperrors.ErrNotFound, err))
	}

	return nil
}

func (sr *SqlcRepository) ListByReleaseId(ctx context.Context, id int64) (map[int64]ArtistRef, error) {
	foundArtists, err := sr.q.GetReleaseArtistsByReleaseId(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make(map[int64]ArtistRef), fmt.Errorf("get artist by release id %d: %w", id, errors.Join(apperrors.ErrNotFound, err))
		}
		return make(map[int64]ArtistRef), fmt.Errorf("get artist by release id %d: %w", id, err)
	}

	trackArtists := make(map[int64]ArtistRef)
	for _, artist := range foundArtists {
		trackArtists[artist.ArtistID] = ArtistRef{
			ID:   artist.ArtistID,
			Name: artist.Name,
		}
	}

	return trackArtists, nil
}
