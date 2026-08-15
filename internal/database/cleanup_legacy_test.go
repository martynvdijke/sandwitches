package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/martynvdijke/sandwitches-go/internal/config"
)

// setupCleanupTestDB creates a fresh DB in a temp dir and runs Init, returning
// the temp dir and config (mirrors setupTestDB but keeps paths handy).
func setupCleanupTestDB(t *testing.T) (string, *config.Config) {
	t.Helper()
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")
	cfg := &config.Config{
		DatabaseFile: dbPath,
		Debug:        false,
		SecretKey:    "testkey",
		LanguageCode: "en",
	}
	Init(cfg)
	return tmp, cfg
}

// seedLegacyTables creates representative Django-era tables plus a data row.
func seedLegacyTables(t *testing.T) {
	t.Helper()
	for _, stmt := range []string{
		`CREATE TABLE django_migrations (id INTEGER PRIMARY KEY, name TEXT)`,
		`CREATE TABLE django_tasks_database_dbtaskresult (id INTEGER PRIMARY KEY, task_path TEXT, status TEXT)`,
		`CREATE TABLE sandwitches_recipe (id INTEGER PRIMARY KEY, title TEXT)`,
		`CREATE TABLE auth_group (id INTEGER PRIMARY KEY, name TEXT)`,
	} {
		if err := DB.Exec(stmt).Error; err != nil {
			t.Fatalf("seed legacy table: %v", err)
		}
	}
	if err := DB.Exec(`INSERT INTO django_migrations (name) VALUES ('0001_initial')`).Error; err != nil {
		t.Fatalf("seed legacy row: %v", err)
	}
}

func TestBackupDatabase(t *testing.T) {
	_, cfg := setupCleanupTestDB(t)

	if err := DB.Exec(`CREATE TABLE dummy (id INTEGER PRIMARY KEY, val TEXT)`).Error; err != nil {
		t.Fatalf("create dummy table: %v", err)
	}
	if err := DB.Exec(`INSERT INTO dummy (val) VALUES ('hello')`).Error; err != nil {
		t.Fatalf("insert dummy row: %v", err)
	}

	backupPath, err := BackupDatabase(cfg.DatabaseFile)
	if err != nil {
		t.Fatalf("BackupDatabase failed: %v", err)
	}
	if backupPath == cfg.DatabaseFile {
		t.Fatal("backup path must differ from original")
	}

	info, err := os.Stat(backupPath)
	if err != nil {
		t.Fatalf("backup file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("backup file is empty")
	}

	// Back up a second time: must succeed with a distinct filename.
	second, err := BackupDatabase(cfg.DatabaseFile)
	if err != nil {
		t.Fatalf("second backup failed: %v", err)
	}
	if second == backupPath {
		t.Fatal("backups must have distinct filenames")
	}
}

func TestBackupDatabaseMissingFile(t *testing.T) {
	_, err := BackupDatabase(filepath.Join(t.TempDir(), "does-not-exist.db"))
	if err == nil {
		t.Fatal("expected error for missing database file")
	}
}

func TestCleanupLegacyTables(t *testing.T) {
	setupCleanupTestDB(t)
	seedLegacyTables(t)

	dropped, err := CleanupLegacyTables(DB)
	if err != nil {
		t.Fatalf("CleanupLegacyTables failed: %v", err)
	}
	if len(dropped) != 4 {
		t.Errorf("expected 4 legacy tables dropped, got %d: %v", len(dropped), dropped)
	}

	var legacy int64
	DB.Raw(`SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'table' AND (name LIKE 'django_%' OR name LIKE 'sandwitches_%' OR name LIKE 'auth_%')`).
		Scan(&legacy)
	if legacy != 0 {
		t.Errorf("legacy tables still present: %d", legacy)
	}

	// Go-side tables must be untouched.
	for _, table := range []string{"users", "recipes", "settings", "go_migrations"} {
		var n int64
		DB.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&n)
		if n != 1 {
			t.Errorf("Go table %q was dropped by cleanup", table)
		}
	}

	// Second run must be a no-op (idempotent via marker).
	again, err := CleanupLegacyTables(DB)
	if err != nil {
		t.Fatalf("second CleanupLegacyTables failed: %v", err)
	}
	if len(again) != 0 {
		t.Errorf("second run should drop nothing, got %v", again)
	}
}

func TestCleanupLegacyBackupThenDrop(t *testing.T) {
	tmp, cfg := setupCleanupTestDB(t)
	seedLegacyTables(t)

	CleanupLegacy(cfg)

	matches, _ := filepath.Glob(filepath.Join(tmp, "*.bak"))
	if len(matches) != 1 {
		t.Fatalf("expected exactly 1 backup file, got %v", matches)
	}

	var legacy int64
	DB.Raw(`SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'table' AND (name LIKE 'django_%' OR name LIKE 'sandwitches_%' OR name LIKE 'auth_%')`).
		Scan(&legacy)
	if legacy != 0 {
		t.Errorf("legacy tables should be gone after backup+cleanup: %d", legacy)
	}
}

func TestCleanupLegacySkipsWhenBackupFails(t *testing.T) {
	_, cfg := setupCleanupTestDB(t)
	seedLegacyTables(t)

	// Point the backup at an unwritable location so it fails; cleanup must
	// then leave everything untouched.
	badCfg := &config.Config{DatabaseFile: "/nonexistent-dir/db.sqlite3"}
	CleanupLegacy(badCfg)

	var legacy int64
	DB.Raw(`SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'table' AND (name LIKE 'django_%' OR name LIKE 'sandwitches_%' OR name LIKE 'auth_%')`).
		Scan(&legacy)
	if legacy != 4 {
		t.Errorf("legacy tables must remain when backup fails, got %d", legacy)
	}

	_ = cfg
}

func TestCleanupLegacyNoLegacyTables(t *testing.T) {
	tmp, cfg := setupCleanupTestDB(t)

	CleanupLegacy(cfg)

	matches, _ := filepath.Glob(filepath.Join(tmp, "*.bak"))
	if len(matches) != 0 {
		t.Errorf("no backup should be created without legacy tables, got %v", matches)
	}
}
