package service

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/vnykmshr/nivo/services/notification/internal/models"
	"github.com/vnykmshr/nivo/services/notification/internal/repository"
	"github.com/vnykmshr/nivo/shared/errors"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// NotificationService handles notification business logic.
type NotificationService struct {
	notifRepo      *repository.NotificationRepository
	templateRepo   *repository.TemplateRepository
	templateEngine *TemplateEngine
	simEngine      *SimulationEngine
}

// NewNotificationService creates a new notification service.
func NewNotificationService(
	notifRepo *repository.NotificationRepository,
	templateRepo *repository.TemplateRepository,
	simConfig SimulationConfig,
) *NotificationService {
	service := &NotificationService{
		notifRepo:      notifRepo,
		templateRepo:   templateRepo,
		templateEngine: NewTemplateEngine(),
	}

	// Initialize simulation engine with the repository
	service.simEngine = NewSimulationEngine(simConfig, notifRepo)

	return service
}

// SendNotification creates and queues a notification for delivery.
func (s *NotificationService) SendNotification(ctx context.Context, req *models.SendNotificationRequest) (*models.SendNotificationResponse, *errors.Error) {
	// Check for duplicate notification using correlation_id
	if req.CorrelationID != nil && *req.CorrelationID != "" {
		existing, err := s.notifRepo.GetByCorrelationID(ctx, *req.CorrelationID)
		if err == nil && existing != nil {
			// Idempotent: return existing notification
			log.Printf("[notification] Duplicate notification request detected (correlation_id=%s), returning existing notification %s",
				*req.CorrelationID, existing.ID)
			return &models.SendNotificationResponse{
				NotificationID: existing.ID,
				Status:         existing.Status,
				QueuedAt:       existing.QueuedAt,
			}, nil
		}
	}

	// Prepare notification
	var subject, body string
	var templateID *string

	// If template is specified, render it
	if req.TemplateID != nil && *req.TemplateID != "" {
		template, err := s.templateRepo.GetByID(ctx, *req.TemplateID)
		if err != nil {
			return nil, err
		}

		// Render subject and body
		if template.SubjectTemplate != "" {
			subject, _ = s.templateEngine.Render(template.SubjectTemplate, req.Variables)
		}
		body, _ = s.templateEngine.Render(template.BodyTemplate, req.Variables)
		templateID = &template.ID
	} else {
		// Use provided subject and body
		subject = req.Subject
		body = req.Body
	}

	// Validate body is not empty
	if body == "" {
		return nil, errors.Validation("notification body cannot be empty")
	}

	// Get metadata
	metadata, err := req.GetMetadata()
	if err != nil {
		return nil, errors.Validation("invalid metadata JSON")
	}

	// Set default priority if not provided
	priority := req.Priority
	if priority == "" {
		priority = models.PriorityNormal
	}

	// Determine source service from context or default
	sourceService := "notification" // Default
	if val := ctx.Value("source_service"); val != nil {
		if src, ok := val.(string); ok {
			sourceService = src
		}
	}

	// Create notification
	notif := &models.Notification{
		ID:            uuid.New().String(),
		UserID:        req.UserID,
		Channel:       req.Channel,
		Type:          req.Type,
		Priority:      priority,
		Recipient:     req.Recipient,
		Subject:       subject,
		Body:          body,
		TemplateID:    templateID,
		Status:        models.StatusQueued,
		CorrelationID: req.CorrelationID,
		SourceService: sourceService,
		Metadata:      metadata,
		RetryCount:    0,
		QueuedAt:      sharedModels.Now(),
		CreatedAt:     sharedModels.Now(),
		UpdatedAt:     sharedModels.Now(),
	}

	// Save to database
	if err := s.notifRepo.Create(ctx, notif); err != nil {
		return nil, err
	}

	log.Printf("[notification] Created notification %s (type=%s, channel=%s, recipient=%s, priority=%s)",
		notif.ID, notif.Type, notif.Channel, notif.Recipient, notif.Priority)

	return &models.SendNotificationResponse{
		NotificationID: notif.ID,
		Status:         notif.Status,
		QueuedAt:       notif.QueuedAt,
	}, nil
}

// GetNotification retrieves a notification by ID.
func (s *NotificationService) GetNotification(ctx context.Context, id string) (*models.Notification, *errors.Error) {
	return s.notifRepo.GetByID(ctx, id)
}

// ListNotifications retrieves notifications with optional filters.
func (s *NotificationService) ListNotifications(ctx context.Context, req *models.ListNotificationsRequest) (*models.ListNotificationsResponse, *errors.Error) {
	notifications, total, err := s.notifRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	return &models.ListNotificationsResponse{
		Notifications: notifications,
		Total:         total,
		Limit:         req.Limit,
		Offset:        req.Offset,
	}, nil
}

// GetStats retrieves notification statistics.
func (s *NotificationService) GetStats(ctx context.Context) (*models.NotificationStats, *errors.Error) {
	return s.notifRepo.GetStats(ctx)
}

// ProcessQueuedNotifications processes queued notifications (called by background worker).
func (s *NotificationService) ProcessQueuedNotifications(ctx context.Context, batchSize int) *errors.Error {
	// Get queued notifications
	notifications, err := s.notifRepo.GetQueuedNotifications(ctx, batchSize)
	if err != nil {
		return err
	}

	if len(notifications) == 0 {
		return nil
	}

	log.Printf("[notification] Processing %d queued notifications", len(notifications))

	// Process each notification
	for _, notif := range notifications {
		// Process in goroutine for concurrency (in production, use worker pool)
		go func(n *models.Notification) {
			if err := s.simEngine.ProcessNotification(ctx, n); err != nil {
				log.Printf("[notification] Error processing notification %s: %v", n.ID, err)
			}
		}(notif)
	}

	return nil
}

// CreateTemplate creates a new notification template.
func (s *NotificationService) CreateTemplate(ctx context.Context, req *models.CreateTemplateRequest) (*models.NotificationTemplate, *errors.Error) {
	metadata, err := req.GetMetadata()
	if err != nil {
		return nil, errors.Validation("invalid metadata JSON")
	}

	template := &models.NotificationTemplate{
		ID:              uuid.New().String(),
		Name:            req.Name,
		Channel:         req.Channel,
		SubjectTemplate: req.SubjectTemplate,
		BodyTemplate:    req.BodyTemplate,
		Version:         1,
		Metadata:        metadata,
		CreatedAt:       sharedModels.Now(),
		UpdatedAt:       sharedModels.Now(),
	}

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, err
	}

	log.Printf("[notification] Created template %s (name=%s, channel=%s)", template.ID, template.Name, template.Channel)
	return template, nil
}

// GetTemplate retrieves a template by ID.
func (s *NotificationService) GetTemplate(ctx context.Context, id string) (*models.NotificationTemplate, *errors.Error) {
	return s.templateRepo.GetByID(ctx, id)
}

// ListTemplates retrieves all templates, optionally filtered by channel.
func (s *NotificationService) ListTemplates(ctx context.Context, channel *models.NotificationChannel) ([]*models.NotificationTemplate, *errors.Error) {
	return s.templateRepo.List(ctx, channel)
}

// UpdateTemplate updates an existing template.
func (s *NotificationService) UpdateTemplate(ctx context.Context, id string, req *models.UpdateTemplateRequest) *errors.Error {
	return s.templateRepo.Update(ctx, id, req)
}

// PreviewTemplate renders a template with provided variables (for testing).
func (s *NotificationService) PreviewTemplate(ctx context.Context, templateID string, variables map[string]interface{}) (*models.PreviewTemplateResponse, *errors.Error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	var subject string
	var subjectVars []string

	if template.SubjectTemplate != "" {
		subject, subjectVars = s.templateEngine.Render(template.SubjectTemplate, variables)
	}

	body, bodyVars := s.templateEngine.Render(template.BodyTemplate, variables)

	// Combine used variables
	allVars := append(subjectVars, bodyVars...)

	return &models.PreviewTemplateResponse{
		Subject:      subject,
		Body:         body,
		RenderedAt:   sharedModels.Now(),
		VariableUsed: allVars,
	}, nil
}

// ReplayNotification re-queues a failed or delivered notification for testing.
func (s *NotificationService) ReplayNotification(ctx context.Context, id string) *errors.Error {
	// Check if notification exists
	_, err := s.notifRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Reset to queued status
	if err := s.notifRepo.UpdateStatus(ctx, id, models.StatusQueued, nil); err != nil {
		return err
	}

	log.Printf("[notification] Replayed notification %s", id)
	return nil
}
