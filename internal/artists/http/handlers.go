package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/artists"
	"github.com/As71er/once/internal/httpx"
	"github.com/As71er/once/internal/utils"
)

type handler struct {
	service artists.Service
}

func NewHandler(service artists.Service) *handler {
	return &handler{
		service,
	}
}

func (h *handler) CreateArtist(w http.ResponseWriter, r *http.Request) {
	var request artists.CreateArtistReq
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.Error(w, r, fmt.Errorf("invalid body request: %w", apperrors.ErrInvalidInput))
		return
	}

	response, err := h.service.CreateArtist(r.Context(), request)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, response)
}

func (h *handler) GetArtist(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	response, err := h.service.GetArtist(r.Context(), id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusOK, response)
}

func (h *handler) ListArtists(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	queryParams := r.URL.Query()

	limitStr := queryParams.Get("limit")
	offsetStr := queryParams.Get("offset")
	limit := utils.ParseInt(limitStr, 10)
	offset := utils.ParseInt(offsetStr, 0)

	response, err := h.service.ListArtists(r.Context(), limit, offset)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusOK, response)
}

func (h *handler) UpdateArtist(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	var request artists.UpdateArtistReq
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.Error(w, r, fmt.Errorf("invalid body request: %w", apperrors.ErrInvalidInput))
		return
	}
	request.ID = id

	response, err := h.service.UpdateArtist(r.Context(), request)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusOK, response)
}

func (h *handler) DeleteArtist(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	err = h.service.DeleteArtist(r.Context(), id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
