package releases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/As71er/once/internal/artists"
	"github.com/As71er/once/internal/utils"
)

type Service interface {
	CreateRelease(ctx context.Context, release CreateReleaseReq) (*CreatedReleaseRes, error)
	GetReleaseById(ctx context.Context, id int) (*ReleaseRes, error)
	GetCover(ctx context.Context, id int) (*CoverArtDesc, error)
	GetCoverThumb(ctx context.Context, id int) (*CoverArtDesc, error)
	ListReleases(ctx context.Context, limit, offset int) (ReleaseListRes, error)
	DeleteRelease(ctx context.Context, id int) error
}

type ReleaseService struct {
	repository       Repository
	artistRepository artists.Repository
	fileManager      utils.FileManager
}

func NewService(repository Repository, artistRepository artists.Repository, fileManager utils.FileManager) *ReleaseService {
	return &ReleaseService{
		repository,
		artistRepository,
		fileManager,
	}
}

func (as *ReleaseService) CreateRelease(ctx context.Context, req CreateReleaseReq) (*CreatedReleaseRes, error) {
	if !ReleaseType(req.ReleaseType).Valid() {
		msg := fmt.Sprintf("Invalid release type. Valid types: %s, %s, %s, %s",
			ReleaseTypeAlbum, ReleaseTypeEP, ReleaseTypeCompilation, ReleaseTypeSingle)
		return nil, errors.New(msg)
	}

	releaseTitle := req.Title
	if releaseTitle == "" {
		releaseTitle = "default"
	}
	dirName := utils.GenerateChecksum(releaseTitle)
	coverName := utils.GenerateChecksumShort(releaseTitle)

	createdCover, err := as.repository.CreateCover(ctx, Cover{
		Name:      coverName,
		ThumbName: fmt.Sprintf("tb-%s", coverName),
	})
	if err != nil {
		return nil, err
	}

	createdRelease, err := as.repository.Create(ctx, Release{
		Title:       req.Title,
		Name:        dirName,
		ReleaseType: req.ReleaseType,
		Date:        req.Date,
		TotalTracks: req.TotalTracks,
		TotalDiscs:  req.TotalDiscs,
		Cover: &Cover{
			ID: createdCover.ID,
		},
	})
	if err != nil {
		return nil, err
	}

	_, err = as.fileManager.CreateDir(dirName)
	if err != nil {
		slog.Error("Failed at processing album", slog.Any("error", err))
	}

	return &CreatedReleaseRes{
		ID:          createdRelease.ID,
		Title:       createdRelease.Title,
		ReleaseType: createdRelease.ReleaseType,
		Date:        createdRelease.Date,
		TotalTracks: createdRelease.TotalTracks,
		TotalDiscs:  createdRelease.TotalDiscs,
		CoverID:     *createdCover.ID,
	}, nil
}

func (s *ReleaseService) GetReleaseById(ctx context.Context, id int) (*ReleaseRes, error) {
	release, err := s.repository.GetById(ctx, int64(id))
	if err != nil {
		return nil, err
	}

	return &ReleaseRes{
		ID:          release.ID,
		Title:       release.Title,
		ReleaseType: release.ReleaseType,
		Date:        release.Date,
		TotalTracks: release.TotalTracks,
		TotalDiscs:  release.TotalDiscs,
		CoverID:     *release.Cover.ID,
	}, nil
}

func (s *ReleaseService) ListReleases(ctx context.Context, limit, offset int) (ReleaseListRes, error) {
	releases, err := s.repository.List(ctx, int64(limit), int64(offset))
	if err != nil {
		return ReleaseListRes{}, err
	}

	var releasesRes []ReleaseRes
	for _, release := range releases {

		releasesRes = append(releasesRes, ReleaseRes{
			ID:          release.ID,
			Title:       release.Title,
			ReleaseType: release.ReleaseType,
			Date:        release.Date,
			TotalTracks: release.TotalTracks,
			TotalDiscs:  release.TotalDiscs,
			CoverID:     *release.Cover.ID,
		})
	}

	return ReleaseListRes{
		Releases: releasesRes,
		Count:    len(releases),
	}, nil
}

func (s *ReleaseService) GetCover(ctx context.Context, id int) (*CoverArtDesc, error) {
	release, err := s.repository.GetReleaseCover(ctx, int64(id))
	if err != nil {
		return nil, err
	}

	file, fileInfo, err := s.fileManager.OpenFile(fmt.Sprintf("%s.png", release.Cover.Name), release.Name)
	if err != nil {
		return nil, err
	}

	coverArt := CoverArtDesc{
		File:    file,
		ModTime: fileInfo.ModTime(),
	}

	return &coverArt, nil
}

func (s *ReleaseService) GetCoverThumb(ctx context.Context, id int) (*CoverArtDesc, error) {
	release, err := s.repository.GetReleaseCover(ctx, int64(id))
	if err != nil {
		return nil, err
	}

	file, fileInfo, err := s.fileManager.OpenFile(fmt.Sprintf("%s.jpg", release.Cover.ThumbName), release.Name)
	if err != nil {
		return nil, err
	}

	coverArt := CoverArtDesc{
		File:    file,
		ModTime: fileInfo.ModTime(),
	}

	return &coverArt, nil
}

func (r *ReleaseService) DeleteRelease(ctx context.Context, id int) error {
	release, err := r.repository.GetById(ctx, int64(id))
	if err != nil {
		return nil
	}

	if err := r.repository.Delete(ctx, int64(id)); err != nil {
		return nil
	}

	if err := r.fileManager.DeleteDirRecursive(release.Name); err != nil {
		return nil
	}

	return nil
}
