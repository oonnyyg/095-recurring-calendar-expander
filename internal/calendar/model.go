package calendar

import (
	"fmt"
	"time"
)

const DateLayout = "2006-01-02"

type Input struct {
	Window     Window     `json:"window"`
	Activities []Activity `json:"activities"`
}

type Window struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Activity struct {
	ID           string      `json:"id"`
	Start        string      `json:"start"`
	DurationDays int         `json:"duration_days"`
	Recurrence   *Recurrence `json:"recurrence,omitempty"`
	Add          []string    `json:"add,omitempty"`
	Exclude      []string    `json:"exclude,omitempty"`
}

type Recurrence struct {
	Freq     string   `json:"freq"` // daily | weekly | monthly
	Interval int      `json:"interval"`
	Count    *int     `json:"count,omitempty"`
	Until    *string  `json:"until,omitempty"`
	Weekdays []string `json:"weekdays,omitempty"` // MO,TU,WE,TH,FR,SA,SU
}

type Instance struct {
	ActivityID string
	Start      time.Time
	End        time.Time
}

func parseDate(s string) (time.Time, error) {
	t, err := time.ParseInLocation(DateLayout, s, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q (want YYYY-MM-DD)", s)
	}
	return t, nil
}

var weekdayNames = map[string]time.Weekday{
	"MO": time.Monday, "TU": time.Tuesday, "WE": time.Wednesday,
	"TH": time.Thursday, "FR": time.Friday, "SA": time.Saturday,
	"SU": time.Sunday,
}

// Validate checks the whole input; any error fails the batch.
func (in *Input) Validate() error {
	ws, err := parseDate(in.Window.Start)
	if err != nil {
		return fmt.Errorf("window.start: %w", err)
	}
	we, err := parseDate(in.Window.End)
	if err != nil {
		return fmt.Errorf("window.end: %w", err)
	}
	if !we.After(ws) {
		return fmt.Errorf("window must have positive length: %s .. %s", in.Window.Start, in.Window.End)
	}
	seen := map[string]bool{}
	for i, a := range in.Activities {
		if a.ID == "" {
			return fmt.Errorf("activities[%d]: empty id", i)
		}
		if seen[a.ID] {
			return fmt.Errorf("duplicate activity id %q", a.ID)
		}
		seen[a.ID] = true
		if err := a.validate(); err != nil {
			return fmt.Errorf("activity %q: %w", a.ID, err)
		}
	}
	return nil
}

func (a *Activity) validate() error {
	start, err := parseDate(a.Start)
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}
	if a.DurationDays < 1 {
		return fmt.Errorf("duration_days must be a positive integer, got %d", a.DurationDays)
	}
	for _, s := range a.Add {
		if _, err := parseDate(s); err != nil {
			return fmt.Errorf("add: %w", err)
		}
	}
	for _, s := range a.Exclude {
		if _, err := parseDate(s); err != nil {
			return fmt.Errorf("exclude: %w", err)
		}
	}
	if a.Recurrence == nil {
		return nil
	}
	r := a.Recurrence
	switch r.Freq {
	case "daily", "weekly", "monthly":
	default:
		return fmt.Errorf("recurrence.freq must be daily, weekly or monthly, got %q", r.Freq)
	}
	if r.Interval < 1 {
		return fmt.Errorf("recurrence.interval must be a positive integer, got %d", r.Interval)
	}
	if (r.Count == nil) == (r.Until == nil) {
		return fmt.Errorf("recurrence must set exactly one of count or until")
	}
	if r.Count != nil && *r.Count < 1 {
		return fmt.Errorf("recurrence.count must be a positive integer, got %d", *r.Count)
	}
	if r.Until != nil {
		until, err := parseDate(*r.Until)
		if err != nil {
			return fmt.Errorf("recurrence.until: %w", err)
		}
		if until.Before(start) {
			return fmt.Errorf("recurrence.until %s is before start %s", *r.Until, a.Start)
		}
	}
	if r.Freq == "weekly" {
		if len(r.Weekdays) == 0 {
			return fmt.Errorf("weekly recurrence requires weekdays")
		}
		seenDays := map[time.Weekday]bool{}
		for _, d := range r.Weekdays {
			wd, ok := weekdayNames[d]
			if !ok {
				return fmt.Errorf("invalid weekday %q (want MO..SU)", d)
			}
			if seenDays[wd] {
				return fmt.Errorf("duplicate weekday %q", d)
			}
			seenDays[wd] = true
		}
		if !seenDays[start.Weekday()] {
			return fmt.Errorf("weekly weekdays must include the start weekday %s", shortWeekday(start.Weekday()))
		}
	}
	return nil
}

func shortWeekday(w time.Weekday) string {
	names := []string{"SU", "MO", "TU", "WE", "TH", "FR", "SA"}
	return names[int(w)]
}
