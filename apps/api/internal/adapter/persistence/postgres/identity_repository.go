package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
)

type IdentityRepository struct {
	pool *pgxpool.Pool
}

func NewIdentityRepository(pool *pgxpool.Pool) *IdentityRepository {
	return &IdentityRepository{pool: pool}
}

func (r *IdentityRepository) CreateUser(ctx context.Context, input identity.CreateUserInput) (identity.User, error) {
	normalizedEmail := identity.NormalizeEmail(input.Email)
	if normalizedEmail == "" {
		return identity.User{}, fmt.Errorf("create user: email is required")
	}
	if strings.TrimSpace(input.PasswordHash) == "" {
		return identity.User{}, fmt.Errorf("create user: password hash is required")
	}
	if strings.TrimSpace(input.Name) == "" {
		return identity.User{}, fmt.Errorf("create user: name is required")
	}

	role := input.Role
	if role == "" {
		role = identity.RoleUser
	}

	query := `
INSERT INTO users (email, password_hash, name, role, is_premium, premium_expired_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, now(), now())
RETURNING id::text, email, password_hash, name, role, is_premium, premium_expired_at, created_at, updated_at
`

	user, err := scanUser(
		r.pool.QueryRow(
			ctx,
			query,
			normalizedEmail,
			strings.TrimSpace(input.PasswordHash),
			strings.TrimSpace(input.Name),
			string(role),
			input.IsPremium,
			nullableTime(input.PremiumExpiredAt),
		),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return identity.User{}, identity.ErrEmailAlreadyRegistered
		}
		return identity.User{}, err
	}

	return user, nil
}

func (r *IdentityRepository) GetUserByID(ctx context.Context, userID string) (identity.User, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return identity.User{}, identity.ErrUserNotFound
	}

	query := `
SELECT id::text, email, password_hash, name, role, is_premium, premium_expired_at, created_at, updated_at
FROM users
WHERE id::text = $1
`

	user, err := scanUser(r.pool.QueryRow(ctx, query, normalizedUserID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.User{}, identity.ErrUserNotFound
		}
		return identity.User{}, err
	}
	return user, nil
}

func (r *IdentityRepository) GetUserByEmail(ctx context.Context, email string) (identity.User, error) {
	normalizedEmail := identity.NormalizeEmail(email)
	if normalizedEmail == "" {
		return identity.User{}, identity.ErrUserNotFound
	}

	query := `
SELECT id::text, email, password_hash, name, role, is_premium, premium_expired_at, created_at, updated_at
FROM users
WHERE email = $1
`

	user, err := scanUser(r.pool.QueryRow(ctx, query, normalizedEmail))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.User{}, identity.ErrUserNotFound
		}
		return identity.User{}, err
	}
	return user, nil
}

func (r *IdentityRepository) UpdatePremiumStatus(
	ctx context.Context,
	userID string,
	isPremium bool,
	premiumExpiredAt *time.Time,
) (identity.User, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return identity.User{}, fmt.Errorf("update premium status: user id is required")
	}

	query := `
UPDATE users
SET is_premium = $2, premium_expired_at = $3, updated_at = now()
WHERE id::text = $1
RETURNING id::text, email, password_hash, name, role, is_premium, premium_expired_at, created_at, updated_at
`

	updated, err := scanUser(
		r.pool.QueryRow(ctx, query, normalizedUserID, isPremium, nullableTime(premiumExpiredAt)),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.User{}, identity.ErrUserNotFound
		}
		return identity.User{}, err
	}
	return updated, nil
}

func (r *IdentityRepository) ListUsers(ctx context.Context) ([]identity.User, error) {
	query := `
SELECT id::text, email, password_hash, name, role, is_premium, premium_expired_at, created_at, updated_at
FROM users
ORDER BY created_at DESC, id DESC
`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	result := make([]identity.User, 0)
	for rows.Next() {
		item, scanErr := scanUser(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users rows: %w", err)
	}

	return result, nil
}

func (r *IdentityRepository) GetPreferences(ctx context.Context, userID string) (identity.Preferences, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return identity.Preferences{}, identity.ErrUserNotFound
	}

	// Ownership validation first: user must exist.
	if _, err := r.GetUserByID(ctx, normalizedUserID); err != nil {
		return identity.Preferences{}, err
	}

	query := `
SELECT keywords, locations, job_types, salary_min, alert_mode, digest_hour, updated_at
FROM user_preferences
WHERE user_id::text = $1
`

	var (
		keywords   []string
		locations  []string
		jobTypes   []string
		salaryMin  int64
		alertMode  string
		digestHour sql.NullInt32
		updatedAt  sql.NullTime
	)

	err := r.pool.QueryRow(ctx, query, normalizedUserID).Scan(
		&keywords,
		&locations,
		&jobTypes,
		&salaryMin,
		&alertMode,
		&digestHour,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.Preferences{
				UserID:     normalizedUserID,
				Keywords:   []string{},
				Locations:  []string{},
				JobTypes:   []string{},
				SalaryMin:  0,
				AlertMode:  identity.NotificationAlertModeInstant,
				DigestHour: nil,
				UpdatedAt:  nil,
			}, nil
		}
		return identity.Preferences{}, fmt.Errorf("get preferences: %w", err)
	}

	var digestHourPtr *int
	if digestHour.Valid {
		value := int(digestHour.Int32)
		digestHourPtr = &value
	}

	var updatedAtPtr *time.Time
	if updatedAt.Valid {
		value := updatedAt.Time.UTC()
		updatedAtPtr = &value
	}

	mode := identity.NotificationAlertMode(strings.TrimSpace(alertMode))
	if mode == "" {
		mode = identity.NotificationAlertModeInstant
	}

	return identity.Preferences{
		UserID:     normalizedUserID,
		Keywords:   append([]string(nil), keywords...),
		Locations:  append([]string(nil), locations...),
		JobTypes:   append([]string(nil), jobTypes...),
		SalaryMin:  salaryMin,
		AlertMode:  mode,
		DigestHour: digestHourPtr,
		UpdatedAt:  updatedAtPtr,
	}, nil
}

func (r *IdentityRepository) SavePreferences(ctx context.Context, preferences identity.Preferences) (identity.Preferences, error) {
	normalizedUserID := strings.TrimSpace(preferences.UserID)
	if normalizedUserID == "" {
		return identity.Preferences{}, fmt.Errorf("save preferences: user id is required")
	}
	if preferences.UpdatedAt == nil {
		return identity.Preferences{}, fmt.Errorf("save preferences: updated_at is required")
	}

	mode := preferences.AlertMode
	if mode == "" {
		mode = identity.NotificationAlertModeInstant
	}

	query := `
WITH selected_user AS (
  SELECT id FROM users WHERE id::text = $1
)
INSERT INTO user_preferences (
  user_id,
  keywords,
  locations,
  job_types,
  salary_min,
  alert_mode,
  digest_hour,
  updated_at
)
SELECT
  selected_user.id,
  $2::text[],
  $3::text[],
  $4::text[],
  $5,
  $6,
  $7,
  $8
FROM selected_user
ON CONFLICT (user_id)
DO UPDATE SET
  keywords = EXCLUDED.keywords,
  locations = EXCLUDED.locations,
  job_types = EXCLUDED.job_types,
  salary_min = EXCLUDED.salary_min,
  alert_mode = EXCLUDED.alert_mode,
  digest_hour = EXCLUDED.digest_hour,
  updated_at = EXCLUDED.updated_at
RETURNING user_id::text, keywords, locations, job_types, salary_min, alert_mode, digest_hour, updated_at
`

	var (
		userID     string
		keywords   []string
		locations  []string
		jobTypes   []string
		salaryMin  int64
		alertMode  string
		digestHour sql.NullInt32
		updatedAt  time.Time
	)

	err := r.pool.QueryRow(
		ctx,
		query,
		normalizedUserID,
		append([]string(nil), preferences.Keywords...),
		append([]string(nil), preferences.Locations...),
		append([]string(nil), preferences.JobTypes...),
		preferences.SalaryMin,
		string(mode),
		nullableInt(preferences.DigestHour),
		preferences.UpdatedAt.UTC(),
	).Scan(
		&userID,
		&keywords,
		&locations,
		&jobTypes,
		&salaryMin,
		&alertMode,
		&digestHour,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.Preferences{}, identity.ErrUserNotFound
		}
		return identity.Preferences{}, fmt.Errorf("save preferences: %w", err)
	}

	var digestHourPtr *int
	if digestHour.Valid {
		value := int(digestHour.Int32)
		digestHourPtr = &value
	}

	updatedAtUTC := updatedAt.UTC()
	return identity.Preferences{
		UserID:     userID,
		Keywords:   append([]string(nil), keywords...),
		Locations:  append([]string(nil), locations...),
		JobTypes:   append([]string(nil), jobTypes...),
		SalaryMin:  salaryMin,
		AlertMode:  identity.NotificationAlertMode(alertMode),
		DigestHour: digestHourPtr,
		UpdatedAt:  &updatedAtUTC,
	}, nil
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(scanner userScanner) (identity.User, error) {
	var (
		result      identity.User
		role        string
		premiumTime sql.NullTime
	)

	err := scanner.Scan(
		&result.ID,
		&result.Email,
		&result.PasswordHash,
		&result.Name,
		&role,
		&result.IsPremium,
		&premiumTime,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return identity.User{}, err
	}

	result.Role = identity.Role(role)
	if premiumTime.Valid {
		value := premiumTime.Time.UTC()
		result.PremiumExpiredAt = &value
	}
	result.CreatedAt = result.CreatedAt.UTC()
	result.UpdatedAt = result.UpdatedAt.UTC()

	return result, nil
}

var _ identity.Repository = (*IdentityRepository)(nil)
