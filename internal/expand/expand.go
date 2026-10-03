package expand

import (
	"sort"
	"time"
)

// Occurrences returns the deduplicated, sorted start dates of an activity
// after applying its recurrence rule, extra dates and exclusions.
func Occurrences(a *Activity) []Date {
	dates := map[string]Date{}
	if a.Recurrence == nil {
		dates[a.Start.String()] = a.Start
	} else {
		for _, d := range ruleDates(a) {
			dates[d.String()] = d
		}
	}
	for _, d := range a.AddDates {
		dates[d.String()] = d
	}
	for _, d := range a.ExcludeDates {
		delete(dates, d.String())
	}
	out := make([]Date, 0, len(dates))
	for _, d := range dates {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Before(out[j]) })
	return out
}

// ruleDates generates dates from the recurrence rule, honouring count or until.
func ruleDates(a *Activity) []Date {
	r := a.Recurrence
	var out []Date
	accept := func(d Date) bool {
		if r.Until != nil && d.After(*r.Until) {
			return false
		}
		out = append(out, d)
		return r.Count == nil || len(out) < *r.Count
	}
	switch r.Type {
	case "daily":
		for d := a.Start; ; d = d.AddDays(r.Interval) {
			if !accept(d) {
				return out
			}
		}
	case "weekly":
		return weeklyDates(a, accept)
	case "monthly":
		return monthlyDates(a, accept)
	}
	return out
}

func weeklyDates(a *Activity, accept func(Date) bool) []Date {
	r := a.Recurrence
	// Week zero starts on the Monday of the start date's week.
	offset := (int(a.Start.Weekday()) - int(time.Monday) + 7) % 7
	weekZero := a.Start.AddDays(-offset)
	days := make([]int, 0, len(r.Weekdays))
	for _, w := range r.Weekdays {
		days = append(days, (int(weekdayNames[w])-int(time.Monday)+7)%7)
	}
	sort.Ints(days)
	var out []Date
	emit := func(d Date) bool {
		if d.Before(a.Start) {
			return true // skip dates before the start day
		}
		if r.Until != nil && d.After(*r.Until) {
			return false
		}
		out = append(out, d)
		return r.Count == nil || len(out) < *r.Count
	}
	for week := 0; ; week += r.Interval {
		base := weekZero.AddDays(week * 7)
		for _, day := range days {
			if !emit(base.AddDays(day)) {
				return out
			}
		}
	}
}

func monthlyDates(a *Activity, accept func(Date) bool) []Date {
	r := a.Recurrence
	day := a.Start.Day()
	year, month := a.Start.Year(), int(a.Start.Month())
	var out []Date
	emit := func(d Date) bool {
		if r.Until != nil && d.After(*r.Until) {
			return false
		}
		out = append(out, d)
		return r.Count == nil || len(out) < *r.Count
	}
	for k := 0; ; k += r.Interval {
		total := year*12 + (month - 1) + k
		y, m := total/12, total%12+1
		last := daysInMonth(y, time.Month(m))
		if day > last {
			// Month lacks this day-of-month: skip without consuming count.
			if r.Until != nil {
				firstOfNext := time.Date(y, time.Month(m)+1, 1, 0, 0, 0, 0, time.UTC)
				if (Date{firstOfNext}).After(*r.Until) {
					return out
				}
			}
			continue
		}
		if !emit(Date{time.Date(y, time.Month(m), day, 0, 0, 0, 0, time.UTC)}) {
			return out
		}
	}
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// Instances expands all activities into instances intersecting the window.
// Results are sorted by start date, then activity ID.
func Instances(in *Input) []Instance {
	var out []Instance
	for i := range in.Activities {
		a := &in.Activities[i]
		for _, start := range Occurrences(a) {
			inst := Instance{ActivityID: a.ID, Start: start, End: start.AddDays(a.DurationDays)}
			if inst.Start.Before(in.Window.End) && inst.End.After(in.Window.Start) {
				out = append(out, inst)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Start.Equal(out[j].Start) {
			return out[i].Start.Before(out[j].Start)
		}
		return out[i].ActivityID < out[j].ActivityID
	})
	return out
}
