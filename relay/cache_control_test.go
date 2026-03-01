package relay

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func TestSystemCacheControlPreservation(t *testing.T) {
	// 模拟用户请求的 JSON（与 request.json 结构一致）
	requestJSON := `{
		"model": "claude-sonnet-4-5-20250929",
		"max_tokens": 100,
		"stream": true,
		"temperature": 0.6,
		"system": [
			{
				"cache_control": {"type": "ephemeral"},
				"text": "You are a helpful assistant with extensive knowledge.",
				"type": "text"
			}
		],
		"messages": [
			{
				"role": "user",
				"content": [
					{
						"text": "Hello, this is a test message with cache control.",
						"type": "text",
						"cache_control": {"type": "ephemeral"}
					}
				]
			}
		]
	}`

	// ===== 步骤1: 模拟 GetAndValidateClaudeRequest (ShouldBindJSON) =====
	var request dto.ClaudeRequest
	err := common.Unmarshal([]byte(requestJSON), &request)
	if err != nil {
		t.Fatalf("步骤1 Unmarshal 失败: %v", err)
	}

	// 检查 System 解析后的类型
	t.Logf("步骤1 - System type: %T", request.System)
	systemJSON1, _ := json.Marshal(request.System)
	t.Logf("步骤1 - System JSON: %s", string(systemJSON1))
	assertContains(t, "步骤1 system", string(systemJSON1), "cache_control")

	// 检查 Messages[0].Content 解析后
	contentJSON1, _ := json.Marshal(request.Messages[0].Content)
	t.Logf("步骤1 - Messages[0].Content JSON: %s", string(contentJSON1))
	assertContains(t, "步骤1 messages", string(contentJSON1), "cache_control")

	// ===== 步骤2: 模拟 common.DeepCopy =====
	copied, err := common.DeepCopy(&request)
	if err != nil {
		t.Fatalf("步骤2 DeepCopy 失败: %v", err)
	}

	systemJSON2, _ := json.Marshal(copied.System)
	t.Logf("步骤2 - DeepCopy后 System JSON: %s", string(systemJSON2))
	assertContains(t, "步骤2 system DeepCopy", string(systemJSON2), "cache_control")

	contentJSON2, _ := json.Marshal(copied.Messages[0].Content)
	t.Logf("步骤2 - DeepCopy后 Messages[0].Content JSON: %s", string(contentJSON2))
	assertContains(t, "步骤2 messages DeepCopy", string(contentJSON2), "cache_control")

	// ===== 步骤3: 模拟 Marshal (ConvertClaudeRequest 后) =====
	jsonData, err := common.Marshal(copied)
	if err != nil {
		t.Fatalf("步骤3 Marshal 失败: %v", err)
	}

	t.Logf("步骤3 - 完整 Marshal JSON 长度: %d", len(jsonData))

	// 解析检查 system 和 messages
	var fullMap map[string]json.RawMessage
	common.Unmarshal(jsonData, &fullMap)

	systemStr := string(fullMap["system"])
	messagesStr := string(fullMap["messages"])
	t.Logf("步骤3 - system: %s", systemStr)
	t.Logf("步骤3 - messages: %.200s...", messagesStr)
	assertContains(t, "步骤3 Marshal system", systemStr, "cache_control")
	assertContains(t, "步骤3 Marshal messages", messagesStr, "cache_control")

	// ===== 步骤4: 模拟 RemoveDisabledFields =====
	jsonDataAfter, err := relaycommon.RemoveDisabledFields(jsonData, dto.ChannelOtherSettings{}, false)
	if err != nil {
		t.Fatalf("步骤4 RemoveDisabledFields 失败: %v", err)
	}

	var fullMap2 map[string]json.RawMessage
	common.Unmarshal(jsonDataAfter, &fullMap2)

	systemStr2 := string(fullMap2["system"])
	messagesStr2 := string(fullMap2["messages"])
	t.Logf("步骤4 - RemoveDisabledFields后 system: %s", systemStr2)
	assertContains(t, "步骤4 RemoveDisabledFields system", systemStr2, "cache_control")
	assertContains(t, "步骤4 RemoveDisabledFields messages", messagesStr2, "cache_control")

	// ===== 最终对比 =====
	t.Log("\n===== 最终对比 =====")
	t.Logf("原始 system cache_control: %v", containsStr(string(systemJSON1), "cache_control"))
	t.Logf("DeepCopy后 system cache_control: %v", containsStr(string(systemJSON2), "cache_control"))
	t.Logf("Marshal后 system cache_control: %v", containsStr(systemStr, "cache_control"))
	t.Logf("RemoveDisabledFields后 system cache_control: %v", containsStr(systemStr2, "cache_control"))
}

func assertContains(t *testing.T, step string, haystack string, needle string) {
	t.Helper()
	if !containsStr(haystack, needle) {
		t.Errorf("❌ %s: 预期包含 '%s' 但未找到!\n  实际值: %s", step, needle, haystack)
	} else {
		t.Logf("✅ %s: 包含 '%s'", step, needle)
	}
}

func containsStr(s, sub string) bool {
	return len(s) > 0 && len(sub) > 0 && fmt.Sprintf("%s", s) != "" && json.Valid([]byte(s)) && contains(s, sub)
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
