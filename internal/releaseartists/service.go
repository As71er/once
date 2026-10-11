package releaseartists

import (
	"context"
)

type Service interface {
	Add(ctx context.Context, releaseID, artistID int) (ReleaseArtist, error)
	Remove(ctx context.Context, releaseID, artistID int) error
	ListByReleaseId(context.Context, int) (ListReleaseArtistRes, error)
}

type ReleaseArtistService struct {
	repository Repository
}

func NewService(repository Repository) *ReleaseArtistService {
	return &ReleaseArtistService{
		repository,
	}
}

func (rs *ReleaseArtistService) Add(ctx context.Context, releaseID, artistID int) (ReleaseArtist, error) {
	releaseArtist, err := rs.repository.ListByReleaseId(ctx, int64(releaseID))
	if err != nil {
		return ReleaseArtist{}, err
	}

	if foundArtist, ok := releaseArtist[int64(artistID)]; ok {
		return ReleaseArtist{
			ReleaseID: int64(releaseID),
			ArtistID:  foundArtist.ID,
		}, nil
	}

	added, err := rs.repository.Add(ctx, int64(releaseID), int64(artistID))
	if err != nil {
		return ReleaseArtist{}, err
	}

	return ReleaseArtist{
		ReleaseID: added.ReleaseID,
		ArtistID:  added.ArtistID,
	}, nil
}

func (rs *ReleaseArtistService) Remove(ctx context.Context, releaseID, artistID int) error {
	err := rs.repository.Remove(ctx, int64(releaseID), int64(artistID))
	if err != nil {
		return err
	}

	return nil
}

func (rs *ReleaseArtistService) ListByReleaseId(ctx context.Context, id int) (ListReleaseArtistRes, error) {
	foundReleaseArtists, err := rs.repository.ListByReleaseId(ctx, int64(id))
	if err != nil {
		return ListReleaseArtistRes{}, err
	}

	var releaseArtists []ArtistRef
	for _, artist := range foundReleaseArtists {
		releaseArtists = append(releaseArtists, ArtistRef{
			ID:   artist.ID,
			Name: artist.Name,
		})
	}

	return ListReleaseArtistRes{
		Artists: releaseArtists,
		Count:   len(releaseArtists),
	}, nil
}
