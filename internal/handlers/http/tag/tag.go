package tag

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/middleware"
	"github.com/kuromii5/chat-bot-shared/validator"
	"github.com/kuromii5/chat-bot-shared/wrapper"
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
	if err := validator.Validate(req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	if err := h.svc.UpdateProfileTags(r.Context(), userID, req.Tags); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.Success(w)
}

type getTagsResponse struct {
	Tags []string `json:"tags"`
}

func (h *Handler) GetProfileTags(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	tags, err := h.svc.GetProfileTags(r.Context(), userID)
	if err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.JSON(w, http.StatusOK, getTagsResponse{Tags: tags})
}
