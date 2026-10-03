package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"example.com/recurring-calendar-expander/internal/calendar"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var inputPath, outDir string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--out":
			if i+1 >= len(args) {
				return fmt.Errorf("--out requires a directory")
			}
			outDir = args[i+1]
			i++
		default:
			if inputPath != "" {
				return fmt.Errorf("unexpected argument %q", args[i])
			}
			inputPath = args[i]
		}
	}
	if inputPath == "" || outDir == "" {
		return fmt.Errorf("usage: go run ./cmd/calendar 输入.json --out 输出目录")
	}

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	var in calendar.Input
	if err := json.Unmarshal(data, &in); err != nil {
		return fmt.Errorf("parsing %s: %w", inputPath, err)
	}
	if err := in.Validate(); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	instances, err := calendar.Instances(&in)
	if err != nil {
		return err
	}
	summary := calendar.Summary(&in, instances)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := writeInstancesCSV(filepath.Join(outDir, "instances.csv"), instances); err != nil {
		return err
	}
	if err := writeSummaryJSON(filepath.Join(outDir, "summary.json"), &in, summary); err != nil {
		return err
	}
	fmt.Printf("wrote %d instances to %s\n", len(instances), outDir)
	return nil
}

func writeInstancesCSV(path string, instances []calendar.Instance) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err := w.Write([]string{"activity_id", "start", "end"}); err != nil {
		return err
	}
	for _, inst := range instances {
		row := []string{
			inst.ActivityID,
			inst.Start.Format(calendar.DateLayout),
			inst.End.Format(calendar.DateLayout),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

type summaryFile struct {
	Window     calendar.Window   `json:"window"`
	Total      int               `json:"total_instances"`
	Activities []activitySummary `json:"activities"`
}

type activitySummary struct {
	ID        string `json:"id"`
	Instances int    `json:"instances"`
}

func writeSummaryJSON(path string, in *calendar.Input, counts map[string]int) error {
	ids := make([]string, 0, len(counts))
	for id := range counts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := summaryFile{Window: in.Window, Total: 0}
	for _, id := range ids {
		out.Activities = append(out.Activities, activitySummary{ID: id, Instances: counts[id]})
		out.Total += counts[id]
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
