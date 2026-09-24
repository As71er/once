package artists

import (
	"context"
	"database/sql"

	"github.com/As71er/once/internal/sqlc"
)

type Repository interface {
	CreateArtist(ctx context.Context, name string) (Artist, error)
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

func (sr *SqlcRepository) CreateArtist(ctx context.Context, name string) (Artist, error) {
	artist, err := sr.q.CreateArtist(ctx, name)
	if err != nil {
		return Artist{}, err
	}

	return Artist{
		ID:   artist.ID,
		Name: artist.Name,
	}, nil
}
