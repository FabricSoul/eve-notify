package monitoring

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/FabricSoul/eve-notify/pkg/logger"
	"github.com/FabricSoul/eve-notify/pkg/subscription"
)

var ( miningFullRegex = regexp.MustCompile(`Ship's cargo hold is full`)
	manualAutopilotRegex = regexp.MustCompile(`Jumping from`)
)

// miningWorker tails a gamelog file and looks for "cargo full" messages.
func (m *characterMonitor) gamelogWorker(ctx context.Context, filePath string, settings *subscription.NotificationSettings) {
	logger.Sugar.Infof("[%d] Mining worker started for file: %s", m.charID, filePath)

	file, err := os.Open(filePath)
	if err != nil {
		logger.Sugar.Errorf("[%d] Failed to open gamelog for mining worker: %v", m.charID, err)
		return
	}
	defer file.Close()

	// Seek to the end of the file to only read new lines.
	_, _ = file.Seek(0, io.SeekEnd)

	reader := bufio.NewReader(file)

	for {
		select {
		case <-ctx.Done():
			logger.Sugar.Infof("[%d] Mining worker stopped for file: %s", m.charID, filePath)
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// No new lines, wait a bit before trying again.
					time.Sleep(500 * time.Millisecond)
					continue
				}
				logger.Sugar.Warnf("[%d] Error reading from gamelog: %v", m.charID, err)
				return // Exit if there's a real error
			}

			line = strings.TrimSpace(line)
			if settings.MiningStorageFull && miningFullRegex.MatchString(line) {
				logger.Sugar.Infof("!!! MINING NOTIFICATION FOR CHAR %d: Cargo is full!", m.charID)

				title := "EVE Notify - Mining"
				message := fmt.Sprintf("Character %d: Your ship's cargo hold is full.", m.charID)
				m.notifSvc.Notify(title, message, true)
			}
		if settings.ManualAutopilot && manualAutopilotRegex.MatchString(line) {
				title := "EVE Notify - Autopilot"
				message := fmt.Sprintf("Character %d: Manually jumping.", m.charID)
				m.notifSvc.Notify(title, message, true) // Autopilot jumps are frequent, maybe no sound
			}
		}
	}
}
