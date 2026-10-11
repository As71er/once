package tracks

import (
	"github.com/As71er/once/internal/releases"
	"github.com/As71er/once/internal/trackartists"
)

type Track struct {
	ID          int64            `json:"id"`
	Title       string           `json:"title"`
	Duration    float64          `json:"duration"`
	Artist      []string         `json:"artists"`
	Track       int64            `json:"track"`
	Disc        int64            `json:"disc"`
	Composer    string           `json:"composer"`
	Lyrics      string           `json:"lyrics"`
	Release     releases.Release `json:"release"`
	AudioFile   AudioFile        `json:"audio_file"`
	TrackArtist []TrackArtistRef `json:"track_artists"`
}

type AudioFile struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Codec      string `json:"codec"`
	BitRate    int64  `json:"bit_rate"`
	BitDepth   int64  `json:"bit_depth"`
	Channels   int64  `json:"channels"`
	SampleRate int64  `json:"sample_rate"`
	Size       int64  `json:"size"`
}

type TrackArtistRef struct {
	ID   int64                    `json:"id"`
	Name string                   `json:"name"`
	Role trackartists.ArtistRoles `json:"role"`
}
