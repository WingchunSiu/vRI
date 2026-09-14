package app

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/WingchunSiu/vRI/internal/store"
)

type RunsResult struct {
	Manifest string    `json:"manifest"`
	Tasks    []string  `json:"tasks"`
	Rows     []RunsRow `json:"rows"`
	Missing  []string  `json:"missing"` // manifest runs with no ingested harbor_job record
}

type RunsRow struct {
	Job     string             `json:"job"`
	Agent   string             `json:"agent"`
	Model   string             `json:"model,omitempty"`
	Rewards map[string]float64 `json:"rewards"` // task id -> reward
}

// Runs builds the reward matrix across all ingested jobs linked (via
// evaluates relations) to the tasks of the given manifest.
func (a *App) Runs(ref string) (*RunsResult, error) {
	m, _, _, err := a.resolveManifest(ref)
	if err != nil {
		return nil, err
	}

	res := &RunsResult{Manifest: m.ID}
	jobRows := map[string]*RunsRow{}
	for _, task := range m.Tasks {
		res.Tasks = append(res.Tasks, task.ID)
		rels, err := a.St.RelationsTo(store.EndpointObject, task.Digest, "evaluates")
		if err != nil {
			return nil, err
		}
		for _, rel := range rels {
			rec, err := a.St.GetRecord(rel.SrcID)
			if err != nil || rec.Type != "harbor_job" {
				continue
			}
			var p harborJobPayload
			if err := json.Unmarshal([]byte(rec.Payload), &p); err != nil {
				continue
			}
			row, ok := jobRows[rec.ID]
			if !ok {
				row = &RunsRow{Job: p.JobName, Agent: p.Agent, Model: p.Model, Rewards: map[string]float64{}}
				jobRows[rec.ID] = row
			}
			for _, t := range p.Trials {
				if t.Task == task.ID && t.Reward != nil {
					row.Rewards[task.ID] = *t.Reward
				}
			}
		}
	}
	for _, row := range jobRows {
		res.Rows = append(res.Rows, *row)
	}
	sort.Slice(res.Rows, func(i, j int) bool { return res.Rows[i].Job < res.Rows[j].Job })

	// Manifest runs whose job digest no ingested record claims.
	claimed := map[string]bool{}
	recs, err := a.St.ListRecords("harbor_job")
	if err != nil {
		return nil, err
	}
	for _, r := range recs {
		var p harborJobPayload
		if err := json.Unmarshal([]byte(r.Payload), &p); err == nil {
			claimed[p.DirDigest] = true
		}
	}
	for _, run := range m.Runs {
		if !claimed[run.JobDigest] {
			res.Missing = append(res.Missing, run.JobPath)
		}
	}

	return res, a.emit(res, func() {
		a.printRunsTable(res)
	})
}

func (a *App) printRunsTable(res *RunsResult) {
	headers := []string{"JOB", "AGENT", "MODEL"}
	headers = append(headers, res.Tasks...)
	var cells [][]string
	for _, row := range res.Rows {
		model := row.Model
		if model == "" {
			model = "-"
		}
		line := []string{row.Job, row.Agent, model}
		for _, taskID := range res.Tasks {
			if r, ok := row.Rewards[taskID]; ok {
				line = append(line, strconv.FormatFloat(r, 'g', -1, 64))
			} else {
				line = append(line, "-")
			}
		}
		cells = append(cells, line)
	}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, line := range cells {
		for i, c := range line {
			if len(c) > widths[i] {
				widths[i] = len(c)
			}
		}
	}
	printLine := func(parts []string) {
		var b strings.Builder
		for i, p := range parts {
			if i > 0 {
				b.WriteString("  ")
			}
			fmt.Fprintf(&b, "%-*s", widths[i], p)
		}
		fmt.Fprintln(a.Out, strings.TrimRight(b.String(), " "))
	}
	printLine(headers)
	for _, line := range cells {
		printLine(line)
	}
	for _, missing := range res.Missing {
		fmt.Fprintf(a.Out, "not ingested: %s\n", missing)
	}
}
