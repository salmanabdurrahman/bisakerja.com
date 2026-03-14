package email

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	notificationapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/notification"
)

type LoggerSender struct {
	logger *slog.Logger
}

func NewLoggerSender(logger *slog.Logger) *LoggerSender {
	if logger == nil {
		logger = slog.Default()
	}
	return &LoggerSender{logger: logger}
}

func (s *LoggerSender) Send(_ context.Context, message notificationapp.EmailMessage) error {
	if strings.TrimSpace(message.To) == "" {
		return errors.New("email recipient is required")
	}
	if strings.TrimSpace(message.Subject) == "" {
		return errors.New("email subject is required")
	}

	s.logger.Info(
		"notification email sent",
		"to", message.To,
		"subject", message.Subject,
	)
	return nil
}
