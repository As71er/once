package releaseartists

type ArtistRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ListReleaseArtistRes struct {
	Artists []ArtistRef `json:"data"`
	Count   int         `json:"count"`
}
