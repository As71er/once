package tracks

import "io"

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

type UploadTracksRes struct {
	Tracks []UploadedTrack `json:"data"`
	Count  int             `json:"count"`
}

type UploadedTrack struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Duration  float64   `json:"duration"`
	Track     int64     `json:"track"`
	Disc      int64     `json:"disc"`
	Composer  string    `json:"composer"`
	ReleaseID int64     `json:"release_id"`
	AudioFile AudioFile `json:"audio_file"`
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
