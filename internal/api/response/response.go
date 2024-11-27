package response

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"songLibrary/internal/storage/postgres"
	"strconv"
)

func GetInfoSong(log *slog.Logger, url string) (postgres.InfoSong, error) {
	const op = "internal.api.response.getInfoSong()"

	var infoSong postgres.InfoSong

	resp, err := http.Get(url)
	if err != nil {
		log.Error("Error making request to external API", "error", err, "operation", op)
		return infoSong, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {

		if err = json.NewDecoder(resp.Body).Decode(&infoSong); err != nil {
			log.Error("Error decoding external API response", "error", err, "operation", op)
			return infoSong, err
		}
	}

	return infoSong, nil
}

func ChangeInfoSong(log *slog.Logger, infoSong postgres.InfoSong, id int) error {
	const op = "internal.api.response.changeInfoSong()"
	url := "http://0.0.0.0:8081/songLibrary/ChangeInfo" + "?id=" + strconv.Itoa(id)

	jsonData, err := json.Marshal(infoSong)
	if err != nil {
		return fmt.Errorf("error marshalling infoSong to JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating new request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error to response info song", "error", err, "operation", op)
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("Error to response info song", "error", err, "operation", op)
		return fmt.Errorf("error: received non-OK status code %d", resp.StatusCode)
	}

	return err
}
