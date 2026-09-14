package tracks

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/As71er/once/internal/utils"
)

type Limits struct {
	MaxUploadMemoryMiB int64
}

type handler struct {
	service Service
	limits  Limits
}

func NewHandler(service Service, limits Limits) *handler {
	return &handler{
		service,
		limits,
	}
}

func (h *handler) GetTrackById(w http.ResponseWriter, r *http.Request) {
	id := utils.ParseInt(r.PathValue("id"), -1)
	if id < 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response, err := h.service.GetTrackById(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *handler) UploadTracks(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	extractCoverStr := queryParams.Get("extractCover")
	coverIdxStr := queryParams.Get("coverIdx")

	extract := utils.ParseBool(extractCoverStr, true)
	coverIdx := utils.ParseInt(coverIdxStr, 0)

	id := utils.ParseInt(r.PathValue("id"), -1)
	if id < 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseMultipartForm(h.limits.MaxUploadMemoryMiB << 20); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	headers := r.MultipartForm.File["tracks"]
	request := make([]UploadTrack, 0, len(headers))

	for _, header := range headers {
		request = append(request, UploadTrack{
			Name: header.Filename,
			Open: func() (io.ReadCloser, error) {
				return header.Open()
			},
		})
	}

	req := UploadTracksReq{
		ReleaseID:    int64(id),
		ExtractCover: extract,
		CoverIdx:     coverIdx,
		Uploaded:     request,
	}

	response, err := h.service.UploadTracks(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) ListTracksRelease(w http.ResponseWriter, r *http.Request) {
	id := utils.ParseInt(r.PathValue("id"), -1)
	if id < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response, err := h.service.ListTracksRelease(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *handler) DeleteTrack(w http.ResponseWriter, r *http.Request) {
	id := utils.ParseInt(r.PathValue("id"), -1)
	if id < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteTrack(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
