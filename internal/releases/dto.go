package releases

import (
	"os"
	"time"
)

type CreateReleaseReq struct {
	Title       string `json:"title"`
	ReleaseType string `json:"release_type"`
	Date        string `json:"date"`
	TotalTracks uint   `json:"total_tracks"`
	TotalDiscs  uint   `json:"total_discs"`
}

// domain Release exposes a file names and thubmnails in the case of Cover
type CreatedReleaseRes struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	ReleaseType string `json:"release_type"`
	Date        string `json:"date"`
	TotalTracks uint   `json:"total_tracks"`
	TotalDiscs  uint   `json:"total_discs"`
	CoverID     int64  `json:"cover_id"`
}

type ReleaseRes struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Duration    float64 `json:"duration"`
	ReleaseType string  `json:"release_type"`
	Date        string  `json:"date"`
	TotalTracks uint    `json:"total_tracks"`
	TotalDiscs  uint    `json:"total_discs"`
	CoverID     int64   `json:"cover_id"`
}

type ReleaseListRes struct {
	Releases []ReleaseRes `json:"data"`
	Count    int          `json:"count"`
}

type CoverArtDesc struct {
	File    *os.File
	ModTime time.Time
}
