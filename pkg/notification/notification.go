package notification

import (
	"bytes"
	_ "embed" // Needed for the //go:embed directive
	"time"

	"github.com/FabricSoul/eve-notify/pkg/logger"
	"github.com/gen2brain/beeep"
	"github.com/hajimehoshi/oto/v2"
)

//go:embed icon.png
var appIcon []byte

//go:embed notify.wav
var notificationSound []byte

// Service provides a wrapper for sending desktop notifications and playing sounds.
type Service struct {
	otoCtx *oto.Context
}

// NewService creates and initializes a new notification service.
func NewService() *Service {
	// Initialize the audio player context once with a standard format.
	// We choose a high-quality, common setting: 48000Hz sample rate, 2 channels (stereo),
	// and 16-bit signed integer format.
	// Your notification.wav file should match these settings.
	otoCtx, ready, err := oto.NewContext(48000, 2, oto.FormatSignedInt16LE)
	if err != nil {
		logger.Sugar.Errorf("Failed to initialize audio context (oto), sound will be disabled: %v", err)
		// Return a service with a nil context so the app doesn't crash.
		return &Service{otoCtx: nil}
	}
	// This channel is used to wait for the audio driver to be ready.
	<-ready

	// Set the app name for notifications globally.
	beeep.AppName = "EVE Notify"
	logger.Sugar.Infoln("Notification and audio services initialized.")
	return &Service{
		otoCtx: otoCtx,
	}
}

// Notify sends a standard desktop notification.
// It can optionally play a custom sound through the default audio device.
func (s *Service) Notify(title, message string, withSound bool) {
	logger.Sugar.Debugf("Sending notification: Title='%s', Message='%s'", title, message)

	if withSound {
		// Play the sound in a separate goroutine so it doesn't block the notification.
		go s.PlaySound()
	}

	// Send the visual notification with our embedded icon.
	err := beeep.Notify(title, message, appIcon)
	if err != nil {
		logger.Sugar.Errorf("Failed to send desktop notification: %v", err)
	}
}

// PlaySound plays the embedded WAV file.
func (s *Service) PlaySound() {
	if s.otoCtx == nil {
		logger.Sugar.Warn("Audio context not available, skipping sound playback.")
		return
	}

	// Create a new reader for the embedded sound data for this playback instance.
	soundReader := bytes.NewReader(notificationSound)

	// Create a new player from our audio context with the sound data.
	player := s.otoCtx.NewPlayer(soundReader)
	defer player.Close()

	// --- THE CORRECTED LOGIC ---
	// Play() starts the playback. It does not block.
	player.Play()

	// We need to keep the function alive until the sound is finished.
	// We can do this by waiting for the player to be done.
	for player.IsPlaying() {
		time.Sleep(100 * time.Millisecond)
	}
	// --- END CORRECTION ---
}
