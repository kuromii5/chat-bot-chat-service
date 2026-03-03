package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

func (pg *postgres) CreateRoom(ctx context.Context, humanID uuid.UUID) (*domain.Room, error) {
	var room domain.Room
	if err := pg.DB.GetContext(ctx, &room, createRoomQuery, humanID); err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}
	return &room, nil
}

func (pg *postgres) ClaimRoom(ctx context.Context, roomID uuid.UUID, aiID uuid.UUID) error {
	tx, err := pg.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("tx begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var id uuid.UUID
	if err := tx.GetContext(ctx, &id, claimRoomQuery, aiID, roomID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrRoomAlreadyClaimed
		}
		return fmt.Errorf("claim room: %w", err)
	}

	payload, err := json.Marshal(domain.RoomClaimedPayload{
		RoomID: roomID,
		AiID:   aiID,
	})
	if err != nil {
		return fmt.Errorf("marshal room claimed payload: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		saveOutboxEventQuery,
		"room",
		roomID,
		domain.EventRoomClaimed,
		payload,
	); err != nil {
		return fmt.Errorf("save outbox event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx commit: %w", err)
	}
	return nil
}

func (pg *postgres) CloseRoom(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error {
	var row struct {
		Status        domain.RoomStatus `db:"status"`
		IsParticipant bool              `db:"is_participant"`
	}
	if err := pg.DB.GetContext(ctx, &row, checkRoomQuery, roomID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrRoomNotFound
		}
		return fmt.Errorf("check room: %w", err)
	}
	if !row.IsParticipant {
		return domain.ErrNotRoomParticipant
	}
	if row.Status == domain.RoomClosed {
		return domain.ErrRoomAlreadyClosed
	}

	if _, err := pg.DB.ExecContext(ctx, closeRoomQuery, roomID, userID); err != nil {
		return fmt.Errorf("close room: %w", err)
	}
	return nil
}

func (pg *postgres) GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.Room, error) {
	var room domain.Room
	if err := pg.DB.GetContext(ctx, &room, getRoomQuery, roomID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRoomNotFound
		}
		return nil, fmt.Errorf("get room: %w", err)
	}
	return &room, nil
}
