package trackartists

type AddTrackArtistReq struct {
	TrackID  int    `json:"track_id"`
	ArtistID int    `json:"artist_id"`
	Role     string `json:"artist_role"`
}

type AddTrackArtistRes struct {
	TrackID    int64  `json:"track_id"`
	ArtistID   int64  `json:"artist_id"`
	ArtistName string `json:"artist_name"`
	Role       string `json:"artist_role"`
}

type UpdateTrackArtistReq struct {
	TrackID  int    `json:"track_id"`
	ArtistID int    `json:"artist_id"`
	Role     string `json:"artist_role"`
}

type RemoveTrackArtistReq struct {
	TrackID  int `json:"track_id"`
	ArtistID int `json:"artist_id"`
}

type TrackArtistSumamry struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type ListTrackArtistRes struct {
	Artists []TrackArtistSumamry `json:"data"`
	Count   int                  `json:"count"`
}
