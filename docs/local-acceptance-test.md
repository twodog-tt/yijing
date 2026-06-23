# 本地验收测试手册

本文档用于在**本地环境**完整验收当前 MVP，不涉及云部署、真实支付或微信小程序。

**验收前确认：**

- 已安装 Docker、Go 1.22+（macOS 建议 1.23+）、Node.js 20+
- 未将真实 `DEEPSEEK_API_KEY` 写入 Git、README 或前端
- 本地测试建议开启：`ENABLE_DEBUG_ROUTES=true`

---

## 1. 本地开发模式启动步骤

### 1.1 准备环境变量

```bash
# 项目根目录
cp .env.example .env

# 前端
cd frontend
cp .env.local.example .env.local
cd ..
```

根目录 `.env` 关键项（本地开发）：

```env
AI_PROVIDER=mock
ENABLE_DEBUG_ROUTES=true
CORS_ALLOWED_ORIGINS=http://localhost:3000
DB_HOST=127.0.0.1
```

`frontend/.env.local`：

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_SITE_URL=http://localhost:3000
```

### 1.2 启动 MySQL

```bash
docker compose up -d mysql
```

验证：

```bash
docker compose ps mysql
# 状态应为 healthy
```

### 1.3 执行数据库迁移

```bash
cd backend
go run ./cmd/migrate
```

期望输出包含 `skip` 或 `done`，最后一行 `migration completed`。

### 1.4 启动后端

macOS 若遇 Go 版本问题：

```bash
unset GOROOT
export PATH="/opt/homebrew/bin:$PATH"
go version   # 应 >= 1.23
```

启动：

```bash
cd backend
go run ./cmd/server
```

期望日志包含：

- `mysql connected`
- `debug routes enabled: true`（若 `.env` 已开启）
- `backend listening on http://localhost:8080`

健康检查：

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/health
```

期望：`"status":"ok"`，`"db":"ok"`。

### 1.5 启动前端

新开终端：

```bash
cd frontend
npm install   # 首次
npm run dev
```

访问：http://localhost:3000

### 1.6 开发模式验收通过标准

| 检查项 | 期望 |
|--------|------|
| 首页可打开 | 显示免责声明 |
| `/health` | `db: ok` |
| 浏览器控制台无 CORS 错误 | 正常请求 API |
| `go test ./...` | 全部通过（可选预检） |

---

## 2. Docker Compose 模式启动步骤

### 2.1 准备

```bash
cp .env.docker.example .env
```

确认 `.env` 中：

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_SITE_URL=http://localhost:3000
ENABLE_DEBUG_ROUTES=true
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

> **注意：** 浏览器访问 API 必须用 `localhost:8080`，不要用 `http://backend:8080`（`backend` 仅容器内 DNS）。

### 2.2 一键启动

```bash
docker compose up --build
```

或后台运行：

```bash
docker compose up --build -d
docker compose logs -f backend
```

### 2.3 启动顺序验证

`backend` 日志应依次出现：

1. `Waiting for MySQL...`
2. `MySQL is available`
3. `Running database migrations...`（`skip` / `done` 若干行）
4. `Starting server...`
5. `backend listening on http://localhost:8080`

### 2.4 Docker 模式验收通过标准

| 服务 | 地址 | 期望 |
|------|------|------|
| 前端 | http://localhost:3000 | 首页正常 |
| 后端 | http://localhost:8080/health | `db: ok` |
| MySQL | localhost:3306 | 容器 healthy |

```bash
docker compose ps
# mysql、backend、frontend 均 Up
```

### 2.5 停止与清理

```bash
docker compose down        # 保留数据卷
docker compose down -v     # 清空 MySQL 数据（重建库时用）
```

---

## 3. Mock 模式测试流程

**配置：** 根目录 `.env` 设置 `AI_PROVIDER=mock`，重启后端。

### 3.1 API 预检（可选）

```bash
curl http://localhost:8080/api/v1/debug/ai-health
```

期望：`"provider":"mock"`

### 3.2 前端完整流程

1. 打开 http://localhost:3000
2. 进入「问事」页 `/ask`
3. 选择事项类型（如「事业」）
4. 输入问题（5–200 字），勾选免责声明
5. 提交 → 出现起卦动画 → 跳转结果页 `/divination/{id}`
6. 确认展示：本卦、变卦、六爻、动爻、免费解读
7. 点击解锁 → 选择 mock 方式（看广告 / 按钮）→ 解锁成功
8. 完整解读出现（本地 mock 模板内容）
9. 点击「生成分享海报」→ 预览与下载
10. 进入「历史记录」`/history` → 可见刚起的卦

### 3.3 Mock 模式通过标准

| 检查项 | 期望 |
|--------|------|
| `ai_provider`（完整解读） | `mock` |
| AI 日志 `status` | `1`（成功） |
| 用户无感知错误 | 全流程顺畅 |

---

## 4. DeepSeek 模式测试流程

**仅在后端 `.env` 配置，切勿写入前端或 Git。**

```env
AI_PROVIDER=deepseek
DEEPSEEK_API_KEY=<你的密钥>
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_MODEL=deepseek-v4-flash
DEEPSEEK_TIMEOUT_SECONDS=60
```

重启后端。

### 4.1 配置检查（不发起真实生成）

```bash
curl http://localhost:8080/api/v1/debug/ai-health
```

期望：

- `"provider":"deepseek"`
- `"api_key_configured":true`
- 响应中**不包含** API Key 明文

### 4.2 前端测试

1. **新起一卦**（不要用已解锁的旧记录，旧记录有 `full_content` 不会重新生成）
2. 完成 mock 解锁
3. 等待完整解读加载（可能需数十秒）
4. 完整解读应为结构化 JSON 字段（summary、overall、action_steps 等）

### 4.3 DeepSeek 模式通过标准

| 检查项 | 期望 |
|--------|------|
| 完整解读 `ai_provider` | `deepseek` |
| AI 日志 | `status=1`，`ai_provider=deepseek`，`duration_ms > 0` |
| 免费解读 | 仍为本地 mock（不走 DeepSeek） |

### 4.4 测完恢复

建议测完后将 `.env` 改回 `AI_PROVIDER=mock`，避免日常开发误耗额度。

---

## 5. Fallback 模式测试流程

验证 DeepSeek 失败时自动降级为 mock，用户仍能拿到完整解读。

### 5.1 配置

```env
AI_PROVIDER=deepseek
DEEPSEEK_API_KEY=          # 留空，或故意填无效值
ENABLE_DEBUG_ROUTES=true
```

重启后端。

### 5.2 测试步骤

1. `curl http://localhost:8080/api/v1/debug/ai-health` → `api_key_configured: false`
2. **新起一卦** → mock 解锁
3. 用户仍应看到完整解读内容
4. 打开 http://localhost:3000/debug/ai-logs

### 5.3 Fallback 通过标准

| 检查项 | 期望 |
|--------|------|
| 完整解读可用 | 有内容，无 500 错误 |
| `interpretation.ai_provider` | `mock_fallback` |
| AI 日志 `status` | `3`（fallback 成功） |
| `fallback_used` | `1` |
| `error_message` | 有摘要（如 api key missing） |
| 统计卡片 `fallback_count` | 增加 |

> **注意：** 已对同一条记录解锁过的，不会重新调用 AI；fallback 测试必须**新起卦**。

---

## 6. 核心用户流程测试清单

按顺序勾选：

| # | 步骤 | 操作 | 期望结果 | 通过 |
|---|------|------|----------|------|
| 1 | 首页 | 访问 `/` | 有免责声明、可进入问事 | ☐ |
| 2 | 关于页 | 访问 `/about` | 合规说明正常显示 | ☐ |
| 3 | 会话 | 首次进入 `/ask` | 自动创建 session（localStorage 有 `yijing_session_key`） | ☐ |
| 4 | 事项类型 | `/ask` 加载 | 显示 5 类：事业/关系/学习/选择/近期状态 | ☐ |
| 5 | 表单校验 | 不填问题提交 | 前端提示，不发起请求 | ☐ |
| 6 | 起卦 | 合法问题提交 | 动画 → 跳转结果页 | ☐ |
| 7 | 结果页 | `/divination/{id}` | 本卦、变卦、爻象、免费解读 | ☐ |
| 8 | 免费解读 | 结果页 | 有文字内容 | ☐ |
| 9 | Mock 解锁 | 点击解锁 | 弹窗 → 解锁成功 | ☐ |
| 10 | 完整解读 | 解锁后 | 结构化完整报告展示 | ☐ |
| 11 | 分享海报 | 生成海报 | 预览、下载 PNG | ☐ |
| 12 | 历史记录 | `/history` | 列表含刚起的卦，可点进详情 | ☐ |
| 13 | 再次进入 | 刷新结果页 | 已解锁状态保持，完整解读仍在 | ☐ |

---

## 11. 今日运势测试流程

前置：已执行 `go run ./cmd/migrate`（含 `sql/006_daily_fortune.sql`），前后端均已启动。

按顺序勾选：

| # | 步骤 | 操作 | 期望结果 | 通过 |
|---|------|------|----------|------|
| 1 | 首页入口 | 访问 `/` | 可见「今日运势」或「查看今日一卦」入口 | ☐ |
| 2 | 进入页面 | 点击入口 | 跳转 `/today`，展示今日一卦说明文案 | ☐ |
| 3 | 首次起卦 | 点击「查看今日一卦」 | 调用 API 成功，跳转 `/divination/{id}` | ☐ |
| 4 | 结果页 | 查看结果 | 本卦/变卦/免费解读正常；解读偏今日节奏与行动提醒，非事业/财运预测 | ☐ |
| 5 | 重复进入 | 返回 `/today` 再次点击 | 提示「你今天已经起过一卦…」，打开**同一天**已有结果（`is_existing=true`） | ☐ |
| 6 | Mock 解锁 | 结果页解锁 | mock 解锁成功，完整解读 JSON 结构正常 | ☐ |
| 7 | 分享海报 | 生成海报 | 标题为「今日一卦」，问题摘要为固定今日主题 | ☐ |
| 8 | 历史记录 | 访问 `/history` | 列表显示 `今日运势｜YYYY-MM-DD`，可进入详情 | ☐ |
| 9 | DeepSeek（可选） | `AI_PROVIDER=deepseek` 且配置 Key 后新起卦 | 完整解读生成成功，不含具体事件预测表述 | ☐ |

### 11.1 API 重复起卦验证

```bash
SESSION="acceptance-daily-$(date +%s)"
DATE="2026-06-23"

# 第一次：is_existing 应为 false
curl -s -X POST http://localhost:8080/api/v1/daily-fortune/today \
  -H "Content-Type: application/json" \
  -d "{\"session_key\":\"$SESSION\",\"local_date\":\"$DATE\"}" | jq '.data.daily_fortune, .data.divination.id'

# 第二次：is_existing 应为 true，divination.id 与第一次相同
curl -s -X POST http://localhost:8080/api/v1/daily-fortune/today \
  -H "Content-Type: application/json" \
  -d "{\"session_key\":\"$SESSION\",\"local_date\":\"$DATE\"}" | jq '.data.daily_fortune, .data.divination.id'
```

### 11.2 跨日验证（可选）

将 `local_date` 改为次日（如 `2026-06-24`），同一 `session_key` 应生成新的 `divination.id`。

### 11.3 验收通过标准

- 同一天同 session 不重复起卦
- 结果页、解锁、海报、历史记录与普通问事流程兼容
- 普通 `/ask` 问事流程不受影响

---

## 7. 高风险问题拦截测试用例

敏感词为**硬拦截**，返回 `code=40002`，不允许起卦。

### 7.1 前端测试（推荐）

在 `/ask` 分别输入包含以下关键词的问题（其余字段合法），提交后应看到拦截提示，**不跳转**结果页：

| 类别 | 示例问题（含关键词） | 期望 |
|------|----------------------|------|
| 医疗 | 我得了癌症要不要手术？ | 拦截，code 40002 |
| 寿命 | 我还能活多久？ | 拦截 |
| 自伤 | 我想自杀怎么办 | 拦截 |
| 投资 | 这只股票会涨停吗 | 拦截 |
| 赌博 | 今天彩票能中吗 | 拦截 |
| 法律 | 这场官司能赢吗 | 拦截 |
| 违法 | 这样做是否违法 | 拦截 |
| 威胁 | 我想报复某人 | 拦截 |

前端提示文案类似：

> 这个问题不适合用卦象方式解读。你可以换成更偏向自我反思、情绪整理或行动选择的问题。

### 7.2 API 测试（可选）

```bash
curl -s -X POST http://localhost:8080/api/v1/divinations \
  -H "Content-Type: application/json" \
  -d '{
    "session_key": "test-session-001",
    "category_id": 1,
    "question": "这只股票会涨停吗",
    "confirm_disclaimer": true
  }'
```

期望：`"code":40002`

### 7.3 边界用例

| 用例 | 期望 |
|------|------|
| 问题少于 5 字 | 前端校验失败，`code=40001` |
| 问题超过 200 字 | 前端校验失败 |
| 正常问题「近期工作是否适合推进」 | 通过，成功起卦 |
| 未勾选免责声明 | 前端拦截 |

---

## 8. 分享海报测试用例

前置：已完成起卦，建议已解锁（海报可展示更完整摘要）。

| # | 测试项 | 操作 | 期望 | 通过 |
|---|--------|------|------|------|
| 1 | 打开发海报 | 结果页点击「生成分享海报」 | 弹出预览模态框 | ☐ |
| 2 | 海报内容 | 查看预览 | 含问题摘要、本卦/变卦、动爻、一句话提示、免责声明 | ☐ |
| 3 | 二维码 | 查看右下角 | 显示真实二维码（非「占位」） | ☐ |
| 4 | 二维码链接 | 手机扫码 | 打开 `http://localhost:3000/divination/{id}` | ☐ |
| 5 | 下载 | 点击「下载海报」 | 下载 PNG 文件，内容与预览一致 | ☐ |
| 6 | 环境变量 | 确认 `NEXT_PUBLIC_SITE_URL` | 与二维码域名一致 | ☐ |

### 二维码 URL 规则

```
${NEXT_PUBLIC_SITE_URL}/divination/${id}
```

默认：`http://localhost:3000/divination/{id}`

### 异常场景

| 场景 | 期望 |
|------|------|
| 二维码生成失败 | 显示占位区域，**不影响**海报下载 |
| 未解锁 | 仍可生成海报（使用免费解读摘要） |

---

## 9. 历史记录分页测试用例

历史页：`/history`，每页 `page_size=20`。

| # | 测试项 | 操作 | 期望 | 通过 |
|---|--------|------|------|------|
| 1 | 空列表 | 新 session 首次访问 | 提示无记录或空状态 | ☐ |
| 2 | 单条记录 | 起卦 1 次后查看 | 列表显示 1 条，信息含问题摘要、卦名、时间 | ☐ |
| 3 | 点击跳转 | 点击某条记录 | 进入对应 `/divination/{id}` | ☐ |
| 4 | 会话隔离 | 换浏览器 / 清 localStorage | 历史为空（不同 session_key） | ☐ |
| 5 | 加载更多 | 记录 > 20 条时点「加载更多」 | 追加下一页，不重复 | ☐ |
| 6 | 无更多 | 全部加载完 | 「加载更多」消失或不可用 | ☐ |

### 快速造数（可选）

同一 `session_key` 连续起卦多次（注意 rate limit，默认 20 次/分钟），或在 `.env` 临时调大 `RATE_LIMIT_PER_MINUTE`。

### API 验证（可选）

```bash
curl "http://localhost:8080/api/v1/divinations?session_key=<你的session_key>&page=1&page_size=20"
```

期望：`code=0`，`data.items` 数组，`data.total` 为总数。

---

## 10. AI 日志测试用例

前置：`ENABLE_DEBUG_ROUTES=true`，后端已重启。

### 10.1 页面测试

访问：http://localhost:3000/debug/ai-logs

| # | 测试项 | 期望 | 通过 |
|---|--------|------|------|
| 1 | 页面顶部提示 | 显示「仅本地调试使用」 | ☐ |
| 2 | 统计卡片 | 显示 total / success / fail / fallback / 平均耗时 | ☐ |
| 3 | 日志列表 | mock 解锁后有一条记录 | ☐ |
| 4 | 状态标签 | 成功=绿，失败=红，fallback=黄 | ☐ |
| 5 | 字段完整 | provider、model、duration、created_at | ☐ |
| 6 | 跳转 | 点击 divination_id 可进结果页 | ☐ |
| 7 | 分页 | 超过 20 条可「加载更多」 | ☐ |

### 10.2 API 测试

```bash
# 日志列表
curl "http://localhost:8080/api/v1/debug/ai-logs?page=1&page_size=20"

# 统计
curl http://localhost:8080/api/v1/debug/ai-stats

# 健康检查
curl http://localhost:8080/api/v1/debug/ai-health
```

### 10.3 Debug 关闭测试

`.env` 设置 `ENABLE_DEBUG_ROUTES=false`，重启后端：

| 测试项 | 期望 |
|--------|------|
| `curl .../debug/ai-logs` | HTTP 404 |
| 前端 `/debug/ai-logs` | 显示「调试接口未启用，仅本地开发环境可用。」 |

测完后改回 `ENABLE_DEBUG_ROUTES=true`。

---

## 11. 常见问题排查

### 11.1 前端报「网络连接失败」

| 可能原因 | 处理 |
|----------|------|
| 后端未启动 | `go run ./cmd/server` 或 `docker compose up backend` |
| API 地址错误 | 检查 `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1` |
| Docker 用了 `backend:8080` | 浏览器必须用 `localhost:8080` |

### 11.2 CORS 错误

| 可能原因 | 处理 |
|----------|------|
| origin 不在白名单 | `.env` 设 `CORS_ALLOWED_ORIGINS=http://localhost:3000` |
| 改了 CORS 未重启后端 | 重启 server |

### 11.3 `db: down` 或连接失败

| 可能原因 | 处理 |
|----------|------|
| MySQL 未启动 | `docker compose up -d mysql` |
| 本地模式 DB_HOST 错误 | 应为 `127.0.0.1`；Docker 后端应为 `mysql` |
| 未 migrate | `cd backend && go run ./cmd/migrate` |

### 11.4 Go 编译/运行错误 `dyld: missing LC_UUID`

```bash
unset GOROOT
export PATH="/opt/homebrew/bin:$PATH"
go version
```

需 Go 1.23+（macOS 15+）。

### 11.5 migrate checksum 冲突

数据库卷与当前 `sql/` 文件不一致。本地可重建：

```bash
docker compose down -v
docker compose up -d mysql
cd backend && go run ./cmd/migrate
```

### 11.6 操作太频繁（42901）

`RATE_LIMIT_PER_MINUTE` 超限。等待 1 分钟，或临时调大限制。

### 11.7 DeepSeek 无响应 / 超时

- 检查 `DEEPSEEK_API_KEY` 是否有效
- 检查网络能否访问 `api.deepseek.com`
- 查看 AI 日志 `error_message`
- 失败应自动 fallback，用户仍有 mock 完整解读

### 11.8 完整解读不更新

同一条 `divination_id` 已有 `full_content` 会跳过重新生成。**新起卦**再测。

### 11.9 Docker daemon 未运行

```
Cannot connect to the Docker daemon
```

启动 Docker Desktop 后重试。

### 11.10 海报二维码是占位

- 确认 `NEXT_PUBLIC_SITE_URL` 已配置
- 刷新页面重试；二维码失败时占位属正常降级

---

## 12. Bug 记录模板

验收时发现问题，请按以下模板记录：

```markdown
### BUG-YYYYMMDD-001

- **发现时间：** YYYY-MM-DD HH:mm
- **测试人：**
- **环境：**
  - [ ] 本地开发模式
  - [ ] Docker Compose 模式
- **相关配置：**
  - AI_PROVIDER: mock / deepseek
  - ENABLE_DEBUG_ROUTES: true / false
- **复现步骤：**
  1.
  2.
  3.
- **期望结果：**
- **实际结果：**
- **接口/页面：**
  - URL 或 API：
  - 响应 code / HTTP 状态：
- **截图/日志：**（如有）
- **严重程度：**
  - [ ] P0 阻断（无法起卦/解锁/访问）
  - [ ] P1 核心功能异常
  - [ ] P2 体验问题
  - [ ] P3 文案/样式
- **是否可稳定复现：** 是 / 否
- **备注：**
```

---

## 附录：建议验收顺序（速查）

```
环境启动 → 健康检查 → Mock 全流程 → 今日运势 → 敏感词拦截 → 历史分页
    → 分享海报 → AI 日志 → Debug 开关 → Fallback → DeepSeek（可选）
    → Docker 全栈复测（可选）
```

---

*文档版本：Phase 9 本地 MVP 验收（含今日运势）· 不含云部署与真实支付*
