package tracks

import (
	"io"
	"os"
	"time"
)

type UploadTracksReq struct {
	ReleaseID    int64
	ExtractCover bool
	CoverIdx     int
	Uploaded     []UploadTrack
}

type UploadTrack struct {
	Name string
	Open func() (io.ReadCloser, error)
}

type IndexedTracksRes struct {
	Tracks   IndexedTracks `json:"data"`
	Warnings []string      `json:"warnings,omitempty"`
}

type IndexedTracks struct {
	Indexed []IndexedTrack        `json:"indexed"`
	Failed  []IgnoredIndexedTrack `json:"failed"`
	Skipped []IgnoredIndexedTrack `json:"skipped"`
}

type IndexedTrack struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Duration  float64   `json:"duration"`
	Track     int64     `json:"track"`
	Disc      int64     `json:"disc"`
	Composer  string    `json:"composer"`
	ReleaseID int64     `json:"release_id"`
	AudioFile AudioFile `json:"audio_file"`
}

type IgnoredIndexedTrack struct {
	File   string `json:"file"`
	Reason string `json:"reason"`
}

type TrackSummary struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Duration    float64 `json:"duration"`
	Track       int64   `json:"track"`
	Disc        int64   `json:"disc"`
	Composer    string  `json:"composer"`
	AudioFileID *int64  `json:"audio_file_id"`
	ReleaseID   *int64  `json:"release_id"`
}

type ListTracksRelease struct {
	ReleaseID *int64         `json:"release_id"`
	CoverID   *int64         `json:"cover_id"`
	Tracks    []TrackSummary `json:"data"`
	Count     int            `json:"count"`
}

type TrackDesc struct {
	File    *os.File
	ModTime time.Time
}
