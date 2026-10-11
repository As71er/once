package tracks

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/metadata"
	"github.com/As71er/once/internal/releases"
	"github.com/As71er/once/internal/utils"
)

type UploadTrackProcessor struct {
	repository  Repository
	fileManager utils.FileManager
	dbWrite     *sync.Mutex
}

func NewUploadTrackProcessor(repository Repository, fileManager utils.FileManager,
	dbWrite *sync.Mutex) *UploadTrackProcessor {
	return &UploadTrackProcessor{
		repository:  repository,
		fileManager: fileManager,
		dbWrite:     dbWrite,
	}
}

type UploadTrackJob struct {
	Dir       string
	ReleaseID int64
	File      UploadTrack
}

type UploadTrackResult struct {
	Track IndexedTrack
	FileStatus
}

type FileStatus string

const (
	FileIndexed FileStatus = "indexed"
	FileSkipped FileStatus = "skipped"
	FileError   FileStatus = "error"
)

func (p *UploadTrackProcessor) Process(ctx context.Context, job UploadTrackJob) (UploadTrackResult, error) {
	dir := job.Dir
	releaseID := job.ReleaseID
	trackFile := job.File

	fileExist, err := p.fileManager.FileExists(trackFile.Name, dir)
	if err != nil {
		return UploadTrackResult{
			FileStatus: FileError,
		}, err
	}
	if fileExist {
		return UploadTrackResult{
			FileStatus: FileSkipped,
		}, nil
	}

	file, err := trackFile.Open()
	if err != nil {
		return UploadTrackResult{
			FileStatus: FileError,
		}, fmt.Errorf("open upload file: %w", errors.Join(apperrors.ErrInvalidInput, err))
	}

	// Metadata extraction for index
	path, err := p.fileManager.WriteFile(file, trackFile.Name, dir)
	file.Close()
	if err != nil {
		p.fileManager.DeleteFile(trackFile.Name, dir)
		return UploadTrackResult{
			FileStatus: FileError,
		}, fmt.Errorf("write upload file: %w", err)
	}

	tags, err := metadata.Extract(path)
	if err != nil {
		p.fileManager.DeleteFile(trackFile.Name, dir)
		return UploadTrackResult{
			FileStatus: FileError,
		}, err
	}

	if tags.Format.FormatName == "" {
		p.fileManager.DeleteFile(trackFile.Name, dir)
		return UploadTrackResult{
			FileStatus: FileError,
		}, fmt.Errorf("could not determine format for %s: %w", trackFile.Name, apperrors.ErrInvalidInput)
	}

	// Make sure the files is written with a correct format extension
	originalExt := filepath.Ext(trackFile.Name)
	baseName := strings.TrimSuffix(trackFile.Name, originalExt)
	normalized := fmt.Sprintf("%s.%s", baseName, tags.Format.FormatName)

	// Actually changed
	if normalized != trackFile.Name {
		fileExist, err := p.fileManager.FileExists(normalized, dir)
		if err != nil {
			return UploadTrackResult{
				FileStatus: FileError,
			}, err
		}
		if fileExist {
			// Revert first write without normalized name
			if err := p.fileManager.DeleteFile(path, dir); err != nil {
				return UploadTrackResult{
					FileStatus: FileError,
				}, err
			}
			return UploadTrackResult{
				FileStatus: FileSkipped,
			}, nil
		}

		if _, err := p.fileManager.RenameFile(trackFile.Name, normalized, dir); err != nil {
			return UploadTrackResult{
				FileStatus: FileError,
			}, err
		}
	}

	p.dbWrite.Lock()
	created, err := p.repository.InsertTrack(ctx, Track{
		Title:    tags.Title,
		Duration: utils.ParseFloat(tags.Streams[0].Duration, 0),
		Artist:   tags.Artist,
		Track:    int64(tags.TrackNumber),
		Disc:     int64(tags.DiscNumber),
		Composer: tags.Composer,
		Lyrics:   tags.Lyrics,
		Release:  releases.Release{ID: releaseID},
	}, AudioFile{
		Name:       normalized,
		Codec:      tags.Streams[0].CodecName,
		BitRate:    int64(utils.ParseInt(tags.Format.BitRate, 0)),
		BitDepth:   int64(utils.ParseInt(tags.Streams[0].BitDepth, 0)),
		Channels:   int64(tags.Streams[0].Channels),
		SampleRate: int64(utils.ParseInt(tags.Streams[0].SampleRate, 0)),
		Size:       int64(utils.ParseInt(tags.Format.FileSize, 0)),
	})
	p.dbWrite.Unlock()
	if err != nil {
		return UploadTrackResult{
			FileStatus: FileError,
		}, err
	}

	indexedTrack := IndexedTrack{
		ID:        created.ID,
		Title:     created.Title,
		Artist:    tags.Artist,
		Duration:  created.Duration,
		Track:     created.Track,
		Disc:      created.Disc,
		Composer:  created.Composer,
		ReleaseID: releaseID,
		AudioFile: created.AudioFile,
	}

	return UploadTrackResult{
		Track:      indexedTrack,
		FileStatus: FileIndexed,
	}, nil
}
