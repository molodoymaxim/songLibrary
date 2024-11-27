package swager

import (
	"github.com/go-chi/chi"
	"log/slog"
	"songLibrary/internal/api"
	"songLibrary/internal/storage/postgres"
)

func InitRoutes(r *chi.Mux, log *slog.Logger, storage *postgres.Storage) {
	r.HandleFunc("/songs/add", api.AddSongHandler(log, storage))
	r.HandleFunc("/songs/change", api.ChangeInfoSongHandler(log, storage))
	r.HandleFunc("/songs/delete", api.DeleteSongHandler(log, storage))
	r.HandleFunc("/songs/text", api.TextSongHandler(log, storage))
	r.HandleFunc("/library", api.LibraryHandler(log, storage))
	r.HandleFunc("/info", api.InfoHandler(log, storage))
}
