package scheduler

import (
	"context"
	"log/slog"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type TaskRepository interface {
	ListRecurring(ctx context.Context) ([]taskdomain.Task, error)
	ExistsForParentAndDate(ctx context.Context, parentID int64, date time.Time) (bool, error)
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
}

type Scheduler struct {
	repo   TaskRepository
	logger *slog.Logger
}

func New(repo TaskRepository, logger *slog.Logger) *Scheduler {
	return &Scheduler{repo: repo, logger: logger}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	s.createTasksForToday(ctx)

	for {
		select {
		case <-ticker.C:
			s.createTasksForToday(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) createTasksForToday(ctx context.Context) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	templates, err := s.repo.ListRecurring(ctx)
	if err != nil {
		s.logger.Error("scheduler: list recurring tasks", "error", err)
		return
	}

	s.logger.Info("scheduler: found templates", "count", len(templates))

	for _, tmpl := range templates {
		if tmpl.Recurrence == nil {
			continue
		}

		s.logger.Info("scheduler: checking template", "id", tmpl.ID, "type", tmpl.Recurrence.Type, "matches", tmpl.Recurrence.Matches(today))

		if !tmpl.Recurrence.Matches(today) {
			continue
		}
		exists, err := s.repo.ExistsForParentAndDate(ctx, tmpl.ID, today)
		if err != nil {
			s.logger.Error("scheduler: check exists", "error", err, "parent_id", tmpl.ID)
			continue
		}

		if exists {
			continue
		}

		now := time.Now().UTC()
		_, err = s.repo.Create(ctx, &taskdomain.Task{
			Title:         tmpl.Title,
			Description:   tmpl.Description,
			Status:        taskdomain.StatusNew,
			ParentID:      &tmpl.ID,
			ScheduledDate: &today,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
		if err != nil {
			s.logger.Error("scheduler: check exists", "error", err, "parent_id", tmpl.ID)
			continue
		}

		s.logger.Info("scheduler: created task", "parent_id", tmpl.ID, "date", today)
	}
}
