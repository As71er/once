package main

import (
	"database/sql"
	"log/slog"
	"os"

	migrations "github.com/As71er/once/data"
	"github.com/As71er/once/internal/config"
	"github.com/As71er/once/internal/sqlc"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	conn, err := sql.Open("sqlite", cfg.DBDSN)
	if err != nil {
		slog.Error("database failed to initialize", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	goose.SetBaseFS(migrations.EmbedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}
	if err := goose.Up(conn, "migrations"); err != nil {
		panic(err)
	}

	querier := sqlc.New(conn)

	api := api{cfg: cfg, db: querier}

	if err := api.run(api.mount()); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
