package notification

import (
	_ "embed" // Needed for the //go:embed directive

	"github.com/FabricSoul/eve-notify/pkg/logger"
	"github.com/gen2brain/beeep"
)

//go:embed icon.png
var appIcon []byte

// Service provides a wrapper for sending desktop notifications.
type Service struct{}

// NewService creates and initializes a new notification service.
func NewService() *Service {
	// Set the app name for notifications globally. This is important for some OSes.
	beeep.AppName = "EVE Notify"
	logger.Sugar.Infoln("Notification service initialized.")
	return &Service{}
}

// Notify sends a standard desktop notification.
// It can optionally include a default sound.
func (s *Service) Notify(title, message string, withSound bool) {
	logger.Sugar.Debugf("Sending notification: Title='%s', Message='%s'", title, message)

	if withSound {
		// Beep first, but don't let a sound error stop the notification.
		err := beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		if err != nil {
			logger.Sugar.Warnf("Failed to play notification sound: %v", err)
		}
	}

	// Send the visual notification with our embedded icon.
	err := beeep.Notify(title, message, appIcon)
	if err != nil {
		logger.Sugar.Errorf("Failed to send desktop notification: %v", err)
	}
}
