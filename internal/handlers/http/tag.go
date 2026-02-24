package http

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/pkg/wrapper"
)

type updateTagsRequest struct {
	Tags []string `json:"tags" validate:"required,min=1"`
}

func (h *Handler) UpdateProfileTags(w http.ResponseWriter, r *http.Request) {
	var req updateTagsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	if !ok {
		wrapper.WrapError(w, r, domain.ErrAccessDenied)
		return
	}
	role, ok := r.Context().Value(UserRoleKey).(domain.Role)
	if !ok {
		wrapper.WrapError(w, r, domain.ErrAccessDenied)
		return
	}
	if role != domain.AI {
		wrapper.WrapError(w, r, domain.ErrAccessDenied)
		return
	}

	if err := h.service.UpdateProfileTags(r.Context(), userID, req.Tags); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.Success(w)
}

type getTagsResponse struct {
	Tags []string `json:"tags"`
}

func (h *Handler) GetProfileTags(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	if !ok {
		wrapper.WrapError(w, r, domain.ErrAccessDenied)
		return
	}
	role, ok := r.Context().Value(UserRoleKey).(domain.Role)
	if !ok {
		wrapper.WrapError(w, r, domain.ErrAccessDenied)
		return
	}
	if role != domain.AI {
		wrapper.WrapError(w, r, domain.ErrAccessDenied)
		return
	}

	tags, err := h.service.GetProfileTags(r.Context(), userID)
	if err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.JSON(w, http.StatusOK, getTagsResponse{Tags: tags})
}
