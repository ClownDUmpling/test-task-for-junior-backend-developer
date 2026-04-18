package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type recurrenceDTO struct {
	Type       taskdomain.RecurrenceType `json:"type"`
	EveryNDays *int                      `json:"every_n_days,omitempty"`
	MonthDays  []int                     `json:"month_days,omitempty"`
	Dates      []time.Time               `json:"dates,omitempty"`
	Parity     *string                   `json:"parity,omitempty"`
}

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Recurrence  *recurrenceDTO    `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID            int64             `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Status        taskdomain.Status `json:"status"`
	Recurrence    *recurrenceDTO    `json:"recurrence,omitempty"`
	ParentID      *int64            `json:"parent_id,omitempty"`
	ScheduledDate *time.Time        `json:"scheduled_date,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	dto := taskDTO{
		ID:            task.ID,
		Title:         task.Title,
		Description:   task.Description,
		Status:        task.Status,
		ParentID:      task.ParentID,
		ScheduledDate: task.ScheduledDate,
		CreatedAt:     task.CreatedAt,
		UpdatedAt:     task.UpdatedAt,
	}

	if task.Recurrence != nil {
		dto.Recurrence = &recurrenceDTO{
			Type:       task.Recurrence.Type,
			EveryNDays: task.Recurrence.EveryNDays,
			MonthDays:  task.Recurrence.MonthDays,
			Dates:      task.Recurrence.Dates,
			Parity:     task.Recurrence.Parity,
		}
	}
	return dto
}
