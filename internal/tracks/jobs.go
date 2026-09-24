package tracks

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/metadata"
	"github.com/As71er/once/internal/releases"
	"github.com/As71er/once/internal/utils"
)

type FileStatus string

const (
	FileIndexed FileStatus = "indexed"
	FileSkipped FileStatus = "skipped"
	FileError   FileStatus = "error"
)

type FileOut struct {
	Track IndexedTrack
	FileStatus
}

type Input struct {
	dir       string
	releaseID int64
	file      UploadTrack
}

func (ts *TrackService) ProcessUploadedFile(ctx context.Context, dir string, releaseID int64, trackFile UploadTrack) (FileOut, error) {
	fileExist, err := ts.fileManager.FileExists(trackFile.Name, dir)
	if err != nil {
		return FileOut{
			FileStatus: FileError,
		}, err
	}
	if fileExist {
		return FileOut{
			FileStatus: FileSkipped,
		}, nil
	}

	file, err := trackFile.Open()
	if err != nil {
		return FileOut{
			FileStatus: FileError,
		}, fmt.Errorf("open upload file: %w", errors.Join(apperrors.ErrInvalidInput, err))
	}

	// Metadata extraction for index
	path, err := ts.fileManager.WriteFile(file, trackFile.Name, dir)
	file.Close()
	if err != nil {
		ts.fileManager.DeleteFile(trackFile.Name, dir)
		return FileOut{
			FileStatus: FileError,
		}, fmt.Errorf("write upload file: %w", err)
	}

	tags, err := metadata.Extract(path)
	if err != nil {
		ts.fileManager.DeleteFile(trackFile.Name, dir)
		return FileOut{
			FileStatus: FileError,
		}, err
	}

	if tags.Format.FormatName == "" {
		ts.fileManager.DeleteFile(trackFile.Name, dir)
		return FileOut{
			FileStatus: FileError,
		}, fmt.Errorf("could not determine format for %s: %w", trackFile.Name, apperrors.ErrInvalidInput)
	}

	// Make sure the files is written with a correct format extension
	originalExt := filepath.Ext(trackFile.Name)
	baseName := strings.TrimSuffix(trackFile.Name, originalExt)
	normalized := fmt.Sprintf("%s.%s", baseName, tags.Format.FormatName)

	// Actually changed
	if normalized != trackFile.Name {
		fileExist, err := ts.fileManager.FileExists(normalized, dir)
		if err != nil {
			return FileOut{
				FileStatus: FileError,
			}, err
		}
		if fileExist {
			// Revert first write without normalized name
			if err := ts.fileManager.DeleteFile(path, dir); err != nil {
				return FileOut{
					FileStatus: FileError,
				}, err
			}
			return FileOut{
				FileStatus: FileSkipped,
			}, nil
		}

		if _, err := ts.fileManager.RenameFile(trackFile.Name, normalized, dir); err != nil {
			return FileOut{
				FileStatus: FileError,
			}, err
		}
	}

	ts.dbWrite.Lock()
	created, err := ts.repository.InsertTrack(ctx, Track{
		Title:    tags.Title,
		Duration: utils.ParseFloat(tags.Streams[0].Duration, 0),
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
	ts.dbWrite.Unlock()
	if err != nil {
		return FileOut{
			FileStatus: FileError,
		}, err
	}

	indexedTrack := IndexedTrack{
		ID:        created.ID,
		Title:     created.Title,
		Duration:  created.Duration,
		Track:     created.Track,
		Disc:      created.Disc,
		Composer:  created.Composer,
		ReleaseID: releaseID,
		AudioFile: created.AudioFile,
	}

	return FileOut{
		Track:      indexedTrack,
		FileStatus: FileIndexed,
	}, nil
}
