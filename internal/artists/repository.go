package artists

import (
	"context"

	"github.com/As71er/once/internal/sqlc"
)

type Repository interface {
	CreateArtist(ctx context.Context, name string) (Artist, error)
}

type SqlcRepository struct {
	db *sqlc.Queries
}

func NewRepository(db *sqlc.Queries) *SqlcRepository {
	return &SqlcRepository{
		db: db,
	}
}

func (sr *SqlcRepository) CreateArtist(ctx context.Context, name string) (Artist, error) {
	artistDb, err := sr.db.CreateArtist(ctx, name)
	if err != nil {
		return Artist{}, err
	}

	return Artist{
		ID:   artistDb.ID,
		Name: artistDb.Name,
	}, nil
}
