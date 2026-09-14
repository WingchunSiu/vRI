package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WingchunSiu/vRI/internal/objects"
	"github.com/WingchunSiu/vRI/internal/store"
)

type fixtureJob struct {
	name   string
	agent  string
	model  string
	reward float64
}

var fixtureJobs = []fixtureJob{
	{name: "oracle-job", agent: "oracle", reward: 1},
	{name: "nop-job", agent: "nop", reward: 0},
	{name: "agent-a-job", agent: "test-agent", model: "model-a", reward: 1},
	{name: "agent-b-job", agent: "test-agent", model: "model-b", reward: 0},
}

func newTestApp(t *testing.T) (*App, *bytes.Buffer) {
	t.Helper()
	st, err := store.Init(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	buf := &bytes.Buffer{}
	return &App{St: st, Out: buf}, buf
}

// makeFixture creates the smallest portable source/task/jobs/manifest tree
// needed by the app behavior tests. It deliberately contains no real session
// data, provider logs, or machine-specific paths persisted in the repository.
func makeFixture(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "eval-fixture")
	writeFile := func(path string, content []byte) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeJSON := func(path string, value any) {
		t.Helper()
		raw, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		writeFile(path, append(raw, '\n'))
	}

	writeFile(filepath.Join(root, "00-source", "source.md"), []byte("synthetic source evidence\n"))
	taskDir := filepath.Join(root, "02-task", "sample-task")
	writeFile(filepath.Join(taskDir, "instruction.md"), []byte("produce the expected artifact\n"))

	for _, job := range fixtureJobs {
		jobDir := filepath.Join(root, "03-runs", job.name)
		trialName := "sample-task__trial"
		writeJSON(filepath.Join(jobDir, "config.json"), map[string]any{
			"job_name":    job.name,
			"environment": map[string]any{"type": "local"},
			"agents":      []any{map[string]any{"name": job.agent, "model_name": job.model}},
			"tasks":       []any{map[string]any{"path": taskDir}},
		})
		writeJSON(filepath.Join(jobDir, "result.json"), map[string]any{"completed": true})
		writeJSON(filepath.Join(jobDir, trialName, "config.json"), map[string]any{"task": "sample-task"})
		writeJSON(filepath.Join(jobDir, trialName, "result.json"), map[string]any{
			"trial_name": trialName,
			"config": map[string]any{
				"agent": map[string]any{"name": job.agent, "model_name": job.model},
			},
			"verifier_result": map[string]any{"rewards": map[string]any{"reward": job.reward}},
		})
		writeFile(filepath.Join(jobDir, trialName, "verifier", "reward.txt"), []byte(fmt.Sprintf("%g\n", job.reward)))
	}

	taskDigest, _, err := objects.DigestDir(taskDir)
	if err != nil {
		t.Fatal(err)
	}
	runs := make([]any, 0, len(fixtureJobs))
	for _, job := range fixtureJobs {
		jobPath := filepath.Join(root, "03-runs", job.name)
		jobDigest, _, err := objects.DigestDir(jobPath)
		if err != nil {
			t.Fatal(err)
		}
		runs = append(runs, map[string]any{
			"job_path":   filepath.ToSlash(filepath.Join("03-runs", job.name)),
			"job_digest": jobDigest,
			"system":     map[string]any{"agent": job.agent, "model": job.model},
			"reward":     job.reward,
			"role":       "test",
		})
	}
	writeJSON(filepath.Join(root, "04-benchmark", "manifest.json"), map[string]any{
		"kind": "benchmark_manifest",
		"id":   "test-benchmark",
		"tasks": []any{map[string]any{
			"id": "sample-task", "path": "02-task/sample-task", "digest": taskDigest,
		}},
		"runs":        runs,
		"limitations": []any{"synthetic test data"},
	})
	return root
}

func putTaskPackage(t *testing.T, a *App, fixture string) string {
	t.Helper()
	res, err := a.ObjectPut(filepath.Join(fixture, "02-task", "sample-task"), "")
	if err != nil {
		t.Fatal(err)
	}
	return res.Digest
}

func ingestFixtureJobs(t *testing.T, a *App, fixture string) {
	t.Helper()
	for _, job := range fixtureJobs {
		j := job.name
		if _, err := a.IngestHarborJob(filepath.Join(fixture, "03-runs", j)); err != nil {
			t.Fatalf("ingesting %s: %v", j, err)
		}
	}
}

func TestImportIdempotent(t *testing.T) {
	a, _ := newTestApp(t)
	fixture := makeFixture(t)
	src := filepath.Join(fixture, "00-source", "source.md")

	first, err := a.Import(src, store.KindSourceExcerpt, "user")
	if err != nil {
		t.Fatal(err)
	}
	if !first.CreatedObject || !first.CreatedRecord {
		t.Fatalf("first import should create object and record: %+v", first)
	}
	second, err := a.Import(src, store.KindSourceExcerpt, "user")
	if err != nil {
		t.Fatal(err)
	}
	if second.CreatedObject || second.CreatedRecord {
		t.Fatalf("second import should create nothing: %+v", second)
	}
	if second.RecordID != first.RecordID || second.Digest != first.Digest {
		t.Fatal("second import should return the same record and digest")
	}
	objs, err := a.St.ListObjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(objs) != 1 {
		t.Fatalf("objects = %d, want 1", len(objs))
	}
	recs, err := a.St.ListRecords("source")
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("source records = %d, want 1", len(recs))
	}
}

func TestIngestIdempotent(t *testing.T) {
	a, _ := newTestApp(t)
	fixture := makeFixture(t)
	putTaskPackage(t, a, fixture)
	jobDir := filepath.Join(fixture, "03-runs", "agent-a-job")

	first, err := a.IngestHarborJob(jobDir)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Created || len(first.Evaluates) != 1 {
		t.Fatalf("first ingest should create record and evaluates link: %+v", first)
	}
	objsAfterFirst, err := a.St.ListObjects()
	if err != nil {
		t.Fatal(err)
	}

	second, err := a.IngestHarborJob(jobDir)
	if err != nil {
		t.Fatal(err)
	}
	if second.Created {
		t.Fatal("re-ingesting the same dir must create nothing")
	}
	if second.RecordID != first.RecordID {
		t.Fatal("re-ingest should return the existing record")
	}
	recs, err := a.St.ListRecords("harbor_job")
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("harbor_job records = %d, want 1", len(recs))
	}
	objsAfterSecond, err := a.St.ListObjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(objsAfterSecond) != len(objsAfterFirst) {
		t.Fatalf("object count changed on re-ingest: %d -> %d", len(objsAfterFirst), len(objsAfterSecond))
	}
	rels, err := a.St.RelationsFrom(store.EndpointRecord, first.RecordID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rels) != 1 || rels[0].Type != "evaluates" {
		t.Fatalf("expected exactly one evaluates relation, got %+v", rels)
	}
}

func TestIngestCorrectness(t *testing.T) {
	a, _ := newTestApp(t)
	fixture := makeFixture(t)
	cases := []struct {
		job   string
		agent string
		model string
		want  float64
	}{
		{"oracle-job", "oracle", "", 1.0},
		{"nop-job", "nop", "", 0.0},
		{"agent-a-job", "test-agent", "model-a", 1.0},
		{"agent-b-job", "test-agent", "model-b", 0.0},
	}
	for _, c := range cases {
		res, err := a.IngestHarborJob(filepath.Join(fixture, "03-runs", c.job))
		if err != nil {
			t.Fatalf("%s: %v", c.job, err)
		}
		if res.Job.Agent != c.agent || res.Job.Model != c.model {
			t.Errorf("%s: agent/model = %s/%s, want %s/%s", c.job, res.Job.Agent, res.Job.Model, c.agent, c.model)
		}
		if len(res.Job.Trials) != 1 || res.Job.Trials[0].Reward == nil || *res.Job.Trials[0].Reward != c.want {
			t.Errorf("%s: trials = %+v, want one trial with reward %v", c.job, res.Job.Trials, c.want)
		}

		rec, err := a.St.GetRecord(res.RecordID)
		if err != nil {
			t.Fatal(err)
		}
		var p harborJobPayload
		if err := json.Unmarshal([]byte(rec.Payload), &p); err != nil {
			t.Fatal(err)
		}
		if p.DirDigest == "" || p.Environment != "local" {
			t.Errorf("%s: payload missing dir digest or environment: %s", c.job, rec.Payload)
		}
		// Every artifact referenced by the record must exist as an object.
		for rel, digest := range p.Artifacts {
			if ok, _ := a.St.HasObject(digest); !ok {
				t.Errorf("%s: artifact %s digest %s not in store", c.job, rel, digest)
			}
		}
		if len(p.Artifacts) != 2 {
			t.Errorf("%s: job-level artifacts = %v, want config.json and result.json only", c.job, p.Artifacts)
		}
		for _, tr := range p.Trials {
			for rel, digest := range tr.Artifacts {
				if ok, _ := a.St.HasObject(digest); !ok {
					t.Errorf("%s: trial artifact %s digest %s not in store", c.job, rel, digest)
				}
			}
			if tr.Artifacts["verifier/reward.txt"] == "" {
				t.Errorf("%s: trial %s missing verifier/reward.txt artifact (got %v)", c.job, tr.Trial, tr.Artifacts)
			}
			if tr.Artifacts["config.json"] == "" || tr.Artifacts["result.json"] == "" {
				t.Errorf("%s: trial %s missing config/result artifacts (got %v)", c.job, tr.Trial, tr.Artifacts)
			}
		}
	}
}

// Reference integrity: writing a relation or a manifest record that points at
// a missing object must fail at write time.
func TestIntegrity(t *testing.T) {
	a, _ := newTestApp(t)
	fixture := makeFixture(t)
	missing := "sha256:" + strings.Repeat("0", 64)

	if _, err := a.Link(store.EndpointRecord, "01JXXXXXXXXXXXXXXXXXXXXXXX", store.EndpointObject, missing, "evaluates", ""); err == nil {
		t.Fatal("relation with missing endpoints should fail")
	} else if !strings.Contains(err.Error(), missing) && !strings.Contains(err.Error(), "01JXXX") {
		t.Fatalf("error should name the missing endpoint: %v", err)
	}

	src, err := a.Import(filepath.Join(fixture, "00-source", "source.md"), store.KindSourceExcerpt, "user")
	if err != nil {
		t.Fatal(err)
	}
	rec, err := a.RecordPut("claim", []byte(`{"text":"x"}`), "user")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Link(store.EndpointObject, src.Digest, store.EndpointRecord, rec.ID, "supports", ""); err != nil {
		t.Fatalf("relation between existing endpoints should succeed: %v", err)
	}
	if _, err := a.Link(store.EndpointObject, missing, store.EndpointRecord, rec.ID, "supports", ""); err == nil {
		t.Fatal("relation with missing object endpoint should fail")
	}

	manifest := `{"id":"m1","tasks":[{"id":"t","path":"p","digest":"` + missing + `"}]}`
	if _, err := a.RecordPut("benchmark_manifest", []byte(manifest), "user"); err == nil {
		t.Fatal("manifest record referencing a missing object should fail")
	} else if !strings.Contains(err.Error(), missing) {
		t.Fatalf("error should name the missing digest: %v", err)
	}
}

func TestReview(t *testing.T) {
	a, buf := newTestApp(t)
	fixture := makeFixture(t)
	if _, err := a.Import(filepath.Join(fixture, "00-source", "source.md"), store.KindSourceExcerpt, "user"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.RecordPut("task_candidate", []byte(`{"task":"sample-task","rationale":"synthetic behavior test"}`), "agent"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.RecordPut("claim", []byte(`{"text":"the task gives no free reward"}`), "agent"); err != nil {
		t.Fatal(err)
	}
	ingestFixtureJobs(t, a, fixture)

	packet, err := a.Review()
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, section := range []string{"## Candidates", "## Claims", "## Runs", "## Open items"} {
		if !strings.Contains(out, section) {
			t.Errorf("review output missing section %q", section)
		}
	}
	for _, job := range []string{"oracle-job", "nop-job", "agent-a-job", "agent-b-job"} {
		if !strings.Contains(out, job) {
			t.Errorf("review output missing job %q", job)
		}
	}
	if len(packet.Candidates) != 1 || len(packet.Claims) != 1 || len(packet.Runs) != 4 {
		t.Errorf("packet counts: candidates=%d claims=%d runs=%d", len(packet.Candidates), len(packet.Claims), len(packet.Runs))
	}
}

func TestReleaseRefusesTamperedManifest(t *testing.T) {
	a, _ := newTestApp(t)
	fixture := makeFixture(t)
	putTaskPackage(t, a, fixture)
	ingestFixtureJobs(t, a, fixture)

	manifestPath := filepath.Join(fixture, "04-benchmark", "manifest.json")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	tampered := "sha256:" + strings.Repeat("0", 64)
	m["tasks"].([]any)[0].(map[string]any)["digest"] = tampered
	tamperedRaw, _ := json.Marshal(m)
	if err := os.WriteFile(manifestPath, tamperedRaw, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = a.Release(manifestPath, "user", "approve")
	if err == nil {
		t.Fatal("release should refuse a manifest with a tampered task digest")
	}
	if !strings.Contains(err.Error(), tampered) {
		t.Fatalf("error should name the dangling digest %s: %v", tampered, err)
	}
	recs, err := a.St.ListRecords("decision")
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 0 {
		t.Fatal("a refused release must not write decision records")
	}
}

func TestReleaseHappyPath(t *testing.T) {
	a, _ := newTestApp(t)
	fixture := makeFixture(t)
	taskDigest := putTaskPackage(t, a, fixture)
	ingestFixtureJobs(t, a, fixture)

	res, err := a.Release(filepath.Join(fixture, "04-benchmark", "manifest.json"), "user", "approve")
	if err != nil {
		t.Fatal(err)
	}
	if res.AlreadyReleased || res.DecisionRecordID == "" {
		t.Fatalf("first release should write a decision: %+v", res)
	}
	if len(res.Includes) != 1 || res.Includes[0] != taskDigest {
		t.Fatalf("includes = %v, want [%s]", res.Includes, taskDigest)
	}

	rels, err := a.St.RelationsFrom(store.EndpointRecord, res.ManifestRecordID)
	if err != nil {
		t.Fatal(err)
	}
	types := map[string]bool{}
	for _, r := range rels {
		types[r.Type] = true
	}
	if !types["includes"] || !types["decided_by"] {
		t.Fatalf("release relations missing: %v", types)
	}

	second, err := a.Release(filepath.Join(fixture, "04-benchmark", "manifest.json"), "user", "approve")
	if err != nil {
		t.Fatal(err)
	}
	if !second.AlreadyReleased {
		t.Fatal("second release should report already released")
	}
	decisions, err := a.St.ListRecords("decision")
	if err != nil {
		t.Fatal(err)
	}
	if len(decisions) != 1 {
		t.Fatalf("decision records = %d, want 1", len(decisions))
	}
}

func TestReleaseNonApprovalDoesNotRelease(t *testing.T) {
	a, buf := newTestApp(t)
	fixture := makeFixture(t)
	putTaskPackage(t, a, fixture)
	ingestFixtureJobs(t, a, fixture)
	manifestPath := filepath.Join(fixture, "04-benchmark", "manifest.json")

	rejected, err := a.Release(manifestPath, "reviewer", "reject")
	if err != nil {
		t.Fatal(err)
	}
	if rejected.AlreadyReleased || rejected.DecisionRecordID == "" {
		t.Fatalf("reject should record a decision without releasing: %+v", rejected)
	}
	rels, err := a.St.RelationsFrom(store.EndpointRecord, rejected.ManifestRecordID)
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range rels {
		if rel.Type == "decided_by" {
			t.Fatal("reject must not create a release binding")
		}
	}
	if strings.Contains(buf.String(), "released test-benchmark") {
		t.Fatalf("reject output claims release: %s", buf.String())
	}

	approved, err := a.Release(manifestPath, "reviewer", "approve")
	if err != nil {
		t.Fatal(err)
	}
	if approved.AlreadyReleased || approved.DecisionRecordID == "" {
		t.Fatalf("approval after rejection should release: %+v", approved)
	}
	rels, err = a.St.RelationsFrom(store.EndpointRecord, approved.ManifestRecordID)
	if err != nil {
		t.Fatal(err)
	}
	foundRelease := false
	for _, rel := range rels {
		if rel.Type == "decided_by" {
			foundRelease = true
		}
	}
	if !foundRelease {
		t.Fatal("approval should create a release binding")
	}
}

func TestReleaseRejectsUnknownOutcome(t *testing.T) {
	a, _ := newTestApp(t)
	if _, err := a.Release("unused", "reviewer", "maybe"); err == nil {
		t.Fatal("unknown outcome should fail before reading a manifest")
	}
}

func TestReleaseRefusesUnIngestedRun(t *testing.T) {
	a, _ := newTestApp(t)
	fixture := makeFixture(t)
	putTaskPackage(t, a, fixture)
	// Ingest only three of the four runs.
	for _, j := range []string{"oracle-job", "nop-job", "agent-a-job"} {
		if _, err := a.IngestHarborJob(filepath.Join(fixture, "03-runs", j)); err != nil {
			t.Fatal(err)
		}
	}
	_, err := a.Release(filepath.Join(fixture, "04-benchmark", "manifest.json"), "user", "approve")
	if err == nil {
		t.Fatal("release should refuse a manifest with an un-ingested run")
	}
	if !strings.Contains(err.Error(), "agent-b-job") {
		t.Fatalf("error should name the un-ingested run: %v", err)
	}
}

func TestRuns(t *testing.T) {
	a, buf := newTestApp(t)
	fixture := makeFixture(t)
	putTaskPackage(t, a, fixture)
	ingestFixtureJobs(t, a, fixture)

	res, err := a.Runs(filepath.Join(fixture, "04-benchmark", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]float64{
		"oracle-job":  1.0,
		"nop-job":     0.0,
		"agent-a-job": 1.0,
		"agent-b-job": 0.0,
	}
	if len(res.Rows) != len(want) {
		t.Fatalf("rows = %d, want %d", len(res.Rows), len(want))
	}
	for _, row := range res.Rows {
		w, ok := want[row.Job]
		if !ok {
			t.Fatalf("unexpected job %q in runs output", row.Job)
		}
		got, ok := row.Rewards["sample-task"]
		if !ok || got != w {
			t.Errorf("%s reward = %v, want %v", row.Job, got, w)
		}
	}
	if len(res.Missing) != 0 {
		t.Errorf("missing = %v, want none", res.Missing)
	}
	out := buf.String()
	for job := range want {
		if !strings.Contains(out, job) {
			t.Errorf("runs table missing %q", job)
		}
	}
}
