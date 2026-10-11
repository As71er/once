package usecases

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/artists"
	"github.com/As71er/once/internal/metadata"
	"github.com/As71er/once/internal/releases"
	"github.com/As71er/once/internal/trackartists"
	"github.com/As71er/once/internal/tracks"
	"github.com/As71er/once/internal/utils"
	"github.com/As71er/once/internal/worker"
)

type UploadTracks struct {
	processor             *tracks.UploadTrackProcessor
	releaseRepository     releases.Repository
	artistRepository      artists.Repository
	trackArtistRepository trackartists.Repository
	fileManager           utils.FileManager
}

func NewUploadTracks(
	processor *tracks.UploadTrackProcessor,
	releaseRepository releases.Repository,
	artistRepository artists.Repository,
	trackArtistRepository trackartists.Repository,
	fileManager utils.FileManager,
) UploadTracks {
	return UploadTracks{
		processor,
		releaseRepository,
		artistRepository,
		trackArtistRepository,
		fileManager,
	}
}

func (uc *UploadTracks) Execute(ctx context.Context, req tracks.UploadTracksReq) (tracks.IndexedTracksRes, error) {
	if len(req.Uploaded) == 0 {
		return tracks.IndexedTracksRes{}, fmt.Errorf("no files uploaded: %w", apperrors.ErrInvalidInput)
	}

	release, err := uc.releaseRepository.GetReleaseCover(ctx, req.ReleaseID)
	if err != nil {
		return tracks.IndexedTracksRes{}, err
	}

	if req.CoverIdx < 0 || req.CoverIdx >= len(req.Uploaded) {
		return tracks.IndexedTracksRes{}, fmt.Errorf("cover index %d out of range: %w", req.CoverIdx, apperrors.ErrInvalidInput)
	}

	dir := uc.fileManager.GetUploadDir(release.Name)

	report, sampleAudioFile, err := uc.runWorkers(ctx, req, release, dir)
	if err != nil {
		return tracks.IndexedTracksRes{}, err
	}

	res := tracks.IndexedTracksRes{Tracks: report}

	err = uc.checkArtists(ctx, report)
	if err != nil {
		return res, err
	}

	if !req.ExtractCover || sampleAudioFile.Name == "" {
		return res, nil
	}

	if err = uc.extractCover(release, dir, sampleAudioFile, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (uc *UploadTracks) runWorkers(ctx context.Context, req tracks.UploadTracksReq,
	release releases.Release, dir string) (tracks.IndexedTracks, tracks.AudioFile, error) {
	jobs := make([]tracks.UploadTrackJob, 0, len(req.Uploaded))
	for _, file := range req.Uploaded {
		jobs = append(jobs, tracks.UploadTrackJob{
			Dir:       dir,
			ReleaseID: release.ID,
			File:      file,
		})
	}

	workers := 2
	if len(req.Uploaded) >= 10 {
		workers = 4
	}

	results := worker.Run(ctx, workers, jobs, uc.processor.Process)

	report := tracks.IndexedTracks{}
	var sampleAudioFile tracks.AudioFile

	for _, r := range results {
		name := r.Job.Value.File.Name

		switch {
		case r.Err != nil:
			report.Failed = append(report.Failed, tracks.IgnoredIndexedTrack{
				File:   name,
				Reason: r.Err.Error(),
			})

		case r.Value.FileStatus == tracks.FileSkipped:
			report.Skipped = append(report.Skipped, tracks.IgnoredIndexedTrack{
				File:   name,
				Reason: "already exists",
			})

		case r.Value.FileStatus == tracks.FileIndexed:
			report.Indexed = append(report.Indexed, r.Value.Track)
			if r.Job.ID == req.CoverIdx {
				sampleAudioFile = r.Value.Track.AudioFile
			}

		default:
			return tracks.IndexedTracks{}, tracks.AudioFile{},
				fmt.Errorf("no results recorded")
		}
	}

	return report, sampleAudioFile, nil
}

func (uc *UploadTracks) checkArtists(ctx context.Context, report tracks.IndexedTracks) error {
	for _, track := range report.Indexed {
		if len(track.Artist) < 1 {
			continue
		}

		foundTrackArtists, err := uc.trackArtistRepository.ListByTrackId(ctx, track.ID)
		if err != nil {
			return err
		}

		for _, name := range track.Artist {
			found, err := uc.artistRepository.GetByName(ctx, name)
			if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
				return err
			}
			artistID := found.ID

			if errors.Is(err, apperrors.ErrNotFound) {
				created, err := uc.artistRepository.Create(ctx, name)
				if err != nil {
					return err
				}
				artistID = created.ID
			}

			role := trackartists.ArtistRolePrimary // default
			for _, ta := range foundTrackArtists {
				if ta.Role == trackartists.ArtistRolePrimary {
					role = trackartists.ArtistRoleFeatured
					break
				}
			}

			_, err = uc.trackArtistRepository.Add(ctx, track.ID, artistID, string(role))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (uc *UploadTracks) extractCover(release releases.Release, dir string,
	sampleAudioFile tracks.AudioFile, res *tracks.IndexedTracksRes) error {
	coverName := fmt.Sprintf("%s.png", release.Cover.Name)

	coverExist, err := uc.fileManager.FileExists(coverName, release.Name)
	if err != nil {
		return err
	}
	if coverExist {
		return nil
	}

	path := filepath.Join(dir, sampleAudioFile.Name)

	writer, _, err := uc.fileManager.GetFileWriter(coverName, release.Name)
	if err != nil {
		return err
	}
	if err = metadata.ExtractCover(path, writer); err != nil {
		writer.Close()
		uc.fileManager.DeleteFile(coverName, release.Name)
		res.Warnings = append(res.Warnings, fmt.Sprintf("cover extraction failed: %v", err))
		return nil
	}
	if err = writer.Close(); err != nil {
		uc.fileManager.DeleteFile(coverName, release.Name)
		res.Warnings = append(res.Warnings, fmt.Sprintf("cover write failed: %v", err))
		return nil
	}

	thumbName := fmt.Sprintf("%s.jpg", release.Cover.ThumbName)
	writerThumb, _, err := uc.fileManager.GetFileWriter(thumbName, release.Name)
	if err != nil {
		res.Warnings = append(res.Warnings, fmt.Sprintf("thumbnail writer failed: %v", err))
		return nil
	}
	if err = metadata.ExtractCoverThumbnail(path, writerThumb); err != nil {
		writerThumb.Close()
		uc.fileManager.DeleteFile(thumbName, release.Name)
		res.Warnings = append(res.Warnings, fmt.Sprintf("cover extraction thumbnail failed: %v", err))
		return nil
	}
	if err = writerThumb.Close(); err != nil {
		uc.fileManager.DeleteFile(thumbName, release.Name)
		res.Warnings = append(res.Warnings, fmt.Sprintf("thumbnail write failed: %v", err))
		return nil
	}

	return nil
}
