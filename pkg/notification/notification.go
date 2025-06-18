package notification

import (
	"bytes"
	_ "embed" // Needed for the //go:embed directive
	"time"

	"fyne.io/fyne/v2" // <-- Import fyne
	"github.com/FabricSoul/eve-notify/pkg/logger"
	"github.com/hajimehoshi/oto/v2"
)

// We no longer need the embedded icon, as Fyne will use the app's main icon.
// We still need the sound file.
//go:embed notify.wav
var notificationSound []byte

// Service now holds a reference to the Fyne app.
type Service struct {
	app    fyne.App
	otoCtx *oto.Context
}

// NewService now accepts the fyne.App object.
func NewService(app fyne.App) *Service {
	// Initialize audio context...
	otoCtx, ready, err := oto.NewContext(48000, 2, oto.FormatSignedInt16LE)
	if err != nil {
		logger.Sugar.Errorf("Failed to initialize audio context (oto), sound will be disabled: %v", err)
		return &Service{app: app, otoCtx: nil}
	}
	<-ready

	logger.Sugar.Infoln("Notification and audio services initialized.")
	return &Service{
		app:    app,
		otoCtx: otoCtx,
	}
}

// Notify now uses fyne.App.SendNotification.
func (s *Service) Notify(title, message string, withSound bool) {
	logger.Sugar.Debugf("Sending Fyne notification: Title='%s', Message='%s'", title, message)

	if withSound {
		go s.PlaySound()
	}

	// Use Fyne's built-in, thread-safe notification system.
	s.app.SendNotification(&fyne.Notification{
		Title:   title,
		Content: message,
	})
}

// PlaySound is unchanged and still correct.
func (s *Service) PlaySound() {
	if s.otoCtx == nil {
		logger.Sugar.Warn("Audio context not available, skipping sound playback.")
		return
	}
	soundReader := bytes.NewReader(notificationSound)
	player := s.otoCtx.NewPlayer(soundReader)
	defer player.Close()
	player.Play()
	for player.IsPlaying() {
		time.Sleep(100 * time.Millisecond)
	}
	logger.Sugar.Debug("Test sound playback finished.")
}
