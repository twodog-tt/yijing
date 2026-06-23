# Phase 2 API 测试（curl）

> 前置：MySQL 已启动并完成 seed。后端默认 `http://localhost:8080`。

## 设计说明：session_key 处理

起卦接口 `POST /api/v1/divinations` 采用 **自动创建会话（upsert）**：

- 若 `session_key` 已存在 → 复用该会话
- 若不存在 → 自动写入 `user_sessions`

**理由**：MVP 减少前端/测试往返；`POST /api/v1/sessions` 仍可用于显式初始化与获取 `session_id`。

---

## 1. Health Check

```bash
curl -s http://localhost:8080/health | jq
curl -s http://localhost:8080/api/v1/health | jq
```

---

## 2. 创建 Session

```bash
curl -s -X POST http://localhost:8080/api/v1/sessions \
  -H 'Content-Type: application/json' \
  -d '{"session_key":"demo-session-001"}' | jq
```

不传 `session_key` 时后端自动生成：

```bash
curl -s -X POST http://localhost:8080/api/v1/sessions \
  -H 'Content-Type: application/json' \
  -d '{}' | jq
```

---

## 3. 查询事项类型

```bash
curl -s http://localhost:8080/api/v1/categories | jq
```

---

## 4. 正常起卦

```bash
curl -s -X POST http://localhost:8080/api/v1/divinations \
  -H 'Content-Type: application/json' \
  -d '{
    "session_key": "demo-session-001",
    "category_id": 1,
    "question": "我现在适不适合继续推进这个 AI 易经小程序？",
    "confirm_disclaimer": true
  }' | jq
```

---

## 5. 高风险问题硬拦截

```bash
curl -s -X POST http://localhost:8080/api/v1/divinations \
  -H 'Content-Type: application/json' \
  -d '{
    "session_key": "demo-session-001",
    "category_id": 1,
    "question": "这只股票下周会涨停吗？",
    "confirm_disclaimer": true
  }' | jq
```

期望：`code: 40002`，且不写入 `divination_record`。

---

## 6. 查询起卦详情

将 `{id}` 替换为上一步返回的 `data.id`：

```bash
curl -s http://localhost:8080/api/v1/divinations/1 | jq
```

---

## 7. 查询历史记录

```bash
curl -s 'http://localhost:8080/api/v1/divinations?session_key=demo-session-001&page=1&page_size=20' | jq
```

---

## 已有数据库时补齐 64 卦

若 MySQL 容器在 Phase 1 已初始化，需手动重跑种子：

```bash
docker compose exec -T mysql mysql -uyijing -pyijingpass yijing < sql/002_seed_hexagrams.sql
```
