package app

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/url"
	"os"
	"path/filepath"

	"github.com/WingchunSiu/vRI/internal/objects"
	"github.com/WingchunSiu/vRI/internal/store"
)

const (
	CompletenessComplete = "complete"
	CompletenessPartial  = "partial"
	CompletenessUnknown  = "unknown"
)

type EvidenceCaptureOptions struct {
	Path         string
	Title        string
	Kind         string
	SourceURI    string
	Completeness string
	Scopes       []string
	Actor        string
}

type EvidenceCaptureResult struct {
	RecordID      string `json:"record_id"`
	ObjectDigest  string `json:"object_digest"`
	CreatedRecord bool   `json:"created_record"`
	CreatedObject bool   `json:"created_object"`
	Directory     bool   `json:"directory"`
}

type sourcePayload struct {
	Title        string   `json:"title"`
	Name         string   `json:"name"`
	SourceURI    string   `json:"source_uri"`
	ObjectDigest string   `json:"object_digest"`
	Kind         string   `json:"kind"`
	MediaType    string   `json:"media_type,omitempty"`
	Completeness string   `json:"completeness"`
	Scopes       []string `json:"scopes,omitempty"`
}

// EvidenceCapture preserves a file or directory as an immutable object and
// records a minimal provenance envelope. It intentionally does not parse a
// provider-specific schema; agents can cite the source record and open the raw
// material when an interpretation needs checking.
func (a *App) EvidenceCapture(opts EvidenceCaptureOptions) (*EvidenceCaptureResult, error) {
	if opts.Path == "" {
		return nil, fmt.Errorf("evidence capture requires a path")
	}
	if opts.Kind == "" {
		opts.Kind = "artifact"
	}
	if opts.Completeness == "" {
		opts.Completeness = CompletenessComplete
	}
	if !validCompleteness(opts.Completeness) {
		return nil, fmt.Errorf("unsupported completeness %q (want complete, partial, or unknown)", opts.Completeness)
	}
	if opts.Actor == "" {
		opts.Actor = "agent"
	}

	abs, err := filepath.Abs(opts.Path)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(abs); err != nil {
		return nil, err
	}
	if opts.Title == "" {
		opts.Title = filepath.Base(abs)
	}
	if opts.SourceURI == "" {
		opts.SourceURI = (&url.URL{Scheme: "file", Path: abs}).String()
	}

	digest, directory, createdObject, err := a.storePath(abs, store.KindEvidence)
	if err != nil {
		return nil, err
	}
	if existing, err := a.findCapturedSource(opts.SourceURI, digest, opts.Kind, opts.Completeness, opts.Scopes); err != nil {
		return nil, err
	} else if existing != nil {
		res := &EvidenceCaptureResult{RecordID: existing.ID, ObjectDigest: digest, Directory: directory}
		return res, a.emit(res, func() {
			fmt.Fprintf(a.Out, "already captured: %s (source %s)\n", digest, existing.ID)
		})
	}

	payload := sourcePayload{
		Title:        opts.Title,
		Name:         filepath.Base(abs),
		SourceURI:    opts.SourceURI,
		ObjectDigest: digest,
		Kind:         opts.Kind,
		Completeness: opts.Completeness,
		Scopes:       append([]string(nil), opts.Scopes...),
	}
	if directory {
		payload.MediaType = "application/vnd.vri.directory"
	} else {
		payload.MediaType = mime.TypeByExtension(filepath.Ext(abs))
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	rec, err := a.putRecord("source", raw, opts.Actor)
	if err != nil {
		return nil, err
	}
	if _, err := a.Link(store.EndpointRecord, rec.ID, store.EndpointObject, digest, "captures", ""); err != nil {
		return nil, err
	}

	res := &EvidenceCaptureResult{
		RecordID: rec.ID, ObjectDigest: digest, CreatedRecord: true,
		CreatedObject: createdObject, Directory: directory,
	}
	return res, a.emit(res, func() {
		fmt.Fprintf(a.Out, "captured %s\n  source %s\n  object %s\n", opts.Title, rec.ID, digest)
	})
}

func validCompleteness(value string) bool {
	return value == CompletenessComplete || value == CompletenessPartial || value == CompletenessUnknown
}

func (a *App) storePath(path, kind string) (digest string, directory, created bool, err error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", false, false, err
	}
	if info.IsDir() {
		digest, rel, created, err := objects.WriteDir(a.St.ObjectsDir(), path)
		if err != nil {
			return "", false, false, err
		}
		_, entries, err := objects.DigestDir(path)
		if err != nil {
			return "", false, false, err
		}
		var size int64
		for _, entry := range entries {
			size += entry.Size
		}
		if err := a.St.PutObject(&store.Object{Digest: digest, Kind: kind, Path: rel, SizeBytes: size, CreatedAt: now()}); err != nil {
			return "", false, false, err
		}
		return digest, true, created, nil
	}

	digest, rel, size, created, err := objects.Write(a.St.ObjectsDir(), path)
	if err != nil {
		return "", false, false, err
	}
	if err := a.St.PutObject(&store.Object{Digest: digest, Kind: kind, Path: rel, SizeBytes: size, CreatedAt: now()}); err != nil {
		return "", false, false, err
	}
	return digest, false, created, nil
}

func (a *App) findCapturedSource(sourceURI, digest, kind, completeness string, scopes []string) (*store.Record, error) {
	recs, err := a.St.ListRecords("source")
	if err != nil {
		return nil, err
	}
	for _, rec := range recs {
		var payload sourcePayload
		if err := json.Unmarshal([]byte(rec.Payload), &payload); err == nil &&
			payload.SourceURI == sourceURI && payload.ObjectDigest == digest &&
			payload.Kind == kind && payload.Completeness == completeness &&
			sameStringSet(payload.Scopes, scopes) {
			return rec, nil
		}
	}
	return nil, nil
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	counts := map[string]int{}
	for _, value := range left {
		counts[value]++
	}
	for _, value := range right {
		counts[value]--
		if counts[value] < 0 {
			return false
		}
	}
	return true
}
