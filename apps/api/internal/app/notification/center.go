package notification

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
)

var (
	ErrInvalidPage            = errors.New("invalid page")
	ErrInvalidLimit           = errors.New("invalid limit")
	ErrNotificationIDRequired = errors.New("notification id is required")
)

type CenterService struct {
	identityRepository     identity.Repository
	notificationRepository notification.Repository
	now                    func() time.Time
}

type ListNotificationsInput struct {
	UserID     string
	Page       int
	Limit      int
	UnreadOnly bool
}

type ListNotificationsResult struct {
	Data         []notification.Notification
	Page         int
	Limit        int
	TotalPages   int
	TotalRecords int
}

func NewCenterService(
	identityRepository identity.Repository,
	notificationRepository notification.Repository,
) *CenterService {
	return &CenterService{
		identityRepository:     identityRepository,
		notificationRepository: notificationRepository,
		now:                    func() time.Time { return time.Now().UTC() },
	}
}

func (s *CenterService) ListNotifications(
	ctx context.Context,
	input ListNotificationsInput,
) (ListNotificationsResult, error) {
	if s.identityRepository == nil || s.notificationRepository == nil {
		return ListNotificationsResult{}, errors.New("notification center dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(input.UserID)
	if normalizedUserID == "" {
		return ListNotificationsResult{}, identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return ListNotificationsResult{}, fmt.Errorf("get user profile: %w", err)
	}

	page := input.Page
	if page == 0 {
		page = 1
	}
	if page < 1 {
		return ListNotificationsResult{}, ErrInvalidPage
	}

	limit := input.Limit
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 {
		return ListNotificationsResult{}, ErrInvalidLimit
	}

	items, err := s.notificationRepository.ListByUser(ctx, normalizedUserID)
	if err != nil {
		return ListNotificationsResult{}, fmt.Errorf("list notifications by user: %w", err)
	}

	filtered := make([]notification.Notification, 0, len(items))
	for _, item := range items {
		if input.UnreadOnly && item.ReadAt != nil {
			continue
		}
		filtered = append(filtered, item)
	}

	totalRecords := len(filtered)
	totalPages := 0
	if totalRecords > 0 {
		totalPages = (totalRecords + limit - 1) / limit
	}

	start := (page - 1) * limit
	if start >= totalRecords {
		return ListNotificationsResult{
			Data:         []notification.Notification{},
			Page:         page,
			Limit:        limit,
			TotalPages:   totalPages,
			TotalRecords: totalRecords,
		}, nil
	}

	end := start + limit
	if end > totalRecords {
		end = totalRecords
	}

	return ListNotificationsResult{
		Data:         filtered[start:end],
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
		TotalRecords: totalRecords,
	}, nil
}

func (s *CenterService) MarkNotificationRead(
	ctx context.Context,
	userID string,
	notificationID string,
) (notification.Notification, error) {
	if s.identityRepository == nil || s.notificationRepository == nil {
		return notification.Notification{}, errors.New("notification center dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return notification.Notification{}, identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return notification.Notification{}, fmt.Errorf("get user profile: %w", err)
	}

	normalizedNotificationID := strings.TrimSpace(notificationID)
	if normalizedNotificationID == "" {
		return notification.Notification{}, ErrNotificationIDRequired
	}

	updated, err := s.notificationRepository.MarkRead(
		ctx,
		normalizedNotificationID,
		normalizedUserID,
		s.now(),
	)
	if err != nil {
		return notification.Notification{}, fmt.Errorf("mark notification read: %w", err)
	}
	return updated, nil
}
