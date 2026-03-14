package notification

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
)

// Matcher represents matcher.
type Matcher struct {
	logger                 *slog.Logger
	jobsRepository         job.Repository
	identityRepository     identity.Repository
	notificationRepository notification.Repository
	queue                  notification.Queue
	batchSize              int
	now                    func() time.Time
}

// MatchSummary summarizes execution details for match.
type MatchSummary struct {
	ProcessedEvents    int
	MatchedUsers       int
	EnqueuedDeliveries int
	DeferredDigest     int
	DuplicateCount     int
	SkippedNonPremium  int
	SkippedNotMatching int
}

// NewMatcher creates a new matcher instance.
func NewMatcher(
	logger *slog.Logger,
	jobsRepository job.Repository,
	identityRepository identity.Repository,
	notificationRepository notification.Repository,
	queue notification.Queue,
	batchSize int,
) *Matcher {
	if logger == nil {
		logger = slog.Default()
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	return &Matcher{
		logger:                 logger,
		jobsRepository:         jobsRepository,
		identityRepository:     identityRepository,
		notificationRepository: notificationRepository,
		queue:                  queue,
		batchSize:              batchSize,
		now:                    func() time.Time { return time.Now().UTC() },
	}
}

// RunOnce runs once.
func (m *Matcher) RunOnce(ctx context.Context) (MatchSummary, error) {
	if m.jobsRepository == nil || m.identityRepository == nil || m.notificationRepository == nil || m.queue == nil {
		return MatchSummary{}, errors.New("matcher dependency is not fully configured")
	}

	events, err := m.queue.DequeueJobEvents(ctx, m.batchSize)
	if err != nil {
		return MatchSummary{}, fmt.Errorf("dequeue job events: %w", err)
	}

	summary := MatchSummary{ProcessedEvents: len(events)}
	for _, event := range events {
		item, getJobErr := m.jobsRepository.GetByID(ctx, event.JobID)
		if getJobErr != nil {
			return summary, fmt.Errorf("get job by id %s: %w", event.JobID, getJobErr)
		}

		users, listUsersErr := m.identityRepository.ListUsers(ctx)
		if listUsersErr != nil {
			return summary, fmt.Errorf("list users: %w", listUsersErr)
		}

		for _, user := range users {
			if !isPremiumActive(user, m.now()) {
				summary.SkippedNonPremium++
				continue
			}

			preferences, getPreferencesErr := m.identityRepository.GetPreferences(ctx, user.ID)
			if getPreferencesErr != nil {
				return summary, fmt.Errorf("get preferences for user %s: %w", user.ID, getPreferencesErr)
			}

			if !matchesPreferences(item, preferences) {
				summary.SkippedNotMatching++
				continue
			}

			created, createErr := m.notificationRepository.CreatePending(ctx, notification.CreateInput{
				UserID:  user.ID,
				JobID:   item.ID,
				Channel: notification.ChannelEmail,
			})
			if createErr != nil {
				if errors.Is(createErr, notification.ErrDuplicateNotification) {
					summary.DuplicateCount++
					continue
				}
				return summary, fmt.Errorf("create pending notification for user %s job %s: %w", user.ID, item.ID, createErr)
			}

			summary.MatchedUsers++
			if shouldDeferToDigest(preferences.AlertMode) {
				summary.DeferredDigest++
				continue
			}
			task := notification.DeliveryTask{
				NotificationID: created.ID,
				UserID:         user.ID,
				UserEmail:      user.Email,
				UserName:       user.Name,
				JobID:          item.ID,
				Channel:        notification.ChannelEmail,
				JobTitle:       item.Title,
				Company:        item.Company,
				Location:       item.Location,
				URL:            item.URL,
			}

			if enqueueErr := m.queue.EnqueueDeliveryTask(ctx, task); enqueueErr != nil {
				return summary, fmt.Errorf("enqueue delivery task for notification %s: %w", created.ID, enqueueErr)
			}
			summary.EnqueuedDeliveries++
		}
	}

	if summary.ProcessedEvents > 0 {
		m.logger.Info(
			"matcher run completed",
			"events", summary.ProcessedEvents,
			"matched_users", summary.MatchedUsers,
			"deliveries_enqueued", summary.EnqueuedDeliveries,
			"deferred_digest", summary.DeferredDigest,
			"duplicates", summary.DuplicateCount,
		)
	}
	return summary, nil
}

func shouldDeferToDigest(mode identity.NotificationAlertMode) bool {
	switch mode {
	case identity.NotificationAlertModeDailyDigest, identity.NotificationAlertModeWeeklyDigest:
		return true
	default:
		return false
	}
}

func isPremiumActive(user identity.User, now time.Time) bool {
	if !user.IsPremium {
		return false
	}
	if user.PremiumExpiredAt == nil {
		return true
	}
	return user.PremiumExpiredAt.After(now)
}

func matchesPreferences(item job.Job, preferences identity.Preferences) bool {
	if !matchesKeywords(item, preferences.Keywords) {
		return false
	}
	if !matchesLocations(item, preferences.Locations) {
		return false
	}
	if !matchesJobTypes(item, preferences.JobTypes) {
		return false
	}
	if !matchesSalary(item, preferences.SalaryMin) {
		return false
	}
	return true
}

func matchesKeywords(item job.Job, keywords []string) bool {
	if len(keywords) == 0 {
		return false
	}
	haystack := strings.ToLower(strings.Join([]string{
		item.Title,
		item.Company,
		item.Description,
	}, " "))
	for _, keyword := range keywords {
		needle := strings.ToLower(strings.TrimSpace(keyword))
		if needle == "" {
			continue
		}
		if strings.Contains(haystack, needle) {
			return true
		}
	}
	return false
}

func matchesLocations(item job.Job, locations []string) bool {
	if len(locations) == 0 {
		return true
	}
	location := strings.ToLower(strings.TrimSpace(item.Location))
	for _, candidate := range locations {
		needle := strings.ToLower(strings.TrimSpace(candidate))
		if needle == "" {
			continue
		}
		if strings.Contains(location, needle) {
			return true
		}
	}
	return false
}

func matchesJobTypes(item job.Job, allowedJobTypes []string) bool {
	if len(allowedJobTypes) == 0 {
		return true
	}
	jobType := strings.ToLower(strings.TrimSpace(extractJobType(item)))
	if jobType == "" {
		return false
	}
	for _, candidate := range allowedJobTypes {
		if jobType == strings.ToLower(strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}

func extractJobType(item job.Job) string {
	if item.RawData == nil {
		return ""
	}
	keys := []string{"job_type", "employment_type", "type"}
	for _, key := range keys {
		rawValue, ok := item.RawData[key]
		if !ok {
			continue
		}
		value, ok := rawValue.(string)
		if !ok {
			continue
		}
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func matchesSalary(item job.Job, requiredSalaryMin int64) bool {
	if requiredSalaryMin <= 0 {
		return true
	}
	if item.SalaryMin == nil {
		return false
	}
	return *item.SalaryMin >= requiredSalaryMin
}
