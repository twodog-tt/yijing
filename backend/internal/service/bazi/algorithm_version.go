package bazi

import (
	"fmt"
	"strings"

	"github.com/wangxintong/yijing/backend/internal/model"
)

var ErrInvalidAlgorithmVersion = fmt.Errorf("%w: invalid algorithm_version", ErrInvalidParams)

// ResolveAlgorithmVersion maps create input to a supported bazi algorithm version.
// Empty input defaults to bazi-simple-v1.
func ResolveAlgorithmVersion(raw string) (string, error) {
	version := strings.TrimSpace(raw)
	if version == "" {
		return model.AlgorithmVersionBaziSimpleV1, nil
	}
	switch version {
	case model.AlgorithmVersionBaziSimpleV1, AlgorithmVersionBaziV2POC:
		return version, nil
	default:
		return "", ErrInvalidAlgorithmVersion
	}
}
