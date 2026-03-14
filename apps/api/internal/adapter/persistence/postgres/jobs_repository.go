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

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

type JobsRepository struct {
	pool *pgxpool.Pool
}

func NewJobsRepository(pool *pgxpool.Pool) *JobsRepository {
	return &JobsRepository{pool: pool}
}

func (r *JobsRepository) UpsertMany(ctx context.Context, source job.Source, inputs []job.UpsertInput) (job.UpsertResult, error) {
	if source == "" {
		return job.UpsertResult{}, errors.New("source is required")
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return job.UpsertResult{}, fmt.Errorf("begin jobs upsert transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	inserted := make([]job.Job, 0, len(inputs))
	duplicateCount := 0

	insertQuery := `
INSERT INTO jobs (
  source,
  original_job_id,
  title,
  company,
  location,
  description,
  salary_min,
  salary_max,
  salary_range,
  url,
  posted_at,
  raw_data,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12::jsonb, now(), now())
ON CONFLICT (source, original_job_id) DO NOTHING
RETURNING id::text, source, original_job_id, title, COALESCE(company, ''), COALESCE(location, ''), COALESCE(description, ''),
          url, salary_min, salary_max, COALESCE(salary_range, ''), posted_at, created_at, updated_at, raw_data
`

	for _, input := range inputs {
		normalizedInput, normalizeErr := normalizeJobInput(input)
		if normalizeErr != nil {
			return job.UpsertResult{}, normalizeErr
		}

		rawData, encodeErr := encodeJSON(normalizedInput.RawData)
		if encodeErr != nil {
			return job.UpsertResult{}, encodeErr
		}

		insertedJob, scanErr := scanJob(
			tx.QueryRow(
				ctx,
				insertQuery,
				string(source),
				normalizedInput.OriginalJobID,
				normalizedInput.Title,
				normalizedInput.Company,
				normalizedInput.Location,
				normalizedInput.Description,
				nullableInt64(normalizedInput.SalaryMin),
				nullableInt64(normalizedInput.SalaryMax),
				normalizedInput.SalaryRange,
				normalizedInput.URL,
				nullableTime(normalizedInput.PostedAt),
				rawData,
			),
		)
		if scanErr != nil {
			if errors.Is(scanErr, pgx.ErrNoRows) {
				duplicateCount++
				continue
			}
			return job.UpsertResult{}, fmt.Errorf("insert job %s/%s: %w", source, normalizedInput.OriginalJobID, scanErr)
		}

		inserted = append(inserted, insertedJob)
	}

	if err := tx.Commit(ctx); err != nil {
		return job.UpsertResult{}, fmt.Errorf("commit jobs upsert transaction: %w", err)
	}
	committed = true

	return job.UpsertResult{
		Inserted:       inserted,
		InsertedCount:  len(inserted),
		DuplicateCount: duplicateCount,
	}, nil
}

func (r *JobsRepository) Search(ctx context.Context, query job.SearchQuery) (job.SearchResult, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}

	sortBy := strings.TrimSpace(query.Sort)
	if sortBy == "" {
		sortBy = "-posted_at"
	}

	conditions := make([]string, 0, 4)
	args := make([]any, 0, 6)

	if query.Q != "" {
		needle := "%" + strings.ToLower(strings.TrimSpace(query.Q)) + "%"
		conditions = append(
			conditions,
			fmt.Sprintf("(LOWER(title) LIKE $%d OR LOWER(company) LIKE $%d OR LOWER(description) LIKE $%d)", len(args)+1, len(args)+1, len(args)+1),
		)
		args = append(args, needle)
	}

	if query.Location != "" {
		needle := "%" + strings.ToLower(strings.TrimSpace(query.Location)) + "%"
		conditions = append(conditions, fmt.Sprintf("LOWER(location) LIKE $%d", len(args)+1))
		args = append(args, needle)
	}

	if query.Source != "" {
		conditions = append(conditions, fmt.Sprintf("source = $%d", len(args)+1))
		args = append(args, string(query.Source))
	}

	if query.HasSalaryMin {
		conditions = append(conditions, fmt.Sprintf("salary_min IS NOT NULL AND salary_min >= $%d", len(args)+1))
		args = append(args, query.SalaryMin)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM jobs" + whereClause
	var totalRecords int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalRecords); err != nil {
		return job.SearchResult{}, fmt.Errorf("count jobs: %w", err)
	}

	totalPages := 0
	if totalRecords > 0 {
		totalPages = (totalRecords + limit - 1) / limit
	}

	offset := (page - 1) * limit
	if offset >= totalRecords {
		return job.SearchResult{
			Data:         []job.Job{},
			Page:         page,
			Limit:        limit,
			TotalPages:   totalPages,
			TotalRecords: totalRecords,
		}, nil
	}

	orderBy, err := mapJobSort(sortBy)
	if err != nil {
		return job.SearchResult{}, err
	}

	queryArgs := append(append([]any(nil), args...), limit, offset)
	searchQuery := `
SELECT id::text, source, original_job_id, title, COALESCE(company, ''), COALESCE(location, ''), COALESCE(description, ''),
       url, salary_min, salary_max, COALESCE(salary_range, ''), posted_at, created_at, updated_at, raw_data
FROM jobs` + whereClause + ` ORDER BY ` + orderBy +
		fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)

	rows, err := r.pool.Query(ctx, searchQuery, queryArgs...)
	if err != nil {
		return job.SearchResult{}, fmt.Errorf("search jobs: %w", err)
	}
	defer rows.Close()

	items := make([]job.Job, 0, limit)
	for rows.Next() {
		item, scanErr := scanJob(rows)
		if scanErr != nil {
			return job.SearchResult{}, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return job.SearchResult{}, fmt.Errorf("search jobs rows: %w", err)
	}

	return job.SearchResult{
		Data:         items,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
		TotalRecords: totalRecords,
	}, nil
}

func (r *JobsRepository) GetByID(ctx context.Context, id string) (job.Job, error) {
	normalizedID := strings.TrimSpace(id)
	if normalizedID == "" {
		return job.Job{}, job.ErrNotFound
	}

	query := `
SELECT id::text, source, original_job_id, title, COALESCE(company, ''), COALESCE(location, ''), COALESCE(description, ''),
       url, salary_min, salary_max, COALESCE(salary_range, ''), posted_at, created_at, updated_at, raw_data
FROM jobs
WHERE id::text = $1
`

	item, err := scanJob(r.pool.QueryRow(ctx, query, normalizedID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return job.Job{}, job.ErrNotFound
		}
		return job.Job{}, fmt.Errorf("get job by id: %w", err)
	}
	return item, nil
}

func (r *JobsRepository) RecordScrapeRun(ctx context.Context, run job.ScrapeRun) error {
	startedAt := run.StartedAt.UTC()
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}

	finishedAt := run.FinishedAt.UTC()
	if finishedAt.IsZero() {
		finishedAt = startedAt
	}

	query := `
INSERT INTO scrape_runs (
  source,
  status,
  error_class,
  error_message,
  fetched_count,
  inserted_count,
  duplicate_count,
  started_at,
  finished_at,
  created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
`

	_, err := r.pool.Exec(
		ctx,
		query,
		string(run.Source),
		string(run.Status),
		strings.TrimSpace(run.ErrorClass),
		strings.TrimSpace(run.ErrorMessage),
		run.FetchedCount,
		run.InsertedCount,
		run.DuplicateCount,
		startedAt,
		finishedAt,
	)
	if err != nil {
		return fmt.Errorf("record scrape run: %w", err)
	}
	return nil
}

func normalizeJobInput(input job.UpsertInput) (job.UpsertInput, error) {
	input.OriginalJobID = strings.TrimSpace(input.OriginalJobID)
	input.Title = strings.TrimSpace(input.Title)
	input.Company = strings.TrimSpace(input.Company)
	input.Location = strings.TrimSpace(input.Location)
	input.Description = strings.TrimSpace(input.Description)
	input.URL = strings.TrimSpace(input.URL)
	input.SalaryRange = strings.TrimSpace(input.SalaryRange)
	if input.RawData == nil {
		input.RawData = map[string]any{}
	}

	if input.OriginalJobID == "" {
		return job.UpsertInput{}, errors.New("original_job_id is required")
	}
	if input.Title == "" {
		return job.UpsertInput{}, errors.New("title is required")
	}
	if input.URL == "" {
		return job.UpsertInput{}, errors.New("url is required")
	}

	return input, nil
}

func mapJobSort(sortBy string) (string, error) {
	switch sortBy {
	case "posted_at":
		return "COALESCE(posted_at, created_at) ASC, id ASC", nil
	case "-created_at":
		return "created_at DESC, id DESC", nil
	case "created_at":
		return "created_at ASC, id ASC", nil
	case "-posted_at":
		fallthrough
	default:
		return "COALESCE(posted_at, created_at) DESC, id DESC", nil
	}
}

type jobScanner interface {
	Scan(dest ...any) error
}

func scanJob(scanner jobScanner) (job.Job, error) {
	var (
		item        job.Job
		source      string
		salaryMin   sql.NullInt64
		salaryMax   sql.NullInt64
		postedAt    sql.NullTime
		rawDataJSON []byte
	)

	err := scanner.Scan(
		&item.ID,
		&source,
		&item.OriginalJobID,
		&item.Title,
		&item.Company,
		&item.Location,
		&item.Description,
		&item.URL,
		&salaryMin,
		&salaryMax,
		&item.SalaryRange,
		&postedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&rawDataJSON,
	)
	if err != nil {
		return job.Job{}, err
	}

	item.Source = job.Source(source)
	if salaryMin.Valid {
		value := salaryMin.Int64
		item.SalaryMin = &value
	}
	if salaryMax.Valid {
		value := salaryMax.Int64
		item.SalaryMax = &value
	}
	if postedAt.Valid {
		value := postedAt.Time.UTC()
		item.PostedAt = &value
	}
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()

	rawData, err := decodeJSON(rawDataJSON)
	if err != nil {
		return job.Job{}, err
	}
	item.RawData = rawData

	return item, nil
}

var _ job.Repository = (*JobsRepository)(nil)
