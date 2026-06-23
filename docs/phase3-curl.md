# Phase 3 API 测试（curl）

> 前置：MySQL 已启动、已执行 seed 与 `sql/004_phase3_interpretation_unlock.sql`（可选）。后端默认 `http://localhost:8080`。

---

## 1. 创建 Session

```bash
curl -s -X POST http://localhost:8080/api/v1/sessions \
  -H 'Content-Type: application/json' \
  -d '{"session_key":"phase3-demo-001"}' | jq
```

---

## 2. 正常起卦（含 free_interpretation）

```bash
curl -s -X POST http://localhost:8080/api/v1/divinations \
  -H 'Content-Type: application/json' \
  -d '{
    "session_key": "phase3-demo-001",
    "category_id": 1,
    "question": "我现在适不适合继续推进这个 AI 易经小程序？",
    "confirm_disclaimer": true
  }' | jq
```

期望：`code=0`，`data.free_interpretation` 非空，`data.unlock_status=0`。

记下返回的 `data.id` 作为 `DIVINATION_ID`。

```bash
export DIVINATION_ID=1
```

---

## 3. 查询免费解读

```bash
curl -s http://localhost:8080/api/v1/divinations/$DIVINATION_ID/interpretation/free | jq
```

期望：`generation_status=1`，`free_content` 与起卦响应一致。

---

## 4. 未解锁查询完整解读（40301）

```bash
curl -s "http://localhost:8080/api/v1/divinations/$DIVINATION_ID/interpretation/full?session_key=phase3-demo-001" | jq
```

期望：`code=40301`。

---

## 5. Mock 解锁

```bash
curl -s -X POST http://localhost:8080/api/v1/divinations/$DIVINATION_ID/unlock \
  -H 'Content-Type: application/json' \
  -d '{
    "session_key": "phase3-demo-001",
    "unlock_type": "mock_ad"
  }' | jq
```

期望：`unlock_status=1`，`mock_transaction_id` 以 `MOCK-` 开头，`full_interpretation` 为结构化 JSON。

---

## 6. 解锁后查询完整解读

```bash
curl -s "http://localhost:8080/api/v1/divinations/$DIVINATION_ID/interpretation/full?session_key=phase3-demo-001" | jq
```

期望：`code=0`，`data.full_content` 含 `summary`、`action_steps`、`disclaimer` 等字段。

---

## 7. 查询起卦详情（unlock_status=1）

```bash
curl -s http://localhost:8080/api/v1/divinations/$DIVINATION_ID | jq
```

期望：`data.unlock_status=1`。

---

## 8. session_key 不匹配不能解锁

```bash
curl -s -X POST http://localhost:8080/api/v1/divinations/$DIVINATION_ID/unlock \
  -H 'Content-Type: application/json' \
  -d '{
    "session_key": "other-user-session",
    "unlock_type": "mock_button"
  }' | jq
```

期望：`code=40301`。

---

## 9. mock_button 解锁类型

对新起卦记录可测试：

```bash
curl -s -X POST http://localhost:8080/api/v1/divinations/$DIVINATION_ID/unlock \
  -H 'Content-Type: application/json' \
  -d '{
    "session_key": "phase3-demo-001",
    "unlock_type": "mock_button"
  }' | jq
```

已解锁记录再次解锁将直接返回已有完整解读（幂等）。
