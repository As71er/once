package tracks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/metadata"
	"github.com/As71er/once/internal/releases"
	"github.com/As71er/once/internal/utils"
	"github.com/As71er/once/internal/worker"
)

type Service interface {
	GetTrackById(ctx context.Context, id int) (*TrackSummary, error)
	UploadTracks(ctx context.Context, req UploadTracksReq) (IndexedTracksRes, error)
	ListTracksRelease(ctx context.Context, id int) (ListTracksRelease, error)
	GetTrack(ctx context.Context, id int) (TrackDesc, error)
	DeleteTrack(ctx context.Context, id int) error
}

type TrackService struct {
	repository        Repository
	releaseRepository releases.Repository
	fileManager       utils.FileManager
	dbWrite           sync.Mutex
}

func NewService(repository Repository, releaseRepository releases.Repository, fileManager utils.FileManager) *TrackService {
	return &TrackService{
		repository:        repository,
		releaseRepository: releaseRepository,
		fileManager:       fileManager,
	}
}

func (ts *TrackService) GetTrackById(ctx context.Context, id int) (*TrackSummary, error) {
	track, err := ts.repository.GetById(ctx, int64(id))
	if err != nil {
		return nil, err
	}

	return &TrackSummary{
		ID:          track.ID,
		Title:       track.Title,
		Duration:    track.Duration,
		Track:       track.Track,
		Disc:        track.Disc,
		Composer:    track.Composer,
		AudioFileID: &track.AudioFile.ID,
		ReleaseID:   &track.Release.ID,
	}, nil
}

// TODO: refactor to a use case
func (ts *TrackService) UploadTracks(ctx context.Context, req UploadTracksReq) (IndexedTracksRes, error) {
	if len(req.Uploaded) == 0 {
		return IndexedTracksRes{}, fmt.Errorf("no files uploaded: %w", apperrors.ErrInvalidInput)
	}

	release, err := ts.releaseRepository.GetReleaseCover(ctx, req.ReleaseID)
	if err != nil {
		return IndexedTracksRes{}, err
	}

	if req.CoverIdx < 0 || req.CoverIdx >= len(req.Uploaded) {
		return IndexedTracksRes{}, fmt.Errorf("cover index %d out of range: %w", req.CoverIdx, apperrors.ErrInvalidInput)
	}

	dir := ts.fileManager.GetUploadDir(release.Name)

	// From interface
	process := func(ctx context.Context, file UploadTrack) (FileOut, error) {
		return ts.ProcessUploadedFile(ctx, dir, release.ID, file)
	}

	inputs := make([]Input, 0, len(req.Uploaded))
	for _, up := range req.Uploaded {
		inputs = append(inputs, Input{
			dir:       dir,
			releaseID: release.ID,
			file:      up,
		})
	}

	var workers = 2
	if len(req.Uploaded) >= 10 {
		workers = 4
	}

	results := worker.Run(ctx, workers, req.Uploaded, process)

	report := IndexedTracks{}
	var sampleAudioFile AudioFile

	for i, res := range results {
		name := inputs[i].file.Name

		switch {
		case res.Err != nil:
			report.Failed = append(report.Failed, IgnoredIndexedTrack{
				File:   name,
				Reason: res.Err.Error(),
			})
		case res.Value.FileStatus == FileSkipped:
			report.Skipped = append(report.Skipped, IgnoredIndexedTrack{
				File:   name,
				Reason: "already exists",
			})
		case res.Value.FileStatus == FileIndexed:
			report.Indexed = append(report.Indexed, res.Value.Track)
			if i == req.CoverIdx {
				sampleAudioFile = res.Value.Track.AudioFile
			}
		default:
			return IndexedTracksRes{}, fmt.Errorf("no results recorded")
		}
	}
	res := IndexedTracksRes{Tracks: report}

	if !req.ExtractCover || sampleAudioFile.Name == "" {
		return res, nil
	}

	coverName := fmt.Sprintf("%s.png", release.Cover.Name)
	coverExist, err := ts.fileManager.FileExists(coverName, release.Name)
	if err != nil {
		return res, err
	}

	if coverExist {
		return res, nil
	}

	path := filepath.Join(dir, sampleAudioFile.Name)

	writer, _, err := ts.fileManager.GetFileWriter(coverName, release.Name)
	if err != nil {
		return res, err
	}
	if err = metadata.ExtractCover(path, writer); err != nil {
		writer.Close()

		res.Warnings = append(res.Warnings, fmt.Sprintf("cover extraction failed: %v", err))
		return res, nil
	}
	writer.Close()

	writerThumb, _, err := ts.fileManager.GetFileWriter(
		fmt.Sprintf("%s.jpg", release.Cover.ThumbName), release.Name)
	if err != nil {
		return res, err
	}
	if err = metadata.ExtractCoverThumbnail(path, writerThumb); err != nil {
		writerThumb.Close()
		res.Warnings = append(res.Warnings, fmt.Sprintf("cover extraction thumbnail failed: %v", err))

		return res, nil
	}
	writerThumb.Close()

	return res, nil
}

func (ts *TrackService) ListTracksRelease(ctx context.Context, id int) (ListTracksRelease, error) {
	release, err := ts.releaseRepository.GetById(ctx, int64(id))
	if err != nil {
		return ListTracksRelease{}, err
	}

	tracks, err := ts.repository.ListTracksRelease(ctx, int64(id))
	if err != nil {
		return ListTracksRelease{}, err
	}

	cover, err := ts.releaseRepository.GetCoverById(ctx, int64(id))
	if err != nil {
		return ListTracksRelease{}, err
	}

	tracksSummary := make([]TrackSummary, 0, len(tracks))

	for _, track := range tracks {
		tracksSummary = append(tracksSummary, TrackSummary{
			ID:          track.ID,
			Title:       track.Title,
			Duration:    track.Duration,
			Track:       track.Track,
			Disc:        track.Disc,
			Composer:    track.Composer,
			AudioFileID: &track.AudioFile.ID,
		})
	}

	return ListTracksRelease{
		ReleaseID: &release.ID,
		CoverID:   cover.ID,
		Tracks:    tracksSummary,
		Count:     len(tracksSummary),
	}, nil
}

func (ts *TrackService) GetTrack(ctx context.Context, id int) (TrackDesc, error) {
	track, err := ts.repository.GetById(ctx, int64(id))
	if err != nil {
		return TrackDesc{}, err
	}

	file, fileInfo, err := ts.fileManager.OpenFile(track.AudioFile.Name, track.Release.Name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return TrackDesc{}, fmt.Errorf("audio file for track %d: %w", id, apperrors.ErrNotFound)
		}
		return TrackDesc{}, fmt.Errorf("open audio file for track %d: %w", id, err)
	}

	return TrackDesc{
		File:    file,
		ModTime: fileInfo.ModTime(),
	}, nil

}

func (ts *TrackService) DeleteTrack(ctx context.Context, id int) error {
	track, err := ts.repository.GetById(ctx, int64(id))
	if err != nil {
		return err
	}

	if err := ts.repository.Delete(ctx, int64(id)); err != nil {
		return err
	}

	if err := ts.repository.DeleteAudioFile(ctx, track.AudioFile.ID); err != nil {
		return err
	}

	if err := ts.fileManager.DeleteFile(track.AudioFile.Name, ts.fileManager.GetUploadDir(track.Release.Name)); err != nil {
		return err
	}

	return nil
}
