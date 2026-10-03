package expand

import (
	"testing"
	"time"
)

func d(s string) Date {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		panic(err)
	}
	return Date{t}
}

func intPtr(n int) *int { return &n }

func datePtr(s string) *Date { dt := d(s); return &dt }

func stringsOf(dates []Date) []string {
	out := make([]string, len(dates))
	for i, dt := range dates {
		out[i] = dt.String()
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestDailyCount(t *testing.T) {
	a := &Activity{
		ID: "a", Start: d("2026-01-01"), DurationDays: 1,
		Recurrence: &Recurrence{Type: "daily", Interval: 3, Count: intPtr(3)},
	}
	got := stringsOf(Occurrences(a))
	want := []string{"2026-01-01", "2026-01-04", "2026-01-07"}
	if !equalStrings(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestDailyUntilInclusive(t *testing.T) {
	a := &Activity{
		ID: "a", Start: d("2026-01-01"), DurationDays: 1,
		Recurrence: &Recurrence{Type: "daily", Interval: 2, Until: datePtr("2026-01-05")},
	}
	got := stringsOf(Occurrences(a))
	want := []string{"2026-01-01", "2026-01-03", "2026-01-05"}
	if !equalStrings(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestWeeklyIntervalSkipsBeforeStart(t *testing.T) {
	// 2026-01-07 is a Wednesday; week zero starts Monday 2026-01-05.
	a := &Activity{
		ID: "a", Start: d("2026-01-07"), DurationDays: 1,
		Recurrence: &Recurrence{
			Type: "weekly", Interval: 2, Count: intPtr(4),
			Weekdays: []string{"MO", "WE"},
		},
	}
	got := stringsOf(Occurrences(a))
	want := []string{"2026-01-07", "2026-01-19", "2026-01-21", "2026-02-02"}
	if !equalStrings(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestMonthlySkipsMissingDayWithoutConsumingCount(t *testing.T) {
	// Jan 31 -> Feb has no 31st (skipped, no count consumed) -> Mar 31.
	a := &Activity{
		ID: "a", Start: d("2026-01-31"), DurationDays: 1,
		Recurrence: &Recurrence{Type: "monthly", Interval: 1, Count: intPtr(3)},
	}
	got := stringsOf(Occurrences(a))
	want := []string{"2026-01-31", "2026-03-31", "2026-05-31"}
	if !equalStrings(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestMonthlyUntilStopsAfterSkippedMonth(t *testing.T) {
	a := &Activity{
		ID: "a", Start: d("2026-01-31"), DurationDays: 1,
		Recurrence: &Recurrence{Type: "monthly", Interval: 1, Until: datePtr("2026-03-30")},
	}
	got := stringsOf(Occurrences(a))
	want := []string{"2026-01-31"}
	if !equalStrings(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestAddAndExcludeDates(t *testing.T) {
	a := &Activity{
		ID: "a", Start: d("2026-01-01"), DurationDays: 1,
		Recurrence:   &Recurrence{Type: "daily", Interval: 7, Count: intPtr(2)},
		AddDates:     []Date{d("2026-01-08"), d("2026-06-01"), d("2026-01-08")},
		ExcludeDates: []Date{d("2026-01-08"), d("2026-12-31")},
	}
	got := stringsOf(Occurrences(a))
	// 2026-01-08 excluded (exclusion wins, count not refilled); 2026-06-01
	// added beyond count; excluding a non-occurring date is a no-op.
	want := []string{"2026-01-01", "2026-06-01"}
	if !equalStrings(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestNoRecurrence(t *testing.T) {
	a := &Activity{ID: "a", Start: d("2026-02-10"), DurationDays: 5}
	got := stringsOf(Occurrences(a))
	if !equalStrings(got, []string{"2026-02-10"}) {
		t.Fatalf("got %v", got)
	}
}
