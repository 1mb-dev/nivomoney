package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/vnykmshr/gopantic/pkg/model"
	"github.com/vnykmshr/nivo/services/notification/internal/models"
	"github.com/vnykmshr/nivo/services/notification/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/response"
)

// NotificationHandler handles notification HTTP requests.
type NotificationHandler struct {
	notifService *service.NotificationService
}

// NewNotificationHandler creates a new notification handler.
func NewNotificationHandler(notifService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notifService: notifService,
	}
}

// SendNotification handles sending a new notification.
// POST /v1/notifications/send
func (h *NotificationHandler) SendNotification(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	req, err := model.ParseInto[models.SendNotificationRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	resp, svcErr := h.notifService.SendNotification(r.Context(), &req)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.Created(w, resp)
}

// GetNotification retrieves a notification by ID.
// GET /v1/notifications/{id}
func (h *NotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		response.Error(w, errors.BadRequest("notification id is required"))
		return
	}

	notif, svcErr := h.notifService.GetNotification(r.Context(), id)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, notif)
}

// ListNotifications retrieves notifications with filters.
// GET /v1/notifications
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req := &models.ListNotificationsRequest{}

	// Parse filters from query params
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		req.UserID = &userID
	}

	if channel := r.URL.Query().Get("channel"); channel != "" {
		ch := models.NotificationChannel(channel)
		req.Channel = &ch
	}

	if notifType := r.URL.Query().Get("type"); notifType != "" {
		nt := models.NotificationType(notifType)
		req.Type = &nt
	}

	if status := r.URL.Query().Get("status"); status != "" {
		st := models.NotificationStatus(status)
		req.Status = &st
	}

	if source := r.URL.Query().Get("source_service"); source != "" {
		req.SourceService = &source
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			req.Limit = parsed
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			req.Offset = parsed
		}
	}

	resp, svcErr := h.notifService.ListNotifications(r.Context(), req)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, resp)
}

// CreateTemplate creates a new notification template.
// POST /v1/templates
func (h *NotificationHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	req, err := model.ParseInto[models.CreateTemplateRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	template, svcErr := h.notifService.CreateTemplate(r.Context(), &req)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.Created(w, template)
}

// GetTemplate retrieves a template by ID.
// GET /v1/templates/{id}
func (h *NotificationHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		response.Error(w, errors.BadRequest("template id is required"))
		return
	}

	template, svcErr := h.notifService.GetTemplate(r.Context(), id)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, template)
}

// ListTemplates retrieves all templates.
// GET /v1/templates
func (h *NotificationHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	var channel *models.NotificationChannel
	if ch := r.URL.Query().Get("channel"); ch != "" {
		c := models.NotificationChannel(ch)
		channel = &c
	}

	templates, svcErr := h.notifService.ListTemplates(r.Context(), channel)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, templates)
}

// UpdateTemplate updates an existing template.
// PUT /v1/templates/{id}
func (h *NotificationHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		response.Error(w, errors.BadRequest("template id is required"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	req, err := model.ParseInto[models.UpdateTemplateRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	if svcErr := h.notifService.UpdateTemplate(r.Context(), id, &req); svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.NoContent(w)
}

// PreviewTemplate previews a template with variables.
// POST /v1/templates/{id}/preview
func (h *NotificationHandler) PreviewTemplate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		response.Error(w, errors.BadRequest("template id is required"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	req, err := model.ParseInto[models.PreviewTemplateRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	preview, svcErr := h.notifService.PreviewTemplate(r.Context(), id, req.Variables)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, preview)
}

// GetStats retrieves notification statistics.
// GET /admin/notifications/stats
func (h *NotificationHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, svcErr := h.notifService.GetStats(r.Context())
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, stats)
}

// ReplayNotification re-queues a notification for replay/testing.
// POST /admin/notifications/{id}/replay
func (h *NotificationHandler) ReplayNotification(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		response.Error(w, errors.BadRequest("notification id is required"))
		return
	}

	if svcErr := h.notifService.ReplayNotification(r.Context(), id); svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.NoContent(w)
}

// Health check endpoint.
// GET /health
func (h *NotificationHandler) Health(w http.ResponseWriter, r *http.Request) {
	response.OK(w, map[string]string{
		"status":  "healthy",
		"service": "notification",
	})
}
