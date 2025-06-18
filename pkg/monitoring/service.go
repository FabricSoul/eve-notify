package monitoring

import (
	"context"
	"sync"

	"github.com/FabricSoul/eve-notify/pkg/config"
	"github.com/FabricSoul/eve-notify/pkg/subscription"
	"github.com/FabricSoul/eve-notify/pkg/logger"
)

// Service is the main monitoring controller.
type Service struct {
	configSvc *config.Service
	subSvc    *subscription.Service
	monitors  map[int64]*characterMonitor
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func NewService(cfg *config.Service, sub *subscription.Service) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		configSvc: cfg,
		subSvc:    sub,
		monitors:  make(map[int64]*characterMonitor),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start begins the main monitoring loop. It should be run in a goroutine.
func (s *Service) Start() {
	logger.Sugar.Infoln("Monitoring service started.")
	s.wg.Add(1)
	defer s.wg.Done()

	for {
		select {
		case charID := <-s.subSvc.Subscribed:
			s.startMonitor(charID)
		case charID := <-s.subSvc.Unsubscribed:
			s.stopMonitor(charID)
		case <-s.ctx.Done():
			logger.Sugar.Infoln("Monitoring service shutting down.")
			return
		}
	}
}

// Stop gracefully shuts down all active monitors.
func (s *Service) Stop() {
	logger.Sugar.Infoln("Stopping monitoring service...")
	s.cancel() // Signal all goroutines to stop.
	s.wg.Wait()  // Wait for the main loop and all monitors to finish.
	logger.Sugar.Infoln("Monitoring service stopped.")
}

func (s *Service) startMonitor(charID int64) {
	if _, exists := s.monitors[charID]; exists {
		logger.Sugar.Warnf("Monitor for character %d already running.", charID)
		return
	}
	logger.Sugar.Infof("Starting monitor for character %d.", charID)
	monitor := newCharacterMonitor(s.ctx, charID, s.configSvc, s.subSvc)
	s.monitors[charID] = monitor

	s.wg.Add(1) // Add to waitgroup for this monitor
	go func() {
		defer s.wg.Done()
		monitor.run()
	}()
}

func (s *Service) stopMonitor(charID int64) {
	monitor, exists := s.monitors[charID]
	if !exists {
		logger.Sugar.Warnf("No monitor found for character %d to stop.", charID)
		return
	}
	logger.Sugar.Infof("Stopping monitor for character %d.", charID)
	monitor.stop()
	delete(s.monitors, charID)
}
