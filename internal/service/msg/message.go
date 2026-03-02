package msg

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

type CreateMessageReq struct {
	UserID  uuid.UUID
	Role    domain.Role
	Content string
	Tags    []string  // Human: required; AI: ignored
	RoomID  uuid.UUID // AI: required (non-zero); Human: ignored
}

func (s *Service) SendMessage(ctx context.Context, req CreateMessageReq) (*domain.Message, error) {
	switch req.Role {
	case domain.Human:
		return s.sendHumanMessage(ctx, req)
	case domain.AI:
		return s.sendAIMessage(ctx, req)
	default:
		return nil, domain.ErrAccessDenied
	}
}

func (s *Service) sendHumanMessage(ctx context.Context, req CreateMessageReq) (*domain.Message, error) {
	if req.RoomID == (uuid.UUID{}) {
		return s.sendHumanNewQuestion(ctx, req)
	}
	return s.sendHumanFollowUp(ctx, req)
}

func (s *Service) sendHumanNewQuestion(ctx context.Context, req CreateMessageReq) (*domain.Message, error) {
	if len(req.Tags) == 0 {
		return nil, domain.ErrInvalidTags
	}
	slices.Sort(req.Tags)
	req.Tags = slices.Compact(req.Tags)

	room, err := s.roomRepo.CreateRoom(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}

	saved, err := s.repo.SaveWithOutbox(ctx, &domain.Message{
		SenderID:   req.UserID,
		SenderRole: domain.Human,
		RoomID:     room.ID,
		Content:    req.Content,
		Tags:       req.Tags,
	}, domain.EventNewQuestion, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("save message: %w", err)
	}

	return saved, nil
}

func (s *Service) sendHumanFollowUp(ctx context.Context, req CreateMessageReq) (*domain.Message, error) {
	room, err := s.roomRepo.GetRoom(ctx, req.RoomID)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}
	if room.Status != domain.RoomActive {
		return nil, domain.ErrRoomNotActive
	}
	if room.HumanID != req.UserID {
		return nil, domain.ErrNotRoomParticipant
	}

	saved, err := s.repo.SaveWithOutbox(ctx, &domain.Message{
		SenderID:   req.UserID,
		SenderRole: domain.Human,
		RoomID:     req.RoomID,
		Content:    req.Content,
		Tags:       pq.StringArray{},
	}, domain.EventFollowUp, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("save message: %w", err)
	}

	return saved, nil
}

func (s *Service) sendAIMessage(ctx context.Context, req CreateMessageReq) (*domain.Message, error) {
	if req.RoomID == (uuid.UUID{}) {
		return nil, domain.ErrRoomRequired
	}

	room, err := s.roomRepo.GetRoom(ctx, req.RoomID)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}
	if room.Status != domain.RoomActive {
		return nil, domain.ErrRoomNotActive
	}
	if room.AIID == nil || *room.AIID != req.UserID {
		return nil, domain.ErrNotRoomParticipant
	}

	lastMsgs, err := s.repo.GetLastMessages(ctx, req.RoomID, domain.AISequentialMessageLimit)
	if err != nil {
		return nil, fmt.Errorf("get last messages: %w", err)
	}
	if err := domain.ValidateAIMsg(lastMsgs); err != nil {
		return nil, fmt.Errorf("validate AI msg: %w", err)
	}

	saved, err := s.repo.SaveWithOutbox(ctx, &domain.Message{
		SenderID:   req.UserID,
		SenderRole: domain.AI,
		RoomID:     req.RoomID,
		Content:    req.Content,
		Tags:       pq.StringArray{},
	}, domain.EventAIReply, room.HumanID)
	if err != nil {
		return nil, fmt.Errorf("save message: %w", err)
	}

	return saved, nil
}
