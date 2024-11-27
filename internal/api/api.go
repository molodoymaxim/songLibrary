package api

import (
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"log/slog"
	"net/http"
	url2 "net/url"
	"songLibrary/internal/api/request"
	"songLibrary/internal/api/response"
	"songLibrary/internal/storage/postgres"
	"strconv"
)

// AddSongHandler godoc
// @Summary Add a new song to the database
// @Description Adds a new song to the library by providing song information
// @Tags songs
// @Accept json
// @Produce json
// @Param song body postgres.Song true "Song Data"
// @Success 200 {object} request.OkResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 500 {object} request.ErrorResponse
// @Router /song/add [post]
func AddSongHandler(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.AddSongHandler()"

		var song postgres.Song
		w.Header().Set("Content-Type", "application/json")

		err := json.NewDecoder(r.Body).Decode(&song)
		if err != nil {
			log.Error("Error decoding request body", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(request.BadRequest("Error decoding request body"))
			return
		}

		url := fmt.Sprintf("http://0.0.0.0:8081/info?group=%s&song=%s", url2.QueryEscape(song.Group), url2.QueryEscape(song.Name))
		infoSong, err := response.GetInfoSong(log, url)
		if err != nil {
			log.Error("Error getting info song in library", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(request.BadRequest("Error decoding request body"))
			return
		}

		if infoSong.ReleaseDate == nil {
			log.Error("Error library don't have this song", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(request.BadRequest("Error library don't have this song"))
			return
		}

		id, err := storage.AddSong(song, log)
		if err != nil {
			log.Error("Error adding song", "error", err, "operation", op)
			w.WriteHeader(http.StatusInternalServerError)
			pgErr, _ := err.(*pq.Error)
			json.NewEncoder(w).Encode(request.InternalServer(pgErr.Message))
			return
		}

		err = response.ChangeInfoSong(log, infoSong, id)
		if err != nil {
			log.Error("Error changing song info", "error", err, "operation", op)
			w.WriteHeader(http.StatusInternalServerError)
			pgErr, _ := err.(*pq.Error)
			json.NewEncoder(w).Encode(request.InternalServer(pgErr.Message))
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(request.Ok())
		log.Info("song successfully added")
		return
	}
}

// ChangeInfoSongHandler godoc
// @Summary Update song information
// @Description Update the information for an existing song by its ID
// @Tags songs
// @Accept json
// @Produce json
// @Param id query int true "Song ID"
// @Param song body postgres.InfoSong true "Updated Song Info"
// @Success 200 {object} request.OkResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 500 {object} request.ErrorResponse
// @Router /song/change [put]
func ChangeInfoSongHandler(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.AddInfoSongHandler()"

		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			log.Error("no id or transmitted incorrectly", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var infoSong postgres.InfoSong
		err = json.NewDecoder(r.Body).Decode(&infoSong)
		if err != nil {
			log.Error("Error decoding request body", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(request.BadRequest("Error decoding request body"))
			return
		}

		status, err := storage.ChangeInfo(id, infoSong, log)
		if err != nil {
			log.Error("Error changing song info", "error", err, "operation", op)
			w.WriteHeader(http.StatusInternalServerError)
			pgErr, _ := err.(*pq.Error)
			json.NewEncoder(w).Encode(request.InternalServer(pgErr.Message))
			return
		}

		w.WriteHeader(status)
		json.NewEncoder(w).Encode(request.Ok())
		log.Info("song successfully changed")
		return
	}
}

// DeleteSongHandler godoc
// @Summary Delete a song
// @Description Delete a song by its ID
// @Tags songs
// @Produce json
// @Param id query int true "Song ID"
// @Success 200 {object} request.OkResponse
// @Failure 400 {object} request.ErrorResponse
// @Failure 500 {object} request.ErrorResponse
// @Router /song/delete [delete]
func DeleteSongHandler(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.DeleteSongHandler()"

		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			log.Error("no id or transmitted incorrectly", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result, err := storage.DeleteSong(id, log)
		if err != nil {
			log.Error("Error deleting song", "error", err, "operation", op)
			w.WriteHeader(http.StatusInternalServerError)
			pgErr, _ := err.(*pq.Error)
			json.NewEncoder(w).Encode(request.InternalServer(pgErr.Message))
			return
		}

		rowsAffected, err := result.RowsAffected()
		if rowsAffected == 0 || err != nil {
			log.Error("Error deleting song, song with this id not found", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(request.BadRequest("Error deleting song, song id not found"))
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(request.Ok())
		log.Info("song successfully deleted")
		return
	}
}

// TextSongHandler godoc
// @Summary Get song lyrics
// @Description Retrieve the lyrics of a song by its ID
// @Tags songs
// @Produce json
// @Param id query int true "Song ID"
// @Success 200 {object} string "Song Lyrics"
// @Failure 400 {object} request.ErrorResponse
// @Failure 500 {object} request.ErrorResponse
// @Router /song/text [get]
func TextSongHandler(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.TextSongHandler()"

		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			log.Error("no id or transmitted incorrectly", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		text, err := storage.GetText(id, log)
		if err != nil {
			log.Error("Error getting song text", "error", err, "operation", op)
			w.WriteHeader(http.StatusInternalServerError)
			pgErr, _ := err.(*pq.Error)
			json.NewEncoder(w).Encode(request.InternalServer(pgErr.Message))
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(text)
		log.Info("lyrics of the song successfully received")
		return
	}
}

// LibraryHandler godoc
// @Summary Get all songs in the library
// @Description Retrieve a list of all songs available in the library
// @Tags library
// @Produce json
// @Success 200 {array} postgres.Song
// @Failure 400 {object} request.ErrorResponse
// @Router /library [get]
func LibraryHandler(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.LibraryHandler()"

		library, err := storage.GetLibrary(log)
		if err != nil {
			log.Error("Error getting library", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(request.BadRequest("Error getting library"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(library)
		log.Info("library successfully received")
	}
}

// InfoHandler godoc
// @Summary Get info about a specific song
// @Description Retrieve detailed information about a song by group and title
// @Tags songs
// @Produce json
// @Param group query string true "Music Group"
// @Param song query string true "Song Name"
// @Success 200 {object} postgres.InfoSong
// @Failure 500 {object} request.ErrorResponse
// @Router /info [get]
func InfoHandler(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.InfoHandler()"

		group := r.URL.Query().Get("group")
		song := r.URL.Query().Get("song")

		info, err := storage.GetInfo(group, song, log)
		if err != nil {
			log.Error("Error getting info", "error", err, "operation", op)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(request.InternalServer("Error getting info"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
		log.Info("info successfully received")
		return
	}
}

// LibraryMainHandler godoc
// @Summary Get the main library information
// @Description Retrieve the main library information
// @Tags library
// @Produce json
// @Success 200 {array} postgres.Song
// @Failure 400 {object} request.ErrorResponse
// @Router /library/main [get]
func LibraryMainHandler(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.api.LibraryHandlerDB()"

		library, err := storage.GetLibraryMain(log)
		if err != nil {
			log.Error("Error getting library", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(request.BadRequest("Error getting library"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(library)
		log.Info("library successfully received")
	}
}
