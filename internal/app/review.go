package app

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/WingchunSiu/vRI/internal/store"
)

type ReviewPacket struct {
	GeneratedAt string       `json:"generated_at"`
	Candidates  []ReviewItem `json:"candidates"`
	Claims      []ReviewItem `json:"claims"`
	Runs        []ReviewRun  `json:"runs"`
	OpenItems   []string     `json:"open_items"`
}

type ReviewItem struct {
	ID             string   `json:"id"`
	Summary        string   `json:"summary"`
	SupportedBy    []string `json:"supported_by,omitempty"`
	ContradictedBy []string `json:"contradicted_by,omitempty"`
}

type ReviewRun struct {
	RecordID    string   `json:"record_id"`
	Job         string   `json:"job"`
	Agent       string   `json:"agent"`
	Model       string   `json:"model,omitempty"`
	Environment string   `json:"environment"`
	Trials      []string `json:"trials"` // "task reward" summaries
}

// Review rebuilds the review packet from the store: proposed candidates,
// claims with their evidence links, ingested runs, and open items.
func (a *App) Review() (*ReviewPacket, error) {
	p := &ReviewPacket{GeneratedAt: now()}

	candidates, err := a.St.ListRecords("task_candidate")
	if err != nil {
		return nil, err
	}
	for _, r := range candidates {
		p.Candidates = append(p.Candidates, ReviewItem{ID: r.ID, Summary: summarize(r.Payload)})
	}

	claims, err := a.St.ListRecords("claim")
	if err != nil {
		return nil, err
	}
	for _, r := range claims {
		item := ReviewItem{ID: r.ID, Summary: summarize(r.Payload)}
		rels, err := a.St.RelationsTo(store.EndpointRecord, r.ID, "")
		if err != nil {
			return nil, err
		}
		for _, rel := range rels {
			ref := rel.SrcID
			switch rel.Type {
			case "supports":
				item.SupportedBy = append(item.SupportedBy, ref)
			case "contradicts":
				item.ContradictedBy = append(item.ContradictedBy, ref)
			}
		}
		p.Claims = append(p.Claims, item)
	}

	jobs, err := a.St.ListRecords("harbor_job")
	if err != nil {
		return nil, err
	}
	for _, r := range jobs {
		var jp harborJobPayload
		if err := json.Unmarshal([]byte(r.Payload), &jp); err != nil {
			continue
		}
		run := ReviewRun{RecordID: r.ID, Job: jp.JobName, Agent: jp.Agent, Model: jp.Model, Environment: jp.Environment}
		for _, t := range jp.Trials {
			reward := "n/a"
			if t.Reward != nil {
				reward = strconv.FormatFloat(*t.Reward, 'g', -1, 64)
			}
			run.Trials = append(run.Trials, fmt.Sprintf("%s reward %s", t.Task, reward))
		}
		p.Runs = append(p.Runs, run)
	}

	manifests, err := a.St.ListRecords("benchmark_manifest")
	if err != nil {
		return nil, err
	}
	for _, r := range manifests {
		var m struct {
			ID          string   `json:"id"`
			Limitations []string `json:"limitations"`
		}
		if err := json.Unmarshal([]byte(r.Payload), &m); err != nil {
			continue
		}
		for _, lim := range m.Limitations {
			p.OpenItems = append(p.OpenItems, fmt.Sprintf("manifest %s: %s", m.ID, lim))
		}
		rels, err := a.St.RelationsFrom(store.EndpointRecord, r.ID)
		if err != nil {
			return nil, err
		}
		decided := false
		for _, rel := range rels {
			if rel.Type == "decided_by" {
				decided = true
			}
		}
		if !decided {
			p.OpenItems = append(p.OpenItems, fmt.Sprintf("manifest %s: release decision pending", m.ID))
		}
	}

	return p, a.emit(p, func() {
		a.printReview(p)
	})
}

func (a *App) printReview(p *ReviewPacket) {
	fmt.Fprintf(a.Out, "# Review packet (generated %s)\n\n", p.GeneratedAt)

	fmt.Fprintln(a.Out, "## Candidates")
	if len(p.Candidates) == 0 {
		fmt.Fprintln(a.Out, "(none)")
	}
	for _, c := range p.Candidates {
		fmt.Fprintf(a.Out, "- %s %s\n", c.ID, c.Summary)
	}
	fmt.Fprintln(a.Out)

	fmt.Fprintln(a.Out, "## Claims")
	if len(p.Claims) == 0 {
		fmt.Fprintln(a.Out, "(none)")
	}
	for _, c := range p.Claims {
		fmt.Fprintf(a.Out, "- %s %s\n", c.ID, c.Summary)
		if len(c.SupportedBy) > 0 {
			fmt.Fprintf(a.Out, "  supported by: %s\n", strings.Join(c.SupportedBy, ", "))
		}
		if len(c.ContradictedBy) > 0 {
			fmt.Fprintf(a.Out, "  contradicted by: %s\n", strings.Join(c.ContradictedBy, ", "))
		}
	}
	fmt.Fprintln(a.Out)

	fmt.Fprintln(a.Out, "## Runs")
	if len(p.Runs) == 0 {
		fmt.Fprintln(a.Out, "(none)")
	}
	for _, r := range p.Runs {
		model := r.Model
		if model == "" {
			model = "-"
		}
		fmt.Fprintf(a.Out, "- %s (agent=%s model=%s env=%s): %s\n",
			r.Job, r.Agent, model, r.Environment, strings.Join(r.Trials, "; "))
	}
	fmt.Fprintln(a.Out)

	fmt.Fprintln(a.Out, "## Open items")
	if len(p.OpenItems) == 0 {
		fmt.Fprintln(a.Out, "(none)")
	}
	for _, item := range p.OpenItems {
		fmt.Fprintf(a.Out, "- %s\n", item)
	}
}
