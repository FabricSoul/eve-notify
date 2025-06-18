package subscription

import (
	"sync"

	"github.com/FabricSoul/eve-notify/pkg/logger"
)

// NotificationSettings holds the state of all checkboxes for a character.
type NotificationSettings struct {
	AllianceChat      bool
	CorpChat          bool
	LocalChat         bool
	MiningStorageFull bool
	NpcAggression     bool
	PlayerAggression  bool
	ManualAutopilot   bool
}

// Service manages the subscription state for all characters. It's thread-safe.
type Service struct {
	// The map key is the character ID. The value holds their settings.
	// The presence of a key indicates a subscription is active.
	subscriptions map[int64]*NotificationSettings
	mu            sync.RWMutex
}

// NewService creates a new, empty subscription service.
func NewService() *Service {
	return &Service{
		subscriptions: make(map[int64]*NotificationSettings),
	}
}

// Subscribe activates notifications for a character with default settings.
func (s *Service) Subscribe(charID int64, settings *NotificationSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.Sugar.Infof("Subscribing character %d with specified settings.", charID)
	s.subscriptions[charID] = settings
}

// Unsubscribe deactivates all notifications for a character.
func (s *Service) Unsubscribe(charID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	logger.Sugar.Infof("Unsubscribing character %d", charID)
	delete(s.subscriptions, charID)
}

// IsSubscribed checks if a character is currently subscribed.
func (s *Service) IsSubscribed(charID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.subscriptions[charID]
	return exists
}

// GetSettings returns the notification settings for a character.
// The boolean return value indicates if the character is subscribed.
func (s *Service) GetSettings(charID int64) (*NotificationSettings, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	settings, exists := s.subscriptions[charID]
	if !exists {
		return nil, false
	}
	// Return a copy to prevent race conditions if the caller modifies it.
	settingsCopy := *settings
	return &settingsCopy, true
}

// UpdateSettings saves a new set of notification settings for a character.
// This should only be called for a character who is already subscribed.
func (s *Service) UpdateSettings(charID int64, newSettings *NotificationSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.subscriptions[charID]; exists {
		logger.Sugar.Debugf("Updating settings for character %d", charID)
		s.subscriptions[charID] = newSettings
	}
}
