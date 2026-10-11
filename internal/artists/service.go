package artists

import (
	"context"
)

type Service interface {
	CreateArtist(context.Context, CreateArtistReq) (CreateArtistRes, error)
	GetArtist(context.Context, int) (ArtistSummary, error)
	ListArtists(ctx context.Context, limit, offset int) (ListArtistsRes, error)
	UpdateArtist(context.Context, UpdateArtistReq) (UpdateArtistRes, error)
	DeleteArtist(context.Context, int) error
}

type ArtistService struct {
	repository Repository
}

func NewService(repository Repository) *ArtistService {
	return &ArtistService{
		repository,
	}
}

func (as *ArtistService) CreateArtist(ctx context.Context, req CreateArtistReq) (CreateArtistRes, error) {
	created, err := as.repository.Create(ctx, req.Name)
	if err != nil {
		return CreateArtistRes{}, err
	}

	return CreateArtistRes{
		ID:   created.ID,
		Name: created.Name,
	}, nil
}

func (as *ArtistService) GetArtist(ctx context.Context, id int) (ArtistSummary, error) {
	artist, err := as.repository.GetById(ctx, int64(id))
	if err != nil {
		return ArtistSummary{}, err

	}

	return ArtistSummary{
		ID:   artist.ID,
		Name: artist.Name,
	}, nil
}

func (as *ArtistService) ListArtists(ctx context.Context, limit, offset int) (ListArtistsRes, error) {
	artists, err := as.repository.List(ctx, int64(limit), int64(offset))
	if err != nil {
		return ListArtistsRes{}, err
	}

	var artistsList []ArtistSummary
	for _, artist := range artists {
		artistsList = append(artistsList, ArtistSummary{
			ID:   artist.ID,
			Name: artist.Name,
		})
	}

	return ListArtistsRes{
		Artists: artistsList,
		Count:   len(artistsList),
	}, nil
}

func (as *ArtistService) UpdateArtist(ctx context.Context, req UpdateArtistReq) (UpdateArtistRes, error) {
	artist, err := as.repository.Update(ctx, int64(req.ID), req.Name)
	if err != nil {
		return UpdateArtistRes{}, err
	}

	return UpdateArtistRes{
		ID:   artist.ID,
		Name: artist.Name,
	}, nil
}

func (as *ArtistService) DeleteArtist(ctx context.Context, id int) error {
	err := as.repository.Delete(ctx, int64(id))
	if err != nil {
		return err
	}

	return nil
}
