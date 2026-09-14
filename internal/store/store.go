// Package store owns the .vri directory: the SQLite database and the three
// tables (objects, records, relations) defined by docs/CORE_DESIGN.md.
// Reference integrity is enforced by the app layer, not here.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const DirName = ".vri"

// Schema is the v0 contract verbatim: three tables, no more.
const schema = `
CREATE TABLE IF NOT EXISTS objects (
  digest     TEXT PRIMARY KEY,
  kind       TEXT NOT NULL,
  path       TEXT NOT NULL,
  size_bytes INTEGER NOT NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS records (
  id         TEXT PRIMARY KEY,
  type       TEXT NOT NULL,
  payload    TEXT NOT NULL,
  actor      TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS relations (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  src_kind   TEXT NOT NULL,
  src_id     TEXT NOT NULL,
  dst_kind   TEXT NOT NULL,
  dst_id     TEXT NOT NULL,
  type       TEXT NOT NULL,
  meta       TEXT,
  created_at TEXT NOT NULL
);
`

// Object kinds defined by the v0 contract.
const (
	KindTaskPackage   = "task_package"
	KindJobArtifact   = "job_artifact"
	KindSourceExcerpt = "source_excerpt"
	KindFile          = "file"
	KindEvidence      = "evidence"
	KindExperience    = "experience_body"
)

// Endpoint kinds for relations.
const (
	EndpointObject = "object"
	EndpointRecord = "record"
)

type Object struct {
	Digest    string `json:"digest"`
	Kind      string `json:"kind"`
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	CreatedAt string `json:"created_at"`
}

type Record struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Payload   string `json:"payload"`
	Actor     string `json:"actor"`
	CreatedAt string `json:"created_at"`
}

type Relation struct {
	ID        int64  `json:"id"`
	SrcKind   string `json:"src_kind"`
	SrcID     string `json:"src_id"`
	DstKind   string `json:"dst_kind"`
	DstID     string `json:"dst_id"`
	Type      string `json:"type"`
	Meta      string `json:"meta,omitempty"`
	CreatedAt string `json:"created_at"`
}

type Store struct {
	db  *sql.DB
	Dir string // the .vri directory
}

func (s *Store) ObjectsDir() string { return filepath.Join(s.Dir, "objects") }

// Init creates a new store under projectDir. It fails if one already exists.
func Init(projectDir string) (*Store, error) {
	dir := filepath.Join(projectDir, DirName)
	if _, err := os.Stat(dir); err == nil {
		return nil, fmt.Errorf("store already exists at %s", dir)
	}
	if err := os.MkdirAll(filepath.Join(dir, "objects"), 0o755); err != nil {
		return nil, err
	}
	return open(dir)
}

// Open opens the store under projectDir, creating the schema if needed.
func Open(projectDir string) (*Store, error) {
	dir := filepath.Join(projectDir, DirName)
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("no vRI store in %s (run `vri init` first)", projectDir)
	}
	return open(dir)
}

func open(dir string) (*Store, error) {
	db, err := sql.Open("sqlite", filepath.Join(dir, "vri.db"))
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrating store: %w", err)
	}
	return &Store{db: db, Dir: dir}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) PutObject(o *Object) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO objects (digest, kind, path, size_bytes, created_at) VALUES (?, ?, ?, ?, ?)`,
		o.Digest, o.Kind, o.Path, o.SizeBytes, o.CreatedAt)
	return err
}

func (s *Store) GetObject(digest string) (*Object, error) {
	row := s.db.QueryRow(`SELECT digest, kind, path, size_bytes, created_at FROM objects WHERE digest = ?`, digest)
	return scanObject(row)
}

func (s *Store) HasObject(digest string) (bool, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM objects WHERE digest = ?`, digest).Scan(&n)
	return n > 0, err
}

func (s *Store) ListObjects() ([]*Object, error) {
	rows, err := s.db.Query(`SELECT digest, kind, path, size_bytes, created_at FROM objects ORDER BY created_at, digest`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Object
	for rows.Next() {
		o, err := scanObject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func scanObject(row interface{ Scan(...any) error }) (*Object, error) {
	var o Object
	if err := row.Scan(&o.Digest, &o.Kind, &o.Path, &o.SizeBytes, &o.CreatedAt); err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *Store) PutRecord(r *Record) error {
	_, err := s.db.Exec(
		`INSERT INTO records (id, type, payload, actor, created_at) VALUES (?, ?, ?, ?, ?)`,
		r.ID, r.Type, r.Payload, r.Actor, r.CreatedAt)
	return err
}

func (s *Store) GetRecord(id string) (*Record, error) {
	row := s.db.QueryRow(`SELECT id, type, payload, actor, created_at FROM records WHERE id = ?`, id)
	return scanRecord(row)
}

func (s *Store) ListRecords(typ string) ([]*Record, error) {
	query := `SELECT id, type, payload, actor, created_at FROM records`
	var rows *sql.Rows
	var err error
	if typ == "" {
		rows, err = s.db.Query(query + ` ORDER BY id`)
	} else {
		rows, err = s.db.Query(query+` WHERE type = ? ORDER BY id`, typ)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Record
	for rows.Next() {
		r, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanRecord(row interface{ Scan(...any) error }) (*Record, error) {
	var r Record
	if err := row.Scan(&r.ID, &r.Type, &r.Payload, &r.Actor, &r.CreatedAt); err != nil {
		return nil, err
	}
	return &r, nil
}

// AddRelation inserts a relation without checking that endpoints resolve;
// callers must validate reference integrity first (app layer does).
func (s *Store) AddRelation(rel *Relation) error {
	res, err := s.db.Exec(
		`INSERT INTO relations (src_kind, src_id, dst_kind, dst_id, type, meta, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rel.SrcKind, rel.SrcID, rel.DstKind, rel.DstID, rel.Type, nullEmpty(rel.Meta), rel.CreatedAt)
	if err != nil {
		return err
	}
	rel.ID, err = res.LastInsertId()
	return err
}

func (s *Store) RelationExists(srcKind, srcID, dstKind, dstID, typ string) (bool, error) {
	var n int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM relations WHERE src_kind = ? AND src_id = ? AND dst_kind = ? AND dst_id = ? AND type = ?`,
		srcKind, srcID, dstKind, dstID, typ).Scan(&n)
	return n > 0, err
}

// RelationsFrom returns relations whose source is the given endpoint.
func (s *Store) RelationsFrom(srcKind, srcID string) ([]*Relation, error) {
	return s.queryRelations(`WHERE src_kind = ? AND src_id = ?`, srcKind, srcID)
}

// RelationsTo returns relations of relType pointing at the given endpoint;
// relType may be "" to match all types.
func (s *Store) RelationsTo(dstKind, dstID, relType string) ([]*Relation, error) {
	if relType == "" {
		return s.queryRelations(`WHERE dst_kind = ? AND dst_id = ?`, dstKind, dstID)
	}
	return s.queryRelations(`WHERE dst_kind = ? AND dst_id = ? AND type = ?`, dstKind, dstID, relType)
}

func (s *Store) queryRelations(where string, args ...any) ([]*Relation, error) {
	rows, err := s.db.Query(
		`SELECT id, src_kind, src_id, dst_kind, dst_id, type, COALESCE(meta, ''), created_at FROM relations `+where+` ORDER BY id`,
		args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Relation
	for rows.Next() {
		var r Relation
		if err := rows.Scan(&r.ID, &r.SrcKind, &r.SrcID, &r.DstKind, &r.DstID, &r.Type, &r.Meta, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &r)
	}
	return out, rows.Err()
}

func nullEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
