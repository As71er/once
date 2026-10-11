package main

import (
	"database/sql"
	"log"
	"net/http"
	"sync"

	"github.com/As71er/once/internal/artists"
	artistshttp "github.com/As71er/once/internal/artists/http"

	"github.com/As71er/once/internal/releaseartists"
	releaseartistshttp "github.com/As71er/once/internal/releaseartists/http"

	"github.com/As71er/once/internal/trackartists"
	trackartistshttp "github.com/As71er/once/internal/trackartists/http"

	"github.com/As71er/once/internal/releases"
	releaseshttp "github.com/As71er/once/internal/releases/http"

	"github.com/As71er/once/internal/tracks"
	trackshttp "github.com/As71er/once/internal/tracks/http"
	trackusecases "github.com/As71er/once/internal/tracks/usecases"

	"github.com/As71er/once/internal/config"
	"github.com/As71er/once/internal/httpx"
	"github.com/As71er/once/internal/utils"
)

type api struct {
	cfg *config.Config
	db  *sql.DB
}

func (app *api) mount() http.Handler {
	// HTTP request multiplexer
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Beep Boop"))
	})

	// Ver
	v1 := http.NewServeMux()

	// Setup
	fileManager := utils.NewLocalArchive(app.cfg.ArchivePath)

	artistRepository := artists.NewRepository(app.db)
	releaseRepository := releases.NewRepository(app.db)
	trackRepository := tracks.NewRepository(app.db)
	trackArtistsRepository := trackartists.NewRepository(app.db)

	artistService := artists.NewService(artistRepository)
	artistHandler := artistshttp.NewHandler(artistService)

	releaseService := releases.NewService(releaseRepository, fileManager)
	releaseHandler := releaseshttp.NewHandler(releaseService)

	releaseArtistRepository := releaseartists.NewRepository(app.db)
	releaseArtistService := releaseartists.NewService(releaseArtistRepository)
	releaseArtistHandler := releaseartistshttp.NewHandler(releaseArtistService)

	trackArtistsService := trackartists.NewService(trackArtistsRepository)
	trackArtistsHandler := trackartistshttp.NewHandler(trackArtistsService)

	trackService := tracks.NewService(trackRepository, releaseRepository, fileManager)
	uploadTrackProcessor := tracks.NewUploadTrackProcessor(trackRepository, fileManager, &sync.Mutex{})
	uploadTracks := trackusecases.NewUploadTracks(
		uploadTrackProcessor,
		releaseRepository,
		artistRepository,
		trackArtistsRepository,
		fileManager,
	)
	trackUseCases := trackusecases.NewUsecases(uploadTracks)
	trackHandler := trackshttp.NewHandler(
		trackService,
		trackshttp.Limits{
			MaxUploadMemoryMiB: app.cfg.MaxUploadMemoryMiB,
		},
		trackUseCases,
	)

	// Routes
	v1.HandleFunc("POST /artists", artistHandler.CreateArtist)
	v1.HandleFunc("GET /artists/{id}", artistHandler.GetArtist)
	v1.HandleFunc("GET /artists", artistHandler.ListArtists)
	v1.HandleFunc("PUT /artists/{id}", artistHandler.UpdateArtist)
	v1.HandleFunc("DELETE /artists/{id}", artistHandler.DeleteArtist)

	v1.HandleFunc("GET /releases", releaseHandler.ListReleases)
	v1.HandleFunc("POST /releases", releaseHandler.CreateRelease)
	v1.HandleFunc("GET /releases/{id}", releaseHandler.GetReleaseById)
	v1.HandleFunc("DELETE /releases/{id}", releaseHandler.DeleteRelease)
	v1.HandleFunc("POST /releases/{releaseId}/artists/{artistId}", releaseArtistHandler.AddReleaseArtist)
	v1.HandleFunc("GET /releases/{id}/artists", releaseArtistHandler.ListReleaseArtist)
	v1.HandleFunc("DELETE /releases/{releaseId}/artists/{artistId}", releaseArtistHandler.RemoveReleaseArtist)

	v1.HandleFunc("GET /releases/{id}/cover", releaseHandler.DownloadReleaseCover)
	v1.HandleFunc("GET /releases/{id}/thumb", releaseHandler.DownloadReleaseCoverThumb)

	v1.HandleFunc("GET /releases/{id}/tracks", trackHandler.ListTracksRelease)
	v1.HandleFunc("POST /releases/{id}/tracks", trackHandler.UploadTracks)

	v1.HandleFunc("GET /tracks/{id}", trackHandler.GetTrackById)
	v1.HandleFunc("GET /tracks/{id}/stream", trackHandler.GetTrack)
	v1.HandleFunc("DELETE /tracks/{id}", trackHandler.DeleteTrack)
	v1.HandleFunc("POST /tracks/{id}/artists", trackArtistsHandler.AddTrackArtist)
	v1.HandleFunc("GET /tracks/{id}/artists", trackArtistsHandler.ListTrackArtistByTrackId)
	v1.HandleFunc("PATCH /tracks/{trackId}/artists/{artistId}", trackArtistsHandler.UpdateTrackArtist)
	v1.HandleFunc("DELETE /tracks/{trackId}/artists/{artistId}", trackArtistsHandler.RemoveTrackArtist)

	// Mount ver
	mux.Handle("/v1/", http.StripPrefix("/v1", v1))

	// Middleware
	var handler http.Handler = mux
	handler = httpx.Recover(handler)
	handler = httpx.Logging(handler)

	return handler
}

func (app *api) run(h http.Handler) error {
	server := http.Server{
		Addr:         app.cfg.Addr,
		Handler:      h,
		WriteTimeout: app.cfg.WriteTimeout,
		ReadTimeout:  app.cfg.ReadTimeout,
		IdleTimeout:  app.cfg.IdleTimeout,
	}

	log.Printf("Server initialized at %s", server.Addr)

	return server.ListenAndServe()
}
