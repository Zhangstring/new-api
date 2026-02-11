package claude

import (
	"regexp"
	"strings"

	"github.com/QuantumNous/new-api/dto"
)

const mcpToolPrefix = "mcp_"

// mcpNameRegex 匹配 JSON 中 "name":"mcp_xxx" 模式，用于响应中去除前缀
var mcpNameRegex = regexp.MustCompile(`"name"\s*:\s*"mcp_([^"]+)"`)

// ApplyMcpToolPrefix 为请求中的工具名添加 mcp_ 前缀
// 处理三个位置：工具定义、消息中的 tool_use 块、tool_choice 中的 name
func ApplyMcpToolPrefix(request *dto.ClaudeRequest) {
	if request == nil {
		return
	}

	// 1. 工具定义：tools[].name
	if tools, ok := request.Tools.([]interface{}); ok {
		for _, tool := range tools {
			switch t := tool.(type) {
			case map[string]interface{}:
				if name, ok := t["name"].(string); ok && name != "" && !strings.HasPrefix(name, mcpToolPrefix) {
					t["name"] = mcpToolPrefix + name
				}
			case *dto.Tool:
				if t.Name != "" && !strings.HasPrefix(t.Name, mcpToolPrefix) {
					t.Name = mcpToolPrefix + t.Name
				}
			}
		}
	}

	// 2. 消息中的 tool_use 块：messages[].content[].name
	for i := range request.Messages {
		if contents, ok := request.Messages[i].Content.([]interface{}); ok {
			for _, c := range contents {
				if m, ok := c.(map[string]interface{}); ok {
					if m["type"] == "tool_use" {
						if name, ok := m["name"].(string); ok && name != "" && !strings.HasPrefix(name, mcpToolPrefix) {
							m["name"] = mcpToolPrefix + name
						}
					}
				}
			}
		}
	}

	// 3. tool_choice 中的 name（type 为 "tool" 时指定具体工具名）
	if tc, ok := request.ToolChoice.(map[string]interface{}); ok {
		if tc["type"] == "tool" {
			if name, ok := tc["name"].(string); ok && name != "" && !strings.HasPrefix(name, mcpToolPrefix) {
				tc["name"] = mcpToolPrefix + name
			}
		}
	}
}

// StripMcpPrefixFromStreamResponse 从流式 OpenAI 响应中去除工具名的 mcp_ 前缀
func StripMcpPrefixFromStreamResponse(response *dto.ChatCompletionsStreamResponse) {
	if response == nil {
		return
	}
	for i := range response.Choices {
		for j := range response.Choices[i].Delta.ToolCalls {
			tc := &response.Choices[i].Delta.ToolCalls[j]
			tc.Function.Name = strings.TrimPrefix(tc.Function.Name, mcpToolPrefix)
		}
	}
}

// StripMcpPrefixFromData 从 JSON 字符串中去除工具名的 mcp_ 前缀
func StripMcpPrefixFromData(data string) string {
	return mcpNameRegex.ReplaceAllString(data, `"name":"$1"`)
}

// StripMcpPrefixFromBytes 从 JSON 字节中去除工具名的 mcp_ 前缀
func StripMcpPrefixFromBytes(data []byte) []byte {
	return mcpNameRegex.ReplaceAll(data, []byte(`"name":"$1"`))
}
