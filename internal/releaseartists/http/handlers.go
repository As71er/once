package http

import (
	"net/http"

	"github.com/As71er/once/internal/httpx"
	"github.com/As71er/once/internal/releaseartists"
)

type handler struct {
	service releaseartists.Service
}

func NewHandler(service releaseartists.Service) *handler {
	return &handler{
		service,
	}
}

func (h *handler) AddReleaseArtist(w http.ResponseWriter, r *http.Request) {
	releaseID, err := httpx.PathID(r, "releaseId")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	artistID, err := httpx.PathID(r, "artistId")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	response, err := h.service.Add(r.Context(), releaseID, artistID)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusOK, response)
}

func (h *handler) RemoveReleaseArtist(w http.ResponseWriter, r *http.Request) {
	releaseID, err := httpx.PathID(r, "releaseId")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	artistID, err := httpx.PathID(r, "artistId")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	err = h.service.Remove(r.Context(), releaseID, artistID)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) ListReleaseArtist(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	response, err := h.service.ListByReleaseId(r.Context(), id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusOK, response)
}
