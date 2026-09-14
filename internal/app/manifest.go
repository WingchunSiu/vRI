package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Manifest is the minimal benchmark-manifest shape needed by the v0 commands.
// Unknown fields are preserved only in the stored record payload, not here.
type Manifest struct {
	ID    string `json:"id"`
	Tasks []struct {
		ID     string `json:"id"`
		Path   string `json:"path"`
		Digest string `json:"digest"`
	} `json:"tasks"`
	Runs []struct {
		JobPath   string `json:"job_path"`
		JobDigest string `json:"job_digest"`
		System    struct {
			Agent string `json:"agent"`
			Model string `json:"model"`
		} `json:"system"`
		Reward float64 `json:"reward"`
		Role   string  `json:"role"`
	} `json:"runs"`
	Limitations []string `json:"limitations"`
}

// resolveManifest loads a manifest from a file path, a record ULID, or the
// manifest's own id field. It returns the parsed
// manifest, its raw JSON, and the source it came from (file path or record
// id; file paths are used to resolve relative job_path entries).
func (a *App) resolveManifest(ref string) (m *Manifest, raw []byte, source string, err error) {
	if _, statErr := os.Stat(ref); statErr == nil {
		m, raw, err := readManifestFile(ref)
		if err != nil {
			return nil, nil, "", err
		}
		abs, err := filepath.Abs(ref)
		if err != nil {
			return nil, nil, "", err
		}
		return m, raw, abs, nil
	}
	if rec, recErr := a.St.GetRecord(ref); recErr == nil && rec.Type == "benchmark_manifest" {
		return parseManifestPayload(rec.Payload, rec.ID)
	}
	rec, findErr := a.findRecordByPayloadField("benchmark_manifest", "id", ref)
	if findErr != nil {
		return nil, nil, "", findErr
	}
	if rec == nil {
		return nil, nil, "", fmt.Errorf("manifest %q not found as file, record id, or manifest id", ref)
	}
	return parseManifestPayload(rec.Payload, rec.ID)
}

func parseManifestPayload(payload, source string) (*Manifest, []byte, string, error) {
	var m Manifest
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		return nil, nil, "", err
	}
	return &m, []byte(payload), source, nil
}

func readManifestFile(path string) (*Manifest, []byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var probe struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(b, &probe); err != nil {
		return nil, nil, fmt.Errorf("parsing manifest %s: %w", path, err)
	}
	if probe.Kind != "" && probe.Kind != "benchmark_manifest" {
		return nil, nil, fmt.Errorf("%s: kind is %q, want benchmark_manifest", path, probe.Kind)
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, nil, err
	}
	return &m, b, nil
}
