package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/martynvdijke/sandwitches-go/internal/config"
)

// setupLogFile prepares the application log file so the Admin → Logs page
// shows live Go logs instead of the stale Django-era log that previous
// deployments left behind.
//
// It mirrors the database cleanup philosophy: if the existing sandwitches.log
// looks like a legacy Python/Django log, it is backed up beside the original
// and a fresh file is started. All Go log output is then written to both
// stdout (visible via `docker logs`) and the log file (visible in the admin
// dashboard).
func setupLogFile(cfg *config.Config) {
	logFile := cfg.LogFile
	if logFile == "" {
		logFile = filepath.Join(cfg.MediaRoot, "sandwitches.log")
	}

	if dir := filepath.Dir(logFile); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Printf("Warning: could not create log directory %s: %v", dir, err)
		}
	}

	backup, err := rotateStaleLog(logFile)
	if err != nil {
		log.Printf("Warning: could not rotate stale log file: %v", err)
	} else if backup != "" {
		log.Printf("Backed up legacy log file to: %s", backup)
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Printf("Warning: could not open log file %s: %v", logFile, err)
		return
	}

	log.SetOutput(io.MultiWriter(os.Stdout, f))
	log.Printf("Logging to %s (and stdout)", logFile)
}

// rotateStaleLog renames a legacy Django/Python log file to a timestamped
// backup path, leaving the original location free for fresh Go logs. It only
// acts when the file content shows Python-era markers, so a normal Go log
// file is never touched. Returns the backup path, or "" when no rotation
// happened.
func rotateStaleLog(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if info.Size() == 0 {
		return "", nil
	}

	// Peek at the first chunk; Django logs carry python/django markers
	// (venv paths, asgiref, django.server, tracebacks).
	const peekSize = 64 * 1024
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	buf := make([]byte, peekSize)
	n, _ := io.ReadFull(f, buf)
	head := strings.ToLower(string(buf[:n]))

	markers := []string{"site-packages", "django", "asgiref", "gunicorn", "python3", "traceback"}
	stale := false
	for _, m := range markers {
		if strings.Contains(head, m) {
			stale = true
			break
		}
	}
	if !stale {
		return "", nil
	}

	backup := fmt.Sprintf("%s.pre-go-cleanup-%s.bak", path, time.Now().Format("20060102T150405.000000000"))
	if err := os.Rename(path, backup); err != nil {
		return "", err
	}
	return backup, nil
}
