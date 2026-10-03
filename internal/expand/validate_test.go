package expand

import "testing"

func baseInput() *Input {
	return &Input{
		Window: Window{Start: d("2026-01-01"), End: d("2026-02-01")},
		Activities: []Activity{
			{ID: "a", Start: d("2026-01-05"), DurationDays: 2},
		},
	}
}

func TestValidateOK(t *testing.T) {
	if err := Validate(baseInput()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateErrors(t *testing.T) {
	cases := map[string]func(*Input){
		"empty window": func(in *Input) { in.Window.End = in.Window.Start },
		"duplicate id": func(in *Input) {
			in.Activities = append(in.Activities, Activity{ID: "a", Start: d("2026-01-01"), DurationDays: 1})
		},
		"zero duration": func(in *Input) { in.Activities[0].DurationDays = 0 },
		"bad interval": func(in *Input) {
			in.Activities[0].Recurrence = &Recurrence{Type: "daily", Interval: 0, Count: intPtr(2)}
		},
		"count and until both set": func(in *Input) {
			in.Activities[0].Recurrence = &Recurrence{Type: "daily", Interval: 1, Count: intPtr(2), Until: datePtr("2026-02-01")}
		},
		"neither count nor until": func(in *Input) {
			in.Activities[0].Recurrence = &Recurrence{Type: "daily", Interval: 1}
		},
		"until before start": func(in *Input) {
			in.Activities[0].Recurrence = &Recurrence{Type: "daily", Interval: 1, Until: datePtr("2026-01-01")}
		},
		"weekly missing start weekday": func(in *Input) {
			in.Activities[0].Recurrence = &Recurrence{Type: "weekly", Interval: 1, Count: intPtr(2), Weekdays: []string{"TU"}}
		},
		"weekly duplicate weekday": func(in *Input) {
			in.Activities[0].Recurrence = &Recurrence{Type: "weekly", Interval: 1, Count: intPtr(2), Weekdays: []string{"MO", "MO"}}
		},
		"weekdays on daily": func(in *Input) {
			in.Activities[0].Recurrence = &Recurrence{Type: "daily", Interval: 1, Count: intPtr(2), Weekdays: []string{"MO"}}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			in := baseInput()
			mutate(in)
			if err := Validate(in); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestInstancesWindowAndSorting(t *testing.T) {
	in := &Input{
		Window: Window{Start: d("2026-01-10"), End: d("2026-01-20")},
		Activities: []Activity{
			// Instance [01-08, 01-12) overlaps window start; kept with full dates.
			{ID: "b", Start: d("2026-01-08"), DurationDays: 4},
			// Instance [01-19, 01-21) overlaps window end.
			{ID: "a", Start: d("2026-01-19"), DurationDays: 2},
			// No overlap.
			{ID: "c", Start: d("2026-02-01"), DurationDays: 1},
		},
	}
	if err := Validate(in); err != nil {
		t.Fatal(err)
	}
	insts := Instances(in)
	if len(insts) != 2 {
		t.Fatalf("got %d instances, want 2", len(insts))
	}
	if insts[0].ActivityID != "b" || insts[0].Start.String() != "2026-01-08" || insts[0].End.String() != "2026-01-12" {
		t.Fatalf("unexpected first instance: %+v", insts[0])
	}
	if insts[1].ActivityID != "a" {
		t.Fatalf("unexpected second instance: %+v", insts[1])
	}
}
