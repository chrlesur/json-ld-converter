package llm

import (
	"context"
)

type AnalysisContext struct {
	PreviousEntities  map[string]string
	PreviousRelations []string
	Summary           string
}

type LLMClient interface {
    Analyze(ctx context.Context, content string, analysisContext *AnalysisContext) (string, *AnalysisContext, error)
}
