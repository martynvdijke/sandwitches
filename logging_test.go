package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/martynvdijke/sandwitches-go/internal/config"
)

func TestRotateStaleLogRotatesDjangoLog(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "sandwitches.log")

	content := "2026-08-12 10:00:00 INFO django.server: ...\n" +
		"/app/.venv/lib/python3.14/site-packages/django/core/handlers/base.py\n" +
		"Traceback (most recent call last):\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	backup, err := rotateStaleLog(path)
	if err != nil {
		t.Fatalf("rotateStaleLog failed: %v", err)
	}
	if backup == "" {
		t.Fatal("expected rotation of a Django-era log")
	}
	if _, err := os.Stat(backup); err != nil {
		t.Errorf("backup file missing: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("original path should be gone after rotation")
	}
}

func TestRotateStaleLogKeepsGoLog(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "sandwitches.log")
	content := "2026/08/15 14:20:30 admin.go:783: message\nDatabase initialized\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	backup, err := rotateStaleLog(path)
	if err != nil {
		t.Fatalf("rotateStaleLog failed: %v", err)
	}
	if backup != "" {
		t.Errorf("Go log should not be rotated, got backup %q", backup)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("original log should remain: %v", err)
	}
}

func TestRotateStaleLogMissingFile(t *testing.T) {
	backup, err := rotateStaleLog(filepath.Join(t.TempDir(), "nope.log"))
	if err != nil {
		t.Fatalf("rotateStaleLog failed: %v", err)
	}
	if backup != "" {
		t.Errorf("expected no rotation for missing file, got %q", backup)
	}
}

func TestSetupLogFileWritesToFile(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{
		MediaRoot: filepath.Join(tmp, "media"),
		LogFile:   filepath.Join(tmp, "media", "sandwitches.log"),
	}

	// Stale Django log present before startup.
	stale := filepath.Join(tmp, "media", "sandwitches.log")
	if err := os.MkdirAll(filepath.Dir(stale), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stale, []byte("2026-08-12 INFO django.server: old\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	setupLogFile(cfg)

	data, err := os.ReadFile(cfg.LogFile)
	if err != nil {
		t.Fatalf("log file not readable: %v", err)
	}
	if strings.Contains(string(data), "django.server") {
		t.Error("stale Django content leaked into new log file")
	}
	if !strings.Contains(string(data), "Logging to") {
		t.Error("expected startup log line in new log file")
	}

	// Backup of the stale log must exist.
	matches, _ := filepath.Glob(filepath.Join(tmp, "media", "*.bak"))
	if len(matches) != 1 {
		t.Errorf("expected 1 backup of stale log, got %v", matches)
	}
}
