package monitoring

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/FabricSoul/eve-notify/pkg/config"
	"github.com/FabricSoul/eve-notify/pkg/logger"
	"github.com/FabricSoul/eve-notify/pkg/notification"
	"github.com/FabricSoul/eve-notify/pkg/subscription"
)

// characterMonitor listens for log file changes for a single character.
type characterMonitor struct {
	charID    int64
	configSvc *config.Service
	subSvc    *subscription.Service
	notifSvc *notification.Service
	ctx       context.Context
	cancel    context.CancelFunc

	// Active log file paths and their cancel functions
	activeGamelogFile    string
	cancelActiveGamelog  context.CancelFunc
	// ... add other logs like Chatlogs here ...
}

func newCharacterMonitor(ctx context.Context, charID int64, cfg *config.Service, sub *subscription.Service, notifi *notification.Service) *characterMonitor {
	// Create a new context for this monitor that is a child of the service's context.
	monitorCtx, monitorCancel := context.WithCancel(ctx)
	return &characterMonitor{
		charID:    charID,
		configSvc: cfg,
		subSvc:    sub,
		notifSvc: notifi,
		ctx:       monitorCtx,
		cancel:    monitorCancel,
	}
}

// run is the main loop for a single character's monitor.
func (m *characterMonitor) run() {
	logger.Sugar.Debugf("[%d] Monitor run loop started.", m.charID)
	ticker := time.NewTicker(10 * time.Second) // Check for new files periodically
	defer ticker.Stop()

	// Initial check
	m.checkForNewLogs()

	for {
		select {
		case <-ticker.C:
			m.checkForNewLogs()
		case <-m.ctx.Done():
			logger.Sugar.Debugf("[%d] Monitor run loop stopping.", m.charID)
			m.stopAllWorkers()
			return
		}
	}
}

func (m *characterMonitor) stop() {
	m.cancel()
}

func (m *characterMonitor) stopAllWorkers() {
	if m.cancelActiveGamelog != nil {
		m.cancelActiveGamelog()
	}
}

// checkForNewLogs finds the latest log files and starts/stops workers as needed.
func (m *characterMonitor) checkForNewLogs() {
	settings, exists := m.subSvc.GetSettings(m.charID)
	if !exists {
		// Should not happen if logic is correct, but a good safeguard.
		m.stop()
		return
	}

	// --- CORRECTED GAMELOG LOGIC ---
	// First, determine if the gamelog worker should be running at all.
	isGamelogMonitoringNeeded := settings.MiningStorageFull || settings.ManualAutopilot // Add future Gamelog settings here

	if isGamelogMonitoringNeeded {
		// If it should be running, find the latest log file.
		logPath := m.configSvc.GetLogPath()
		gamelogDir := filepath.Join(logPath, "Gamelogs")
		latestGamelog := m.findLatestLog(gamelogDir, fmt.Sprintf(`^\d{8}_\d{6}_%d\.txt$`, m.charID))

		// Only act if we found a file AND it's a different one than we're currently watching.
		if latestGamelog != "" && latestGamelog != m.activeGamelogFile {
			logger.Sugar.Infof("[%d] New gamelog detected for monitoring: %s", m.charID, filepath.Base(latestGamelog))

			// Stop the old worker if it's running.
			if m.cancelActiveGamelog != nil {
				m.cancelActiveGamelog()
			}

			// Start the new, generalized worker.
			workerCtx, workerCancel := context.WithCancel(m.ctx)
			m.activeGamelogFile = latestGamelog
			m.cancelActiveGamelog = workerCancel
			go m.gamelogWorker(workerCtx, latestGamelog, settings)
		}
	} else {
		// If no Gamelog monitoring is needed, ensure the worker is stopped.
		if m.cancelActiveGamelog != nil {
			logger.Sugar.Infof("[%d] Disabling gamelog worker as no relevant notifications are active.", m.charID)
			m.cancelActiveGamelog()
			m.cancelActiveGamelog = nil
			m.activeGamelogFile = ""
		}
	}

	// Future: Add checks for Chatlogs here in a similar `if/else` block.
}

// findLatestLog scans a directory for files matching a pattern and returns the path of the most recent one.
func (m *characterMonitor) findLatestLog(dir, pattern string) string {
	files, err := os.ReadDir(dir)
	if err != nil {
		logger.Sugar.Errorf("Failed to read directory %s: %v", dir, err)
		return ""
	}

	re := regexp.MustCompile(pattern)
	var latestFile string
	var latestTime time.Time

	for _, file := range files {
		if !file.IsDir() && re.MatchString(file.Name()) {
			info, err := file.Info()
			if err != nil { continue }

			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestFile = filepath.Join(dir, file.Name())
			}
		}
	}
	return latestFile
}
