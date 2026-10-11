package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/httpx"
	"github.com/As71er/once/internal/trackartists"
)

type handler struct {
	service trackartists.Service
}

func NewHandler(service trackartists.Service) *handler {
	return &handler{
		service,
	}
}

func (h *handler) AddTrackArtist(w http.ResponseWriter, r *http.Request) {
	var request trackartists.AddTrackArtistReq
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.Error(w, r, fmt.Errorf("invalid body request: %w", apperrors.ErrInvalidInput))
		return
	}

	response, err := h.service.Add(r.Context(), request)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusOK, response)
}

func (h *handler) UpdateTrackArtist(w http.ResponseWriter, r *http.Request) {
	trackID, err := httpx.PathID(r, "trackId")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	artistID, err := httpx.PathID(r, "artistId")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	var request trackartists.UpdateTrackArtistReq
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.Error(w, r, fmt.Errorf("invalid body request: %w", apperrors.ErrInvalidInput))
		return
	}
	request.TrackID = trackID
	request.ArtistID = artistID

	err = h.service.Update(r.Context(), request)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) RemoveTrackArtist(w http.ResponseWriter, r *http.Request) {
	trackID, err := httpx.PathID(r, "trackId")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	artistID, err := httpx.PathID(r, "artistId")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	err = h.service.Remove(r.Context(), trackID, artistID)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) ListTrackArtistByTrackId(w http.ResponseWriter, r *http.Request) {
	trackID, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	response, err := h.service.ListByTrackId(r.Context(), trackID)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusOK, response)
}
