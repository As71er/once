package releases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/sqlc"
)

type Repository interface {
	Create(context.Context, Release) (Release, error)
	GetById(context.Context, int64) (Release, error)
	Delete(context.Context, int64) error
	List(ctx context.Context, limit, offset int64) ([]Release, error)
	CreateCover(context.Context, Cover) (Cover, error)
	GetCoverById(context.Context, int64) (Cover, error)
	GetReleaseCover(context.Context, int64) (Release, error)
	UpdateReleaseCover(context.Context, int64, int64) error
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

func (sr *SqlcRepository) Create(ctx context.Context, release Release) (Release, error) {
	var coverID sql.NullInt64
	if release.Cover != nil {
		coverID = sql.NullInt64{Int64: *release.Cover.ID, Valid: true}
	}

	releaseDB, err := sr.q.CreateRelease(ctx, sqlc.CreateReleaseParams{
		Title:       release.Title,
		Name:        release.Name,
		ReleaseType: release.ReleaseType,
		Date:        sql.NullString{String: release.Date, Valid: true},
		TotalTracks: int64(release.TotalTracks),
		TotalDiscs:  int64(release.TotalDiscs),
		CoverID:     coverID,
	})
	if err != nil {
		return Release{}, fmt.Errorf("create release %w", errors.Join(apperrors.ErrCreation, err))
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
	releaseDB, err := sr.q.GetReleaseById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Release{}, fmt.Errorf("get release with id %d: %w", id, errors.Join(apperrors.ErrNotFound, err))
		}
		return Release{}, fmt.Errorf("get release with id %d: %w", id, err)
	}

	foundArtists, err := sr.q.GetReleaseArtistsByReleaseId(ctx, id)
	if err != nil {
		return Release{}, fmt.Errorf("get release artists with id %d: %w", id, err)
	}

	var releaseArtists []ReleaseArtistRef
	for _, artist := range foundArtists {
		releaseArtists = append(releaseArtists, ReleaseArtistRef{
			ID:   artist.ArtistID,
			Name: artist.Name,
		})
	}

	return Release{
		ID:            releaseDB.ID,
		Title:         releaseDB.Title,
		Name:          releaseDB.Name,
		ReleaseType:   releaseDB.ReleaseType,
		Date:          releaseDB.Date.String,
		TotalTracks:   uint(releaseDB.TotalTracks),
		TotalDiscs:    uint(releaseDB.TotalDiscs),
		TotalDuration: uint(releaseDB.TotalDuration.Float64),
		Cover: &Cover{
			ID: &releaseDB.CoverID.Int64,
		},
		Artists: releaseArtists,
	}, nil
}

func (sr *SqlcRepository) Delete(ctx context.Context, id int64) error {
	if err := sr.q.DeleteRelease(ctx, id); err != nil {
		return fmt.Errorf("delete release with id %d: %w", id, errors.Join(apperrors.ErrDeletion, err))

	}

	return nil
}

func (sr *SqlcRepository) List(ctx context.Context, limit, offset int64) ([]Release, error) {
	releasesDB, err := sr.q.ListReleases(ctx, sqlc.ListReleasesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list releases: %w", err)
	}

	var releases []Release
	for _, release := range releasesDB {
		releases = append(releases, Release{
			ID:            release.ID,
			Title:         release.Title,
			ReleaseType:   release.ReleaseType,
			Date:          release.Date.String,
			TotalTracks:   uint(release.TotalTracks),
			TotalDiscs:    uint(release.TotalDiscs),
			TotalDuration: uint(release.TotalDuration.Float64),
			Cover: &Cover{
				ID: &release.CoverID.Int64,
			},
		})
	}

	return releases, nil
}

func (sr *SqlcRepository) CreateCover(ctx context.Context, cover Cover) (Cover, error) {
	coverDB, err := sr.q.CreateCover(ctx, sqlc.CreateCoverParams{
		Name:      cover.Name,
		ThumbName: cover.ThumbName,
	})
	if err != nil {
		return Cover{}, fmt.Errorf("create cover: %w", errors.Join(apperrors.ErrCreation, err))
	}

	return Cover{
		ID:        &coverDB.ID,
		Name:      coverDB.Name,
		ThumbName: cover.ThumbName,
	}, nil
}

func (sr *SqlcRepository) GetCoverById(ctx context.Context, id int64) (Cover, error) {
	coverDB, err := sr.q.GetCoverById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Cover{}, fmt.Errorf("get cover with id %d: %w", id, errors.Join(apperrors.ErrNotFound, err))
		}
		return Cover{}, fmt.Errorf("get cover with id %d: %w", id, err)
	}

	return Cover{
		ID:        &coverDB.ID,
		Name:      coverDB.Name,
		ThumbName: coverDB.ThumbName,
	}, nil
}

func (sr *SqlcRepository) GetReleaseCover(ctx context.Context, id int64) (Release, error) {
	release, err := sr.q.GetReleaseCover(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Release{}, fmt.Errorf("get cover for release with id %d: %w", id, errors.Join(apperrors.ErrNotFound, err))
		}
		return Release{}, fmt.Errorf("get cover for release with id %d: %w", id, err)
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
	if err := sr.q.UpdateReleaseCover(ctx, sqlc.UpdateReleaseCoverParams{
		CoverID: sql.NullInt64{Int64: coverID, Valid: true},
		ID:      releaseID,
	}); err != nil {
		return fmt.Errorf("update cover for release with id %d: %w", releaseID, err)
	}

	return nil
}
