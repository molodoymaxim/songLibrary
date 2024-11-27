package postgres

import (
	"database/sql"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Library struct {
	Songs Songs `json:"songs"`
}

type Songs struct {
	Song     Song     `json:"song"`
	InfoSong InfoSong `json:"info_song"`
}

type Song struct {
	Group string `json:"group"`
	Name  string `json:"song"`
}

type InfoSong struct {
	ReleaseDate *time.Time `json:"releaseDate"`
	Text        string     `json:"text"`
	Link        string     `json:"link"`
}

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) AddSong(song Song, log *slog.Logger) (int, error) {
	const op = "storage.postgres.AddSong()"

	query := `INSERT INTO song (song, music_group) VALUES ($1, $2) returning id`

	var id int

	err := s.db.QueryRow(query, song.Name, song.Group).Scan(&id)
	if err != nil {
		log.Error("Error to insert", op)
		return http.StatusBadRequest, err
	}

	query = `INSERT INTO infosong (id_song) VALUES ($1)`

	_, err = s.db.Exec(query, id)
	if err != nil {
		log.Error("Error to insert", op)
		return http.StatusBadRequest, err
	}

	return id, nil
}

func (s *Storage) ChangeInfo(id int, info InfoSong, log *slog.Logger) (int, error) {
	const op = "storage.postgres.AddInfo()"

	query := `
		UPDATE InfoSong
		SET 
		    releaseDate = COALESCE($1, releaseDate),
		    text = COALESCE($2, text),
		    link = COALESCE($3, link)
		WHERE id_song = $4;
	`

	_, err := s.db.Exec(query, info.ReleaseDate, info.Text, info.Link, id)
	if err != nil {
		log.Error("Error to update", op)
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

func (s *Storage) DeleteSong(id int, log *slog.Logger) (sql.Result, error) {
	const op = "storage.postgres.DeleteInfo()"

	query := `DELETE FROM Song WHERE id = $1;`

	res, err := s.db.Exec(query, id)
	if err != nil {
		log.Error("Error to delete", op)
	}

	return res, nil
}

func (s *Storage) GetText(id int, log *slog.Logger) (string, error) {
	const op = "storage.postgres.GetText()"

	query := `SELECT text FROM infosong WHERE id_song = $1;`

	var text string

	err := s.db.QueryRow(query, id).Scan(&text)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn("No song text found", "id_song", id, "operation", op)
			return "", nil // Можно вернуть пустую строку или специальное сообщение
		}
		log.Error("Error getting song text", "error", err, "operation", op)
		return "", err
	}

	return text, nil
}

func (s *Storage) GetLibrary(log *slog.Logger) ([]Library, error) {

	const op = "storage.postgres.GetLibrary()"

	query := `SELECT s.id, s.music_group, s.song, i.text, i.releasedate, i.link
				FROM song s
				JOIN infosong i ON s.id = i.id_song;
				`

	var library []Library

	rows, err := s.db.Query(query)
	if err != nil {
		log.Error("Error to get songs", op)
	}

	for rows.Next() {
		var lib Library
		var id int64
		err = rows.Scan(&id,
			&lib.Songs.Song.Group,
			&lib.Songs.Song.Name,
			&lib.Songs.InfoSong.Text,
			&lib.Songs.InfoSong.ReleaseDate,
			&lib.Songs.InfoSong.Link)
		if err != nil {
			log.Error("Error to get songs", op)
			return nil, err
		}

		library = append(library, lib)
	}

	return library, nil
}

func (s *Storage) GetInfo(song, group string, log *slog.Logger) (InfoSong, error) {

	const op = "storage.postgres.GetInfo()"

	query := `SELECT text, releasedate, link FROM Library WHERE music_group = $1 AND song = $2;`

	var infoSong InfoSong

	rows, err := s.db.Query(query, song, group)

	if err != nil {
		log.Error("Error to get songs", op)
	}

	for rows.Next() {
		err = rows.Scan(&infoSong.Text,
			&infoSong.ReleaseDate,
			&infoSong.Link)
		if err != nil {
			log.Error("Error to get songs", op)
			return InfoSong{}, err
		}
	}

	return infoSong, nil
}

func (s *Storage) GetLibraryMain(log *slog.Logger) ([]Library, error) {

	const op = "storage.postgres.GetLibraryMain()"

	query := `SELECT music_group, song, text, releasedate, link FROM library;`

	var library []Library

	rows, err := s.db.Query(query)
	if err != nil {
		log.Error("Error to get songs", op)
	}

	for rows.Next() {
		var lib Library
		err = rows.Scan(
			&lib.Songs.Song.Group,
			&lib.Songs.Song.Name,
			&lib.Songs.InfoSong.Text,
			&lib.Songs.InfoSong.ReleaseDate,
			&lib.Songs.InfoSong.Link)
		if err != nil {
			log.Error("Error to get songs", op)
			return nil, err
		}

		library = append(library, lib)
	}

	return library, nil
}

func (s *Storage) CreateTable(log *slog.Logger) {
	const op = "storage.postgres.CreateTable()"

	createLibraryTable := `
    CREATE TABLE IF NOT EXISTS Library(
	id serial PRIMARY KEY,
	music_group varchar(53) NOT NULL ,
	song varchar(50) NOT NULL ,
	text text NOT NULL ,
	releasedate date NOT NULL ,
	link varchar(70) NOT NULL ,
	UNIQUE(music_group, song)
	);`

	createSongTable := `
    CREATE TABLE IF NOT EXISTS song(
	id serial PRIMARY KEY,
	music_group varchar(53) NOT NULL ,
	song varchar(50) NOT NULL ,
	UNIQUE(music_group, song)
	);`

	createInfoSongTable := `
    CREATE TABLE IF NOT EXISTS infosong(
	id serial PRIMARY KEY,
	id_song int references song(id) ON DELETE CASCADE,
	releasedate date ,
	text text ,
	link varchar(70)
	);`

	_, err := s.db.Exec(createLibraryTable)
	if err != nil {
		log.Error("Error to create library table", op)
	}

	_, err = s.db.Exec(createSongTable)
	if err != nil {
		log.Error("Error to create song table", op)
	}

	_, err = s.db.Exec(createInfoSongTable)
	if err != nil {
		log.Error("Error to create infosong table", op)
	}

	return
}

func (s *Storage) MigrateLibrary(log *slog.Logger) {
	const op = "storage.postgres.executeSQLFile()"

	initPath := os.Getenv("INIT_PATH")
	if initPath == "" {
		log.Error("No INIT_PATH environment variable found", op)
	}

	sqlBytes, err := ioutil.ReadFile(initPath)
	if err != nil {
		log.Error("Error to read init.sql", op)
	}
	sqlStatement := string(sqlBytes)

	_, err = s.db.Exec(sqlStatement)
	if err != nil {
		log.Error("Error to execute sql", op)
	}
}
