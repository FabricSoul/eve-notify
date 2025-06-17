package esi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/FabricSoul/eve-notify/pkg/logger"
)

// ESI API endpoint for character details
const characterEndpoint = "https://esi.evetech.net/latest/characters/%d/?datasource=tranquility"

// Client is a basic HTTP client for ESI.
var client = &http.Client{Timeout: 10 * time.Second}

// CharacterResponse models the part of the ESI response we care about.
type CharacterResponse struct {
	Name string `json:"name"`
}

// GetCharacterName fetches a character's name from the ESI API using their ID.
func GetCharacterName(id int64) (string, error) {
	url := fmt.Sprintf(characterEndpoint, id)
	logger.Sugar.Debugf("Fetching character name from ESI: %s", url)

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to make request to ESI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ESI returned non-200 status: %s", resp.Status)
	}

	var charResp CharacterResponse
	if err := json.NewDecoder(resp.Body).Decode(&charResp); err != nil {
		return "", fmt.Errorf("failed to decode ESI response: %w", err)
	}

	return charResp.Name, nil
}
