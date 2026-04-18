package task

import "time"

type RecurrenceType string

const (
	RecurrenceDaily          RecurrenceType = "daily"
	RecurrenceMonthly        RecurrenceType = "monthly"
	RecurrenceSpecifiedDates RecurrenceType = "specific_dates"
	RecurrenceEvenOdd        RecurrenceType = "even_odd"
)

type Recurrence struct {
	Type       RecurrenceType
	EveryNDays *int
	MonthDays  []int
	Dates      []time.Time
	Parity     *string // "even"/"odd"
	BaseDate   time.Time
}

func (r Recurrence) Matches(date time.Time) bool {
	switch r.Type {
	case RecurrenceDaily:
		diff := int(date.Sub(r.BaseDate).Hours() / 24)
		return diff >= 0 && diff%*r.EveryNDays == 0
	case RecurrenceMonthly:
		for _, d := range r.MonthDays {
			if d == date.Day() {
				return true
			}
		}
	case RecurrenceSpecifiedDates:
		for _, d := range r.Dates {
			if d.Equal(date) {
				return true
			}
		}
	case RecurrenceEvenOdd:
		if *r.Parity == "even" {
			return date.Day()%2 == 0
		}
		return date.Day()%2 != 0
	}
	return false
}
