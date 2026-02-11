package claude

import (
	"strings"

	"github.com/QuantumNous/new-api/dto"
)

// ToolsAlreadyHaveMcpPrefix 检查请求中的工具是否已有 mcp_ 前缀
// 只要有任一工具带 mcp_ 前缀即视为真实 CLI 请求，跳过伪装避免双重前缀
func ToolsAlreadyHaveMcpPrefix(request *dto.ClaudeRequest) bool {
	if request == nil {
		return false
	}
	tools, ok := request.Tools.([]interface{})
	if !ok || len(tools) == 0 {
		return false
	}
	for _, tool := range tools {
		switch t := tool.(type) {
		case map[string]interface{}:
			if name, ok := t["name"].(string); ok && strings.HasPrefix(name, mcpToolPrefix) {
				return true
			}
		case *dto.Tool:
			if strings.HasPrefix(t.Name, mcpToolPrefix) {
				return true
			}
		}
	}
	return false
}
