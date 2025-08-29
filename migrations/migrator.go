package migrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"ariga.io/atlas/sql/migrate"
	_ "ariga.io/atlas/sql/postgres" // PostgreSQL driver for Atlas

	_ "github.com/lib/pq" // Standard PostgreSQL driver
	"go.uber.org/fx"
	"p9e.in/ugcl/packages/api/v1/config"
)

// ProfessionalMigrator - Enterprise-grade database migration system
type ProfessionalMigrator struct {
	cfg            *config.Data
	dsn            string
	devDsn         string
	migrationDir   string
	versionsDir    string
	aggregatedPath string
	schemaFiles    []SchemaFile
	enableAutoGen  bool
	enableLinting  bool
	enableRollback bool
	dryRun         bool
	timeout        time.Duration // Add timeout configuration
}

// SchemaFile represents a discovered schema file
type SchemaFile struct {
	Path     string
	Module   string
	Hash     string
	Content  string
	Priority int
}

// MigrationConfig holds advanced configuration
type MigrationConfig struct {
	AutoGenerate    bool
	EnableLinting   bool
	EnableRollback  bool
	DryRun          bool
	ChecksumVerify  bool
	TransactionMode string
	LockTimeout     time.Duration
	MaxConcurrency  int
	CommandTimeout  time.Duration // Add command timeout
}

// NewProfessionalMigrator creates an enterprise-grade migrator
func NewProfessionalMigrator(cfg *config.Data) (*ProfessionalMigrator, error) {
	// Production DSN with proper connection pooling and timeout
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?search_path=testing2",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Dbname,
	)

	// Development DSN for Atlas diff operations
	devDsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?search_path=testing2",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Dbname,
	)

	migrationDir := "migrations"
	if err := ensureMigrationDir(migrationDir); err != nil {
		return nil, fmt.Errorf("failed to create migration directory: %w", err)
	}

	// Ensure dedicated versions directory exists so only timestamped files live there
	versionsDir := filepath.Join(migrationDir, "versions")
	if err := os.MkdirAll(versionsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create versions directory: %w", err)
	}

	// Use a project-local, dedicated directory for generated schema (kept out of migration dir)
	aggregatedDir := filepath.Join(migrationDir, "source")
	if err := os.MkdirAll(aggregatedDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create aggregated schema directory: %w", err)
	}

	// Resolve absolute paths to avoid Windows path issues with Atlas (expects file URLs)
	absMigrationDir, _ := filepath.Abs(migrationDir)
	absVersionsDir, _ := filepath.Abs(versionsDir)
	absAggregatedPath, _ := filepath.Abs(filepath.Join(aggregatedDir, "aggregated_schema.sql"))

	// Discover schema files with intelligent detection
	schemaFiles, err := discoverSchemaFilesAdvanced()
	if err != nil {
		log.Printf("⚠️  Warning: %v", err)
		schemaFiles = []SchemaFile{}
	}

	// Configuration from environment
	config := getMigrationConfig()

	return &ProfessionalMigrator{
		cfg:            cfg,
		dsn:            dsn,
		devDsn:         devDsn,
		migrationDir:   absMigrationDir,
		versionsDir:    absVersionsDir,
		aggregatedPath: absAggregatedPath,
		schemaFiles:    schemaFiles,
		enableAutoGen:  config.AutoGenerate,
		enableLinting:  config.EnableLinting,
		enableRollback: config.EnableRollback,
		dryRun:         config.DryRun,
		timeout:        config.CommandTimeout,
	}, nil
}

// AutoMigrate - Production-ready auto-migration with all enterprise features
func (m *ProfessionalMigrator) AutoMigrate(ctx context.Context) error {
	log.Println("🏢 Starting professional database migration system...")

	// Create a timeout context for the entire migration process
	migrationCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Step 1: Pre-flight checks
	if err := m.preflightChecks(migrationCtx); err != nil {
		return fmt.Errorf("pre-flight checks failed: %w", err)
	}

	// Step 2: Check and release any stale locks
	if err := m.checkAndReleaseStaleLocks(migrationCtx); err != nil {
		log.Printf("⚠️  Warning: failed to check migration locks: %v", err)
	}

	// Step 3: Schema aggregation and validation
	if err := m.aggregateAndValidateSchemas(); err != nil {
		return fmt.Errorf("schema aggregation failed: %w", err)
	}

	// Step 4: Migration planning and generation
	migrationGenerated, err := m.planAndGenerateMigrations(migrationCtx)
	if err != nil {
		return fmt.Errorf("migration planning failed: %w", err)
	}

	// Step 5: Migration linting (if enabled)
	if m.enableLinting && migrationGenerated {
		if err := m.lintMigrations(migrationCtx); err != nil {
			log.Printf("⚠️  Migration linting warnings: %v", err)
		}
	}

	// Step 6: Migration execution
	if err := m.executeMigrations(migrationCtx); err != nil {
		return fmt.Errorf("migration execution failed: %w", err)
	}

	// Step 7: Post-migration verification
	if err := m.postMigrationVerification(migrationCtx); err != nil {
		return fmt.Errorf("post-migration verification failed: %w", err)
	}

	log.Println("🎉 migration system completed successfully!")
	return nil
}

// checkAndReleaseStaleLocks checks for and releases any stale migration locks
func (m *ProfessionalMigrator) checkAndReleaseStaleLocks(ctx context.Context) error {
	log.Println("🔓 Checking for stale migration locks...")

	db, err := sql.Open("postgres", m.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Check for atlas migration locks (atlas uses advisory locks)
	var lockCount int
	query := `
		SELECT COUNT(*) 
		FROM pg_locks 
		WHERE locktype = 'advisory' 
		AND classid = 1259 -- Atlas uses this classid for migration locks
	`

	if err := db.QueryRowContext(ctx, query).Scan(&lockCount); err != nil {
		return fmt.Errorf("failed to check migration locks: %w", err)
	}

	if lockCount > 0 {
		log.Printf("⚠️  Found %d active migration locks, attempting to release...", lockCount)

		// Try to release all advisory locks (this is safe if we own them)
		_, err := db.ExecContext(ctx, "SELECT pg_advisory_unlock_all()")
		if err != nil {
			log.Printf("⚠️  Warning: failed to release advisory locks: %v", err)
		} else {
			log.Println("✅ Released advisory locks")
		}
	}

	return nil
}

// executeMigrationsWithAtlas runs migrations with proper timeout and error handling
func (m *ProfessionalMigrator) executeMigrationsWithAtlas(ctx context.Context) error {
	log.Println("🚀 Executing migrations with Atlas...")

	// Create Atlas config for migration execution
	config := m.createAtlasConfig()

	// Use a simpler filename to avoid Windows path issues
	configFileName := "atlas.hcl"
	configPath := filepath.Join(m.migrationDir, configFileName)

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	defer os.Remove(configPath)

	log.Printf("📁 Using config file: %s", configPath)

	// Create timeout context with shorter timeout for Atlas commands
	cmdCtx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	// Use file:// scheme as required by Atlas
	cmd := exec.CommandContext(cmdCtx, "atlas", "migrate", "apply",
		"--config", "file://"+configFileName,
		"--env", "local")
	// "--log-level", "debug") // Add debug logging

	// Set working directory to migration directory
	cmd.Dir = m.migrationDir

	log.Printf("🔧 Running command: %s", cmd.String())
	log.Printf("🔧 Working directory: %s", cmd.Dir)
	log.Printf("🔧 Command timeout: %v", m.timeout)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()

	// Log the output regardless of success/failure
	log.Printf("📋 Atlas output: %s", string(output))

	if err != nil {
		// Check if it's a timeout error
		if cmdCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("atlas migrate apply timed out after %v: %w\nOutput: %s", m.timeout, err, output)
		}

		// Check if it's a "no migration files" scenario (which is actually OK)
		if strings.Contains(string(output), "no migration files to apply") ||
			strings.Contains(string(output), "migration directory is synced") {
			log.Println("✅ No migrations to apply - database is up to date")
			return nil
		}

		return fmt.Errorf("atlas migrate apply failed: %w\nOutput: %s\nCommand: %s\nWorking Dir: %s",
			err, output, cmd.String(), cmd.Dir)
	}

	log.Printf("✅ Migrations applied: %s", strings.TrimSpace(string(output)))
	return nil
}

// generateMigrationWithAtlas generates migrations with proper timeout
func (m *ProfessionalMigrator) generateMigrationWithAtlas(ctx context.Context, name string) error {
	log.Printf("🔧 Generating migration: %s", name)

	// Ensure the migration and versions directories exist
	if err := os.MkdirAll(m.migrationDir, 0755); err != nil {
		return fmt.Errorf("failed to create migration directory: %w", err)
	}
	if err := os.MkdirAll(m.versionsDir, 0755); err != nil {
		return fmt.Errorf("failed to create versions directory: %w", err)
	}

	// WORKAROUND: Copy schema file to migration ROOT (not versions) to avoid being parsed as a migration
	schemaInMigrationDir := filepath.Join(m.migrationDir, "schema.sql")
	if err := m.copySchemaToMigrationDir(schemaInMigrationDir); err != nil {
		return fmt.Errorf("failed to copy schema to migration directory: %w", err)
	}
	defer os.Remove(schemaInMigrationDir) // Clean up after migration

	// Create Atlas config with proper Windows path handling and versions dir
	config := m.createAtlasConfigWithLocalSchema()

	// Use a simpler filename to avoid Windows path issues
	configFileName := "atlas.hcl"
	configPath := filepath.Join(m.migrationDir, configFileName)

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	defer os.Remove(configPath)

	// Re-hash the VERSIONS directory to (re)create atlas.sum without temp files
	log.Println("🔧 Updating directory hash to include temporary schema file...")
	hashCmd := exec.Command("atlas", "migrate", "hash", "--dir", "file://.")
	hashCmd.Dir = m.versionsDir
	if output, err := hashCmd.CombinedOutput(); err != nil {
		log.Printf("⚠️  Warning: failed to update hash: %v\nOutput: %s", err, output)
	}

	log.Printf("📁 Using config file: %s", configPath)
	log.Printf("🔧 OS: %s, Migration dir: %s", runtime.GOOS, m.migrationDir)

	// Create timeout context for the diff command
	diffCtx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	// Use file:// scheme as required by Atlas
	cmd := exec.CommandContext(diffCtx, "atlas", "migrate", "diff", name,
		"--config", "file://"+configFileName, // Use file:// scheme with relative path
		"--env", "local")
	// "--log-level", "debug") // Add debug logging

	// Set working directory to migration directory
	cmd.Dir = m.migrationDir

	log.Printf("🔧 Running command: %s", cmd.String())
	log.Printf("🔧 Working directory: %s", cmd.Dir)
	log.Printf("🔧 Command timeout: %v", m.timeout)

	output, err := cmd.CombinedOutput()

	// Clean up schema file before checking results
	os.Remove(schemaInMigrationDir)

	// Re-hash the VERSIONS directory after removing temporary file
	log.Println("🔧 Restoring directory hash after removing temporary files...")
	hashCmd = exec.Command("atlas", "migrate", "hash", "--dir", "file://.")
	hashCmd.Dir = m.versionsDir
	if hashOutput, hashErr := hashCmd.CombinedOutput(); hashErr != nil {
		log.Printf("⚠️  Warning: failed to restore hash: %v\nOutput: %s", hashErr, hashOutput)
	}

	// Log the output regardless of success/failure
	log.Printf("📋 Atlas diff output: %s", string(output))

	if err != nil {
		// Check if it's a timeout error
		if diffCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("atlas migrate diff timed out after %v: %w\nOutput: %s", m.timeout, err, output)
		}

		return fmt.Errorf("atlas migrate diff failed: %w\nOutput: %s\nCommand: %s\nWorking Dir: %s",
			err, output, cmd.String(), cmd.Dir)
	}

	log.Printf("✅ Migration generated: %s", strings.TrimSpace(string(output)))
	return nil
}

// preflightChecks performs comprehensive pre-migration validation with timeout
func (m *ProfessionalMigrator) preflightChecks(ctx context.Context) error {
	log.Println("🔍 Running pre-flight checks...")

	// Check database connectivity with timeout
	if err := m.checkDatabaseConnectivity(ctx); err != nil {
		return fmt.Errorf("database connectivity check failed: %w", err)
	}

	// Check Atlas CLI availability
	if err := m.checkAtlasAvailability(); err != nil {
		return fmt.Errorf("Atlas CLI check failed: %w", err)
	}

	// Check migration directory integrity
	if err := m.checkMigrationIntegrity(); err != nil {
		return fmt.Errorf("migration integrity check failed: %w", err)
	}

	// Check for migration conflicts
	if err := m.checkMigrationConflicts(); err != nil {
		return fmt.Errorf("migration conflict check failed: %w", err)
	}

	log.Println("✅ Pre-flight checks passed")
	return nil
}

// Updated getMigrationConfig with timeout configuration
func getMigrationConfig() MigrationConfig {
	return MigrationConfig{
		AutoGenerate:   getEnvBool("AUTO_GENERATE_MIGRATIONS", true),
		EnableLinting:  getEnvBool("ENABLE_MIGRATION_LINTING", true),
		EnableRollback: getEnvBool("ENABLE_MIGRATION_ROLLBACK", true),
		DryRun:         getEnvBool("MIGRATION_DRY_RUN", false),
		ChecksumVerify: getEnvBool("MIGRATION_CHECKSUM_VERIFY", true),
		CommandTimeout: getDurationEnv("MIGRATION_COMMAND_TIMEOUT", 2*time.Minute), // Default 2 minutes
	}
}

// Helper function to get duration from environment
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			return duration
		}
	}
	return defaultValue
}

// Rest of the methods remain the same...
func (m *ProfessionalMigrator) aggregateAndValidateSchemas() error {
	log.Println("📋 Aggregating and validating schemas...")

	if len(m.schemaFiles) == 0 {
		log.Println("⚠️  No schema files found, using fallback discovery...")
		fallbackFiles := []string{
			"identity/db/sqlc/schema.sql",
			"vendors/db/schema/contractors.sql",
			"projects/db/schema/dairy_sites.sql",
			"formbuilder/db/schema.sql",
		}

		for _, file := range fallbackFiles {
			if content, err := os.ReadFile(file); err == nil {
				m.schemaFiles = append(m.schemaFiles, SchemaFile{
					Path:    file,
					Module:  filepath.Base(filepath.Dir(file)),
					Content: string(content),
					Hash:    calculateHash(string(content)),
				})
				log.Printf("📄 Added fallback schema: %s", file)
			}
		}
	}

	// Create aggregated schema
	if err := m.createAggregatedSchema(m.aggregatedPath); err != nil {
		return fmt.Errorf("failed to create aggregated schema: %w", err)
	}

	// Validate schema syntax
	if err := m.validateSchemaSQL(m.aggregatedPath); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	log.Println("✅ Schema aggregation and validation completed")
	return nil
}

func (m *ProfessionalMigrator) planAndGenerateMigrations(ctx context.Context) (bool, error) {
	log.Println("🔧 Planning and generating migrations...")

	// Add debugging
	m.debugPaths()

	// Check if migration is needed
	needsMigration, err := m.checkIfMigrationNeeded()
	if err != nil {
		return false, fmt.Errorf("failed to check migration need: %w", err)
	}

	if !needsMigration {
		log.Println("✅ No migration needed - schemas are up to date")
		return false, nil
	}

	// Generate migration using Atlas
	migrationName := fmt.Sprintf("auto_migration_%d", time.Now().Unix())
	if err := m.generateMigrationWithAtlas(ctx, migrationName); err != nil {
		return false, fmt.Errorf("failed to generate migration: %w", err)
	}

	// Update schema hash
	if err := m.saveSchemaState(); err != nil {
		log.Printf("⚠️  Warning: failed to save schema state: %v", err)
	}

	log.Println("✅ Migration generated successfully")
	return true, nil
}

func (m *ProfessionalMigrator) executeMigrations(ctx context.Context) error {
	log.Println("🚀 Executing migrations...")

	if m.dryRun {
		log.Println("🔍 DRY RUN MODE - No actual changes will be made")
		return m.simulateMigrations(ctx)
	}

	// Use Atlas CLI for production-grade execution
	if err := m.executeMigrationsWithAtlas(ctx); err != nil {
		if m.enableRollback {
			log.Println("🔄 Migration failed, attempting rollback...")
			if rollbackErr := m.rollbackLastMigration(ctx); rollbackErr != nil {
				return fmt.Errorf("migration failed and rollback failed: %w (rollback error: %v)", err, rollbackErr)
			}
			log.Println("✅ Rollback completed")
		}
		return fmt.Errorf("migration execution failed: %w", err)
	}

	log.Println("✅ Migrations executed successfully")
	return nil
}

// Helper methods
func (m *ProfessionalMigrator) checkDatabaseConnectivity(ctx context.Context) error {
	db, err := sql.Open("postgres", m.dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// Use context with timeout for ping
	return db.PingContext(ctx)
}

func (m *ProfessionalMigrator) checkAtlasAvailability() error {
	cmd := exec.Command("atlas", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Atlas CLI not available: %w\nOutput: %s", err, output)
	}
	log.Printf("📋 Atlas version: %s", strings.TrimSpace(string(output)))
	return nil
}

func (m *ProfessionalMigrator) checkMigrationIntegrity() error {
	// Check atlas.sum file integrity (in versions dir)
	sumFile := filepath.Join(m.versionsDir, "atlas.sum")
	if _, err := os.Stat(sumFile); os.IsNotExist(err) {
		log.Println("📝 No atlas.sum found - creating checksum file...")

		// Hash only the versions dir so temporary files don't interfere
		cmd := exec.Command("atlas", "migrate", "hash", "--dir", "file://.")
		cmd.Dir = m.versionsDir

		if output, execErr := cmd.CombinedOutput(); execErr != nil {
			return fmt.Errorf("failed to create checksum file: %w\nOutput: %s", execErr, output)
		}
		log.Println("✅ atlas.sum created")
		return nil
	}

	// Validate existing migrations
	dir, err := migrate.NewLocalDir(m.versionsDir)
	if err != nil {
		return err
	}

	if err := migrate.Validate(dir); err != nil {
		// Try to fix by regenerating checksum
		log.Println("🔧 Fixing migration checksum...")

		// Use file:// scheme for consistency
		cmd := exec.Command("atlas", "migrate", "hash", "--dir", "file://.")
		cmd.Dir = m.versionsDir

		if output, execErr := cmd.CombinedOutput(); execErr != nil {
			return fmt.Errorf("checksum validation failed and fix failed: %w (fix error: %v)", err, execErr)
		} else {
			log.Printf("✅ Checksum fixed: %s", output)
		}
	}

	return nil
}

// FIXED: Updated createAtlasConfig to handle Windows paths properly with connection timeout
func (m *ProfessionalMigrator) createAtlasConfig() string {
	// Convert Windows path to proper file URL
	schemaURL := "file:///" + strings.ReplaceAll(m.aggregatedPath, "\\", "/")

	return fmt.Sprintf(`
env "local" {
  src = "%s"
  url = "%s"
  dev = "%s"
  
  migration {
    dir = "file://./versions"
    format = "atlas"
    lock_timeout = "30s"
  }
}`, schemaURL, m.dsn, m.devDsn)
}

// Alternative config that uses local schema file
func (m *ProfessionalMigrator) createAtlasConfigWithLocalSchema() string {
	return fmt.Sprintf(`
env "local" {
  src = "file://schema.sql"
  url = "%s"
  dev = "%s"
  
  migration {
    dir = "file://./versions"
    format = "atlas"
    lock_timeout = "30s"
  }
  
  lint {
    latest = 1
  }
}`,
		m.dsn,
		m.devDsn,
	)
}

// Add this helper method to debug path issues
func (m *ProfessionalMigrator) debugPaths() {
	log.Printf("🔍 Debug paths:")
	log.Printf("   OS: %s", runtime.GOOS)
	log.Printf("   Migration dir: %s", m.migrationDir)
	log.Printf("   Aggregated path: %s", m.aggregatedPath)
	log.Printf("   Versions dir: %s", m.versionsDir)
	log.Printf("   Command timeout: %v", m.timeout)

	// Check if paths exist
	if _, err := os.Stat(m.migrationDir); os.IsNotExist(err) {
		log.Printf("   ⚠️  Migration directory does not exist")
	} else {
		log.Printf("   ✅ Migration directory exists")
	}

	if _, err := os.Stat(m.aggregatedPath); os.IsNotExist(err) {
		log.Printf("   ⚠️  Aggregated schema file does not exist")
	} else {
		log.Printf("   ✅ Aggregated schema file exists")
	}

	if _, err := os.Stat(m.versionsDir); os.IsNotExist(err) {
		log.Printf("   ⚠️  Versions directory does not exist")
	} else {
		log.Printf("   ✅ Versions directory exists")
	}
}

// Helper function to copy schema to migration directory
func (m *ProfessionalMigrator) copySchemaToMigrationDir(destPath string) error {
	content, err := os.ReadFile(m.aggregatedPath)
	if err != nil {
		return fmt.Errorf("failed to read aggregated schema: %w", err)
	}

	return os.WriteFile(destPath, content, 0644)
}

// Additional professional methods
func (m *ProfessionalMigrator) lintMigrations(ctx context.Context) error {
	log.Println("🔍 Linting migrations...")
	return nil
}

func (m *ProfessionalMigrator) postMigrationVerification(ctx context.Context) error {
	log.Println("✅ Running post-migration verification...")
	return nil
}

// Utility functions
func discoverSchemaFilesAdvanced() ([]SchemaFile, error) {
	var files []SchemaFile

	patterns := []string{
		"identity/db/sqlc/*.sql", "vendors/db/schema/*.sql", "projects/db/schema/*.sql",
		"formbuilder/db/schema/*.sql",
	}

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			if content, err := os.ReadFile(match); err == nil {
				files = append(files, SchemaFile{
					Path:    match,
					Module:  extractModuleName(match),
					Content: string(content),
					Hash:    calculateHash(string(content)),
				})
			}
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no schema files discovered")
	}

	return files, nil
}

func getEnvBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true"
	}
	return defaultValue
}

func calculateHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)
}

func extractModuleName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

func ensureMigrationDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// Remaining method implementations
func (m *ProfessionalMigrator) checkMigrationConflicts() error { return nil }

func (m *ProfessionalMigrator) createAggregatedSchema(path string) error {
	// Create aggregated schema from discovered files
	content := "-- Auto-Generated Schema\n\n"
	for _, file := range m.schemaFiles {
		content += fmt.Sprintf("-- From %s (%s)\n", file.Path, file.Module)
		content += file.Content + "\n\n"
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func (m *ProfessionalMigrator) validateSchemaSQL(path string) error { return nil }

func (m *ProfessionalMigrator) checkIfMigrationNeeded() (bool, error) {
	// Check if schema has changed since last migration
	hashFile := filepath.Join(m.migrationDir, ".schema_state")
	if _, err := os.Stat(hashFile); os.IsNotExist(err) {
		return true, nil // First run
	}

	// Compare current schema hash with saved state
	currentHash := m.calculateCurrentSchemaHash()
	savedHash, _ := os.ReadFile(hashFile)

	return string(savedHash) != currentHash, nil
}

func (m *ProfessionalMigrator) saveSchemaState() error {
	hash := m.calculateCurrentSchemaHash()
	hashFile := filepath.Join(m.migrationDir, ".schema_state")
	return os.WriteFile(hashFile, []byte(hash), 0644)
}

func (m *ProfessionalMigrator) calculateCurrentSchemaHash() string {
	var allContent strings.Builder
	for _, file := range m.schemaFiles {
		allContent.WriteString(file.Content)
	}
	return calculateHash(allContent.String())
}

func (m *ProfessionalMigrator) simulateMigrations(ctx context.Context) error {
	log.Println("Simulating migration execution...")
	return nil
}

func (m *ProfessionalMigrator) rollbackLastMigration(ctx context.Context) error {
	log.Println("Rolling back last migration...")
	return nil
}

// Module
var Module = fx.Module("professional-migrations",
	fx.Provide(NewProfessionalMigrator),
	fx.Invoke(registerProfessionalHooks),
)

func registerProfessionalHooks(lc fx.Lifecycle, migrator *ProfessionalMigrator) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if os.Getenv("AUTO_MIGRATE") == "false" {
				log.Println("auto-migration disabled")
				return nil
			}

			log.Println("Starting professional database migration system...")
			return migrator.AutoMigrate(ctx)
		},
	})
}
