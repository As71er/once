package tracks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/releases"
	"github.com/As71er/once/internal/utils"
)

type Service interface {
	GetTrackById(context.Context, int) (*TrackSummary, error)
	ListTracksRelease(context.Context, int) (ListTracksReleaseRes, error)
	GetTrack(context.Context, int) (TrackDesc, error)
	DeleteTrack(context.Context, int) error
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
		Artists:     track.TrackArtist,
		AudioFileID: &track.AudioFile.ID,
		ReleaseID:   &track.Release.ID,
	}, nil
}

func (ts *TrackService) ListTracksRelease(ctx context.Context, id int) (ListTracksReleaseRes, error) {
	release, err := ts.releaseRepository.GetById(ctx, int64(id))
	if err != nil {
		return ListTracksReleaseRes{}, err
	}

	tracks, err := ts.repository.ListTracksRelease(ctx, int64(id))
	if err != nil {
		return ListTracksReleaseRes{}, err
	}

	cover, err := ts.releaseRepository.GetCoverById(ctx, int64(id))
	if err != nil {
		return ListTracksReleaseRes{}, err
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
			Artists:     track.TrackArtist,
		})
	}

	return ListTracksReleaseRes{
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
