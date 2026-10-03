package expand

import (
	"fmt"
	"time"
)

var weekdayNames = map[string]time.Weekday{
	"MO": time.Monday, "TU": time.Tuesday, "WE": time.Wednesday,
	"TH": time.Thursday, "FR": time.Friday, "SA": time.Saturday,
	"SU": time.Sunday,
}

// Validate checks the whole input; any error fails the entire batch.
func Validate(in *Input) error {
	if !in.Window.End.After(in.Window.Start) {
		return fmt.Errorf("window end %s must be after start %s", in.Window.End, in.Window.Start)
	}
	seen := map[string]bool{}
	for i := range in.Activities {
		a := &in.Activities[i]
		if a.ID == "" {
			return fmt.Errorf("activity #%d: id must not be empty", i)
		}
		if seen[a.ID] {
			return fmt.Errorf("duplicate activity id %q", a.ID)
		}
		seen[a.ID] = true
		if a.DurationDays < 1 {
			return fmt.Errorf("activity %q: duration_days must be a positive integer", a.ID)
		}
		if a.Recurrence != nil {
			if err := validateRecurrence(a); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateRecurrence(a *Activity) error {
	r := a.Recurrence
	if r.Type != "daily" && r.Type != "weekly" && r.Type != "monthly" {
		return fmt.Errorf("activity %q: recurrence type must be daily, weekly or monthly", a.ID)
	}
	if r.Interval < 1 {
		return fmt.Errorf("activity %q: recurrence interval must be a positive integer", a.ID)
	}
	if (r.Count == nil) == (r.Until == nil) {
		return fmt.Errorf("activity %q: exactly one of count or until must be set", a.ID)
	}
	if r.Count != nil && *r.Count < 1 {
		return fmt.Errorf("activity %q: count must be a positive integer", a.ID)
	}
	if r.Until != nil && r.Until.Before(a.Start) {
		return fmt.Errorf("activity %q: until %s is before start %s", a.ID, r.Until, a.Start)
	}
	if r.Type == "weekly" {
		if len(r.Weekdays) == 0 {
			return fmt.Errorf("activity %q: weekly recurrence requires weekdays", a.ID)
		}
		seen := map[string]bool{}
		containsStart := false
		for _, w := range r.Weekdays {
			wd, ok := weekdayNames[w]
			if !ok {
				return fmt.Errorf("activity %q: invalid weekday %q", a.ID, w)
			}
			if seen[w] {
				return fmt.Errorf("activity %q: duplicate weekday %q", a.ID, w)
			}
			seen[w] = true
			if wd == a.Start.Weekday() {
				containsStart = true
			}
		}
		if !containsStart {
			return fmt.Errorf("activity %q: weekdays must contain the start weekday", a.ID)
		}
	} else if len(r.Weekdays) > 0 {
		return fmt.Errorf("activity %q: weekdays only allowed for weekly recurrence", a.ID)
	}
	return nil
}
