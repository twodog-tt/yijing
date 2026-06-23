package interpretation

import (
	"testing"

	"github.com/wangxintong/yijing/backend/internal/model"
)

func TestHasExistingFullContent(t *testing.T) {
	full := `{"summary":"x"}`
	if hasExistingFullContent(nil) {
		t.Fatal("nil should be false")
	}
	if hasExistingFullContent(&model.Interpretation{}) {
		t.Fatal("empty should be false")
	}
	if !hasExistingFullContent(&model.Interpretation{FullContent: &full}) {
		t.Fatal("existing content should be true")
	}
}
