package migrate

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const bootstrapSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
  id          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  version     VARCHAR(32)  NOT NULL                COMMENT '迁移版本号，如 001',
  filename    VARCHAR(255) NOT NULL                COMMENT '迁移文件名',
  checksum    VARCHAR(64)  NOT NULL                COMMENT '文件内容 SHA256',
  executed_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '执行时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_schema_migrations_filename (filename),
  KEY idx_schema_migrations_version (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SQL 迁移执行记录表';
`

var versionPattern = regexp.MustCompile(`^(\d+)_`)

type Runner struct {
	db     *sql.DB
	sqlDir string
}

func NewRunner(db *sql.DB, sqlDir string) *Runner {
	return &Runner{db: db, sqlDir: sqlDir}
}

func ResolveSQLDir(configured string) (string, error) {
	if configured != "" {
		if st, err := os.Stat(configured); err == nil && st.IsDir() {
			return configured, nil
		}
		return "", fmt.Errorf("SQL_DIR not found: %s", configured)
	}

	candidates := []string{"../sql", "../../sql", "sql", "/sql"}
	for _, dir := range candidates {
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			abs, err := filepath.Abs(dir)
			if err != nil {
				return dir, nil
			}
			return abs, nil
		}
	}
	return "", fmt.Errorf("sql directory not found; set SQL_DIR or run from project root/backend")
}

func (r *Runner) Run(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, bootstrapSQL); err != nil {
		return fmt.Errorf("bootstrap schema_migrations: %w", err)
	}

	files, err := listMigrationFiles(r.sqlDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := r.applyFile(ctx, file); err != nil {
			return err
		}
	}
	return nil
}

func listMigrationFiles(sqlDir string) ([]string, error) {
	entries, err := os.ReadDir(sqlDir)
	if err != nil {
		return nil, fmt.Errorf("read sql dir: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".sql") && versionPattern.MatchString(name) {
			files = append(files, filepath.Join(sqlDir, name))
		}
	}
	sort.Strings(files)
	return files, nil
}

func (r *Runner) applyFile(ctx context.Context, path string) error {
	filename := filepath.Base(path)
	version := extractVersion(filename)

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", filename, err)
	}
	checksum := sha256Hex(content)

	existingChecksum, found, err := r.getChecksum(ctx, filename)
	if err != nil {
		return err
	}
	if found {
		if existingChecksum == checksum {
			fmt.Printf("skip %s (already applied)\n", filename)
			return nil
		}
		return fmt.Errorf("migration %s checksum mismatch: database has %s, file has %s", filename, existingChecksum, checksum)
	}

	fmt.Printf("apply %s ...\n", filename)
	if err := r.execSQL(ctx, string(content)); err != nil {
		return fmt.Errorf("apply migration %s: %w", filename, err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO schema_migrations (version, filename, checksum, executed_at)
		VALUES (?, ?, ?, ?)
	`, version, filename, checksum, time.Now())
	if err != nil {
		return fmt.Errorf("record migration %s: %w", filename, err)
	}
	fmt.Printf("done %s\n", filename)
	return nil
}

func (r *Runner) getChecksum(ctx context.Context, filename string) (string, bool, error) {
	var checksum string
	err := r.db.QueryRowContext(ctx, `
		SELECT checksum FROM schema_migrations WHERE filename = ? LIMIT 1
	`, filename).Scan(&checksum)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return checksum, true, nil
}

func (r *Runner) execSQL(ctx context.Context, script string) error {
	statements := splitSQLStatements(script)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("exec failed: %w\nstatement: %s", err, truncate(stmt, 200))
		}
	}
	return nil
}

func splitSQLStatements(script string) []string {
	lines := strings.Split(script, "\n")
	var (
		buf        strings.Builder
		statements []string
	)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		buf.WriteString(line)
		buf.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, buf.String())
			buf.Reset()
		}
	}
	if tail := strings.TrimSpace(buf.String()); tail != "" {
		statements = append(statements, tail)
	}
	return statements
}

func extractVersion(filename string) string {
	m := versionPattern.FindStringSubmatch(filename)
	if len(m) > 1 {
		return m[1]
	}
	return "000"
}

func sha256Hex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// SplitSQLStatements exposes statement splitting for tests.
func SplitSQLStatements(script string) []string {
	return splitSQLStatements(script)
}

// ExtractVersion exposes migration version parsing for tests.
func ExtractVersion(filename string) string {
	return extractVersion(filename)
}
