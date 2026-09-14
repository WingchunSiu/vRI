package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/WingchunSiu/vRI/internal/store"
)

const (
	ExperienceProposed   = "proposed"
	ExperienceActive     = "active"
	ExperienceContested  = "contested"
	ExperienceSuperseded = "superseded"
	ExperienceRetired    = "retired"
)

type ExperienceOptions struct {
	BodyPath string
	BodyName string
	Title    string
	Status   string
	Scopes   []string
	Evidence []string
	Actor    string
}

type ExperienceResult struct {
	RecordID    string   `json:"record_id"`
	BodyDigest  string   `json:"body_digest"`
	Status      string   `json:"status"`
	Evidence    []string `json:"evidence"`
	Revises     string   `json:"revises,omitempty"`
	CreatedBody bool     `json:"created_body"`
}

type experiencePayload struct {
	Title      string   `json:"title"`
	BodyDigest string   `json:"body_digest"`
	BodyName   string   `json:"body_name"`
	Status     string   `json:"status"`
	Scopes     []string `json:"scopes,omitempty"`
	Evidence   []string `json:"evidence"`
	Revises    string   `json:"revises,omitempty"`
}

// ExperiencePropose creates a revisable interpretation backed by at least one
// existing evidence record or object. The prose body is stored separately so
// later revisions never rewrite the evidence or an earlier interpretation.
func (a *App) ExperiencePropose(opts ExperienceOptions) (*ExperienceResult, error) {
	if opts.BodyPath == "" {
		return nil, fmt.Errorf("experience propose requires a body file")
	}
	if len(opts.Evidence) == 0 {
		return nil, fmt.Errorf("experience propose requires at least one evidence reference")
	}
	if len(opts.Scopes) == 0 {
		return nil, fmt.Errorf("experience propose requires at least one scope")
	}
	if opts.Status == "" {
		opts.Status = ExperienceProposed
	}
	return a.writeExperience(opts, "")
}

// ExperienceRevise appends a new version and links it to the previous record.
// Omitted body, title, scope, and evidence values are inherited. The previous
// record remains unchanged.
func (a *App) ExperienceRevise(previousID string, opts ExperienceOptions) (*ExperienceResult, error) {
	previous, err := a.St.GetRecord(previousID)
	if err != nil || previous.Type != "experience" {
		return nil, fmt.Errorf("experience %s not found", previousID)
	}
	var old experiencePayload
	if err := json.Unmarshal([]byte(previous.Payload), &old); err != nil {
		return nil, fmt.Errorf("experience %s has invalid payload: %w", previousID, err)
	}
	if opts.BodyPath == "" {
		opts.BodyPath = "@" + old.BodyDigest
		opts.BodyName = old.BodyName
	}
	if opts.Title == "" {
		opts.Title = old.Title
	}
	if opts.Status == "" {
		opts.Status = old.Status
	}
	if len(opts.Scopes) == 0 {
		opts.Scopes = append([]string(nil), old.Scopes...)
	}
	if len(opts.Evidence) == 0 {
		opts.Evidence = append([]string(nil), old.Evidence...)
	}
	return a.writeExperience(opts, previousID)
}

func (a *App) writeExperience(opts ExperienceOptions, revises string) (*ExperienceResult, error) {
	if !validExperienceStatus(opts.Status) {
		return nil, fmt.Errorf("unsupported experience status %q", opts.Status)
	}
	if opts.Actor == "" {
		opts.Actor = "agent"
	}

	// Validate all references before writing anything, so a typo cannot leave a
	// partially linked experience record behind.
	type endpoint struct{ kind, id string }
	endpoints := make([]endpoint, 0, len(opts.Evidence))
	for _, ref := range opts.Evidence {
		kind := store.EndpointRecord
		if strings.HasPrefix(ref, "sha256:") {
			kind = store.EndpointObject
		}
		if err := a.validateEndpoint(kind, ref); err != nil {
			return nil, fmt.Errorf("evidence %s: %w", ref, err)
		}
		endpoints = append(endpoints, endpoint{kind: kind, id: ref})
	}

	var bodyDigest string
	createdBody := false
	if strings.HasPrefix(opts.BodyPath, "@sha256:") {
		bodyDigest = strings.TrimPrefix(opts.BodyPath, "@")
		if err := a.validateEndpoint(store.EndpointObject, bodyDigest); err != nil {
			return nil, err
		}
	} else {
		info, err := os.Stat(opts.BodyPath)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			return nil, fmt.Errorf("experience body must be a file")
		}
		bodyDigest, _, createdBody, err = a.storePath(opts.BodyPath, store.KindExperience)
		if err != nil {
			return nil, err
		}
		if opts.Title == "" {
			opts.Title = filepath.Base(opts.BodyPath)
		}
		opts.BodyName = filepath.Base(opts.BodyPath)
	}
	if opts.Title == "" {
		opts.Title = "experience"
	}

	payload := experiencePayload{
		Title:      opts.Title,
		BodyDigest: bodyDigest,
		BodyName:   opts.BodyName,
		Status:     opts.Status,
		Scopes:     append([]string(nil), opts.Scopes...),
		Evidence:   append([]string(nil), opts.Evidence...),
		Revises:    revises,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	rec, err := a.putRecord("experience", raw, opts.Actor)
	if err != nil {
		return nil, err
	}
	if _, err := a.Link(store.EndpointRecord, rec.ID, store.EndpointObject, bodyDigest, "has_body", ""); err != nil {
		return nil, err
	}
	for _, ep := range endpoints {
		if _, err := a.Link(store.EndpointRecord, rec.ID, ep.kind, ep.id, "supported_by", ""); err != nil {
			return nil, err
		}
	}
	if revises != "" {
		if _, err := a.Link(store.EndpointRecord, rec.ID, store.EndpointRecord, revises, "revises", ""); err != nil {
			return nil, err
		}
	}

	res := &ExperienceResult{
		RecordID: rec.ID, BodyDigest: bodyDigest, Status: opts.Status,
		Evidence: append([]string(nil), opts.Evidence...), Revises: revises, CreatedBody: createdBody,
	}
	return res, a.emit(res, func() {
		fmt.Fprintf(a.Out, "experience %s (%s)\n  body %s\n", rec.ID, opts.Status, bodyDigest)
		if revises != "" {
			fmt.Fprintf(a.Out, "  revises %s\n", revises)
		}
	})
}

func validExperienceStatus(value string) bool {
	switch value {
	case ExperienceProposed, ExperienceActive, ExperienceContested, ExperienceSuperseded, ExperienceRetired:
		return true
	default:
		return false
	}
}
