package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WingchunSiu/vRI/internal/store"
)

func TestEvidenceContextSurvivesSourceRemoval(t *testing.T) {
	a, _ := newTestApp(t)
	sourceDir := filepath.Join(t.TempDir(), "session")
	writeFixtureFile(t, filepath.Join(sourceDir, "notes.md"), "A verifier crash can masquerade as model failure.\n")
	writeFixtureFile(t, filepath.Join(sourceDir, "nested", "result.json"), `{"reward":0}`)

	first, err := a.EvidenceCapture(EvidenceCaptureOptions{
		Path: sourceDir, Title: "verifier investigation", Kind: "session", Completeness: CompletenessPartial,
		Scopes: []string{"verifier-repair"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !first.CreatedRecord || !first.CreatedObject || !first.Directory {
		t.Fatalf("unexpected first capture result: %+v", first)
	}
	second, err := a.EvidenceCapture(EvidenceCaptureOptions{
		Path: sourceDir, Title: "verifier investigation", Kind: "session", Completeness: CompletenessPartial,
		Scopes: []string{"verifier-repair"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if second.CreatedRecord || second.CreatedObject || second.RecordID != first.RecordID {
		t.Fatalf("same source and content should be idempotent: %+v", second)
	}
	rescoped, err := a.EvidenceCapture(EvidenceCaptureOptions{
		Path: sourceDir, Title: "verifier investigation", Kind: "session", Completeness: CompletenessPartial,
		Scopes: []string{"additional-scope"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !rescoped.CreatedRecord || rescoped.CreatedObject || rescoped.RecordID == first.RecordID {
		t.Fatalf("changed provenance should append a record while reusing content: %+v", rescoped)
	}

	if err := os.RemoveAll(sourceDir); err != nil {
		t.Fatal(err)
	}
	query, err := a.ContextQuery("masquerade", "", []string{"verifier-repair"}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(query.Hits) != 1 || query.Hits[0].Ref != first.RecordID || query.Hits[0].Locator != "notes.md" {
		t.Fatalf("unexpected context hits: %+v", query.Hits)
	}
	wrongScope, err := a.ContextQuery("masquerade", "", []string{"unrelated"}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(wrongScope.Hits) != 0 {
		t.Fatalf("scope filter returned unrelated context: %+v", wrongScope.Hits)
	}

	opened, err := a.ContextOpen(first.RecordID, "notes.md", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(opened.Content, "model failure") {
		t.Fatalf("opened content = %q", opened.Content)
	}

	out := filepath.Join(t.TempDir(), "context")
	materialized, err := a.ContextMaterialize([]string{first.RecordID}, out)
	if err != nil {
		t.Fatal(err)
	}
	if len(materialized.Entries) != 1 || materialized.Entries[0].ObjectDigest != first.ObjectDigest {
		t.Fatalf("unexpected materialization: %+v", materialized)
	}
	got, err := os.ReadFile(filepath.Join(out, materialized.Entries[0].Path, "notes.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "A verifier crash can masquerade as model failure.\n" {
		t.Fatalf("materialized content = %q", got)
	}
	if _, err := os.Stat(filepath.Join(out, "manifest.json")); err != nil {
		t.Fatalf("materialized context has no manifest: %v", err)
	}
	if _, err := a.ContextMaterialize([]string{first.RecordID}, out); err == nil {
		t.Fatal("materialize should refuse an existing destination")
	}
}

func TestExperienceIsEvidenceBackedAndRevisable(t *testing.T) {
	a, _ := newTestApp(t)
	evidencePath := filepath.Join(t.TempDir(), "lock.json")
	writeFixtureFile(t, evidencePath, `{"task":{"digest":"sha256:runtime"}}`)
	evidence, err := a.EvidenceCapture(EvidenceCaptureOptions{Path: evidencePath, Kind: "run-lock"})
	if err != nil {
		t.Fatal(err)
	}
	body := filepath.Join(t.TempDir(), "runtime-identity.md")
	writeFixtureFile(t, body, "Use the execution-time task identity from the run lock.\n")

	if _, err := a.ExperiencePropose(ExperienceOptions{BodyPath: body}); err == nil {
		t.Fatal("experience without evidence should be rejected")
	}
	if _, err := a.ExperiencePropose(ExperienceOptions{
		BodyPath: body, Evidence: []string{evidence.RecordID},
	}); err == nil {
		t.Fatal("experience without scope should be rejected")
	}
	if _, err := a.ExperiencePropose(ExperienceOptions{
		BodyPath: body, Evidence: []string{"missing-record"}, Scopes: []string{"harbor"},
	}); err == nil {
		t.Fatal("experience with a missing evidence reference should be rejected")
	}
	records, err := a.St.ListRecords("experience")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Fatalf("rejected proposals wrote %d records", len(records))
	}

	proposed, err := a.ExperiencePropose(ExperienceOptions{
		BodyPath: body, Title: "Trust runtime identity", Evidence: []string{evidence.RecordID},
		Scopes: []string{"harbor", "task-identity"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if proposed.Status != ExperienceProposed {
		t.Fatalf("proposal status = %q", proposed.Status)
	}
	assertRelationTypes(t, a, proposed.RecordID, "has_body", "supported_by")

	revised, err := a.ExperienceRevise(proposed.RecordID, ExperienceOptions{Status: ExperienceActive})
	if err != nil {
		t.Fatal(err)
	}
	if revised.RecordID == proposed.RecordID || revised.BodyDigest != proposed.BodyDigest || revised.Revises != proposed.RecordID {
		t.Fatalf("unexpected revision: %+v", revised)
	}
	assertRelationTypes(t, a, revised.RecordID, "has_body", "supported_by", "revises")

	old, err := a.St.GetRecord(proposed.RecordID)
	if err != nil {
		t.Fatal(err)
	}
	var oldPayload experiencePayload
	if err := json.Unmarshal([]byte(old.Payload), &oldPayload); err != nil {
		t.Fatal(err)
	}
	if oldPayload.Status != ExperienceProposed {
		t.Fatalf("revision mutated old status to %q", oldPayload.Status)
	}
	current, err := a.St.GetRecord(revised.RecordID)
	if err != nil {
		t.Fatal(err)
	}
	var currentPayload experiencePayload
	if err := json.Unmarshal([]byte(current.Payload), &currentPayload); err != nil {
		t.Fatal(err)
	}
	if currentPayload.Status != ExperienceActive || currentPayload.BodyName != "runtime-identity.md" {
		t.Fatalf("unexpected revised payload: %+v", currentPayload)
	}

	query, err := a.ContextQuery("execution-time task identity", "experience", []string{"harbor"}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(query.Hits) != 1 || query.Hits[0].Ref != revised.RecordID || query.Hits[0].Status != ExperienceActive {
		t.Fatalf("query should return only the active revision, got %+v", query.Hits)
	}

	retired, err := a.ExperienceRevise(revised.RecordID, ExperienceOptions{Status: ExperienceRetired})
	if err != nil {
		t.Fatal(err)
	}
	assertRelationTypes(t, a, retired.RecordID, "has_body", "supported_by", "revises")
	query, err = a.ContextQuery("execution-time task identity", "experience", []string{"harbor"}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(query.Hits) != 0 {
		t.Fatalf("retired experience should stay out of default context, got %+v", query.Hits)
	}
}

func assertRelationTypes(t *testing.T, a *App, recordID string, want ...string) {
	t.Helper()
	relations, err := a.St.RelationsFrom(store.EndpointRecord, recordID)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, relation := range relations {
		got[relation.Type] = true
	}
	for _, relationType := range want {
		if !got[relationType] {
			t.Errorf("record %s missing %s relation: %+v", recordID, relationType, relations)
		}
	}
}

func writeFixtureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
