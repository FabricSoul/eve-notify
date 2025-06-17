package character

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"github.com/FabricSoul/eve-notify/pkg/config"
	"github.com/FabricSoul/eve-notify/pkg/esi"
	"github.com/FabricSoul/eve-notify/pkg/logger"
)

// Character represents a single EVE Online character found in the logs.
type Character struct {
	ID       int64
	Name     string
	LastSeen time.Time // Used for sorting
}

// Service handles all character-related logic.
type Service struct {
	prefs       fyne.Preferences
	configSvc *config.Service
}

// NewService creates a new character service.
func NewService(app fyne.App, cfg *config.Service) *Service {
	return &Service{
		prefs:       app.Preferences(),
		configSvc: cfg,
	}
}

// logFileRegex matches EVE log files that contain a character ID.
// It captures the timestamp and the character ID.
var logFileRegex = regexp.MustCompile(`^(\d{8}_\d{6})_(\d+)\.txt$`)

// GetCharacters discovers characters from the log files.
func (s *Service) GetCharacters() ([]*Character, error) {
	logPath := s.configSvc.GetLogPath()
	if logPath == "" {
		return nil, fmt.Errorf("EVE log path is not configured")
	}

	gamelogsPath := filepath.Join(logPath, "Gamelogs")
	logger.Sugar.Infof("Scanning for characters in: %s", gamelogsPath)

	files, err := os.ReadDir(gamelogsPath)
	if err != nil {
		return nil, fmt.Errorf("could not read Gamelogs directory: %w", err)
	}

	// Use a map to store the most recent log time for each unique character ID.
	charLatestTime := make(map[int64]time.Time)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		matches := logFileRegex.FindStringSubmatch(file.Name())
		if len(matches) != 3 {
			continue // Filename doesn't match the pattern with an ID.
		}

		// Parse timestamp and character ID
		timestampStr, charIDStr := matches[1], matches[2]
		charID, _ := strconv.ParseInt(charIDStr, 10, 64)
		logTime, err := time.Parse("20060102_150405", timestampStr)
		if err != nil {
			logger.Sugar.Warnf("Could not parse timestamp from log file '%s': %v", file.Name(), err)
			continue
		}

		// If this log is more recent, update the character's last seen time.
		if logTime.After(charLatestTime[charID]) {
			charLatestTime[charID] = logTime
		}
	}

	// Build the final list of characters.
	var characters []*Character
	for id, lastSeen := range charLatestTime {
		name := s.getCharacterName(id)
		characters = append(characters, &Character{
			ID:       id,
			Name:     name,
			LastSeen: lastSeen,
		})
	}

	// Sort characters by LastSeen time, descending (most recent first).
	sort.Slice(characters, func(i, j int) bool {
		return characters[i].LastSeen.After(characters[j].LastSeen)
	})

	logger.Sugar.Infof("Found %d unique characters.", len(characters))
	return characters, nil
}

// getCharacterName retrieves a character's name, using a cache first.
func (s *Service) getCharacterName(id int64) string {
	cacheKey := fmt.Sprintf("char_name_%d", id)

	// 1. Check the cache (fyne.Preferences)
	cachedName := s.prefs.String(cacheKey)
	if cachedName != "" {
		logger.Sugar.Debugf("Cache hit for character ID %d: %s", id, cachedName)
		return cachedName
	}

	// 2. If not in cache, fetch from ESI
	logger.Sugar.Infof("Cache miss for character ID %d. Fetching from ESI.", id)
	name, err := esi.GetCharacterName(id)
	if err != nil {
		logger.Sugar.Errorf("Failed to get name for character ID %d from ESI: %v", id, err)
		// Return a placeholder so the app doesn't break
		return fmt.Sprintf("Character %d", id)
	}

	// 3. Save to cache for next time
	s.prefs.SetString(cacheKey, name)
	return name
}
