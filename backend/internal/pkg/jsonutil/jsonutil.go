package jsonutil

import (
	"encoding/json"
	"fmt"
	"strings"
)

var requiredFullReportFields = []string{
	"summary",
	"overall",
	"current_state",
	"opportunity",
	"risk",
	"action_steps",
	"emotion_reminder",
	"reflection_questions",
	"disclaimer",
}

// ValidateJSONString checks if s is valid JSON object.
func ValidateJSONString(s string) error {
	var obj map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &obj); err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("json is not an object")
	}
	return nil
}

// ExtractJSONObjectFromText tries to parse s as JSON or extract first {...} block.
func ExtractJSONObjectFromText(s string) (string, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return "", fmt.Errorf("empty content")
	}

	if err := ValidateJSONString(trimmed); err == nil {
		return trimmed, nil
	}

	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)
	if err := ValidateJSONString(trimmed); err == nil {
		return trimmed, nil
	}

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		candidate := trimmed[start : end+1]
		if err := ValidateJSONString(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no valid json object found")
}

// EnsureRequiredFields validates and fills missing fields with defaults.
func EnsureRequiredFields(jsonStr string) (string, error) {
	var obj map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return "", err
	}

	defaultDisclaimer := "本内容仅供娱乐和传统文化参考，不构成现实决策建议。"

	for _, key := range requiredFullReportFields {
		if _, ok := obj[key]; !ok {
			switch key {
			case "action_steps", "reflection_questions":
				obj[key] = []any{}
			case "disclaimer":
				obj[key] = defaultDisclaimer
			default:
				obj[key] = ""
			}
		}
	}

	if steps, ok := obj["action_steps"].([]any); !ok || len(steps) == 0 {
		obj["action_steps"] = []any{
			"先整理已知事实与未知项，再决定下一步。",
			"设定一个低成本试探动作，验证方向。",
			"安排固定复盘时间，记录感受与依据。",
		}
	}
	if qs, ok := obj["reflection_questions"].([]any); !ok || len(qs) == 0 {
		obj["reflection_questions"] = []any{
			"我真正想改善的是什么？",
			"哪些风险是真实的，哪些来自焦虑？",
			"三个月后回看，我希望今天做了什么？",
		}
	}

	for _, key := range requiredFullReportFields {
		val := obj[key]
		switch key {
		case "action_steps", "reflection_questions":
			arr, ok := val.([]any)
			if !ok || len(arr) == 0 {
				return "", fmt.Errorf("field %s invalid", key)
			}
		default:
			str, ok := val.(string)
			if !ok || strings.TrimSpace(str) == "" {
				if key == "disclaimer" {
					obj[key] = defaultDisclaimer
					continue
				}
				return "", fmt.Errorf("field %s missing or empty", key)
			}
		}
	}

	out, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
