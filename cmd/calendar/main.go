package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"example.com/recurring-calendar-expander/internal/expand"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	inputPath, outDir, err := parseArgs(os.Args[1:])
	if err != nil {
		return err
	}
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	var in expand.Input
	if err := json.Unmarshal(data, &in); err != nil {
		return fmt.Errorf("parsing input: %w", err)
	}
	if err := expand.Validate(&in); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}
	instances := expand.Instances(&in)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := writeCSV(filepath.Join(outDir, "instances.csv"), instances); err != nil {
		return err
	}
	return writeSummary(filepath.Join(outDir, "summary.json"), &in, instances)
}

// parseArgs accepts "<input.json> --out <dir>" with flags in any position.
func parseArgs(args []string) (inputPath, outDir string, err error) {
	var positional []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--out" {
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--out requires a value")
			}
			outDir = args[i+1]
			i++
			continue
		}
		positional = append(positional, args[i])
	}
	if len(positional) != 1 || outDir == "" {
		return "", "", fmt.Errorf("usage: go run ./cmd/calendar <input.json> --out <output-dir>")
	}
	return positional[0], outDir, nil
}

type summaryEntry struct {
	ActivityID string `json:"activity_id"`
	Count      int    `json:"count"`
}

type summary struct {
	WindowStart string         `json:"window_start"`
	WindowEnd   string         `json:"window_end"`
	Total       int            `json:"total_instances"`
	Activities  []summaryEntry `json:"activities"`
}

func writeCSV(path string, instances []expand.Instance) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"activity_id", "start", "end"}); err != nil {
		return err
	}
	for _, inst := range instances {
		if err := w.Write([]string{inst.ActivityID, inst.Start.String(), inst.End.String()}); err != nil {
			return err
		}
	}
	return w.Error()
}

func writeSummary(path string, in *expand.Input, instances []expand.Instance) error {
	counts := map[string]int{}
	for _, inst := range instances {
		counts[inst.ActivityID]++
	}
	ids := make([]string, 0, len(in.Activities))
	for _, a := range in.Activities {
		ids = append(ids, a.ID)
	}
	sort.Strings(ids)
	s := summary{
		WindowStart: in.Window.Start.String(),
		WindowEnd:   in.Window.End.String(),
		Total:       len(instances),
		Activities:  make([]summaryEntry, 0, len(ids)),
	}
	for _, id := range ids {
		s.Activities = append(s.Activities, summaryEntry{ActivityID: id, Count: counts[id]})
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
