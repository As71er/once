package trackartists

type TrackArtist struct {
	TrackID    int64       `json:"track_id"`
	ArtistID   int64       `json:"artist_id"`
	ArtistName string      `json:"artist_name"`
	Role       ArtistRoles `json:"role"`
}

type ArtistRoles string

const (
	ArtistRolePrimary  ArtistRoles = "primary"
	ArtistRoleFeatured ArtistRoles = "featured"
)

func (r ArtistRoles) Valid() bool {
	switch r {
	case ArtistRolePrimary, ArtistRoleFeatured:
		return true
	default:
		return false
	}
}
