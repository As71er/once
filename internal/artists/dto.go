package artists

type CreateArtistReq struct {
	Name string `json:"name"`
}

type CreateArtistRes struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ArtistSummary struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ListArtistsRes struct {
	Artists []ArtistSummary `json:"data"`
	Count   int             `json:"count"`
}

type UpdateArtistReq struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	// ...other future fields (?)
	// roles not included from models
}

type UpdateArtistRes struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// same as the req
}
