package main

import (
	"log"
	"net/http"

	"github.com/As71er/once/internal/artists"
	"github.com/As71er/once/internal/config"
	"github.com/As71er/once/internal/releases"
	"github.com/As71er/once/internal/sqlc"
	"github.com/As71er/once/internal/tracks"
	"github.com/As71er/once/internal/utils"
)

type api struct {
	cfg *config.Config
	db  *sqlc.Queries
}

func (app *api) mount() http.Handler {
	// HTTP request multiplexer
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Beep Boop"))
	})

	// Ver
	v1 := http.NewServeMux()

	// Middlewares

	// Setup
	fileManager := utils.NewLocalArchive(app.cfg.ArchivePath)

	artistRepository := artists.NewRepository(app.db)
	releaseRepository := releases.NewRepository(app.db)
	trackRepository := tracks.NewRepository(app.db)

	releaseService := releases.NewService(releaseRepository, artistRepository, fileManager)
	releaseHandler := releases.NewHandler(releaseService)

	trackService := tracks.NewService(trackRepository, releaseRepository, fileManager)
	trackHandler := tracks.NewHandler(trackService, tracks.Limits{MaxUploadMemoryMiB: app.cfg.MaxUploadMemoryMiB})

	// Routes
	v1.HandleFunc("GET /releases", releaseHandler.ListReleases)
	v1.HandleFunc("POST /releases", releaseHandler.CreateRelease)
	v1.HandleFunc("GET /releases/{id}", releaseHandler.GetReleaseById) // TODO: add relationships refs as use case (artist + tracks)
	v1.HandleFunc("DELETE /releases/{id}", releaseHandler.DeleteRelease)

	v1.HandleFunc("GET /releases/{id}/cover", releaseHandler.DownloadReleaseCover)
	v1.HandleFunc("GET /releases/{id}/thumb", releaseHandler.DownloadReleaseCoverThumb)

	v1.HandleFunc("GET /releases/{id}/tracks", trackHandler.ListTracksRelease)
	v1.HandleFunc("POST /releases/{id}/tracks", trackHandler.UploadTracks) // Check service comment

	v1.HandleFunc("GET /tracks/{id}", trackHandler.GetTrackById) // TODO: same as release getbyid
	v1.HandleFunc("GET /tracks/{id}/stream", trackHandler.GetTrack)
	v1.HandleFunc("DELETE /tracks/{id}", trackHandler.DeleteTrack)

	// Mount ver
	mux.Handle("/v1/", http.StripPrefix("/v1", v1))

	return mux
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
