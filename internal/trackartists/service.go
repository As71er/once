package trackartists

import (
	"context"
	"fmt"

	"github.com/As71er/once/internal/apperrors"
)

type Service interface {
	Add(context.Context, AddTrackArtistReq) (AddTrackArtistRes, error)
	Update(ctx context.Context, req UpdateTrackArtistReq) error
	Remove(ctx context.Context, trackID, artistID int) error
	ListByTrackId(context.Context, int) (ListTrackArtistRes, error)
}

type TrackArtistService struct {
	repository Repository
}

func NewService(repository Repository) *TrackArtistService {
	return &TrackArtistService{
		repository,
	}
}

func (ta *TrackArtistService) Add(ctx context.Context, req AddTrackArtistReq) (AddTrackArtistRes, error) {
	isValidRole := ArtistRoles.Valid(ArtistRoles(req.Role))
	if !isValidRole {
		return AddTrackArtistRes{}, fmt.Errorf("add track artist with invalid role %s: %w", req.Role, apperrors.ErrInvalidInput)
	}

	trackArtists, err := ta.repository.ListByTrackId(ctx, int64(req.TrackID))
	if err != nil {
		return AddTrackArtistRes{}, err
	}

	if foundArtist, ok := trackArtists[int64(req.ArtistID)]; ok {
		return AddTrackArtistRes{
			TrackID:    int64(req.TrackID),
			ArtistID:   foundArtist.ArtistID,
			ArtistName: foundArtist.ArtistName,
			Role:       string(foundArtist.Role),
		}, nil
	}

	fmt.Println(req.Role)
	trackArtistInserted, err := ta.repository.Add(ctx, int64(req.TrackID), int64(req.ArtistID), req.Role)
	if err != nil {
		return AddTrackArtistRes{}, err
	}

	return AddTrackArtistRes{
		TrackID:    trackArtistInserted.TrackID,
		ArtistID:   trackArtistInserted.ArtistID,
		ArtistName: trackArtistInserted.ArtistName,
		Role:       string(trackArtistInserted.Role),
	}, nil
}

func (ta *TrackArtistService) Update(ctx context.Context, req UpdateTrackArtistReq) error {
	isValidRole := ArtistRoles.Valid(ArtistRoles(req.Role))
	if !isValidRole {
		return fmt.Errorf("update track artist with invalid role %s: %w", req.Role, apperrors.ErrInvalidInput)
	}

	err := ta.repository.Update(ctx, int64(req.TrackID), int64(req.ArtistID), req.Role)
	if err != nil {
		return err
	}

	return nil
}

func (ta *TrackArtistService) Remove(ctx context.Context, trackID, artistID int) error {
	err := ta.repository.Remove(ctx, int64(trackID), int64(artistID))
	if err != nil {
		return err
	}

	return nil
}

func (ta *TrackArtistService) ListByTrackId(ctx context.Context, id int) (ListTrackArtistRes, error) {
	foundTrackArtists, err := ta.repository.ListByTrackId(ctx, int64(id))
	if err != nil {
		return ListTrackArtistRes{}, err
	}

	var trackArtists []TrackArtistSumamry
	for _, artist := range foundTrackArtists {
		trackArtists = append(trackArtists, TrackArtistSumamry{
			ID:   artist.ArtistID,
			Name: artist.ArtistName,
			Role: string(artist.Role),
		})
	}

	return ListTrackArtistRes{
		Artists: trackArtists,
		Count:   len(trackArtists),
	}, nil

}
