package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/martynvdijke/sandwitches-go/internal/config"
	"gorm.io/gorm"
)

// legacyCleanupMigration is the name recorded in go_migrations once the
// legacy Django tables have been dropped. It makes the cleanup one-time.
const legacyCleanupMigration = "drop_legacy_django_tables"

// legacyTablePrefixes match every table created by the former Django app.
// None of the current Go models use these prefixes, so matching them cannot
// accidentally drop application data.
var legacyTablePrefixes = []string{"django_", "sandwitches_", "auth_"}

// goMigration tracks one-time schema migrations performed by this Go app.
type goMigration struct {
	Name      string    `gorm:"primaryKey"`
	AppliedAt time.Time
}

// CleanupLegacy backs up the database and drops leftover Django tables.
//
// It is deliberately safe:
//   - No backup, no drop: if the backup cannot be created and verified the
//     cleanup is skipped and the database is left untouched.
//   - One-time only: a marker row in go_migrations records that the cleanup
//     ran, so subsequent startups are no-ops.
//   - Idempotent: DROP TABLE IF EXISTS is harmless even if the marker write
//     was interrupted between the drops and the marker insert.
func CleanupLegacy(cfg *config.Config) {
	legacy, err := legacyTables(DB)
	if err != nil {
		log.Printf("Legacy cleanup skipped: %v", err)
		return
	}
	if len(legacy) == 0 {
		log.Println("No legacy Django tables found; cleanup not needed")
		return
	}

	backupPath, err := BackupDatabase(cfg.DatabaseFile)
	if err != nil {
		log.Printf("Legacy cleanup skipped: backup failed: %v", err)
		return
	}

	dropped, err := CleanupLegacyTables(DB)
	if err != nil {
		log.Printf("Legacy cleanup failed after backup (%s): %v", backupPath, err)
		return
	}
	log.Printf("Removed legacy Django tables: %s (backup: %s)", strings.Join(dropped, ", "), backupPath)
}

// BackupDatabase creates a timestamped, verified copy of the SQLite database
// next to the original (e.g. db.sqlite3.pre-go-cleanup-20260815T120000.bak).
// It uses VACUUM INTO so the snapshot is transactionally consistent even with
// a live WAL database. The backup is verified to be a readable SQLite file
// before it is returned.
func BackupDatabase(dbPath string) (string, error) {
	if dbPath == "" {
		return "", fmt.Errorf("empty database path")
	}
	if info, err := os.Stat(dbPath); err != nil || info.IsDir() {
		return "", fmt.Errorf("database file not found: %s", dbPath)
	}

	backupPath := dbPath + ".pre-go-cleanup-" + time.Now().Format("20060102T150405.000000000") + ".bak"

	conn, err := sql.Open("sqlite3", dbPath+"?_busy_timeout=10000")
	if err != nil {
		return "", fmt.Errorf("open database for backup: %w", err)
	}
	defer conn.Close()

	// VACUUM INTO cannot run inside a transaction and does not accept bound
	// parameters, so escape the path for the SQL literal.
	escaped := strings.ReplaceAll(backupPath, "'", "''")
	if _, err := conn.Exec("VACUUM INTO '" + escaped + "'"); err != nil {
		return "", fmt.Errorf("create backup: %w", err)
	}

	info, err := os.Stat(backupPath)
	if err != nil || info.Size() == 0 {
		return "", fmt.Errorf("backup file missing or empty: %s", backupPath)
	}

	verify, err := sql.Open("sqlite3", backupPath)
	if err != nil {
		return "", fmt.Errorf("open backup for verification: %w", err)
	}
	defer verify.Close()
	var tables int
	if err := verify.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table'").Scan(&tables); err != nil {
		return "", fmt.Errorf("backup verification failed: %w", err)
	}

	return backupPath, nil
}

// CleanupLegacyTables drops every table matching the legacy Django prefixes
// and records the migration marker. It is a no-op once the marker exists.
func CleanupLegacyTables(db *gorm.DB) ([]string, error) {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS go_migrations (
		name TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL
	)`).Error; err != nil {
		return nil, fmt.Errorf("ensure go_migrations table: %w", err)
	}

	var marked int64
	if err := db.Model(&goMigration{}).Where("name = ?", legacyCleanupMigration).Count(&marked).Error; err != nil {
		return nil, fmt.Errorf("check migration marker: %w", err)
	}
	if marked > 0 {
		return nil, nil
	}

	legacy, err := legacyTables(db)
	if err != nil {
		return nil, err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, t := range legacy {
			if err := tx.Exec("DROP TABLE IF EXISTS " + quoteIdent(t)).Error; err != nil {
				return fmt.Errorf("drop %s: %w", t, err)
			}
		}
		if err := tx.Exec(
			"INSERT OR IGNORE INTO go_migrations (name, applied_at) VALUES (?, ?)",
			legacyCleanupMigration, time.Now(),
		).Error; err != nil {
			return fmt.Errorf("record migration marker: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return legacy, nil
}

// legacyTables lists current tables matching the legacy Django prefixes,
// sorted for deterministic output.
func legacyTables(db *gorm.DB) ([]string, error) {
	var tables []string
	if err := db.Raw("SELECT name FROM sqlite_master WHERE type = 'table'").Scan(&tables).Error; err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}

	var legacy []string
	for _, t := range tables {
		for _, p := range legacyTablePrefixes {
			if strings.HasPrefix(t, p) {
				legacy = append(legacy, t)
				break
			}
		}
	}
	sort.Strings(legacy)
	return legacy, nil
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
