package artists

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/sqlc"
)

type Repository interface {
	Create(context.Context, string) (Artist, error)
	GetById(context.Context, int64) (Artist, error)
	GetByName(context.Context, string) (Artist, error)
	Update(context.Context, int64, string) (Artist, error)
	Delete(context.Context, int64) error
	List(ctx context.Context, limit, offset int64) ([]Artist, error)
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

func (sr *SqlcRepository) Create(ctx context.Context, name string) (Artist, error) {
	artist, err := sr.q.CreateArtist(ctx, name)
	if err != nil {
		return Artist{}, fmt.Errorf("create artist %w", errors.Join(apperrors.ErrCreation, err))
	}

	return Artist{
		ID:   artist.ID,
		Name: artist.Name,
	}, nil
}

func (sr *SqlcRepository) GetByName(ctx context.Context, name string) (Artist, error) {
	artist, err := sr.q.GetArtistByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Artist{}, fmt.Errorf("get artist by name %s: %w", name, errors.Join(apperrors.ErrNotFound, err))
		}
		return Artist{}, fmt.Errorf("get artist by name %s: %w", name, err)
	}

	return Artist{
		ID:   artist.ID,
		Name: artist.Name,
	}, nil
}

func (sr *SqlcRepository) GetById(ctx context.Context, id int64) (Artist, error) {
	artist, err := sr.q.GetArtistById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Artist{}, fmt.Errorf("get artist by id %d: %w", id, errors.Join(apperrors.ErrNotFound, err))
		}
		return Artist{}, fmt.Errorf("get artist by id %d: %w", id, err)
	}

	return Artist{
		ID:   artist.ID,
		Name: artist.Name,
	}, nil
}

func (sr *SqlcRepository) Update(ctx context.Context, id int64, name string) (Artist, error) {
	artist, err := sr.q.UpdateArtist(ctx, sqlc.UpdateArtistParams{
		Name: name,
		ID:   id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Artist{}, fmt.Errorf("update artist id %d: %w", id, errors.Join(apperrors.ErrNotFound, err))
		}
		return Artist{}, fmt.Errorf("update artist id %d: %w", id, err)
	}

	return Artist{
		ID:   artist.ID,
		Name: artist.Name,
	}, nil
}

func (sr *SqlcRepository) Delete(ctx context.Context, id int64) error {
	rowsDeleted, err := sr.q.DeleteArtist(ctx, id)
	if err != nil {
		return fmt.Errorf("delete artist id %d: %w", id, errors.Join(apperrors.ErrDeletion, err))
	}

	if rowsDeleted < 1 {
		return fmt.Errorf("delete artist id %d: %w", id, errors.Join(apperrors.ErrNotFound, err))
	}

	return nil
}

func (sr *SqlcRepository) List(ctx context.Context, limit, offset int64) ([]Artist, error) {
	artists, err := sr.q.ListArtists(ctx, sqlc.ListArtistsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Artist{}, fmt.Errorf("list artists: %w", errors.Join(apperrors.ErrNotFound, err))
		}
		return []Artist{}, fmt.Errorf("list artists: %w", err)
	}

	var result []Artist
	for _, artist := range artists {
		result = append(result, Artist{
			ID:   artist.ID,
			Name: artist.Name,
		})
	}

	return result, nil
}
