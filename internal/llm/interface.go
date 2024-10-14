package llm

import (
	"context"
)

type LLMClient interface {
	Analyze(ctx context.Context, content string) (string, error)
}
