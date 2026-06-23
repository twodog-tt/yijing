# Phase 8：生产化安全收口

本文档说明上线前需要在本地 MVP 中完成的安全与工程化配置。

## 为什么 debug 接口不能生产开放

`/api/v1/debug/*` 与前端 `/debug/ai-logs` 会暴露：

- AI 调用日志（provider、耗时、错误摘要）
- DeepSeek 配置状态（不含 Key，但仍属内部信息）
- 聚合统计

这些信息可帮助攻击者了解系统行为，不应在公网默认暴露。

**做法：** 生产环境设置 `ENABLE_DEBUG_ROUTES=false`（默认值）。仅本地开发 `.env` 中设为 `true`。

## CORS 如何配置

后端通过 `CORS_ALLOWED_ORIGINS` 控制跨域白名单，逗号分隔，**不使用 `*`**。

本地开发：

```env
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

生产示例：

```env
CORS_ALLOWED_ORIGINS=https://your-domain.com,https://www.your-domain.com
```

不在白名单的 `Origin` 将不会收到 `Access-Control-Allow-Origin`，浏览器跨域请求会被拒绝。

## Rate limit 如何配置

```env
ENABLE_RATE_LIMIT=true
RATE_LIMIT_PER_MINUTE=20
```

- 作用范围：`POST /api/v1/divinations`、`POST /api/v1/divinations/{id}/unlock`
- 优先按请求体 `session_key` 限流；无 session_key 时按客户端 IP
- 超限返回 `code=42901`，`message=操作太频繁，请稍后再试`
- 当前为**内存限流**，单进程有效；生产多实例建议换 Redis 分布式限流

### 测试限流

将 `RATE_LIMIT_PER_MINUTE=2`，同一 `session_key` 连续起卦 3 次，第 3 次应收到 42901。

## Migration 如何执行

迁移工具：`backend/cmd/migrate`

```bash
cd backend
go run ./cmd/migrate
```

- 自动创建 `schema_migrations` 表
- 按文件名顺序执行 `sql/001_*.sql` … `sql/005_*.sql`
- 已执行且 checksum 一致的文件会跳过
- 文件内容变更但已执行过会报错，防止静默漂移

环境变量 `SQL_DIR` 可指定 SQL 目录（Docker 内为 `/sql`）。

Docker 启动时 entrypoint 顺序：**等待 MySQL → migrate → server**。

### 已有数据库卷

若 MySQL 卷由 `docker-entrypoint-initdb.d` 初始化过，再次 migrate 通常安全（脚本含 `IF NOT EXISTS` / `ON DUPLICATE KEY UPDATE`）。若遇 checksum 冲突，需评估是否手动标记 migration 或重建本地库。

## 生产环境需要哪些变量

参考根目录 `.env.production.example` 与 `frontend/.env.production.example`。

关键项：

| 变量 | 生产建议 |
|------|----------|
| `APP_ENV` | `production` |
| `ENABLE_DEBUG_ROUTES` | `false` |
| `CORS_ALLOWED_ORIGINS` | 真实前端域名 |
| `DEEPSEEK_API_KEY` | 仅运行时注入，不入 Git |
| `DATABASE_DSN` 或 `DB_*` | 生产数据库连接 |
| `NEXT_PUBLIC_API_BASE_URL` | 公网 API 地址 |
| `NEXT_PUBLIC_SITE_URL` | 公网站点地址（海报二维码） |

## 上线前检查清单

- [ ] `ENABLE_DEBUG_ROUTES=false`
- [ ] `CORS_ALLOWED_ORIGINS` 已改为真实域名，无 `*`
- [ ] `DEEPSEEK_API_KEY` 未出现在 Git、README、前端构建产物
- [ ] 已执行 `go run ./cmd/migrate` 或部署流程包含 migrate
- [ ] `ENABLE_RATE_LIMIT=true`，阈值符合预期
- [ ] 前端 `NEXT_PUBLIC_*` 指向生产域名并已重新 build
- [ ] Docker / 裸机健康检查 `GET /health` 返回 `db: ok`
- [ ] 完整用户流程冒烟：问事 → 起卦 → 免费解读 → mock 解锁 → 海报分享 → 历史
- [ ] DeepSeek 模式（若启用）验证 fallback：`AI_PROVIDER=deepseek` + 无效 Key → 新起卦解锁 → 用户仍有解读，日志 `status=3`

## DeepSeek fallback 验证

1. 后端 `.env`：`AI_PROVIDER=deepseek`，`DEEPSEEK_API_KEY` 留空
2. `ENABLE_DEBUG_ROUTES=true` 时访问 `/debug/ai-logs` 或 `GET /api/v1/debug/ai-stats`
3. 新起卦 → mock 解锁
4. 确认 `fallback_count` 增加，`ai_provider=mock_fallback`

---

Private / MVP — 本地开发与上线前收口
