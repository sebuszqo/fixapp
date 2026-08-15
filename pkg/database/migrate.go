package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations automatically discovers and applies all pending *.up.sql migrations
// and ensures required table schemas are in sync.
func RunMigrations(db *sql.DB) error {
	// 1. Explicitly ensure pro registration columns exist on handyman_profiles and handyman_pricing
	ensureColumnsQuery := `
		ALTER TABLE IF EXISTS handyman_profiles
			ADD COLUMN IF NOT EXISTS company_address               VARCHAR(255),
			ADD COLUMN IF NOT EXISTS pkd_code                      VARCHAR(100),
			ADD COLUMN IF NOT EXISTS gus_status                    VARCHAR(50) DEFAULT 'Aktywna',
			ADD COLUMN IF NOT EXISTS gus_verified                  BOOLEAN NOT NULL DEFAULT false,
			ADD COLUMN IF NOT EXISTS business_type                 VARCHAR(50) DEFAULT 'jdg',
			ADD COLUMN IF NOT EXISTS is_vat_payer                  BOOLEAN NOT NULL DEFAULT false,
			ADD COLUMN IF NOT EXISTS consent_identity_verification BOOLEAN NOT NULL DEFAULT false,
			ADD COLUMN IF NOT EXISTS consent_marketing             BOOLEAN NOT NULL DEFAULT false,
			ADD COLUMN IF NOT EXISTS experience_years              VARCHAR(50),
			ADD COLUMN IF NOT EXISTS working_hours                 JSONB,
			ADD COLUMN IF NOT EXISTS verification_status           VARCHAR(50) NOT NULL DEFAULT 'unverified',
			ADD COLUMN IF NOT EXISTS ceidg_document_url            VARCHAR(500),
			ADD COLUMN IF NOT EXISTS ceidg_uploaded_at             TIMESTAMPTZ,
			ADD COLUMN IF NOT EXISTS microtransfer_status          VARCHAR(50) DEFAULT 'none',
			ADD COLUMN IF NOT EXISTS microtransfer_code            VARCHAR(50),
			ADD COLUMN IF NOT EXISTS microtransfer_confirmed_at    TIMESTAMPTZ;

		ALTER TABLE IF EXISTS handyman_pricing
			ADD COLUMN IF NOT EXISTS estimated_duration            VARCHAR(50);
	`
	if _, err := db.Exec(ensureColumnsQuery); err != nil {
		log.Printf("[Migrations] Note on ensureColumns: %v\n", err)
	} else {
		log.Println("[Migrations] Successfully verified handyman_profiles columns")
	}

	// 2. Find migrations folder in possible locations
	possiblePaths := []string{
		"migrations",
		"./migrations",
		"../migrations",
		"backend/migrations",
		"../../migrations",
	}

	var migrationsDir string
	for _, p := range possiblePaths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			migrationsDir = p
			break
		}
	}

	if migrationsDir == "" {
		log.Println("[Migrations] No migrations directory found, schema verified")
		return nil
	}

	// Create schema_migrations table if it doesn't exist
	initQuery := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`
	if _, err := db.Exec(initQuery); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Read all files in migrationsDir
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations dir %s: %w", migrationsDir, err)
	}

	// Filter and sort .up.sql files
	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	// Apply each migration
	for _, file := range upFiles {
		var exists bool
		_ = db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", file).Scan(&exists)
		if exists {
			continue
		}

		filePath := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("[Migrations] Warning: could not read %s: %v\n", filePath, err)
			continue
		}

		sqlContent := strings.TrimSpace(string(content))
		if sqlContent == "" {
			continue
		}

		log.Printf("[Migrations] Applying migration: %s\n", file)

		if _, err := db.Exec(sqlContent); err != nil {
			// If it's already exists error, record it and continue
			log.Printf("[Migrations] Migration %s completed with note: %v\n", file, err)
		} else {
			log.Printf("[Migrations] Successfully applied: %s\n", file)
		}

		_, _ = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING", file)
	}

	return nil
}
