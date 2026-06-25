package bazi

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/wangxintong/yijing/backend/internal/config"
	"github.com/wangxintong/yijing/backend/internal/model"
)

// FullReportGenerator produces bazi full reports via DeepSeek with template fallback.
type FullReportGenerator struct {
	cfg      *config.Config
	deepseek *deepSeekFullGenerator
}

func NewFullReportGenerator(cfg *config.Config) *FullReportGenerator {
	return &FullReportGenerator{
		cfg:      cfg,
		deepseek: newDeepSeekFullGenerator(cfg),
	}
}

func (g *FullReportGenerator) Generate(
	ctx context.Context,
	analysisID int64,
	resultPayload json.RawMessage,
	freeContent string,
) (content string, aiProvider string, err error) {
	promptInput, err := buildFullReportPromptInput(resultPayload, freeContent)
	if err != nil {
		return "", "", err
	}

	if g != nil && g.deepseek != nil && g.deepseek.enabled() {
		start := time.Now()
		aiContent, aiErr := g.deepseek.generate(ctx, analysisID, promptInput)
		duration := time.Since(start).Milliseconds()
		if aiErr == nil {
			log.Printf("[ai] analysis_id=%d module=bazi provider=deepseek model=%s duration_ms=%d success=true",
				analysisID, modelName(g.cfg), duration)
			return aiContent, model.AIProviderDeepSeek, nil
		}
		log.Printf("[ai] analysis_id=%d module=bazi provider=deepseek model=%s duration_ms=%d success=false error=%q",
			analysisID, modelName(g.cfg), duration, truncateAIErr(aiErr.Error()))
	}

	templateContent, err := BuildFullContent(resultPayload, freeContent)
	if err != nil {
		return "", "", err
	}
	return templateContent, model.AIProviderTemplateFallback, nil
}

func modelName(cfg *config.Config) string {
	if cfg == nil {
		return "unknown"
	}
	name := strings.TrimSpace(cfg.DeepSeekModel)
	if name == "" {
		return "unknown"
	}
	return name
}

func truncateAIErr(message string) string {
	message = strings.ReplaceAll(message, "\n", " ")
	if len(message) > 120 {
		return message[:120] + "…"
	}
	return message
}
