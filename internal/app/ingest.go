package app

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/WingchunSiu/vRI/internal/harbor"
	"github.com/WingchunSiu/vRI/internal/objects"
	"github.com/WingchunSiu/vRI/internal/store"
)

type IngestResult struct {
	RecordID  string      `json:"record_id"`
	Created   bool        `json:"created"` // false when the dir was already ingested
	Job       *harbor.Job `json:"job"`
	DirDigest string      `json:"dir_digest"`
	Evaluates []string    `json:"evaluates"` // task object digests linked this run
}

type harborJobPayload struct {
	JobName     string            `json:"job_name"`
	SourcePath  string            `json:"source_path"`
	DirDigest   string            `json:"dir_digest"`
	Agent       string            `json:"agent"`
	Model       string            `json:"model,omitempty"`
	Environment string            `json:"environment"`
	TaskPaths   []string          `json:"task_paths"`
	Artifacts   map[string]string `json:"artifacts"`
	Trials      []trialPayload    `json:"trials"`
}

type trialPayload struct {
	Trial     string            `json:"trial"`
	Task      string            `json:"task"`
	Reward    *float64          `json:"reward,omitempty"`
	Exception string            `json:"exception,omitempty"`
	Artifacts map[string]string `json:"artifacts"`
}

// IngestHarborJob parses a Harbor job dir, copies its semantic artifact set
// into the object store, and creates one harbor_job record plus evaluates
// relations to any task objects already in the store. Re-ingesting the same
// directory is a no-op that returns the existing record.
func (a *App) IngestHarborJob(dir string) (*IngestResult, error) {
	job, err := harbor.ParseJob(dir)
	if err != nil {
		return nil, err
	}

	existing, err := a.findRecordByPayloadField("harbor_job", "source_path", job.Dir)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		res := &IngestResult{RecordID: existing.ID, Created: false, Job: job}
		return res, a.emit(res, func() {
			fmt.Fprintf(a.Out, "already ingested: %s (record %s)\n", job.Dir, existing.ID)
		})
	}

	dirDigest, _, err := objects.DigestDir(job.Dir)
	if err != nil {
		return nil, err
	}

	payload := harborJobPayload{
		JobName:     job.Name,
		SourcePath:  job.Dir,
		DirDigest:   dirDigest,
		Agent:       job.Agent,
		Model:       job.Model,
		Environment: job.Environment,
		TaskPaths:   job.TaskPaths,
		Artifacts:   map[string]string{},
	}

	artifactPaths, err := harbor.SemanticArtifacts(job)
	if err != nil {
		return nil, err
	}
	trialSet := map[string]bool{}
	for _, t := range job.Trials {
		trialSet[t.Name] = true
	}
	trialArtifacts := map[string]map[string]string{}
	for _, rel := range artifactPaths {
		digest, blobRel, size, _, err := objects.Write(a.St.ObjectsDir(), filepath.Join(job.Dir, filepath.FromSlash(rel)))
		if err != nil {
			return nil, fmt.Errorf("copying artifact %s: %w", rel, err)
		}
		if err := a.St.PutObject(&store.Object{Digest: digest, Kind: store.KindJobArtifact, Path: blobRel, SizeBytes: size, CreatedAt: now()}); err != nil {
			return nil, err
		}
		// Job-level artifacts live at the job root; trial artifacts under
		// <trial>/.... rel is slash-joined (built by SemanticArtifacts).
		if top, _, nested := strings.Cut(rel, "/"); nested && trialSet[top] {
			m, ok := trialArtifacts[top]
			if !ok {
				m = map[string]string{}
				trialArtifacts[top] = m
			}
			m[rel[len(top)+1:]] = digest
		} else {
			payload.Artifacts[rel] = digest
		}
	}
	for _, t := range job.Trials {
		payload.Trials = append(payload.Trials, trialPayload{
			Trial:     t.Name,
			Task:      t.Task,
			Reward:    t.Reward,
			Exception: t.Exception,
			Artifacts: trialArtifacts[t.Name],
		})
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	rec, err := a.putRecord("harbor_job", raw, "vri")
	if err != nil {
		return nil, err
	}

	res := &IngestResult{RecordID: rec.ID, Created: true, Job: job, DirDigest: dirDigest}
	for _, taskPath := range job.TaskPaths {
		taskDigest, _, err := objects.DigestDir(taskPath)
		if err != nil {
			// The task source may be gone or unreadable; the job is still
			// ingested, just without an evaluates link.
			continue
		}
		ok, err := a.St.HasObject(taskDigest)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		linked, err := a.Link(store.EndpointRecord, rec.ID, store.EndpointObject, taskDigest, "evaluates", "")
		if err != nil {
			return nil, err
		}
		if linked {
			res.Evaluates = append(res.Evaluates, taskDigest)
		}
	}

	return res, a.emit(res, func() {
		fmt.Fprintf(a.Out, "ingested %s: record %s\n", job.Name, rec.ID)
		fmt.Fprintf(a.Out, "  agent=%s model=%s environment=%s dir=%s\n", job.Agent, job.Model, job.Environment, dirDigest)
		for _, t := range payload.Trials {
			reward := "n/a"
			if t.Reward != nil {
				reward = fmt.Sprintf("%v", *t.Reward)
			}
			fmt.Fprintf(a.Out, "  trial %s task=%s reward=%s\n", t.Trial, t.Task, reward)
		}
		for _, d := range res.Evaluates {
			fmt.Fprintf(a.Out, "  evaluates %s\n", d)
		}
	})
}
