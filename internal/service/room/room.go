package room

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) ClaimRoom(ctx context.Context, roomID uuid.UUID, aiID uuid.UUID) error {
	if err := s.repo.ClaimRoom(ctx, roomID, aiID); err != nil {
		return fmt.Errorf("claim room: %w", err)
	}
	if err := s.binder.BindRoomToAI(ctx, roomID, aiID); err != nil {
		return fmt.Errorf("bind room to AI queue: %w", err)
	}
	return nil
}

func (s *Service) CloseRoom(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error {
	if err := s.repo.CloseRoom(ctx, roomID, userID); err != nil {
		return fmt.Errorf("close room: %w", err)
	}
	return nil
}
