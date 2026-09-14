package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/WingchunSiu/vRI/internal/objects"
	"github.com/WingchunSiu/vRI/internal/store"
)

const maxContextSearchBytes = 2 << 20

type ContextQueryResult struct {
	Query string       `json:"query"`
	Hits  []ContextHit `json:"hits"`
}

type ContextHit struct {
	Ref          string   `json:"ref"`
	Type         string   `json:"type"`
	Status       string   `json:"status,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	Title        string   `json:"title"`
	ObjectDigest string   `json:"object_digest,omitempty"`
	Locator      string   `json:"locator"`
	Snippet      string   `json:"snippet"`
}

// ContextQuery searches record metadata and textual content captured by
// source and experience records. The first vertical deliberately scans the
// small local store; a rebuildable FTS index can replace this projection when
// measured corpus size or latency requires it.
func (a *App) ContextQuery(query, recordType string, scopes []string, limit int) (*ContextQueryResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("context query requires non-empty text")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		return nil, fmt.Errorf("context query limit must not exceed 200")
	}

	records, err := a.St.ListRecords(recordType)
	if err != nil {
		return nil, err
	}
	result := &ContextQueryResult{Query: query}
	for _, rec := range records {
		visible, err := a.contextRecordVisible(rec)
		if err != nil {
			return nil, err
		}
		if !visible {
			continue
		}
		if !recordMatchesScopes(rec, scopes) {
			continue
		}
		digest, title := recordContextObject(rec)
		matchedObject := false
		if digest != "" {
			hits, err := a.searchContextObject(rec, digest, title, query, limit-len(result.Hits))
			if err != nil {
				return nil, err
			}
			if len(hits) > 0 {
				matchedObject = true
				result.Hits = append(result.Hits, hits...)
			}
		}
		if len(result.Hits) >= limit {
			break
		}
		if !matchedObject && containsFold(rec.Type+"\n"+rec.Payload, query) {
			result.Hits = append(result.Hits, ContextHit{
				Ref: rec.ID, Type: rec.Type, Status: recordContextStatus(rec), Scopes: recordContextScopes(rec), Title: title, ObjectDigest: digest,
				Locator: "record", Snippet: snippet(rec.Payload, query),
			})
		}
		if len(result.Hits) >= limit {
			break
		}
	}

	return result, a.emit(result, func() {
		if len(result.Hits) == 0 {
			fmt.Fprintln(a.Out, "no matching context")
			return
		}
		for _, hit := range result.Hits {
			typeAndStatus := hit.Type
			if hit.Status != "" {
				typeAndStatus += "/" + hit.Status
			}
			fmt.Fprintf(a.Out, "%s  %-18s %s  [%s]\n  %s\n", hit.Ref, typeAndStatus, hit.Title, hit.Locator, hit.Snippet)
		}
	})
}

// contextRecordVisible keeps superseded versions and deliberately retired
// experience out of the default working set. They remain directly addressable
// for audit and can be listed through the generic record command.
func (a *App) contextRecordVisible(rec *store.Record) (bool, error) {
	if rec.Type != "experience" {
		return true, nil
	}
	revisions, err := a.St.RelationsTo(store.EndpointRecord, rec.ID, "revises")
	if err != nil {
		return false, err
	}
	if len(revisions) > 0 {
		return false, nil
	}
	status := recordContextStatus(rec)
	return status != ExperienceRetired && status != ExperienceSuperseded, nil
}

func (a *App) searchContextObject(rec *store.Record, digest, title, query string, limit int) ([]ContextHit, error) {
	if limit <= 0 {
		return nil, nil
	}
	obj, err := a.St.GetObject(digest)
	if err != nil {
		return nil, fmt.Errorf("record %s references missing object %s", rec.ID, digest)
	}
	root, err := objects.Read(a.St.ObjectsDir(), obj.Path)
	if err != nil {
		return nil, err
	}
	if !objects.IsDirManifest(root) {
		if searchableText(root) && containsFold(string(root), query) {
			return []ContextHit{{
				Ref: rec.ID, Type: rec.Type, Status: recordContextStatus(rec), Scopes: recordContextScopes(rec), Title: title, ObjectDigest: digest,
				Locator: "content", Snippet: snippet(string(root), query),
			}}, nil
		}
		return nil, nil
	}

	entries, err := objects.DirManifestEntries(root)
	if err != nil {
		return nil, err
	}
	var hits []ContextHit
	for _, entry := range entries {
		if entry.Size > maxContextSearchBytes {
			continue
		}
		rel, err := objects.BlobPath(entry.Digest)
		if err != nil {
			return nil, err
		}
		content, err := objects.Read(a.St.ObjectsDir(), rel)
		if err != nil {
			return nil, fmt.Errorf("reading %s from object %s: %w", entry.Path, digest, err)
		}
		if !searchableText(content) || !containsFold(string(content), query) {
			continue
		}
		hits = append(hits, ContextHit{
			Ref: rec.ID, Type: rec.Type, Status: recordContextStatus(rec), Scopes: recordContextScopes(rec), Title: title, ObjectDigest: digest,
			Locator: entry.Path, Snippet: snippet(string(content), query),
		})
		if len(hits) >= limit {
			break
		}
	}
	return hits, nil
}

type ContextOpenResult struct {
	Ref            string          `json:"ref"`
	RecordID       string          `json:"record_id,omitempty"`
	RecordType     string          `json:"record_type,omitempty"`
	Payload        json.RawMessage `json:"payload,omitempty"`
	ObjectDigest   string          `json:"object_digest,omitempty"`
	Locator        string          `json:"locator,omitempty"`
	Content        string          `json:"content,omitempty"`
	Entries        []string        `json:"entries,omitempty"`
	MaterializedTo string          `json:"materialized_to,omitempty"`
}

// ContextOpen resolves either a record ID or object digest. Directory objects
// can be listed, opened at one relative path, or extracted with outPath.
func (a *App) ContextOpen(ref, locator, outPath string) (*ContextOpenResult, error) {
	rec, obj, err := a.resolveContextRef(ref)
	if err != nil {
		return nil, err
	}
	result := &ContextOpenResult{Ref: ref, Locator: locator}
	if rec != nil {
		result.RecordID = rec.ID
		result.RecordType = rec.Type
		result.Payload = json.RawMessage(rec.Payload)
	}
	if obj == nil {
		return result, a.emit(result, func() { a.printContextOpen(result) })
	}
	result.ObjectDigest = obj.Digest
	root, err := objects.Read(a.St.ObjectsDir(), obj.Path)
	if err != nil {
		return nil, err
	}

	content := root
	if objects.IsDirManifest(root) {
		entries, err := objects.DirManifestEntries(root)
		if err != nil {
			return nil, err
		}
		if locator == "" {
			if outPath != "" {
				if err := requireAbsent(outPath); err != nil {
					return nil, err
				}
				if err := objects.ExtractDir(a.St.ObjectsDir(), root, outPath); err != nil {
					return nil, err
				}
				abs, _ := filepath.Abs(outPath)
				result.MaterializedTo = abs
			} else {
				for _, entry := range entries {
					result.Entries = append(result.Entries, entry.Path)
				}
			}
			return result, a.emit(result, func() { a.printContextOpen(result) })
		}
		content, err = a.readDirectoryEntry(entries, locator)
		if err != nil {
			return nil, err
		}
	} else if locator != "" {
		return nil, fmt.Errorf("object %s is a file; --path is only valid for directory objects", obj.Digest)
	}

	if outPath != "" {
		if err := writeContextFile(outPath, content); err != nil {
			return nil, err
		}
		abs, _ := filepath.Abs(outPath)
		result.MaterializedTo = abs
	} else if searchableText(content) {
		result.Content = string(content)
	} else {
		return nil, fmt.Errorf("context %s is binary; pass --out to write it", ref)
	}
	return result, a.emit(result, func() { a.printContextOpen(result) })
}

func (a *App) readDirectoryEntry(entries []objects.DirEntry, locator string) ([]byte, error) {
	clean := filepath.ToSlash(filepath.Clean(locator))
	if clean == "." || strings.HasPrefix(clean, "../") || filepath.IsAbs(locator) {
		return nil, fmt.Errorf("invalid context path %q", locator)
	}
	for _, entry := range entries {
		if entry.Path != clean {
			continue
		}
		rel, err := objects.BlobPath(entry.Digest)
		if err != nil {
			return nil, err
		}
		return objects.Read(a.St.ObjectsDir(), rel)
	}
	return nil, fmt.Errorf("context path %q not found", locator)
}

func (a *App) printContextOpen(result *ContextOpenResult) {
	if result.RecordID != "" {
		fmt.Fprintf(a.Out, "record %s (%s)\n", result.RecordID, result.RecordType)
		var pretty json.RawMessage = result.Payload
		if b, err := json.MarshalIndent(result.Payload, "", "  "); err == nil {
			pretty = b
		}
		fmt.Fprintln(a.Out, string(pretty))
	}
	if result.ObjectDigest != "" {
		fmt.Fprintf(a.Out, "object %s\n", result.ObjectDigest)
	}
	if result.MaterializedTo != "" {
		fmt.Fprintf(a.Out, "materialized to %s\n", result.MaterializedTo)
		return
	}
	for _, entry := range result.Entries {
		fmt.Fprintln(a.Out, entry)
	}
	if result.Content != "" {
		fmt.Fprint(a.Out, result.Content)
		if !strings.HasSuffix(result.Content, "\n") {
			fmt.Fprintln(a.Out)
		}
	}
}

type MaterializedContext struct {
	Path    string                     `json:"path"`
	Entries []MaterializedContextEntry `json:"entries"`
}

type MaterializedContextEntry struct {
	Ref          string `json:"ref"`
	RecordID     string `json:"record_id,omitempty"`
	ObjectDigest string `json:"object_digest,omitempty"`
	Path         string `json:"path"`
}

// ContextMaterialize creates a disposable, self-describing working set. It
// refuses an existing destination so it cannot silently mix stale context
// with the selected references.
func (a *App) ContextMaterialize(refs []string, outPath string) (*MaterializedContext, error) {
	if len(refs) == 0 {
		return nil, fmt.Errorf("context materialize requires at least one reference")
	}
	if outPath == "" {
		return nil, fmt.Errorf("context materialize requires --out")
	}
	abs, err := filepath.Abs(outPath)
	if err != nil {
		return nil, err
	}
	if err := requireAbsent(abs); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return nil, err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(abs), ".vri-context-")
	if err != nil {
		return nil, err
	}
	keep := false
	defer func() {
		if !keep {
			_ = os.RemoveAll(tmp)
		}
	}()

	result := &MaterializedContext{Path: abs}
	for i, ref := range refs {
		rec, obj, err := a.resolveContextRef(ref)
		if err != nil {
			return nil, err
		}
		label := fmt.Sprintf("%03d-%s", i+1, contextLabel(rec, ref))
		dst := filepath.Join(tmp, label)
		entry := MaterializedContextEntry{Ref: ref, Path: label}
		if rec != nil {
			entry.RecordID = rec.ID
		}
		if obj == nil {
			if err := os.MkdirAll(dst, 0o755); err != nil {
				return nil, err
			}
			pretty, _ := json.MarshalIndent(json.RawMessage(rec.Payload), "", "  ")
			if err := os.WriteFile(filepath.Join(dst, "record.json"), append(pretty, '\n'), 0o644); err != nil {
				return nil, err
			}
			result.Entries = append(result.Entries, entry)
			continue
		}

		entry.ObjectDigest = obj.Digest
		root, err := objects.Read(a.St.ObjectsDir(), obj.Path)
		if err != nil {
			return nil, err
		}
		if objects.IsDirManifest(root) {
			if err := objects.ExtractDir(a.St.ObjectsDir(), root, dst); err != nil {
				return nil, err
			}
		} else {
			if err := os.MkdirAll(dst, 0o755); err != nil {
				return nil, err
			}
			name := contextFilename(rec)
			if err := os.WriteFile(filepath.Join(dst, name), root, 0o644); err != nil {
				return nil, err
			}
		}
		result.Entries = append(result.Entries, entry)
	}

	manifest, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(tmp, "manifest.json"), append(manifest, '\n'), 0o644); err != nil {
		return nil, err
	}
	if err := os.Rename(tmp, abs); err != nil {
		return nil, err
	}
	keep = true
	return result, a.emit(result, func() {
		fmt.Fprintf(a.Out, "materialized %d context reference(s) to %s\n", len(result.Entries), abs)
	})
}

func (a *App) resolveContextRef(ref string) (*store.Record, *store.Object, error) {
	if strings.HasPrefix(ref, "sha256:") {
		obj, err := a.St.GetObject(ref)
		if err != nil {
			return nil, nil, fmt.Errorf("object %s not found", ref)
		}
		return nil, obj, nil
	}
	rec, err := a.St.GetRecord(ref)
	if err != nil {
		return nil, nil, fmt.Errorf("record %s not found", ref)
	}
	digest, _ := recordContextObject(rec)
	if digest == "" {
		return rec, nil, nil
	}
	obj, err := a.St.GetObject(digest)
	if err != nil {
		return nil, nil, fmt.Errorf("record %s references missing object %s", rec.ID, digest)
	}
	return rec, obj, nil
}

func recordContextObject(rec *store.Record) (digest, title string) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(rec.Payload), &payload); err != nil {
		return "", rec.Type
	}
	for _, field := range []string{"body_digest", "object_digest", "excerpt"} {
		if value, ok := payload[field].(string); ok && value != "" {
			digest = value
			break
		}
	}
	if value, ok := payload["title"].(string); ok && value != "" {
		title = value
	} else {
		title = summarize(rec.Payload)
	}
	return digest, title
}

func recordContextStatus(rec *store.Record) string {
	var payload map[string]any
	if err := json.Unmarshal([]byte(rec.Payload), &payload); err != nil {
		return ""
	}
	status, _ := payload["status"].(string)
	return status
}

func recordContextScopes(rec *store.Record) []string {
	var payload struct {
		Scope  string   `json:"scope"`
		Scopes []string `json:"scopes"`
	}
	if err := json.Unmarshal([]byte(rec.Payload), &payload); err != nil {
		return nil
	}
	if payload.Scope != "" {
		payload.Scopes = append(payload.Scopes, payload.Scope)
	}
	return payload.Scopes
}

func recordMatchesScopes(rec *store.Record, requested []string) bool {
	if len(requested) == 0 {
		return true
	}
	have := map[string]bool{}
	for _, scope := range recordContextScopes(rec) {
		have[scope] = true
	}
	for _, scope := range requested {
		if !have[scope] {
			return false
		}
	}
	return true
}

func searchableText(content []byte) bool {
	return len(content) <= maxContextSearchBytes && utf8.Valid(content) && !bytes.ContainsRune(content, '\x00')
}

func containsFold(value, query string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(query))
}

func snippet(value, query string) string {
	for _, line := range strings.Split(value, "\n") {
		if containsFold(line, query) {
			return truncateText(strings.Join(strings.Fields(line), " "), 220)
		}
	}
	return truncateText(strings.Join(strings.Fields(value), " "), 220)
}

func truncateText(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "..."
}

func requireAbsent(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("destination already exists: %s", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeContextFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func contextLabel(rec *store.Record, ref string) string {
	if rec != nil {
		return safeContextName(rec.Type + "-" + rec.ID[:min(8, len(rec.ID))])
	}
	trimmed := strings.TrimPrefix(ref, "sha256:")
	return "object-" + trimmed[:min(8, len(trimmed))]
}

func contextFilename(rec *store.Record) string {
	if rec != nil {
		var payload map[string]any
		if json.Unmarshal([]byte(rec.Payload), &payload) == nil {
			for _, field := range []string{"body_name", "name", "title"} {
				if name, ok := payload[field].(string); ok && name != "" {
					return safeContextName(filepath.Base(name))
				}
			}
		}
	}
	return "content"
}

func safeContextName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "." || value == ".." {
		return "context"
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	name := strings.Trim(b.String(), "-")
	if name == "" {
		return "context"
	}
	return name
}
