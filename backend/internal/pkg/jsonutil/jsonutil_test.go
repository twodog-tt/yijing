package jsonutil

import (
	"strings"
	"testing"
)

const sampleJSON = `{
  "summary": "测试总结",
  "overall": "总体判断内容",
  "current_state": "当前处境内容",
  "opportunity": "机会点内容",
  "risk": "风险点内容",
  "action_steps": ["建议1", "建议2", "建议3"],
  "emotion_reminder": "情绪提醒内容",
  "reflection_questions": ["问题1", "问题2", "问题3"],
  "disclaimer": "本内容仅供娱乐和传统文化参考，不构成现实决策建议。"
}`

func TestValidateJSONString(t *testing.T) {
	if err := ValidateJSONString(sampleJSON); err != nil {
		t.Fatalf("expected valid json: %v", err)
	}
	if err := ValidateJSONString("not json"); err == nil {
		t.Fatal("expected error for invalid json")
	}
}

func TestExtractJSONObjectFromText(t *testing.T) {
	wrapped := "说明文字\n```json\n" + sampleJSON + "\n```\n尾部"
	got, err := ExtractJSONObjectFromText(wrapped)
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	if !strings.Contains(got, `"summary"`) {
		t.Fatalf("unexpected extract: %s", got)
	}
}

func TestEnsureRequiredFields(t *testing.T) {
	in := `{"summary":"a","overall":"b","current_state":"c","opportunity":"d","risk":"e","action_steps":["1"],"emotion_reminder":"f","reflection_questions":["q"]}`
	out, err := EnsureRequiredFields(in)
	if err != nil {
		t.Fatalf("ensure failed: %v", err)
	}
	if !strings.Contains(out, `"disclaimer"`) {
		t.Fatal("disclaimer should be filled")
	}
}

func TestEnsureRequiredFieldsRejectsEmptyOverall(t *testing.T) {
	in := `{"summary":"a","overall":"","current_state":"c","opportunity":"d","risk":"e","action_steps":["1"],"emotion_reminder":"f","reflection_questions":["q"],"disclaimer":"x"}`
	_, err := EnsureRequiredFields(in)
	if err == nil {
		t.Fatal("expected error for empty overall")
	}
}
