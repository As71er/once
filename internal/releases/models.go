package releases

type Release struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Name        string `json:"name"`
	ReleaseType string `json:"release_type"`
	Date        string `json:"date"`
	TotalTracks uint   `json:"total_tracks"`
	TotalDiscs  uint   `json:"total_discs"`
	Cover       *Cover `json:"cover"`
}

type Cover struct {
	ID        *int64 `json:"id"`
	Name      string `json:"name"`
	ThumbName string `json:"thumb"`
}

type ReleaseType string

const (
	ReleaseTypeAlbum       ReleaseType = "album"
	ReleaseTypeEP          ReleaseType = "ep"
	ReleaseTypeCompilation ReleaseType = "compilation"
	ReleaseTypeSingle      ReleaseType = "single"
)

func (r ReleaseType) Valid() bool {
	switch r {
	case ReleaseTypeAlbum, ReleaseTypeEP, ReleaseTypeCompilation, ReleaseTypeSingle:
		return true
	default:
		return false
	}
}
