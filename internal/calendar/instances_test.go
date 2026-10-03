package calendar

import (
	"reflect"
	"testing"
)

func TestInstancesWindowIntersection(t *testing.T) {
	in := &Input{
		Window: Window{Start: "2026-02-01", End: "2026-03-01"},
		Activities: []Activity{
			// Spans the window start: kept with full dates.
			{ID: "span", Start: "2026-01-30", DurationDays: 5},
			// Entirely before the window: dropped.
			{ID: "gone", Start: "2026-01-01", DurationDays: 3},
			// Ends exactly at window start (half-open): dropped.
			{ID: "edge", Start: "2026-01-28", DurationDays: 4},
			// Starts exactly at window end (half-open): dropped.
			{ID: "tail", Start: "2026-03-01", DurationDays: 2},
			// Zero hits but still present in summary.
			{ID: "empty", Start: "2026-06-01", DurationDays: 1},
		},
	}
	if err := in.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	insts, err := Instances(in)
	if err != nil {
		t.Fatalf("Instances: %v", err)
	}
	if len(insts) != 1 {
		t.Fatalf("got %d instances, want 1: %+v", len(insts), insts)
	}
	inst := insts[0]
	if inst.ActivityID != "span" ||
		inst.Start.Format(DateLayout) != "2026-01-30" ||
		inst.End.Format(DateLayout) != "2026-02-04" {
		t.Fatalf("unexpected instance: %+v", inst)
	}
	sum := Summary(in, insts)
	want := map[string]int{"span": 1, "gone": 0, "edge": 0, "tail": 0, "empty": 0}
	if !reflect.DeepEqual(sum, want) {
		t.Fatalf("summary got %v want %v", sum, want)
	}
}

func TestInstancesSortedByStartThenID(t *testing.T) {
	in := &Input{
		Window: Window{Start: "2026-01-01", End: "2026-02-01"},
		Activities: []Activity{
			{ID: "b", Start: "2026-01-10", DurationDays: 1},
			{ID: "a", Start: "2026-01-10", DurationDays: 1},
			{ID: "c", Start: "2026-01-05", DurationDays: 1},
		},
	}
	insts, err := Instances(in)
	if err != nil {
		t.Fatalf("Instances: %v", err)
	}
	var order []string
	for _, inst := range insts {
		order = append(order, inst.Start.Format(DateLayout)+"/"+inst.ActivityID)
	}
	want := []string{"2026-01-05/c", "2026-01-10/a", "2026-01-10/b"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("got %v want %v", order, want)
	}
}

func TestValidationFailures(t *testing.T) {
	cases := map[string]Input{
		"empty window": {
			Window:     Window{Start: "2026-02-01", End: "2026-02-01"},
			Activities: []Activity{{ID: "a", Start: "2026-01-01", DurationDays: 1}},
		},
		"duplicate id": {
			Window: Window{Start: "2026-01-01", End: "2026-02-01"},
			Activities: []Activity{
				{ID: "a", Start: "2026-01-01", DurationDays: 1},
				{ID: "a", Start: "2026-01-02", DurationDays: 1},
			},
		},
		"zero duration": {
			Window:     Window{Start: "2026-01-01", End: "2026-02-01"},
			Activities: []Activity{{ID: "a", Start: "2026-01-01", DurationDays: 0}},
		},
		"count and until both set": {
			Window: Window{Start: "2026-01-01", End: "2026-02-01"},
			Activities: []Activity{{ID: "a", Start: "2026-01-01", DurationDays: 1,
				Recurrence: &Recurrence{Freq: "daily", Interval: 1, Count: intPtr(2), Until: strPtr("2026-01-10")}}},
		},
		"neither count nor until": {
			Window: Window{Start: "2026-01-01", End: "2026-02-01"},
			Activities: []Activity{{ID: "a", Start: "2026-01-01", DurationDays: 1,
				Recurrence: &Recurrence{Freq: "daily", Interval: 1}}},
		},
		"until before start": {
			Window: Window{Start: "2026-01-01", End: "2026-03-01"},
			Activities: []Activity{{ID: "a", Start: "2026-02-01", DurationDays: 1,
				Recurrence: &Recurrence{Freq: "daily", Interval: 1, Until: strPtr("2026-01-15")}}},
		},
		"weekly missing start weekday": {
			Window: Window{Start: "2026-01-01", End: "2026-02-01"},
			Activities: []Activity{{ID: "a", Start: "2026-01-07", DurationDays: 1, // Wednesday
				Recurrence: &Recurrence{Freq: "weekly", Interval: 1, Count: intPtr(2), Weekdays: []string{"MO"}}}},
		},
		"weekly duplicate weekday": {
			Window: Window{Start: "2026-01-01", End: "2026-02-01"},
			Activities: []Activity{{ID: "a", Start: "2026-01-07", DurationDays: 1,
				Recurrence: &Recurrence{Freq: "weekly", Interval: 1, Count: intPtr(2), Weekdays: []string{"WE", "WE"}}}},
		},
		"zero interval": {
			Window: Window{Start: "2026-01-01", End: "2026-02-01"},
			Activities: []Activity{{ID: "a", Start: "2026-01-01", DurationDays: 1,
				Recurrence: &Recurrence{Freq: "daily", Interval: 0, Count: intPtr(2)}}},
		},
		"bad date": {
			Window:     Window{Start: "2026-01-01", End: "2026-02-01"},
			Activities: []Activity{{ID: "a", Start: "2026-13-01", DurationDays: 1}},
		},
	}
	for name, in := range cases {
		if err := in.Validate(); err == nil {
			t.Errorf("%s: expected validation error, got nil", name)
		}
	}
}
