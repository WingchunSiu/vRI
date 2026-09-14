// Package objects implements vRI's content-addressed file storage and
// deterministic directory digests.
package objects

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const DigestPrefix = "sha256:"

// DirEntry is one file's contribution to a directory digest.
type DirEntry struct {
	Path   string `json:"path"` // slash-separated path relative to the directory root
	Digest string `json:"digest"`
	Size   int64  `json:"size"`
}

// dirManifest is the stored content of a directory object. The digest of a
// directory is computed from the (path, file digest) stream, not from this
// manifest; the manifest exists so `object get` can reconstruct the tree.
type dirManifest struct {
	Marker string     `json:"vri_dir_manifest"` // magic marker, always "v0"
	Files  []DirEntry `json:"files"`
}

const dirManifestMarker = "v0"

// DigestFile returns the content digest of a single file.
func DigestFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return DigestPrefix + hex.EncodeToString(h.Sum(nil)), nil
}

// DigestDir hashes a directory as the sorted-by-path sequence of
// (relative path, NUL, hex file digest, NUL). Changing this algorithm changes
// persisted object identities and requires an explicit migration.
func DigestDir(root string) (string, []DirEntry, error) {
	var entries []DirEntry
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("objects: %s is not a regular file", path)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		fd, err := DigestFile(path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		entries = append(entries, DirEntry{Path: filepath.ToSlash(rel), Digest: fd, Size: info.Size()})
		return nil
	})
	if err != nil {
		return "", nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	h := sha256.New()
	for _, e := range entries {
		h.Write([]byte(e.Path))
		h.Write([]byte{0})
		h.Write([]byte(strings.TrimPrefix(e.Digest, DigestPrefix)))
		h.Write([]byte{0})
	}
	return DigestPrefix + hex.EncodeToString(h.Sum(nil)), entries, nil
}

// BlobPath maps a digest to its path relative to the objects/ directory:
// the first two hex chars become a fan-out subdirectory.
func BlobPath(digest string) (string, error) {
	hexPart, ok := strings.CutPrefix(digest, DigestPrefix)
	if !ok || len(hexPart) != 64 {
		return "", fmt.Errorf("objects: malformed digest %q", digest)
	}
	if _, err := hex.DecodeString(hexPart); err != nil {
		return "", fmt.Errorf("objects: malformed digest %q", digest)
	}
	return filepath.Join(hexPart[:2], hexPart[2:]), nil
}

// Write copies a file into the object store. It returns the digest, the
// blob path relative to objectsDir, the size, and whether the blob was new.
func Write(objectsDir, src string) (digest, rel string, size int64, created bool, err error) {
	digest, err = DigestFile(src)
	if err != nil {
		return "", "", 0, false, err
	}
	rel, err = BlobPath(digest)
	if err != nil {
		return "", "", 0, false, err
	}
	info, err := os.Stat(src)
	if err != nil {
		return "", "", 0, false, err
	}
	size = info.Size()
	dst := filepath.Join(objectsDir, rel)
	if _, err := os.Stat(dst); err == nil {
		return digest, rel, size, false, nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", "", 0, false, err
	}
	if err := copyFile(dst, src); err != nil {
		return "", "", 0, false, err
	}
	return digest, rel, size, true, nil
}

// WriteDir stores every file under root as a blob and writes a directory
// manifest blob addressed by the directory digest.
func WriteDir(objectsDir, root string) (digest, rel string, created bool, err error) {
	digest, entries, err := DigestDir(root)
	if err != nil {
		return "", "", false, err
	}
	for _, e := range entries {
		if _, _, _, _, err := Write(objectsDir, filepath.Join(root, filepath.FromSlash(e.Path))); err != nil {
			return "", "", false, err
		}
	}
	rel, err = BlobPath(digest)
	if err != nil {
		return "", "", false, err
	}
	dst := filepath.Join(objectsDir, rel)
	if _, err := os.Stat(dst); err == nil {
		return digest, rel, false, nil
	}
	manifest, err := json.Marshal(dirManifest{Marker: dirManifestMarker, Files: entries})
	if err != nil {
		return "", "", false, err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", "", false, err
	}
	if err := os.WriteFile(dst, manifest, 0o644); err != nil {
		return "", "", false, err
	}
	return digest, rel, true, nil
}

// Read returns the blob content at a store-relative path.
func Read(objectsDir, rel string) ([]byte, error) {
	return os.ReadFile(filepath.Join(objectsDir, rel))
}

// IsDirManifest reports whether blob content is a stored directory manifest.
func IsDirManifest(content []byte) bool {
	var m dirManifest
	if err := json.Unmarshal(content, &m); err != nil {
		return false
	}
	return m.Marker == dirManifestMarker
}

// DirManifestEntries returns the files referenced by a directory manifest.
// It is used by context views to inspect a captured directory without first
// materializing the whole tree. The returned slice is a copy.
func DirManifestEntries(content []byte) ([]DirEntry, error) {
	var m dirManifest
	if err := json.Unmarshal(content, &m); err != nil {
		return nil, err
	}
	if m.Marker != dirManifestMarker {
		return nil, fmt.Errorf("objects: not a directory manifest")
	}
	return append([]DirEntry(nil), m.Files...), nil
}

// ExtractDir rebuilds a directory tree from a manifest blob.
func ExtractDir(objectsDir string, manifestContent []byte, outDir string) error {
	var m dirManifest
	if err := json.Unmarshal(manifestContent, &m); err != nil {
		return err
	}
	if m.Marker != dirManifestMarker {
		return fmt.Errorf("objects: not a directory manifest")
	}
	for _, e := range m.Files {
		rel, err := BlobPath(e.Digest)
		if err != nil {
			return err
		}
		content, err := Read(objectsDir, rel)
		if err != nil {
			return fmt.Errorf("objects: manifest references missing blob %s: %w", e.Digest, err)
		}
		dst := filepath.Join(outDir, filepath.FromSlash(e.Path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
