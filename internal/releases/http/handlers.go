package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/As71er/once/internal/apperrors"
	"github.com/As71er/once/internal/httpx"
	"github.com/As71er/once/internal/releases"
	"github.com/As71er/once/internal/utils"
)

type handler struct {
	service releases.Service
}

func NewHandler(service releases.Service) *handler {
	return &handler{
		service,
	}
}

func (h *handler) CreateRelease(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var release releases.CreateReleaseReq
	if err := json.NewDecoder(r.Body).Decode(&release); err != nil {
		httpx.Error(w, r, fmt.Errorf("invalid body request: %w", apperrors.ErrInvalidInput))
		return
	}

	response, err := h.service.CreateRelease(r.Context(), release)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, response)
}

func (h *handler) GetReleaseById(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	response, err := h.service.GetReleaseById(r.Context(), id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusOK, response)
}

func (h *handler) ListReleases(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	queryParams := r.URL.Query()

	limitStr := queryParams.Get("limit")
	offsetStr := queryParams.Get("offset")
	limit := utils.ParseInt(limitStr, 10)
	offset := utils.ParseInt(offsetStr, 0)

	response, err := h.service.ListReleases(r.Context(), limit, offset)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusOK, response)
}

func (h *handler) DownloadReleaseCover(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	coverArt, err := h.service.GetCover(r.Context(), id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	defer coverArt.File.Close()

	http.ServeContent(w, r, coverArt.File.Name(), coverArt.ModTime, coverArt.File)
}

func (h *handler) DownloadReleaseCoverThumb(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	coverArt, err := h.service.GetCoverThumb(r.Context(), id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	defer coverArt.File.Close()

	http.ServeContent(w, r, coverArt.File.Name(), coverArt.ModTime, coverArt.File)
}

func (h *handler) DeleteRelease(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	if err := h.service.DeleteRelease(r.Context(), id); err != nil {
		httpx.Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
