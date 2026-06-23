package migrate_test

import (
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
}
