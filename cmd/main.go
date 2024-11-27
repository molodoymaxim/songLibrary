package main

import (
	"github.com/go-chi/chi"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	"os"
	"songLibrary/internal/api"
	"songLibrary/internal/config"
	"songLibrary/internal/storage"
	"songLibrary/internal/storage/postgres"
	"songLibrary/internal/swager"
)

const (
	envLocal = "local"
	encDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting api", slog.String("key", cfg.Env))
	log.Debug("debug message enable")

	db := storage.Connection(log)
	router := chi.NewRouter()

	storageDB := postgres.NewStorage(db)
	log.Info("db connection successful")

	storageDB.CreateTable(log)
	storageDB.MigrateLibrary(log)

	swager.InitRoutes(router, log, storageDB)

	router.Mount("/swagger", httpSwagger.WrapHandler)

	router.Post("/EffectiveMobile/AddSong", api.AddSongHandler(log, storageDB))
	router.Post("/EffectiveMobile/ChangeInfo", api.ChangeInfoSongHandler(log, storageDB))
	router.Delete("/EffectiveMobile/DeleteSong", api.DeleteSongHandler(log, storageDB))
	router.Get("/EffectiveMobile/TextSong", api.TextSongHandler(log, storageDB))
	router.Get("/EffectiveMobile/Library", api.LibraryHandler(log, storageDB))
	router.Get("/EffectiveMobile/info", api.InfoHandler(log, storageDB))

	router.Get("/Library", api.LibraryMainHandler(log, storageDB))

	err := http.ListenAndServe(cfg.Address, router)
	if err != nil {
		log.Error("Error starting server", err)
	}

}

func setupLogger(env string) *slog.Logger {

	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case encDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
