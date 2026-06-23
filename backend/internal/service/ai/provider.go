package ai

import "context"

type Provider interface {
	Name() string
	GenerateFullInterpretation(ctx context.Context, input GenerateInput) (*GenerateOutput, error)
}
