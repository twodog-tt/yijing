# AI 易经问事 · 本地 MVP

传统文化学习与趣味卦象解读工具。用户输入问题、选择事项类型后自动起卦，展示本卦、变卦、动爻，并提供免费解读与完整 AI 解读（mock / DeepSeek）。另提供「今日运势」入口：每天为同一访客生成一卦，整理当日状态、节奏与行动提醒。

> **仅供娱乐和传统文化参考**，不构成任何预测、医疗、法律或投资建议。

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Next.js + TypeScript + Tailwind CSS |
| 后端 | Go |
| 数据库 | MySQL 8.0 |
| AI | mock / DeepSeek（完整解读，失败自动 fallback） |
| 支付 | mock（看广告/点击解锁） |

## 项目结构

```
yijing/
├── backend/
│   ├── cmd/server/       # API 服务
│   └── cmd/migrate/      # SQL 迁移工具
├── frontend/
├── sql/                  # 按序号排列的迁移脚本
├── docker-compose.yml
├── .env.example
├── .env.docker.example
└── .env.production.example
```

详细安全说明见 [docs/phase8-production-hardening.md](docs/phase8-production-hardening.md)。

## 文档

| 文档 | 说明 |
|------|------|
| [docs/ai-agent-workflow.md](docs/ai-agent-workflow.md) | **ChatGPT / Cursor / Codex 协作边界与 review 流程** |
| [docs/local-acceptance-test.md](docs/local-acceptance-test.md) | 本地 MVP 验收手册 |
| [docs/phase8-production-hardening.md](docs/phase8-production-hardening.md) | 生产安全收口 |
| [docs/phase9-aliyun-deployment-plan.md](docs/phase9-aliyun-deployment-plan.md) | 阿里云部署前方案 |
| [docs/phase10-aliyun-internal-test-deploy.md](docs/phase10-aliyun-internal-test-deploy.md) | ECS 内测部署指南 |
| [docs/miniprogram-dev.md](docs/miniprogram-dev.md) | 微信小程序开发 |

---

## 方式 A：本地开发启动

### 1. 环境准备

- Docker（MySQL）
- Go 1.22+（macOS 建议 Go 1.23+）
- Node.js 20+

```bash
cp .env.example .env
```

### 2. 启动 MySQL 并迁移

```bash
docker compose up -d mysql
cd backend
go run ./cmd/migrate
```

### 3. 启动后端

```bash
unset GOROOT
export PATH="/opt/homebrew/bin:$PATH"
cd backend
go run ./cmd/server
```

默认 `http://localhost:8080`

### 4. 启动前端

```bash
cd frontend
cp .env.local.example .env.local
npm install
npm run dev
```

访问 `http://localhost:3000`

---

## 方式 B：Docker Compose 一键启动

```bash
cp .env.docker.example .env
docker compose up --build
```

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost:3000 |
| 后端 | http://localhost:8080/health |
| MySQL | localhost:3306 |

**Docker 网络说明：**

- 后端容器内连接 MySQL：`mysql:3306`
- 浏览器访问 API：仍用 `http://localhost:8080/api/v1`（不要用 `http://backend:8080`）
- 启动顺序：等待 MySQL → `migrate` → `server`

---

## Migration

```bash
cd backend
go run ./cmd/migrate
```

- 记录表：`schema_migrations`
- 按 `sql/001_*.sql` … `sql/006_*.sql` 顺序执行
- 已执行文件自动跳过

---

## Debug 路由开关

```env
ENABLE_DEBUG_ROUTES=true   # 本地 .env.example 默认
ENABLE_DEBUG_ROUTES=false  # 生产 .env.production.example
```

关闭后 `/api/v1/debug/*` 不注册；前端 `/debug/ai-logs` 显示「调试接口未启用」。

**仅本地开启，生产必须关闭。**

---

## CORS 配置

```env
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

支持逗号分隔多个 origin，不使用 `*`。上线时改为真实前端域名。

---

## Rate Limit

```env
ENABLE_RATE_LIMIT=true
RATE_LIMIT_PER_MINUTE=20
```

限流 `POST /divinations` 与 `POST /divinations/{id}/unlock`，优先按 `session_key`，否则按 IP。超限 `code=42901`。

当前为内存限流；生产多实例建议 Redis。

---

## Mock / DeepSeek / Fallback 测试

**Mock：** `AI_PROVIDER=mock`

**DeepSeek：** 仅后端 `.env` 配置 `DEEPSEEK_API_KEY`（勿入 Git/前端）

**Fallback：** `AI_PROVIDER=deepseek` + 空 Key → 新起卦 → mock 解锁 → 日志 `status=3`

**AI 日志：** `ENABLE_DEBUG_ROUTES=true` 时访问 http://localhost:3000/debug/ai-logs

---

## 今日运势

首页或 `/today` 进入「今日一卦」。同一 `session_key` + 同一本地日期（`YYYY-MM-DD`）每天只生成一条记录；重复进入会打开当天已有结果。

| 项目 | 说明 |
|------|------|
| 前端页面 | `/today` |
| API | `POST /api/v1/daily-fortune/today` |
| 固定问题 | 我今天的整体状态和行动节奏如何？ |
| 事项类型 | `category_id = 6`（今日运势） |
| 结果页 | 复用 `/divination/{id}`（解锁、海报、历史记录同问事流程） |

本地快速验证：

```bash
# 首次生成
curl -s -X POST http://localhost:8080/api/v1/daily-fortune/today \
  -H "Content-Type: application/json" \
  -d '{"session_key":"test-daily-001","local_date":"2026-06-23"}'

# 同一天再次请求，应返回 is_existing=true 且相同 divination id
curl -s -X POST http://localhost:8080/api/v1/daily-fortune/today \
  -H "Content-Type: application/json" \
  -d '{"session_key":"test-daily-001","local_date":"2026-06-23"}'
```

详细验收步骤见 [docs/local-acceptance-test.md](docs/local-acceptance-test.md) 第 11 节。

---

## 生产环境上线前检查清单

- [ ] `ENABLE_DEBUG_ROUTES=false`
- [ ] `CORS_ALLOWED_ORIGINS` 为真实域名
- [ ] API Key 不在 Git / README / 前端
- [ ] 已执行 migrate
- [ ] Rate limit 已启用
- [ ] 前端 `NEXT_PUBLIC_*` 已按环境 build
- [ ] 全流程冒烟测试通过

完整清单见 [docs/phase8-production-hardening.md](docs/phase8-production-hardening.md)。

---

## 当前进度

- [x] Phase 1–6：核心功能、DeepSeek
- [x] Phase 7：Docker 全栈、海报二维码、AI 日志
- [x] Phase 8：安全收口、migration、CORS、限流、配置分环境
- [x] Phase 9：今日运势（`/today`、`daily_fortunes` 映射、解读适配）

## License

Private / MVP — 本地开发用途
