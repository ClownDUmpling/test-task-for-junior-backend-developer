package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, created_at, updated_at,
		recurrence_type, recurrence_every_n_days, recurrence_month_days,recurrence_dates,
		recurrence_parity, scheduled_date, parent_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, title, description, status, created_at, updated_at,
		recurrence_type, recurrence_every_n_days, recurrence_month_days,
		recurrence_dates, recurrence_parity, scheduled_date, parent_id
	`

	var (
		recurrenceType *string
		everyNdays     *int
		monthDays      []int
		dates          []time.Time
		parity         *string
	)

	if task.Recurrence != nil {
		rt := string(task.Recurrence.Type)
		recurrenceType = &rt
		everyNdays = task.Recurrence.EveryNDays
		monthDays = task.Recurrence.MonthDays
		dates = task.Recurrence.Dates
		parity = task.Recurrence.Parity
	}

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt,
		recurrenceType, everyNdays, monthDays, dates,
		parity, task.ScheduledDate, task.ParentID)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at,
		recurrence_type, recurrence_every_n_days, recurrence_month_days,recurrence_dates,
		recurrence_parity, scheduled_date, parent_id
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4,
			recurrence_type = $5,
			recurrence_every_n_date = $6,
			recurrence_month_days = $7,
			recurrence_dates = $8
			recurrence_parity = $9,
			scheduled_date = $10,
			parent_id = $11
		WHERE id = $12
		RETURNING id, title, description, status, created_at, updated_at, recurrence_type, recurrence_every_n_date,
		recurrence_month_days, recurrence_dates, recurrence_parity,
		scheduled_date, parent_id
	`
	var (
		recurrenceType *string
		everyNdays     *int
		monthDays      []int
		dates          []time.Time
		parity         *string
	)

	if task.Recurrence != nil {
		rt := string(task.Recurrence.Type)
		recurrenceType = &rt
		everyNdays = task.Recurrence.EveryNDays
		monthDays = task.Recurrence.MonthDays
		dates = task.Recurrence.Dates
		parity = task.Recurrence.Parity
	}
	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt,
		recurrenceType, everyNdays, monthDays, dates,
		parity, task.ScheduledDate, task.ParentID)

	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at,
			recurrence_type, recurrence_every_n_days, recurrence_month_days,
			recurrence_dates, recurrence_parity, scheduled_date, parent_id
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repository) ListRecurring(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at,
			recurrence_type, recurrence_every_n_days, recurrence_month_days,
			recurrence_dates, recurrence_parity, scheduled_date, parent_id
		FROM tasks
		WHERE recurrence_type IS NOT NULL AND parent_id IS NULL
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, err
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task           taskdomain.Task
		status         string
		recurrenceType *string
		everyNdays     *int
		monthDays      []int
		dates          []time.Time
		parity         *string
		scheduledDate  *time.Time
		parentID       *int64
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.CreatedAt,
		&task.UpdatedAt,
		&recurrenceType,
		&everyNdays,
		&monthDays,
		&dates,
		&parity,
		&scheduledDate,
		&parentID,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	task.ParentID = parentID
	task.ScheduledDate = scheduledDate

	if recurrenceType != nil {
		task.Recurrence = &taskdomain.Recurrence{
			Type:       taskdomain.RecurrenceType(*recurrenceType),
			EveryNDays: everyNdays,
			MonthDays:  monthDays,
			Dates:      dates,
			Parity:     parity,
			BaseDate:   task.CreatedAt,
		}
	}
	return &task, nil
}

func (r *Repository) ExistsForParentAndDate(ctx context.Context, parentID int64, date time.Time) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM tasks
			WHERE parent_id = $1 AND scheduled_date = $2
		)
	`
	var exists bool
	if err := r.pool.QueryRow(ctx, query, parentID, date).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
