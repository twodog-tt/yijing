# Phase 6：DeepSeek AI 解读配置与测试

> `DEEPSEEK_API_KEY` 只能配置在后端 `.env`，**不要**写入前端、README 真实值或 Git 提交。

---

## 1. 配置 `.env`

在项目根目录或 `backend/` 运行目录放置 `.env`：

```env
AI_PROVIDER=mock
DEEPSEEK_API_KEY=
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_MODEL=deepseek-v4-flash
DEEPSEEK_TIMEOUT_SECONDS=60
DEEPSEEK_MAX_OUTPUT_TOKENS=1800
```

复制模板：

```bash
cp .env.example .env
```

---

## 2. mock 模式（默认）

```env
AI_PROVIDER=mock
```

行为：

- 免费解读：本地 mock 模板（起卦时生成）
- 完整解读：unlock 时使用 mock 模板
- `interpretation.ai_provider` = `mock`

测试流程与 Phase 3/4 相同，无需 API Key。

```bash
cd backend
AI_PROVIDER=mock go run ./cmd/server
```

---

## 3. deepseek 模式

```env
AI_PROVIDER=deepseek
DEEPSEEK_API_KEY=你的密钥
DEEPSEEK_MODEL=deepseek-v4-flash
```

行为：

- 免费解读：仍为本地 mock（不调用 DeepSeek）
- 完整解读：unlock 时调用 DeepSeek Chat Completions
- 成功：`interpretation.ai_provider` = `deepseek`
- 失败：自动降级 mock，`ai_provider` = `mock_fallback`

```bash
cd backend
go run ./cmd/server
```

启动日志会显示：`ai provider configured: deepseek`

---

## 4. 测试 unlock 后完整解读

```bash
# 1. 创建 session
curl -s -X POST http://localhost:8080/api/v1/sessions \
  -H 'Content-Type: application/json' \
  -d '{"session_key":"deepseek-test-001"}' | jq

# 2. 起卦
curl -s -X POST http://localhost:8080/api/v1/divinations \
  -H 'Content-Type: application/json' \
  -d '{
    "session_key":"deepseek-test-001",
    "category_id":1,
    "question":"我现在适不适合继续推进这个 AI 易经小程序？",
    "confirm_disclaimer":true
  }' | jq

export DIVINATION_ID=1

# 3. mock 解锁
curl -s -X POST http://localhost:8080/api/v1/divinations/$DIVINATION_ID/unlock \
  -H 'Content-Type: application/json' \
  -d '{"session_key":"deepseek-test-001","unlock_type":"mock_ad"}' | jq

# 4. 查询完整解读（含 ai_provider）
curl -s "http://localhost:8080/api/v1/divinations/$DIVINATION_ID/interpretation/full?session_key=deepseek-test-001" | jq
```

---

## 5. 如何判断 ai_provider

| 值 | 含义 |
|----|------|
| `mock` | mock 模板生成 |
| `deepseek` | DeepSeek 成功返回 |
| `mock_fallback` | DeepSeek 失败，已降级 mock |

查看方式：

```bash
# API 响应
curl -s ".../interpretation/full?session_key=..." | jq '.data.ai_provider'

# 数据库
docker compose exec mysql mysql -uyijing -pyijingpass yijing \
  -e "SELECT divination_id, ai_provider, generation_status FROM interpretation ORDER BY id DESC LIMIT 5;"
```

---

## 6. 验证 fallback

将 API Key 留空但设置 `AI_PROVIDER=deepseek`：

```env
AI_PROVIDER=deepseek
DEEPSEEK_API_KEY=
```

unlock 后应：

- 仍返回完整解读 JSON
- `ai_provider` = `mock_fallback`
- 后端日志含 `fallback=mock_fallback`

---

## 7. 常见错误

| 现象 | 原因 | 处理 |
|------|------|------|
| `mock_fallback` | API Key 缺失、超时、非 JSON 响应 | 检查 Key、网络、模型名 |
| 解锁很慢 | DeepSeek 生成耗时 | 调大 `DEEPSEEK_TIMEOUT_SECONDS` |
| 前端展示异常 | `full_content` 非合法 JSON | 后端应已降级；查日志 |
| 重复调用 DeepSeek | 不应发生 | 已有 `full_content` 时跳过生成 |

---

## 8. 单元测试

```bash
cd backend
go test ./...
```

不发起真实 DeepSeek 请求，使用 `httptest` 模拟 API。
