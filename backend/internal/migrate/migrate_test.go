package migrate_test

import (
	"os"
	"strings"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/migrate"
)

func TestSplitSQLStatementsSkipsComments(t *testing.T) {
	script := `
-- comment line
USE yijing;
CREATE TABLE IF NOT EXISTS demo (id INT);
`
	stmts := migrate.SplitSQLStatements(script)
	if len(stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d: %#v", len(stmts), stmts)
	}
	if !strings.Contains(stmts[0], "USE yijing") {
		t.Fatalf("unexpected first statement: %s", stmts[0])
	}
}

func TestExtractVersion(t *testing.T) {
	if got := migrate.ExtractVersion("005_phase7_ai_logs.sql"); got != "005" {
		t.Fatalf("expected 005, got %s", got)
	}
	if got := migrate.ExtractVersion("007_analysis_records.sql"); got != "007" {
		t.Fatalf("expected 007, got %s", got)
	}
}

func TestAnalysisRecordsMigrationIsIdempotent(t *testing.T) {
	content, err := os.ReadFile("../../../sql/007_analysis_records.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}

	script := string(content)
	if !strings.Contains(script, "CREATE TABLE IF NOT EXISTS analysis_records") {
		t.Fatalf("expected CREATE TABLE IF NOT EXISTS analysis_records")
	}
	upper := strings.ToUpper(script)
	for _, keyword := range []string{"ALTER TABLE", "DROP TABLE", "TRUNCATE"} {
		if strings.Contains(upper, keyword) {
			t.Fatalf("migration must not contain %s", keyword)
		}
	}

	stmts := migrate.SplitSQLStatements(script)
	if len(stmts) < 2 {
		t.Fatalf("expected at least USE + CREATE statements, got %d", len(stmts))
	}
}
