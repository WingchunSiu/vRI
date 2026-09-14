// Package app implements the vri CLI commands against an opened store.
// Handlers return structured results and write human-readable (or --json)
// output to App.Out; reference integrity is enforced here, at write time.
package app

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/WingchunSiu/vRI/internal/store"
)

type App struct {
	St   *store.Store
	Out  io.Writer
	JSON bool
}

func Open(projectDir string, out io.Writer, jsonOut bool) (*App, error) {
	st, err := store.Open(projectDir)
	if err != nil {
		return nil, err
	}
	return &App{St: st, Out: out, JSON: jsonOut}, nil
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func newID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), ulid.Monotonic(rand.Reader, 0)).String()
}

// emit writes v as indented JSON in --json mode, otherwise calls text.
func (a *App) emit(v any, text func()) error {
	if a.JSON {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(a.Out, string(b))
		return err
	}
	text()
	return nil
}

// validateEndpoint implements the v0 reference-integrity rule: every
// relation endpoint must resolve to an existing row at write time.
func (a *App) validateEndpoint(kind, id string) error {
	switch kind {
	case store.EndpointObject:
		ok, err := a.St.HasObject(id)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("relation endpoint object %s does not exist in the store", id)
		}
	case store.EndpointRecord:
		if _, err := a.St.GetRecord(id); err != nil {
			return fmt.Errorf("relation endpoint record %s does not exist in the store", id)
		}
	default:
		return fmt.Errorf("unknown endpoint kind %q (want object|record)", kind)
	}
	return nil
}

// Link adds a relation after validating both endpoints, skipping duplicates.
func (a *App) Link(srcKind, srcID, dstKind, dstID, typ, meta string) (bool, error) {
	if err := a.validateEndpoint(srcKind, srcID); err != nil {
		return false, err
	}
	if err := a.validateEndpoint(dstKind, dstID); err != nil {
		return false, err
	}
	exists, err := a.St.RelationExists(srcKind, srcID, dstKind, dstID, typ)
	if err != nil || exists {
		return false, err
	}
	return true, a.St.AddRelation(&store.Relation{
		SrcKind: srcKind, SrcID: srcID, DstKind: dstKind, DstID: dstID,
		Type: typ, Meta: meta, CreatedAt: now(),
	})
}

// putRecord validates and inserts a record.
func (a *App) putRecord(typ string, payload json.RawMessage, actor string) (*store.Record, error) {
	if !json.Valid(payload) {
		return nil, fmt.Errorf("payload is not valid JSON")
	}
	if err := a.validatePayload(typ, payload); err != nil {
		return nil, err
	}
	compact, err := compactJSON(payload)
	if err != nil {
		return nil, err
	}
	r := &store.Record{ID: newID(), Type: typ, Payload: compact, Actor: actor, CreatedAt: now()}
	if err := a.St.PutRecord(r); err != nil {
		return nil, err
	}
	return r, nil
}

// validatePayload checks embedded references for record types whose payloads
// name objects. A benchmark_manifest must only reference task digests that
// exist in the store.
func (a *App) validatePayload(typ string, payload json.RawMessage) error {
	if typ != "benchmark_manifest" {
		return nil
	}
	var m struct {
		Tasks []struct {
			ID     string `json:"id"`
			Digest string `json:"digest"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal(payload, &m); err != nil {
		return fmt.Errorf("benchmark_manifest payload: %w", err)
	}
	for _, t := range m.Tasks {
		ok, err := a.St.HasObject(t.Digest)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("benchmark_manifest task %q references missing object %s", t.ID, t.Digest)
		}
	}
	return nil
}

func compactJSON(b json.RawMessage) (string, error) {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return "", err
	}
	out, err := json.Marshal(v)
	return string(out), err
}

// findRecordByPayloadField scans records of typ for one whose payload has
// field == value. Used for idempotency keys (harbor_job.source_path,
// benchmark_manifest.id) since payload schemas are app-level.
func (a *App) findRecordByPayloadField(typ, field, value string) (*store.Record, error) {
	recs, err := a.St.ListRecords(typ)
	if err != nil {
		return nil, err
	}
	for _, r := range recs {
		var p map[string]any
		if err := json.Unmarshal([]byte(r.Payload), &p); err != nil {
			continue
		}
		if p[field] == value {
			return r, nil
		}
	}
	return nil, nil
}
