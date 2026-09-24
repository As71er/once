package tracks

import (
	"io"
	"net/http"

	"github.com/As71er/once/internal/httpx"
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
	id, err := httpx.PathID(r)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	response, err := h.service.GetTrackById(r.Context(), id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusOK, response)
}

func (h *handler) GetTrack(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	track, err := h.service.GetTrack(r.Context(), id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	defer track.File.Close()

	http.ServeContent(w, r, track.File.Name(), track.ModTime, track.File)
}

func (h *handler) UploadTracks(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	extractCoverStr := queryParams.Get("extractCover")
	coverIdxStr := queryParams.Get("coverIdx")

	extract := utils.ParseBool(extractCoverStr, true)
	coverIdx := utils.ParseInt(coverIdxStr, 0)

	id, err := httpx.PathID(r)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	if err := r.ParseMultipartForm(h.limits.MaxUploadMemoryMiB << 20); err != nil {
		httpx.Error(w, r, err)
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
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusOK, response)
}

func (h *handler) ListTracksRelease(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	response, err := h.service.ListTracksRelease(r.Context(), id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusOK, response)
}

func (h *handler) DeleteTrack(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}

	if err := h.service.DeleteTrack(r.Context(), id); err != nil {
		httpx.Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
