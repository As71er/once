package usecases

import (
	"context"

	"github.com/As71er/once/internal/tracks"
)

type TrackUseCases interface {
	UploadTracks(context.Context, tracks.UploadTracksReq) (tracks.IndexedTracksRes, error)
}

type trackUseCases struct {
	uploadTracks UploadTracks
}

func NewUsecases(uploadTracks UploadTracks) TrackUseCases {
	return &trackUseCases{
		uploadTracks,
	}
}

func (ut *trackUseCases) UploadTracks(ctx context.Context, req tracks.UploadTracksReq) (tracks.IndexedTracksRes, error) {
	return ut.uploadTracks.Execute(ctx, req)
}
