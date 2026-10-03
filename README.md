# Recurring Calendar Expander

Expands recurring calendar activities into concrete instances and reports
those intersecting a query window. Gregorian `YYYY-MM-DD` dates only; no
times or timezones. Standard library only.

## Usage

```
go run ./cmd/calendar <input.json> --out <output-dir>
```

Writes `instances.csv` and `summary.json` into the output directory. Any
invalid definition fails the whole batch and no results are written.

## Input format

```json
{
  "window": {"start": "2026-01-01", "end": "2026-03-01"},
  "activities": [
    {
      "id": "standup",
      "start": "2026-01-05",
      "duration_days": 1,
      "recurrence": {
        "type": "weekly",
        "interval": 1,
        "weekdays": ["MO", "WE"],
        "until": "2026-02-28"
      },
      "add_dates": ["2026-01-31"],
      "exclude_dates": ["2026-01-26"]
    }
  ]
}
```

- `window`: half-open query interval `[start, end)`; must have positive length.
- `id`: unique activity identifier.
- `start`: first occurrence date (`YYYY-MM-DD`).
- `duration_days`: positive integer; instance spans `[start, start+duration)`.
- `recurrence` (optional): omit for a one-off activity.
  - `type`: `daily`, `weekly` or `monthly`.
  - `interval`: positive integer step.
  - Termination: exactly one of `count` (positive integer, counts valid rule
    dates including the start day) or `until` (inclusive end date, not earlier
    than `start`).
  - `weekly` also requires `weekdays`: a non-empty set without duplicates
    (`MO`..`SU`) containing the start weekday. The week of the start date
    (from its Monday) is week zero; every `interval`-th week is selected and
    set dates enumerated, skipping dates before `start`.
  - `monthly` keeps the start day-of-month; months lacking that day are
    skipped without consuming `count`.
- `add_dates`: extra occurrences, merged and deduplicated after rule
  expansion; not limited by `count`/`until`.
- `exclude_dates`: removed last; exclusions win and do not refill `count`;
  excluding a non-occurring date is a no-op.

## Output

`instances.csv` — one row per instance intersecting the window (instances
keep their full start/end dates), sorted by start date then activity id:

```
activity_id,start,end
standup,2026-01-05,2026-01-06
```

`summary.json` — per-activity hit counts (zero-hit activities included):

```json
{
  "window_start": "2026-01-01",
  "window_end": "2026-03-01",
  "total_instances": 1,
  "activities": [{"activity_id": "standup", "count": 1}]
}
```

## Tests

```
go test ./...
```

See `examples/input.json` for a runnable sample.
