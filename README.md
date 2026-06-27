# 文易传统文化

微信小程序 + Go 后端项目，以传统文化模型为载体，提供学习参考、趣味解读、自我观察与行动节奏整理。

当前核心模块包括问事起卦、今日一卦、八字简析、奇门问事、统一历史记录、分享卡片 / 长图和首页模块引导。后端完整报告优先使用 DeepSeek，失败时自动走模板 fallback。

> **仅供传统文化学习与自我反思参考**，不构成预测、改命、医疗、法律、投资或其他现实决策建议。

## 技术栈

| 层级 | 技术 |
|------|------|
| 微信小程序 | `miniprogram/` |
| Web/H5 | Next.js + TypeScript + Tailwind CSS，位于 `frontend/` |
| 后端 | Go |
| 数据库 | MySQL 8.0 |
| AI | DeepSeek / mock / template fallback |
| 部署 | Docker Compose、Nginx、阿里云 ECS |

## 当前模块

| 模块 | 状态 |
|------|------|
| 问事起卦 / 今日一卦 | 后端起卦、免费解读、完整解析、历史与长图 |
| 八字简析 | 默认 `bazi-simple-v1`；内部灰度 `bazi-v2-poc` |
| 奇门问事 | 默认 `qimen-simple-v1`；内部灰度 `qimen-v2-poc` / `qimen-v2-professional` |
| 历史记录 | 聚合问事 / 八字 / 奇门，支持筛选与删除 |
| 分享卡片 / 长图 | 本地 Canvas 摘要化生成，不展示隐私字段或完整报告 |
| 首页模块引导 | 三模块入口、场景选择与合规边界 |

普通小程序 / Web 创建流程不传 `algorithm_version`，普通用户不展示算法选择。内部算法只用于内部记录详情页条件展示。

## 项目结构

```
yijing/
├── backend/
│   ├── cmd/server/       # API 服务
│   └── cmd/migrate/      # SQL 迁移工具
├── frontend/
├── miniprogram/           # 微信小程序页面、组件、utils
├── docs/                  # 阶段设计、验收记录与上下文
├── scripts/               # release 前检查脚本
├── deploy/                # Nginx 与 ECS 脚本
├── sql/                  # 按序号排列的迁移脚本
├── docker-compose.yml
├── docker-compose.prod.yml
├── .env.example
├── .env.docker.example
└── .env.production.example
```

详细项目规则见 [AGENTS.md](AGENTS.md)，当前 Codex 上下文见 [docs/CODEX_PROJECT_CONTEXT.md](docs/CODEX_PROJECT_CONTEXT.md)。

## 文档

| 文档 | 说明 |
|------|------|
| [docs/ai-agent-workflow.md](docs/ai-agent-workflow.md) | **ChatGPT / Cursor / Codex 协作边界与 review 流程** |
| [docs/CODEX_PROJECT_CONTEXT.md](docs/CODEX_PROJECT_CONTEXT.md) | Codex 当前项目上下文、阶段状态与阻塞项 |
| [docs/release-checklist.md](docs/release-checklist.md) | 体验版 / release 前可复用检查清单 |
| [scripts/README.md](scripts/README.md) | release 前检查脚本说明 |
| [docs/main-program-module-roadmap.md](docs/main-program-module-roadmap.md) | 主程序模块路线图，说明今日一卦、六爻、八字、奇门、广告解锁、统一报告中心的阶段规划 |
| [docs/bazi-qimen-extension-design.md](docs/bazi-qimen-extension-design.md) | 八字 / 奇门扩展设计与阶段交付 |
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
- 按 `sql/001_*.sql` … `sql/007_*.sql` 顺序执行
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

## DeepSeek / Fallback 测试

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

## Release 回归检查

当前 release 前脚本：

```bash
bash scripts/check-miniprogram-static.sh
bash scripts/check-release-privacy.sh
bash scripts/check-api-smoke.sh
git diff --check
git status --short
```

TEST1.1 后 `check-api-smoke.sh` 当前期望结果为 `15 PASS / 0 FAIL`，覆盖 health、sessions、八字 v1/v2、八字 v2 未知时辰、奇门 v1/poc/professional、非法 algorithm_version、analysis `free_unlock`、问事起卦 create/unlock。

八字未知时辰 API 必须显式传 `birth_hour_unknown=true`；只省略 `birth_hour_branch` 会按参数错误处理。脚本会检查未知时辰报告包含安全说明且不伪造干支时柱。

---

## 体验版 / 生产环境上线前检查清单

- [ ] `ENABLE_DEBUG_ROUTES=false`
- [ ] `CORS_ALLOWED_ORIGINS` 为真实域名
- [ ] API Key 不在 Git / README / 前端
- [ ] 已执行 migrate
- [ ] Rate limit 已启用
- [ ] 前端 `NEXT_PUBLIC_*` 已按环境 build
- [ ] release 回归脚本通过
- [ ] 微信 DevTools / 真机真实 UI 勾选完成
- [ ] ICP 备案完成
- [ ] `https://api.wenyiapp.cn/api/v1/health` 可用
- [ ] 微信 request 合法域名配置完成

当前 dev API 为 `http://123.57.48.214/api/v1`；正式 HTTPS API 域名尚未完成前，不建议上传体验版，不提审。

完整清单见 [docs/release-checklist.md](docs/release-checklist.md)。

---

## 当前进度

- [x] Phase 1–6：核心功能、DeepSeek
- [x] Phase 7：Docker 全栈、海报二维码、AI 日志
- [x] Phase 8：安全收口、migration、CORS、限流、配置分环境
- [x] Phase 9：今日运势（`/today`、`daily_fortunes` 映射、解读适配）
- [x] 八字 / 奇门 analysis API、完整报告、解锁、长图与历史记录
- [x] 首页模块引导、加载 / 错误 / 空状态与防重复提交优化
- [x] 内部算法灰度：`bazi-v2-poc`、`qimen-v2-poc`、`qimen-v2-professional`
- [x] TEST1 / TEST1.1：release 静态检查、隐私检查、API smoke 与未知时辰回归

当前主要阻塞：备案、HTTPS API 域名、微信 request 合法域名、微信 DevTools / 真机 UI 勾选。

## License

Private
