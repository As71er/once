package releases

import (
	"encoding/json"
	"net/http"

	"github.com/As71er/once/internal/utils"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service,
	}
}

func (h *handler) CreateRelease(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var release CreateReleaseReq
	err := json.NewDecoder(r.Body).Decode(&release)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response, err := h.service.CreateRelease(r.Context(), release)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *handler) GetReleaseById(w http.ResponseWriter, r *http.Request) {
	id := utils.ParseInt(r.PathValue("id"), -1)
	if id < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response, err := h.service.GetReleaseById(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *handler) DownloadReleaseCover(w http.ResponseWriter, r *http.Request) {
	id := utils.ParseInt(r.PathValue("id"), -1)
	if id < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	coverArt, err := h.service.GetCover(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer coverArt.File.Close()

	http.ServeContent(w, r, coverArt.File.Name(), coverArt.ModTime, coverArt.File)
}

func (h *handler) DownloadReleaseCoverThumb(w http.ResponseWriter, r *http.Request) {
	id := utils.ParseInt(r.PathValue("id"), -1)
	if id < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	coverArt, err := h.service.GetCoverThumb(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer coverArt.File.Close()

	http.ServeContent(w, r, coverArt.File.Name(), coverArt.ModTime, coverArt.File)
}

func (h *handler) DeleteRelease(w http.ResponseWriter, r *http.Request) {
	id := utils.ParseInt(r.PathValue("id"), -1)
	if id < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteRelease(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
