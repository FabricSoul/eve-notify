// /eve-notify/pkg/config/config.go
package config

import (
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"github.com/FabricSoul/eve-notify/pkg/logger"
)

// Keys for storing preferences. Using constants prevents typos.
const (
	keyLogPath = "eve_log_path"
)

// Service provides a structured way to interact with app preferences.
type Service struct {
	prefs fyne.Preferences
}

// NewService creates a new configuration service.
func NewService(app fyne.App) *Service {
	return &Service{
		prefs: app.Preferences(),
	}
}

// Init ensures that essential preferences are set, using defaults
// or auto-detection if they don't exist.
func (s *Service) Init() {
	// If the path is already set, we're done.
	if s.prefs.String(keyLogPath) != "" {
		logger.Sugar.Infoln("EVE log path is already configured.")
		return
	}

	logger.Sugar.Infoln("EVE log path not set. Attempting auto-detection.")
	detectedPath := s.findDefaultEveLogPath()
	s.SetLogPath(detectedPath) // Save the path (even if empty)

	if detectedPath == "" {
		logger.Sugar.Warnln("Auto-detection failed. User must set the path manually.")
	}
}

// GetLogPath returns the configured EVE Online log path.
func (s *Service) GetLogPath() string {
	return s.prefs.String(keyLogPath)
}

func (s *Service) RestoreDefaultLogPath() string {
	logger.Sugar.Infoln("Restoring default log path.")
	detectedPath := s.findDefaultEveLogPath()
	s.SetLogPath(detectedPath)
	return detectedPath
}

// SetLogPath saves the EVE Online log path.
func (s *Service) SetLogPath(path string) {
	s.prefs.SetString(keyLogPath, path)
	logger.Sugar.Infof("Set EVE log path to: %s", path)
}

// findDefaultEveLogPath tries to find the default EVE Online log directory.
func (s *Service) findDefaultEveLogPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Sugar.Errorf("Could not get user home directory: %v", err)
		return ""
	}

	var logPath string
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		logPath = filepath.Join(homeDir, "Documents", "EVE", "logs", "Gamelogs")
	} else {
		logger.Sugar.Warnf("Unsupported OS for auto-detection: %s", runtime.GOOS)
		return ""
	}

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		logger.Sugar.Warnf("Auto-detected EVE log path does not exist: %s", logPath)
		return ""
	}

	logger.Sugar.Infof("Auto-detected EVE log path: %s", logPath)
	return logPath
}
