package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/WingchunSiu/vRI/internal/objects"
	"github.com/WingchunSiu/vRI/internal/store"
)

type ReleaseResult struct {
	ManifestRecordID string   `json:"manifest_record_id"`
	DecisionRecordID string   `json:"decision_record_id,omitempty"`
	Includes         []string `json:"includes"` // task object digests bound
	AlreadyReleased  bool     `json:"already_released"`
}

// Release validates a manifest against the store and, if every reference
// resolves, writes the manifest, includes relations, and review decision. An
// approval also writes the decided_by release binding; other outcomes do not.
// Any dangling reference refuses the operation with an error naming each
// problem.
func (a *App) Release(manifestPath, reviewer, outcome string) (*ReleaseResult, error) {
	if outcome != "approve" && outcome != "reject" && outcome != "more-evidence" {
		return nil, fmt.Errorf("unsupported review outcome %q (want approve, reject, or more-evidence)", outcome)
	}

	m, raw, source, err := a.resolveManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	var problems []string
	for _, task := range m.Tasks {
		ok, err := a.St.HasObject(task.Digest)
		if err != nil {
			return nil, err
		}
		if !ok {
			problems = append(problems, fmt.Sprintf("task %s: object %s not in store", task.ID, task.Digest))
		}
	}

	// job_path entries may be relative to the benchmark root rather than the
	// manifest's own directory. resolveJobDir searches the manifest directory
	// and its ancestors.
	claimed, err := a.ingestedDigests()
	if err != nil {
		return nil, err
	}
	for _, run := range m.Runs {
		jobDir, found := resolveJobDir(source, run.JobPath)
		if !found {
			problems = append(problems, fmt.Sprintf("run %s: job directory not found", run.JobPath))
			continue
		}
		actual, _, err := objects.DigestDir(jobDir)
		if err != nil {
			return nil, err
		}
		if actual != run.JobDigest {
			problems = append(problems, fmt.Sprintf("run %s: digest mismatch (manifest %s, recomputed %s)", run.JobPath, run.JobDigest, actual))
			continue
		}
		if !claimed[run.JobDigest] {
			problems = append(problems, fmt.Sprintf("run %s: no ingested harbor_job record (ingest with `vri ingest harbor-job`)", run.JobPath))
		}
	}
	if len(problems) > 0 {
		return nil, fmt.Errorf("release refused: %d dangling reference(s):\n  %s", len(problems), strings.Join(problems, "\n  "))
	}

	res := &ReleaseResult{}
	manifestRec, err := a.findRecordByPayloadField("benchmark_manifest", "id", m.ID)
	if err != nil {
		return nil, err
	}
	if manifestRec == nil {
		manifestRec, err = a.putRecord("benchmark_manifest", json.RawMessage(raw), "user")
		if err != nil {
			return nil, err
		}
	}
	res.ManifestRecordID = manifestRec.ID

	for _, task := range m.Tasks {
		if _, err := a.Link(store.EndpointRecord, manifestRec.ID, store.EndpointObject, task.Digest, "includes", ""); err != nil {
			return nil, err
		}
		res.Includes = append(res.Includes, task.Digest)
	}

	rels, err := a.St.RelationsFrom(store.EndpointRecord, manifestRec.ID)
	if err != nil {
		return nil, err
	}
	for _, rel := range rels {
		if rel.Type == "decided_by" {
			res.AlreadyReleased = true
			break
		}
	}
	if !res.AlreadyReleased {
		decision, _ := json.Marshal(map[string]string{
			"outcome":         outcome,
			"reviewer":        reviewer,
			"scope":           m.ID,
			"manifest_record": manifestRec.ID,
		})
		rec, err := a.putRecord("decision", decision, "user")
		if err != nil {
			return nil, err
		}
		res.DecisionRecordID = rec.ID
		if outcome == "approve" {
			if _, err := a.Link(store.EndpointRecord, manifestRec.ID, store.EndpointRecord, rec.ID, "decided_by", ""); err != nil {
				return nil, err
			}
		}
	}

	return res, a.emit(res, func() {
		if res.AlreadyReleased {
			fmt.Fprintf(a.Out, "manifest %s already released (record %s)\n", m.ID, res.ManifestRecordID)
			return
		}
		if outcome == "approve" {
			fmt.Fprintf(a.Out, "released %s\n  manifest record %s\n  decision record %s (%s by %s)\n",
				m.ID, res.ManifestRecordID, res.DecisionRecordID, outcome, reviewer)
			for _, d := range res.Includes {
				fmt.Fprintf(a.Out, "  includes %s\n", d)
			}
			return
		}
		fmt.Fprintf(a.Out, "recorded %s for %s\n  manifest record %s\n  decision record %s (%s by %s)\n",
			outcome, m.ID, res.ManifestRecordID, res.DecisionRecordID, outcome, reviewer)
	})
}

// resolveJobDir resolves a manifest job_path against the directory of the
// manifest file and up to 3 ancestors, taking the first match. A manifest
// loaded from a record has no filesystem base and resolves nothing.
func resolveJobDir(manifestSource, jobPath string) (string, bool) {
	info, err := os.Stat(manifestSource)
	if err != nil || info.IsDir() {
		return "", false
	}
	dir := filepath.Dir(manifestSource)
	for range 4 {
		cand := filepath.Join(dir, filepath.FromSlash(jobPath))
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

// ingestedDigests returns the dir digests of all ingested harbor_job records.
func (a *App) ingestedDigests() (map[string]bool, error) {
	recs, err := a.St.ListRecords("harbor_job")
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, r := range recs {
		var p harborJobPayload
		if err := json.Unmarshal([]byte(r.Payload), &p); err == nil && p.DirDigest != "" {
			out[p.DirDigest] = true
		}
	}
	return out, nil
}
