package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOpenMSXHelpLastUpdatedLabel(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)
	t.Setenv("LOCALAPPDATA", tmp)
	t.Setenv("LocalAppData", tmp)

	if got := openMSXHelpLastUpdatedLabel(); !strings.Contains(got, "n/a") {
		t.Fatalf("expected n/a label without cache, got=%q", got)
	}

	cacheDir := openMSXHelpCacheDir()
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}
	file := filepath.Join(cacheDir, "setup-guide.html")
	if err := os.WriteFile(file, []byte("# cached"), 0o644); err != nil {
		t.Fatalf("write cache file: %v", err)
	}

	if got := openMSXHelpLastUpdatedLabel(); strings.Contains(got, "n/a") {
		t.Fatalf("expected timestamp label, got=%q", got)
	}
}

func TestPruneOpenMSXHelpCacheByAge(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)
	t.Setenv("LOCALAPPDATA", tmp)
	t.Setenv("LocalAppData", tmp)

	cacheDir := openMSXHelpCacheDir()
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}
	oldFile := filepath.Join(cacheDir, "old.html")
	newFile := filepath.Join(cacheDir, "new.html")
	if err := os.WriteFile(oldFile, []byte("old"), 0o644); err != nil {
		t.Fatalf("write old: %v", err)
	}
	if err := os.WriteFile(newFile, []byte("new"), 0o644); err != nil {
		t.Fatalf("write new: %v", err)
	}
	now := time.Now()
	oldTime := now.Add(-45 * 24 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes old: %v", err)
	}

	removed, err := pruneOpenMSXHelpCacheByAge(30*24*time.Hour, now)
	if err != nil {
		t.Fatalf("prune failed: %v", err)
	}
	if removed != 1 {
		t.Fatalf("removed=%d, want 1", removed)
	}
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Fatalf("old cache file should be deleted")
	}
	if _, err := os.Stat(newFile); err != nil {
		t.Fatalf("new cache file should remain: %v", err)
	}
}

func TestEnsureOpenMSXHelpFileUsesOfflineCacheWhenFetchFails(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)
	t.Setenv("LOCALAPPDATA", tmp)
	t.Setenv("LocalAppData", tmp)

	cacheDir := openMSXHelpCacheDir()
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}
	doc := openMSXRemoteHelpDocs[0]
	cachedFile := openMSXHelpFilePath(doc)
	if err := os.WriteFile(cachedFile, []byte("<html>cached</html>"), 0o644); err != nil {
		t.Fatalf("write cached html: %v", err)
	}
	// Simulate stale cache by back-dating the file
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(cachedFile, old, old); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	// Use a fake doc that will fail to connect
	fakeDOc := remoteHelpDoc{Title: "fake", URL: "http://127.0.0.1:1", Slug: doc.Slug}
	path, source, err := ensureOpenMSXHelpFile(fakeDOc, false)
	if err != nil {
		t.Fatalf("expected offline fallback, got error: %v", err)
	}
	if source != "cache-offline" {
		t.Fatalf("expected source=cache-offline, got=%q", source)
	}
	if !strings.HasSuffix(path, ".html") {
		t.Fatalf("expected .html path, got=%q", path)
	}
}
