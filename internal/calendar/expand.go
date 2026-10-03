package calendar

import (
	"sort"
	"time"
)

// ExpandDates returns the deduplicated, sorted occurrence start dates of an
// activity: rule dates merged with extra dates, minus excluded dates.
func ExpandDates(a *Activity) ([]time.Time, error) {
	start, err := parseDate(a.Start)
	if err != nil {
		return nil, err
	}
	set := map[time.Time]bool{}
	if a.Recurrence == nil {
		set[start] = true
	} else {
		for _, d := range expandRule(a.Recurrence, start) {
			set[d] = true
		}
	}
	for _, s := range a.Add {
		d, err := parseDate(s)
		if err != nil {
			return nil, err
		}
		set[d] = true
	}
	for _, s := range a.Exclude {
		d, err := parseDate(s)
		if err != nil {
			return nil, err
		}
		delete(set, d)
	}
	dates := make([]time.Time, 0, len(set))
	for d := range set {
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	return dates, nil
}

func expandRule(r *Recurrence, start time.Time) []time.Time {
	var until time.Time
	if r.Until != nil {
		until, _ = parseDate(*r.Until)
	}
	limit := -1
	if r.Count != nil {
		limit = *r.Count
	}
	var out []time.Time
	accept := func(d time.Time) bool {
		if r.Until != nil && d.After(until) {
			return false
		}
		out = append(out, d)
		return limit < 0 || len(out) < limit
	}
	switch r.Freq {
	case "daily":
		for k := 0; ; k++ {
			if !accept(start.AddDate(0, 0, k*r.Interval)) {
				return out
			}
		}
	case "weekly":
		// Week zero is the Monday of the week containing start.
		monday := start.AddDate(0, 0, -int((int(start.Weekday())+6)%7))
		days := make([]time.Weekday, 0, len(r.Weekdays))
		for _, s := range r.Weekdays {
			days = append(days, weekdayNames[s])
		}
		sort.Slice(days, func(i, j int) bool { return days[i] < days[j] })
		for w := 0; ; w += r.Interval {
			weekMon := monday.AddDate(0, 0, 7*w)
			if r.Until != nil && weekMon.After(until) {
				return out
			}
			for _, wd := range days {
				d := weekMon.AddDate(0, 0, (int(wd)+6)%7)
				if d.Before(start) {
					continue
				}
				if !accept(d) {
					return out
				}
			}
		}
	case "monthly":
		day := start.Day()
		for m := 0; ; m += r.Interval {
			first := time.Date(start.Year(), start.Month()+time.Month(m), 1, 0, 0, 0, 0, time.UTC)
			last := first.AddDate(0, 1, -1)
			if r.Until != nil && first.After(until) {
				return out
			}
			if day > last.Day() {
				continue // month lacks the day; does not consume count
			}
			if !accept(time.Date(first.Year(), first.Month(), day, 0, 0, 0, 0, time.UTC)) {
				return out
			}
		}
	}
	return out
}

// Instances builds half-open [start, start+duration) instances intersecting
// the half-open query window, sorted by start date then activity id.
func Instances(in *Input) ([]Instance, error) {
	ws, err := parseDate(in.Window.Start)
	if err != nil {
		return nil, err
	}
	we, err := parseDate(in.Window.End)
	if err != nil {
		return nil, err
	}
	var out []Instance
	for i := range in.Activities {
		a := &in.Activities[i]
		dates, err := ExpandDates(a)
		if err != nil {
			return nil, err
		}
		for _, d := range dates {
			end := d.AddDate(0, 0, a.DurationDays)
			if d.Before(we) && end.After(ws) {
				out = append(out, Instance{ActivityID: a.ID, Start: d, End: end})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Start.Equal(out[j].Start) {
			return out[i].Start.Before(out[j].Start)
		}
		return out[i].ActivityID < out[j].ActivityID
	})
	return out, nil
}

// Summary counts window hits per activity, keeping zero-hit activities.
func Summary(in *Input, instances []Instance) map[string]int {
	counts := make(map[string]int, len(in.Activities))
	for _, a := range in.Activities {
		counts[a.ID] = 0
	}
	for _, inst := range instances {
		counts[inst.ActivityID]++
	}
	return counts
}
