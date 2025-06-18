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
	"github.com/FabricSoul/eve-notify/pkg/subscription"
	"github.com/FabricSoul/eve-notify/pkg/logger"
)

// Character represents a single EVE Online character found in the logs.
type Character struct {
	ID       int64
	Name     string
	LastSeen time.Time // Used for sorting
	IsSubscribed bool
}

// Service handles all character-related logic.
type Service struct {
	prefs       fyne.Preferences
	configSvc *config.Service
	subSvc    *subscription.Service
}

// NewService creates a new character service.
func NewService(app fyne.App, cfg *config.Service, subSvc *subscription.Service) *Service {
	return &Service{
		prefs:       app.Preferences(),
		configSvc: cfg,
		subSvc: subSvc,
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

	charLatestTime := make(map[int64]time.Time)
	for _, file := range files {
		if file.IsDir() { continue }
		matches := logFileRegex.FindStringSubmatch(file.Name())
		if len(matches) != 3 { continue }
		timestampStr, charIDStr := matches[1], matches[2]
		charID, _ := strconv.ParseInt(charIDStr, 10, 64)
		logTime, err := time.Parse("20060102_150405", timestampStr)
		if err != nil { continue }
		if logTime.After(charLatestTime[charID]) {
			charLatestTime[charID] = logTime
		}
	}

	var characters []*Character
	for id, lastSeen := range charLatestTime {
		name := s.getCharacterName(id)
		characters = append(characters, &Character{
			ID:           id,
			Name:         name,
			LastSeen:     lastSeen,
			IsSubscribed: s.subSvc.IsSubscribed(id),
		})
	}

	// Sort by subscription status (true first), then by last seen time (desc).
	sort.Slice(characters, func(i, j int) bool {
		if characters[i].IsSubscribed != characters[j].IsSubscribed {
			return characters[i].IsSubscribed // true comes before false
		}
		// If both have the same subscription status, sort by LastSeen
		return characters[i].LastSeen.After(characters[j].LastSeen)
	})

	logger.Sugar.Infof("Found and sorted %d unique characters.", len(characters))
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
