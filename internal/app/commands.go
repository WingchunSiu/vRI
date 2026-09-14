package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/WingchunSiu/vRI/internal/objects"
	"github.com/WingchunSiu/vRI/internal/store"
)

// ---- vri import ----

type ImportResult struct {
	Digest        string `json:"digest"`
	CreatedObject bool   `json:"created_object"`
	RecordID      string `json:"record_id"`
	CreatedRecord bool   `json:"created_record"`
}

// Import copies a source file into the object store and creates a source
// record. It is idempotent by digest: re-importing the same content creates
// neither a new object nor a new record.
func (a *App) Import(path, kind, actor string) (*ImportResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("import takes a file; use `vri object put` for directories")
	}
	digest, rel, size, created, err := objects.Write(a.St.ObjectsDir(), path)
	if err != nil {
		return nil, err
	}
	if err := a.St.PutObject(&store.Object{Digest: digest, Kind: kind, Path: rel, SizeBytes: size, CreatedAt: now()}); err != nil {
		return nil, err
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	res := &ImportResult{Digest: digest, CreatedObject: created}
	existing, err := a.findRecordByPayloadField("source", "excerpt", digest)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		res.RecordID = existing.ID
	} else {
		payload, _ := json.Marshal(map[string]string{
			"origin":      abs,
			"excerpt":     digest,
			"redaction":   "",
			"imported_as": kind,
		})
		rec, err := a.putRecord("source", payload, actor)
		if err != nil {
			return nil, err
		}
		res.RecordID = rec.ID
		res.CreatedRecord = true
	}

	return res, a.emit(res, func() {
		if !res.CreatedObject && !res.CreatedRecord {
			fmt.Fprintf(a.Out, "already imported: %s (record %s)\n", digest, res.RecordID)
			return
		}
		fmt.Fprintf(a.Out, "imported %s as %s\nsource record %s\n", path, digest, res.RecordID)
	})
}

// ---- vri object ----

type ObjectPutResult struct {
	Digest  string `json:"digest"`
	Kind    string `json:"kind"`
	Created bool   `json:"created"`
}

// ObjectPut stores a file or directory content-addressed. Directories get a
// manifest blob addressed by the directory digest; kind defaults to
// task_package for directories and file for files.
func (a *App) ObjectPut(path, kind string) (*ObjectPutResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	res := &ObjectPutResult{Kind: kind}
	if info.IsDir() {
		digest, rel, created, err := objects.WriteDir(a.St.ObjectsDir(), path)
		if err != nil {
			return nil, err
		}
		res.Digest, res.Created = digest, created
		if res.Kind == "" {
			res.Kind = store.KindTaskPackage
		}
		_, entries, err := objects.DigestDir(path)
		if err != nil {
			return nil, err
		}
		var total int64
		for _, e := range entries {
			total += e.Size
		}
		if err := a.St.PutObject(&store.Object{Digest: digest, Kind: res.Kind, Path: rel, SizeBytes: total, CreatedAt: now()}); err != nil {
			return nil, err
		}
	} else {
		digest, rel, size, created, err := objects.Write(a.St.ObjectsDir(), path)
		if err != nil {
			return nil, err
		}
		res.Digest, res.Created = digest, created
		if res.Kind == "" {
			res.Kind = store.KindFile
		}
		if err := a.St.PutObject(&store.Object{Digest: digest, Kind: res.Kind, Path: rel, SizeBytes: size, CreatedAt: now()}); err != nil {
			return nil, err
		}
	}
	return res, a.emit(res, func() {
		verb := "stored"
		if !res.Created {
			verb = "already present"
		}
		fmt.Fprintf(a.Out, "%s %s (%s)\n", res.Digest, verb, res.Kind)
	})
}

// ObjectGet writes an object to outPath, or to stdout for plain files when
// outPath is empty. Directory manifests require --out.
func (a *App) ObjectGet(digest, outPath string) error {
	obj, err := a.St.GetObject(digest)
	if err != nil {
		return fmt.Errorf("object %s not found", digest)
	}
	content, err := objects.Read(a.St.ObjectsDir(), obj.Path)
	if err != nil {
		return err
	}
	if objects.IsDirManifest(content) {
		if outPath == "" {
			return fmt.Errorf("object %s is a directory; pass --out <dir>", digest)
		}
		if err := objects.ExtractDir(a.St.ObjectsDir(), content, outPath); err != nil {
			return err
		}
		fmt.Fprintf(a.Out, "extracted %s to %s\n", digest, outPath)
		return nil
	}
	if outPath == "" {
		_, err := a.Out.Write(content)
		return err
	}
	if err := os.WriteFile(outPath, content, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(a.Out, "wrote %s to %s\n", digest, outPath)
	return nil
}

func (a *App) ObjectList() ([]*store.Object, error) {
	objs, err := a.St.ListObjects()
	if err != nil {
		return nil, err
	}
	return objs, a.emit(objs, func() {
		for _, o := range objs {
			fmt.Fprintf(a.Out, "%s  %-14s %8d  %s\n", o.Digest, o.Kind, o.SizeBytes, o.CreatedAt)
		}
	})
}

// ---- vri record ----

// RecordPut creates a typed record from JSON supplied inline or via file.
func (a *App) RecordPut(typ string, payload []byte, actor string) (*store.Record, error) {
	if typ == "" {
		return nil, fmt.Errorf("record put requires --type")
	}
	rec, err := a.putRecord(typ, json.RawMessage(payload), actor)
	if err != nil {
		return nil, err
	}
	return rec, a.emit(rec, func() {
		fmt.Fprintf(a.Out, "record %s (%s)\n", rec.ID, rec.Type)
	})
}

func (a *App) RecordGet(id string) (*store.Record, error) {
	rec, err := a.St.GetRecord(id)
	if err != nil {
		return nil, fmt.Errorf("record %s not found", id)
	}
	return rec, a.emit(rec, func() {
		fmt.Fprintf(a.Out, "%s %s (%s, %s)\n", rec.ID, rec.Type, rec.Actor, rec.CreatedAt)
		var pretty json.RawMessage = json.RawMessage(rec.Payload)
		if b, err := json.MarshalIndent(json.RawMessage(rec.Payload), "", "  "); err == nil {
			pretty = b
		}
		fmt.Fprintln(a.Out, string(pretty))
	})
}

func (a *App) RecordList(typ string) ([]*store.Record, error) {
	recs, err := a.St.ListRecords(typ)
	if err != nil {
		return nil, err
	}
	return recs, a.emit(recs, func() {
		for _, r := range recs {
			fmt.Fprintf(a.Out, "%s  %-18s %-6s %s\n", r.ID, r.Type, r.Actor, summarize(r.Payload))
		}
	})
}

// summarize renders a one-line preview of a record payload for listings.
func summarize(payload string) string {
	var p map[string]any
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return payload
	}
	for _, k := range []string{"task", "title", "name", "text", "statement", "job_name", "id", "origin"} {
		if v, ok := p[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	if len(payload) > 72 {
		return payload[:72] + "..."
	}
	return payload
}
