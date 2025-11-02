package service

import (
	"context"

	"go.uber.org/zap"
)

// Ping ping
func (s *Service) Ping(ctx context.Context) error {
	if err := s.store.Ping(ctx); err != nil {
		s.logging.Error("db error", zap.Error(err))
		return err
	}
	return nil
}
