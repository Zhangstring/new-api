package claude

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/gin-gonic/gin"
)

const Context1mBetaHeader = "context-1m-2025-08-07"

// Context1mDisabledError is returned when context-1m is disabled but client requested it.
// Callers should check for this error type to return 400 + skip retry.
type Context1mDisabledError struct{}

func (e *Context1mDisabledError) Error() string {
	return "1M context window is disabled on this channel. " +
		"Please remove the context-1m beta header or use a channel with 1M context enabled"
}

// GetContext1mPreference reads the 1M context preference from channel settings.
// Returns "inherit" if not set or not found.
func GetContext1mPreference(c *gin.Context) string {
	settings, ok := common.GetContextKeyType[dto.ChannelOtherSettings](
		c, constant.ContextKeyChannelOtherSetting,
	)
	if !ok || settings.Context1mPreference == "" {
		return "inherit"
	}
	return settings.Context1mPreference
}

// clientRequestsContext1m checks if the anthropic-beta header contains a context-1m flag.
func clientRequestsContext1m(beta string) bool {
	if beta == "" {
		return false
	}
	for _, part := range strings.Split(beta, ",") {
		if strings.HasPrefix(strings.TrimSpace(part), "context-1m") {
			return true
		}
	}
	return false
}

// ensureContext1mBeta ensures the anthropic-beta header contains the context-1m flag.
func ensureContext1mBeta(beta string) string {
	if beta == "" {
		return Context1mBetaHeader
	}
	if clientRequestsContext1m(beta) {
		return beta
	}
	return beta + "," + Context1mBetaHeader
}

// ApplyContext1mPreference processes the anthropic-beta header based on channel preference.
//   - "disabled": returns Context1mDisabledError if client requested context-1m
//   - "force_enable": injects context-1m header if not present
//   - "inherit" (default): passes through as-is
func ApplyContext1mPreference(c *gin.Context, anthropicBeta string) (string, error) {
	pref := GetContext1mPreference(c)

	switch pref {
	case "disabled":
		if clientRequestsContext1m(anthropicBeta) {
			return "", &Context1mDisabledError{}
		}
		return anthropicBeta, nil

	case "force_enable":
		return ensureContext1mBeta(anthropicBeta), nil

	default: // "inherit"
		return anthropicBeta, nil
	}
}
