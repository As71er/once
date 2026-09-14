package releases

import (
	"context"
	"database/sql"

	"github.com/As71er/once/internal/sqlc"
)

type Repository interface {
	Create(ctx context.Context, release Release) (Release, error)
	GetById(ctx context.Context, id int64) (Release, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int64) ([]Release, error)
	GetDuration(ctx context.Context, id int64) (float64, error)
	CreateCover(ctx context.Context, cover Cover) (Cover, error)
	GetCoverById(ctx context.Context, id int64) (Cover, error)
	GetReleaseCover(ctx context.Context, id int64) (Release, error)
	UpdateReleaseCover(ctx context.Context, releaseID, coverID int64) error
}

type SqlcRepository struct {
	db *sqlc.Queries
}

func NewRepository(db *sqlc.Queries) *SqlcRepository {
	return &SqlcRepository{
		db: db,
	}
}

func (sr *SqlcRepository) Create(ctx context.Context, release Release) (Release, error) {
	var coverID sql.NullInt64
	if release.Cover != nil {
		coverID = sql.NullInt64{Int64: *release.Cover.ID, Valid: true}
	}

	releaseDB, err := sr.db.CreateRelease(ctx, sqlc.CreateReleaseParams{
		Title:       release.Title,
		Name:        release.Name,
		ReleaseType: release.ReleaseType,
		Date:        sql.NullString{String: release.Date, Valid: true},
		TotalTracks: int64(release.TotalTracks),
		TotalDiscs:  int64(release.TotalDiscs),
		CoverID:     coverID,
	})
	if err != nil {
		return Release{}, err
	}

	return Release{
		ID:          releaseDB.ID,
		Title:       releaseDB.Title,
		Name:        releaseDB.Name,
		ReleaseType: releaseDB.ReleaseType,
		Date:        releaseDB.Date.String,
		TotalTracks: uint(releaseDB.TotalTracks),
		TotalDiscs:  uint(releaseDB.TotalDiscs),
		Cover: &Cover{
			ID: &releaseDB.CoverID.Int64,
		},
	}, nil
}

func (sr *SqlcRepository) GetById(ctx context.Context, id int64) (Release, error) {
	releaseDB, err := sr.db.GetReleaseById(ctx, id)
	if err != nil {
		return Release{}, err
	}

	return Release{
		ID:          releaseDB.ID,
		Title:       releaseDB.Title,
		Name:        releaseDB.Name,
		ReleaseType: releaseDB.ReleaseType,
		Date:        releaseDB.Date.String,
		TotalTracks: uint(releaseDB.TotalTracks),
		TotalDiscs:  uint(releaseDB.TotalDiscs),
		Cover: &Cover{
			ID: &releaseDB.CoverID.Int64,
		},
	}, nil
}

func (sr *SqlcRepository) Delete(ctx context.Context, id int64) error {
	if err := sr.db.DeleteRelease(ctx, id); err != nil {
		return err
	}

	return nil
}

func (sr *SqlcRepository) List(ctx context.Context, limit, offset int64) ([]Release, error) {
	releasesDB, err := sr.db.ListReleases(ctx, sqlc.ListReleasesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	var releases []Release
	for _, release := range releasesDB {
		releases = append(releases, Release{
			ID:          release.ID,
			Title:       release.Title,
			ReleaseType: release.ReleaseType,
			Date:        release.Date.String,
			TotalTracks: uint(release.TotalTracks),
			TotalDiscs:  uint(release.TotalDiscs),
			Cover: &Cover{
				ID: &release.CoverID.Int64,
			},
		})
	}

	return releases, nil
}

func (sr *SqlcRepository) GetDuration(ctx context.Context, id int64) (float64, error) {
	duration, err := sr.db.GetReleaseDuration(ctx, id)
	if err != nil {
		return 0, err
	}

	return duration.Float64, nil
}

func (sr *SqlcRepository) CreateCover(ctx context.Context, cover Cover) (Cover, error) {
	coverDB, err := sr.db.CreateCover(ctx, sqlc.CreateCoverParams{
		Name:      cover.Name,
		ThumbName: cover.ThumbName,
	})
	if err != nil {
		return Cover{}, err
	}

	return Cover{
		ID:        &coverDB.ID,
		Name:      coverDB.Name,
		ThumbName: cover.ThumbName,
	}, nil
}

func (sr *SqlcRepository) GetCoverById(ctx context.Context, id int64) (Cover, error) {
	coverDB, err := sr.db.GetCoverById(ctx, id)
	if err != nil {
		return Cover{}, err
	}

	return Cover{
		ID:        &coverDB.ID,
		Name:      coverDB.Name,
		ThumbName: coverDB.ThumbName,
	}, nil
}

func (sr *SqlcRepository) GetReleaseCover(ctx context.Context, id int64) (Release, error) {
	release, err := sr.db.GetReleaseCover(ctx, id)
	if err != nil {
		return Release{}, err
	}

	return Release{
		ID:   release.ID,
		Name: release.Name,
		Cover: &Cover{
			ID:        &release.ID_2,
			Name:      release.Name_2,
			ThumbName: release.ThumbName,
		}}, nil
}

func (sr *SqlcRepository) UpdateReleaseCover(ctx context.Context, releaseID, coverID int64) error {
	if err := sr.db.UpdateReleaseCover(ctx, sqlc.UpdateReleaseCoverParams{
		CoverID: sql.NullInt64{Int64: coverID, Valid: true},
		ID:      releaseID,
	}); err != nil {
		return err
	}

	return nil
}
