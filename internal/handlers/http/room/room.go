package room

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/middleware"
	"github.com/kuromii5/chat-bot-shared/wrapper"
)

func (h *Handler) ClaimRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := uuid.Parse(chi.URLParam(r, "roomID"))
	if err != nil {
		wrapper.WrapError(w, r, domain.ErrRoomNotFound)
		return
	}

	aiID, _ := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	if err := h.svc.ClaimRoom(r.Context(), roomID, aiID); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.NoContent(w)
}

func (h *Handler) CloseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := uuid.Parse(chi.URLParam(r, "roomID"))
	if err != nil {
		wrapper.WrapError(w, r, domain.ErrRoomNotFound)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	if err := h.svc.CloseRoom(r.Context(), roomID, userID); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.NoContent(w)
}
