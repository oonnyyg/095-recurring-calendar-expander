package expand

import (
	"encoding/json"
	"fmt"
	"time"
)

const dateLayout = "2006-01-02"

// Date is a Gregorian calendar date without time or timezone.
type Date struct {
	time.Time
}

func ParseDate(s string) (Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return Date{}, fmt.Errorf("invalid date %q (want YYYY-MM-DD)", s)
	}
	return Date{t}, nil
}

func (d Date) String() string { return d.Format(dateLayout) }

func (d Date) AddDays(n int) Date { return Date{d.AddDate(0, 0, n)} }

func (d Date) Before(o Date) bool { return d.Time.Before(o.Time) }

func (d Date) After(o Date) bool { return d.Time.After(o.Time) }

func (d Date) Equal(o Date) bool { return d.Time.Equal(o.Time) }

func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("date must be a string: %w", err)
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// Window is a half-open query interval [Start, End).
type Window struct {
	Start Date `json:"start"`
	End   Date `json:"end"`
}

// Recurrence describes how an activity repeats. Exactly one of Count or
// Until must be set to terminate the rule.
type Recurrence struct {
	Type     string   `json:"type"` // "daily", "weekly" or "monthly"
	Interval int      `json:"interval"`
	Count    *int     `json:"count,omitempty"`
	Until    *Date    `json:"until,omitempty"`
	Weekdays []string `json:"weekdays,omitempty"` // weekly only, e.g. "MO"
}

// Activity is one calendar activity definition.
type Activity struct {
	ID           string      `json:"id"`
	Start        Date        `json:"start"`
	DurationDays int         `json:"duration_days"`
	Recurrence   *Recurrence `json:"recurrence,omitempty"`
	AddDates     []Date      `json:"add_dates,omitempty"`
	ExcludeDates []Date      `json:"exclude_dates,omitempty"`
}

// Input is the top-level JSON document.
type Input struct {
	Window     Window     `json:"window"`
	Activities []Activity `json:"activities"`
}

// Instance is one expanded occurrence of an activity.
type Instance struct {
	ActivityID string
	Start      Date
	End        Date // exclusive
}
