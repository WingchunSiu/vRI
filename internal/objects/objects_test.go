package objects

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func makeTestDir(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "tree")
	writeTestFile(t, filepath.Join(root, "a.txt"), "alpha\n")
	writeTestFile(t, filepath.Join(root, "nested", "b.txt"), "beta\n")
	return root
}

func TestDigestDirGolden(t *testing.T) {
	root := makeTestDir(t)
	got, entries, err := DigestDir(root)
	if err != nil {
		t.Fatal(err)
	}
	want := "sha256:c5f6b3639afd3475e30d4d5757fcb53c7d26172398a156ecd6bcaf42f2be03b1"
	if got != want {
		t.Fatalf("directory digest = %s, want %s", got, want)
	}
	if len(entries) != 2 || entries[0].Path != "a.txt" || entries[1].Path != "nested/b.txt" {
		t.Fatalf("unexpected digest entries: %+v", entries)
	}
}

func TestDigestDirChangesWithContent(t *testing.T) {
	root := makeTestDir(t)
	before, _, err := DigestDir(root)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(root, "nested", "b.txt"), "changed\n")
	after, _, err := DigestDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if before == after {
		t.Fatal("directory digest did not change with file content")
	}
}

func TestWriteAndReadRoundtrip(t *testing.T) {
	objectsDir := t.TempDir()
	src := filepath.Join(t.TempDir(), "source.md")
	writeTestFile(t, src, "portable source\n")

	digest, rel, _, created, err := Write(objectsDir, src)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("first write should report created")
	}
	_, _, _, created2, err := Write(objectsDir, src)
	if err != nil {
		t.Fatal(err)
	}
	if created2 {
		t.Fatal("second write of same content should not create")
	}
	content, err := Read(objectsDir, rel)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "portable source\n" {
		t.Fatal("roundtrip content mismatch")
	}
	if IsDirManifest(content) {
		t.Fatal("plain file misdetected as dir manifest")
	}
	if _, err := BlobPath("notadigest"); err == nil {
		t.Fatal("malformed digest accepted")
	}
	fd, err := DigestFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if fd != digest {
		t.Fatalf("DigestFile %s != Write digest %s", fd, digest)
	}
}

func TestWriteDirExtractRoundtrip(t *testing.T) {
	objectsDir := t.TempDir()
	src := makeTestDir(t)

	digest, _, created, err := WriteDir(objectsDir, src)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("first write should report created")
	}
	rel, err := BlobPath(digest)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := Read(objectsDir, rel)
	if err != nil {
		t.Fatal(err)
	}
	if !IsDirManifest(manifest) {
		t.Fatal("dir blob is not a manifest")
	}
	out := filepath.Join(t.TempDir(), "extracted")
	if err := ExtractDir(objectsDir, manifest, out); err != nil {
		t.Fatal(err)
	}
	redigest, _, err := DigestDir(out)
	if err != nil {
		t.Fatal(err)
	}
	if redigest != digest {
		t.Fatalf("extracted dir digest %s, want %s", redigest, digest)
	}
}
