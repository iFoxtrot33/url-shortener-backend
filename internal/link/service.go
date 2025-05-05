package link

import (
	"time"

	"github.com/rs/zerolog"
)

type LinkService struct {
	Repository *LinkRepository
	Logger     *zerolog.Logger
	stopChan   chan struct{}
}

func NewLinkService(repository *LinkRepository, logger *zerolog.Logger) *LinkService {
	return &LinkService{
		Repository: repository,
		Logger:     logger,
		stopChan:   make(chan struct{}),
	}
}

func (s *LinkService) Start() {
	s.Logger.Info().Msg("Starting link service")

	go s.runLifetimeManager()
}

func (s *LinkService) Stop() {
	s.Logger.Info().Msg("Stopping link service")
	close(s.stopChan)
}

func (s *LinkService) runLifetimeManager() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processLifetimeUpdate()
		case <-s.stopChan:
			s.Logger.Info().Msg("Lifetime manager stopped")
			return
		}
	}
}

func (s *LinkService) processLifetimeUpdate() {
	s.Logger.Info().Msg("Processing lifetime updates for links")

	err := s.Repository.DecrementLifetimeForAllLinks()
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to decrement lifetime for links")
		return
	}

	err = s.Repository.DeleteExpiredLinks()
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to delete expired links")
		return
	}

	s.Logger.Info().Msg("Lifetime update completed successfully")
}
