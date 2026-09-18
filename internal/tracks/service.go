package tracks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/metadata"
	"github.com/As71er/once/internal/releases"
	"github.com/As71er/once/internal/utils"
)

type Service interface {
	GetTrackById(ctx context.Context, id int) (*TrackSummary, error)
	UploadTracks(ctx context.Context, req UploadTracksReq) (UploadTracksRes, error)
	ListTracksRelease(ctx context.Context, id int) (ListTracksRelease, error)
	GetTrack(ctx context.Context, id int) (TrackDesc, error)
	DeleteTrack(ctx context.Context, id int) error
}

type TrackService struct {
	repository        Repository
	releaseRepository releases.Repository
	fileManager       utils.FileManager
}

func NewService(repository Repository, releaseRepository releases.Repository, fileManager utils.FileManager) *TrackService {
	return &TrackService{
		repository,
		releaseRepository,
		fileManager,
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
func (ts *TrackService) UploadTracks(ctx context.Context, req UploadTracksReq) (UploadTracksRes, error) {
	var tracks []UploadedTrack

	if len(req.Uploaded) == 0 {
		return UploadTracksRes{}, fmt.Errorf("no files uploaded: %w", apperrors.ErrInvalidInput)
	}

	release, err := ts.releaseRepository.GetReleaseCover(ctx, req.ReleaseID)
	if err != nil {
		return UploadTracksRes{}, err
	}

	// Check cover index to extract option
	if req.CoverIdx < 0 || req.CoverIdx >= len(req.Uploaded) {
		return UploadTracksRes{}, fmt.Errorf("cover index %d out of range: %w", req.CoverIdx, apperrors.ErrInvalidInput)
	}

	dir := ts.fileManager.GetUploadDir(release.Name)
	var sampleAudioFile AudioFile
	for i, uploaded := range req.Uploaded {
		fileExist, err := ts.fileManager.FileExists(uploaded.Name, dir)
		if err != nil {
			return UploadTracksRes{}, err
		}
		if fileExist {
			continue
		}

		file, err := uploaded.Open()
		if err != nil {
			return UploadTracksRes{}, fmt.Errorf("open upload file: %w", errors.Join(apperrors.ErrInvalidInput, err))
		}

		path, err := ts.fileManager.WriteFile(file, uploaded.Name, dir)
		if err != nil {
			file.Close()
			return UploadTracksRes{}, fmt.Errorf("write upload file: %w", err)
		}

		tags, err := metadata.Extract(path)
		if err != nil {
			return UploadTracksRes{}, err
		}

		// Make sure the files is written with a correct format extension
		ext := filepath.Ext(tags.Format.FileName)
		_, format, _ := strings.Cut(ext, ".")

		fileName := strings.TrimSuffix(uploaded.Name, ext)
		normalized := fmt.Sprintf("%s.%s", fileName, tags.Format.FormatName)

		if format != tags.Format.FormatName {
			fileExist, err = ts.fileManager.FileExists(normalized, dir)
			if err != nil {
				return UploadTracksRes{}, err
			}
			if fileExist {
				// Revert first write
				if err := ts.fileManager.DeleteFile(path, dir); err != nil {
					return UploadTracksRes{}, err
				}
				continue
			}
		}

		_, err = ts.fileManager.RenameFile(uploaded.Name, normalized, dir)
		if err != nil {
			return UploadTracksRes{}, err
		}

		createdAudioFile, err := ts.repository.CreateAudioFile(ctx, AudioFile{
			Name:       normalized,
			Codec:      tags.Streams[0].CodecName,
			BitRate:    int64(utils.ParseInt(tags.Format.BitRate, 0)),
			BitDepth:   int64(utils.ParseInt(tags.Streams[0].BitDepth, 0)),
			Channels:   int64(tags.Streams[0].Channels),
			SampleRate: int64(utils.ParseInt(tags.Streams[0].SampleRate, 0)),
			Size:       int64(utils.ParseInt(tags.Format.FileSize, 0)),
		})
		if err != nil {
			return UploadTracksRes{}, err
		}

		createdTrack, err := ts.repository.Create(ctx, Track{
			Title:     tags.Title,
			Duration:  utils.ParseFloat(tags.Streams[0].Duration, 0),
			Track:     int64(tags.TrackNumber),
			Disc:      int64(tags.DiscNumber),
			Composer:  tags.Composer,
			Lyrics:    tags.Lyrics,
			Release:   releases.Release{ID: release.ID},
			AudioFile: AudioFile{ID: createdAudioFile.ID},
		})
		if err != nil {
			return UploadTracksRes{}, err
		}

		track := UploadedTrack{
			ID:        createdTrack.ID,
			Title:     createdTrack.Title,
			Duration:  createdTrack.Duration,
			Track:     createdTrack.Track,
			Disc:      createdTrack.Disc,
			Composer:  createdTrack.Composer,
			ReleaseID: release.ID,
			AudioFile: createdAudioFile,
		}

		tracks = append(tracks, track)

		if i == req.CoverIdx {
			sampleAudioFile = createdAudioFile
		}
	}

	// Check extract cover option
	res := UploadTracksRes{Tracks: tracks, Count: len(tracks)}
	if !req.ExtractCover {
		return res, nil
	}

	// Arbitrary main thumbnail
	coverExist, err := ts.fileManager.FileExists(fmt.Sprintf("%s.png", release.Cover.Name), release.Name)
	if err != nil {
		return UploadTracksRes{}, err
	}

	if !coverExist {
		path := filepath.Join(dir, sampleAudioFile.Name)

		writer, _, err := ts.fileManager.GetFileWriter(fmt.Sprintf("%s.png", release.Cover.Name), release.Name)
		if err != nil {
			return UploadTracksRes{}, err
		}
		if err = metadata.ExtractCover(path, writer); err != nil {
			return UploadTracksRes{}, fmt.Errorf("extract cover: %w", err)
		}
		writer.Close()

		writerThumb, _, err := ts.fileManager.GetFileWriter(fmt.Sprintf("%s.jpg", release.Cover.ThumbName), release.Name)
		if err != nil {
			return UploadTracksRes{}, err
		}
		if err = metadata.ExtractCoverThumbnail(path, writerThumb); err != nil {
			return UploadTracksRes{}, fmt.Errorf("extract cover thumbnail: %w", err)
		}
		writerThumb.Close()
	}

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
