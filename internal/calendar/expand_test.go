package calendar

import (
	"reflect"
	"testing"
)

func intPtr(n int) *int { return &n }

func strPtr(s string) *string { return &s }

func expandIDs(t *testing.T, a Activity) []string {
	t.Helper()
	if err := a.validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	dates, err := ExpandDates(&a)
	if err != nil {
		t.Fatalf("ExpandDates: %v", err)
	}
	out := make([]string, len(dates))
	for i, d := range dates {
		out[i] = d.Format(DateLayout)
	}
	return out
}

func TestDailyWithCount(t *testing.T) {
	a := Activity{
		ID: "d", Start: "2026-01-30", DurationDays: 1,
		Recurrence: &Recurrence{Freq: "daily", Interval: 2, Count: intPtr(3)},
	}
	got := expandIDs(t, a)
	want := []string{"2026-01-30", "2026-02-01", "2026-02-03"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestWeeklySkipsBeforeStart(t *testing.T) {
	// 2026-01-07 is a Wednesday; week zero starts Monday 2026-01-05.
	a := Activity{
		ID: "w", Start: "2026-01-07", DurationDays: 1,
		Recurrence: &Recurrence{Freq: "weekly", Interval: 1, Count: intPtr(4), Weekdays: []string{"MO", "WE"}},
	}
	got := expandIDs(t, a)
	want := []string{"2026-01-07", "2026-01-12", "2026-01-14", "2026-01-19"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestWeeklyIntervalAndUntil(t *testing.T) {
	a := Activity{
		ID: "w", Start: "2026-01-05", DurationDays: 1, // Monday
		Recurrence: &Recurrence{Freq: "weekly", Interval: 2, Until: strPtr("2026-02-15"), Weekdays: []string{"MO", "FR"}},
	}
	got := expandIDs(t, a)
	want := []string{"2026-01-05", "2026-01-09", "2026-01-19", "2026-01-23", "2026-02-02", "2026-02-06"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestMonthlySkipsMissingDayWithoutConsumingCount(t *testing.T) {
	// Jan 31 -> Feb has no 31st (skipped, no count consumed) -> Mar 31.
	a := Activity{
		ID: "m", Start: "2026-01-31", DurationDays: 1,
		Recurrence: &Recurrence{Freq: "monthly", Interval: 1, Count: intPtr(2)},
	}
	got := expandIDs(t, a)
	want := []string{"2026-01-31", "2026-03-31"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestMonthlyInterval(t *testing.T) {
	a := Activity{
		ID: "m", Start: "2026-01-15", DurationDays: 1,
		Recurrence: &Recurrence{Freq: "monthly", Interval: 2, Count: intPtr(3)},
	}
	got := expandIDs(t, a)
	want := []string{"2026-01-15", "2026-03-15", "2026-05-15"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestAddExcludeMerge(t *testing.T) {
	a := Activity{
		ID: "x", Start: "2026-01-01", DurationDays: 1,
		Recurrence: &Recurrence{Freq: "daily", Interval: 1, Count: intPtr(3)},
		Add:     []string{"2026-01-02", "2026-03-10"}, // 01-02 duplicates a rule date
		Exclude: []string{"2026-01-03", "2026-12-25"}, // 12-25 never occurs: no-op
	}
	got := expandIDs(t, a)
	want := []string{"2026-01-01", "2026-01-02", "2026-03-10"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestNoRecurrence(t *testing.T) {
	a := Activity{ID: "one", Start: "2026-05-01", DurationDays: 4}
	got := expandIDs(t, a)
	want := []string{"2026-05-01"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
