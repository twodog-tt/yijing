# Phase 9：阿里云部署前方案整理

> **文档性质：** 部署前规划与购买前检查清单  
> **不包含：** 真实密钥、实际购买操作、线上部署执行  
> **适用阶段：** MVP 本地验收已通过，尚未购买阿里云资源、尚未准备域名与 ICP 备案

---

## 1. 当前 MVP 功能清单

### 1.1 核心问事流程

| 功能 | 说明 | 路由 / 接口 |
|------|------|-------------|
| 首页 | 免责声明、立即起卦、今日运势入口 | `/` |
| 问事起卦 | 选择事项类型、输入问题、三枚硬币起卦 | `/ask`、`POST /api/v1/divinations` |
| 起卦动画 | API 返回后基于真实 `lines` 播放复盘动画 | 前端 `CastingOverlay` |
| 结果页 | 本卦、变卦、六爻、免费解读 | `/divination/{id}` |
| Mock 解锁 | 模拟看广告/点击解锁完整解读 | `POST /api/v1/divinations/{id}/unlock` |
| DeepSeek 完整解读 | 后端调用 DeepSeek，失败自动 fallback mock | `AI_PROVIDER=deepseek` |
| 分享海报 | PNG 下载，二维码指向结果页 | 结果页内生成 |
| 历史记录 | 分页列表、进入详情 | `/history` |
| 关于与合规 | 免责声明、起卦说明 | `/about` |

### 1.2 今日运势（Phase 9 业务）

| 功能 | 说明 |
|------|------|
| 今日一卦入口 | 首页、`/today` |
| 每日一次 | 同 `session_key` + 本地日期只生成一条 |
| 固定问题与类型 | `category_id=6`，复用结果页与解锁流程 |
| 已有结果 | `is_existing=true` 时跳过完整起卦动画 |

### 1.3 工程与安全能力（Phase 7–8）

| 功能 | 说明 |
|------|------|
| Docker Compose 全栈 | MySQL + backend + frontend |
| SQL Migration | `backend/cmd/migrate`，`sql/001`–`006` |
| AI 调用日志 | `ai_generation_logs` 表 |
| Debug 路由开关 | `ENABLE_DEBUG_ROUTES`（生产必须 `false`） |
| CORS 白名单 | `CORS_ALLOWED_ORIGINS` |
| 内存限流 | 起卦 / 解锁接口，`RATE_LIMIT_PER_MINUTE` |
| 敏感词硬拦截 | `code=40002` |
| 上海时区 | 后端 `Asia/Shanghai`，API 时间 `+08:00` |

### 1.4 当前明确不做

- 真实支付
- 真实广告
- 微信小程序
- 云部署（本阶段仅方案）

---

## 2. 当前技术栈

| 层级 | 技术 | 版本 / 说明 |
|------|------|-------------|
| 前端 | Next.js + TypeScript + Tailwind CSS | Node 20，standalone 构建 |
| 后端 | Go | `cmd/server`、`cmd/migrate` |
| 数据库 | MySQL | 8.0，utf8mb4 |
| AI | Mock / DeepSeek | 仅完整解读；Key 仅后端 |
| 容器化 | Docker Compose | 三服务：mysql、backend、frontend |
| 反向代理 | 本地未配置 | 生产建议 Nginx |
| 时区 | Asia/Shanghai | UTC+8 |

### 2.1 仓库关键路径

```text
yijing/
├── frontend/          # Next.js
├── backend/           # Go API
├── sql/               # 001–006 迁移脚本
├── docker-compose.yml
├── .env.production.example
└── docs/
```

### 2.2 容器端口（默认）

| 服务 | 容器端口 | 宿主机默认 |
|------|----------|------------|
| frontend | 3000 | 3000 |
| backend | 8080 | 8080 |
| mysql | 3306 | 3306 |

---

## 3. 推荐部署架构

### 3.1 总体示意（稍稳方案）

```text
                    Internet
                        │
                   [ 443 / 80 ]
                        │
                   ┌────▼────┐
                   │  Nginx  │  SSL 终止、反代、静态缓存（可选）
                   └──┬──┬───┘
          www.xxx.com  │  │  api.xxx.com
                       │  │
              ┌────────▼  └────────┐
              │       ECS          │
              │  ┌──────────────┐  │
              │  │ frontend:3000│  │
              │  │ backend:8080 │  │  ← 仅内网 / 127.0.0.1 暴露
              │  └──────────────┘  │
              └─────────┬──────────┘
                        │ 内网 / VPC
              ┌─────────▼──────────┐
              │  RDS MySQL 8.0     │  ← 不开放公网
              └────────────────────┘
                        │
              ┌─────────▼──────────┐
              │  DeepSeek API      │  公网 HTTPS 出站
              └────────────────────┘
```

### 3.2 前端部署方式

- **构建：** `npm run build`（Docker 多阶段构建，standalone 模式）
- **运行：** Node 进程 `node server.js`，监听 `3000`
- **生产要点：**
  - `NEXT_PUBLIC_API_BASE_URL`、`NEXT_PUBLIC_SITE_URL` 在 **build 时** 写入（Docker `ARG`）
  - 域名变更需 **重新 build** 前端镜像
  - 由 Nginx 将 `www.xxx.com` 反代到 `127.0.0.1:3000`

### 3.3 后端部署方式

- **构建：** Go 静态编译 `server` + `migrate`
- **启动顺序：** 等待 MySQL → `./migrate` → `./server`
- **生产要点：**
  - 监听 `8080`，建议 **不对公网直接暴露**
  - 由 Nginx 将 `api.xxx.com` 反代到 `127.0.0.1:8080`
  - `ENABLE_DEBUG_ROUTES=false`
  - `CORS_ALLOWED_ORIGINS` 设为前端域名

### 3.4 MySQL 部署方式

**方案 A（最小）：** Docker Compose 内 MySQL 容器 + 数据卷  
**方案 B（稍稳）：** 阿里云 RDS MySQL 8.0，ECS 通过内网连接

共同要求：

- 字符集 `utf8mb4`
- 执行 `sql/001`–`006`（优先用 `migrate` 工具，避免重复执行冲突）
- **禁止** 3306 对公网开放

### 3.5 Nginx / HTTPS / 域名

- 一台 ECS 上安装 Nginx，统一入口 80 / 443
- 80 跳转 443
- 免费或付费 SSL 证书绑定域名
- `www.xxx.com` → frontend；`api.xxx.com` → backend
- 海报二维码 `NEXT_PUBLIC_SITE_URL` 必须与 `www` 域名一致

---

## 4. 阿里云资源清单

| 资源 | 用途 | 是否必须 | 备注 |
|------|------|----------|------|
| **ECS** | 跑 Nginx、Docker（frontend + backend） | 是 | 内测 2C4G 起 |
| **安全组** | 控制入站端口 | 是 | 见第 12 节 |
| **云盘（系统盘 + 数据盘）** | 系统、Docker 镜像、日志 | 是 | 最小方案 MySQL 数据也在云盘卷 |
| **域名** | 用户访问、API、海报二维码 | 公网正式访问时必须 | 需实名 |
| **SSL 证书** | HTTPS | 公网正式访问时必须 | 可与域名一起申请免费 DV 证书 |
| **RDS MySQL** | 托管数据库 | 可选（稍稳方案推荐） | 自动备份、内网访问 |
| **OSS** | 海报/静态资源 CDN | 可选 | 当前海报前端生成，非必须 |
| **日志服务 SLS** | 集中日志 | 可选 | 初期可用 `docker logs` + 文件轮转 |
| **VPC / 交换机** | 内网隔离 | RDS 场景推荐 | ECS 与 RDS 同 VPC |
| **弹性公网 IP** | 固定入口 IP | 推荐 | 绑定 ECS |
| **备案（ICP）** | 大陆服务器 + 域名 Web 服务 | 使用大陆节点公网 Web 时必须 | 见第 13 节 |

---

## 5. 最小部署方案

**适用：** 个人内测、演示、极低流量验证成本优先。

```text
单台 ECS（2 vCPU / 4 GiB / 40–60 GiB SSD）
    └── Docker Compose
            ├── mysql:8.0      （仅绑定 127.0.0.1:3306 或不映射宿主机端口）
            ├── backend:8080   （127.0.0.1:8080）
            └── frontend:3000  （127.0.0.1:3000）
    └── Nginx（宿主机）
            ├── 443 → 127.0.0.1:3000
            └── api 子域 → 127.0.0.1:8080
```

**优点：** 成本低、与本地 `docker-compose.yml` 最接近、上手快。  
**缺点：** MySQL 与业务抢资源；单点故障；备份需自建；性能余量小。

**内测注意：**

- 修改 `docker-compose.yml` 生产覆盖：MySQL **不要** `ports: 3306:3306` 映射到 `0.0.0.0`
- 使用 `.env.production`（不提交 Git）注入真实密码与 DeepSeek Key

---

## 6. 稍稳部署方案

**适用：** 小范围正式内测、希望数据与计算分离。

```text
ECS（2–4 vCPU / 4–8 GiB）
    ├── Nginx（80/443）
    ├── Docker: frontend + backend
    └── 不跑 MySQL 容器

RDS MySQL 8.0（基础版 / 1C2G 起，按量或包月）
    └── 仅 VPC 内网地址，白名单放 ECS 安全组

DeepSeek
    └── 后端出站访问 api.deepseek.com
```

**优点：** 数据库自动备份、升级方便、3306 不经过 ECS。  
**缺点：** 月费高于最小方案；需配置 VPC、RDS 白名单、连接串。

---

## 7. 域名规划

假设主品牌域名为 `xxx.com`（示例，非真实购买）：

| 域名 | 用途 | 指向 |
|------|------|------|
| `www.xxx.com` | 用户访问 Web / H5 | Nginx → frontend:3000 |
| `xxx.com` | 根域 | 301 跳转到 `www.xxx.com`（推荐） |
| `api.xxx.com` | 后端 API | Nginx → backend:8080 |

**前端环境变量对应：**

```env
NEXT_PUBLIC_SITE_URL=https://www.xxx.com
NEXT_PUBLIC_API_BASE_URL=https://api.xxx.com/api/v1
```

**后端 CORS：**

```env
CORS_ALLOWED_ORIGINS=https://www.xxx.com,https://xxx.com
```

**海报二维码：** 扫码打开 `https://www.xxx.com/divination/{id}`

---

## 8. HTTPS 和 Nginx 反代规划

### 8.1 证书

- 阿里云免费 DV 证书（单域名或通配符按预算选择）
- `www.xxx.com` 与 `api.xxx.com` 各需证书，或使用多域名 / 通配符证书
- 证书部署在 Nginx，容器内 frontend/backend **可不配 TLS**

### 8.2 Nginx 配置要点（示意，非最终文件）

```nginx
# www.xxx.com → frontend
server {
    listen 443 ssl http2;
    server_name www.xxx.com;

    ssl_certificate     /etc/nginx/ssl/www.fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/www.key;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# api.xxx.com → backend
server {
    listen 443 ssl http2;
    server_name api.xxx.com;

    ssl_certificate     /etc/nginx/ssl/api.fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/api.key;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 8.3 HTTP 跳转

- `80` 全站跳转 `301` → `https`

### 8.4 健康检查

- 内网：`curl http://127.0.0.1:8080/health`
- 公网：`curl https://api.xxx.com/health`（部署后验证）

---

## 9. 环境变量清单

### 9.1 后端生产环境变量

参考 `.env.production.example`：

| 变量 | 必填 | 生产建议值 / 说明 |
|------|------|-------------------|
| `APP_ENV` | 是 | `production` |
| `BACKEND_PORT` | 是 | `8080` |
| `DATABASE_DSN` 或 `DB_*` | 是 | RDS 内网地址；强密码 |
| `SQL_DIR` | 容器内 | `/sql` |
| `AI_PROVIDER` | 是 | `deepseek` 或内测期 `mock` |
| `DEEPSEEK_API_KEY` | DeepSeek 时 | **仅服务器环境变量 / 密钥管理，不入 Git** |
| `DEEPSEEK_BASE_URL` | 否 | `https://api.deepseek.com` |
| `DEEPSEEK_MODEL` | 否 | `deepseek-v4-flash` |
| `DEEPSEEK_TIMEOUT_SECONDS` | 否 | `60` |
| `DEEPSEEK_MAX_OUTPUT_TOKENS` | 否 | `1800` |
| `ENABLE_DEBUG_ROUTES` | 是 | **`false`** |
| `CORS_ALLOWED_ORIGINS` | 是 | 前端 HTTPS 域名，逗号分隔 |
| `ENABLE_RATE_LIMIT` | 是 | `true` |
| `RATE_LIMIT_PER_MINUTE` | 否 | `20`（可按流量调整） |

Docker Compose 额外项（最小方案）：

| 变量 | 说明 |
|------|------|
| `MYSQL_ROOT_PASSWORD` | 强密码，仅 Compose 内部 |
| `MYSQL_PASSWORD` / `DB_PASSWORD` | 应用连接密码 |

### 9.2 前端生产环境变量

参考 `frontend/.env.production.example`（**build 时注入**）：

| 变量 | 示例 | 说明 |
|------|------|------|
| `NEXT_PUBLIC_API_BASE_URL` | `https://api.xxx.com/api/v1` | 浏览器调 API |
| `NEXT_PUBLIC_SITE_URL` | `https://www.xxx.com` | 海报二维码域名 |

### 9.3 DeepSeek Key 管理

| 做法 | 推荐 |
|------|------|
| 存放在 ECS 环境变量 / `.env`（权限 600） | 内测可用 |
| 阿里云 KMS / 密钥管理服务 | 正式期推荐 |
| 写入前端 `NEXT_PUBLIC_*` | **禁止** |
| 写入 Git / 文档 | **禁止** |
| 日志打印 Key | **禁止** |

轮换策略：泄露或人员变动时更换 Key，重启 backend 容器生效。

---

## 10. MySQL 数据备份方案

### 10.1 最小方案（Compose MySQL）

```bash
# 每日 cron 示例（在 ECS 上执行，路径按实际调整）
docker exec yijing-mysql mysqldump -u root -p"$MYSQL_ROOT_PASSWORD" \
  --single-transaction --routines yijing \
  | gzip > /backup/yijing_$(date +%Y%m%d).sql.gz
```

- 保留 7–30 天滚动
- 备份文件同步到 OSS（可选）
- 定期做 **恢复演练**（恢复到临时库验证）

### 10.2 RDS 方案

- 开启自动备份（建议 ≥ 7 天）
- 重要变更前手动快照
- 记录连接串与 migration 版本（`schema_migrations` 表）

### 10.3 迁移版本记录

```sql
SELECT * FROM schema_migrations ORDER BY version;
```

上线前应确认已执行至 `006_daily_fortune.sql`。

---

## 11. 日志查看方案

### 11.1 当前 MVP 能力

| 日志类型 | 位置 | 生产建议 |
|----------|------|----------|
| 容器标准输出 | `docker compose logs -f backend` | 初期够用 |
| AI 调用记录 | 表 `ai_generation_logs` | 禁 debug 公网后仅 DB / 内网查 |
| Nginx 访问 / 错误 | `/var/log/nginx/` | 分析流量、排查 4xx/5xx |
| MySQL 慢查询 | RDS 控制台 / 慢日志 | 流量上升后开启 |

### 11.2 推荐初期做法

```bash
docker compose logs -f --tail=200 backend
docker compose logs -f --tail=200 frontend
```

### 11.3 可选升级

- 阿里云 SLS 收集 Docker / Nginx 日志
- `logrotate` 防止磁盘打满
- 告警：5xx 激增、磁盘 > 80%、RDS CPU 高

**生产务必：** `ENABLE_DEBUG_ROUTES=false`，避免 `/api/v1/debug/*` 暴露 AI 日志接口。

---

## 12. 安全组端口规划

| 端口 | 协议 | 来源 | 是否公网开放 | 说明 |
|------|------|------|--------------|------|
| **22** | TCP | 你的办公 IP / 跳板机 | 限制 IP，非 `0.0.0.0/0` | SSH 管理 |
| **80** | TCP | `0.0.0.0/0` | 是 | HTTP 跳转 HTTPS |
| **443** | TCP | `0.0.0.0/0` | 是 | HTTPS 入口 |
| **8080** | TCP | — | **否** | 仅 `127.0.0.1` + Nginx 反代 |
| **3000** | TCP | — | **否** | 仅 `127.0.0.1` + Nginx 反代 |
| **3306** | TCP | — | **否** | 仅 Docker 内网或 RDS VPC；**禁止公网** |

RDS 安全组：仅允许 ECS 安全组访问 3306。

---

## 13. ICP 备案前置说明

在中国大陆节点使用域名提供 **Web 网站服务**，通常需要 **ICP 备案**。

### 13.1 备案前你能做的

- 本地 / 内网继续开发与验收
- 编写部署文档、Nginx 模板、备份脚本
- 在阿里云购买 ECS（部分场景可先 IP 访问做技术验证，但不适合正式对外）
- 注册域名（实名）

### 13.2 通常需要备案后才能正式做的

- 域名 `www.xxx.com` 解析到大陆 ECS 并提供 HTTPS 网站
- 对外宣传推广、公开运营
- 微信小程序 request 合法域名（需 HTTPS + 备案域名，见第 14 节）

### 13.3 备案准备材料（个人 / 企业按主体准备）

- 域名证书（阿里云注册较顺）
- 主体证件（身份证 / 营业执照）
- 阿里云备案服务号（ECS 等）
- 网站名称、核验照片等（按管局要求）

**周期参考：** 约 1–3 周（各地管局不同），建议尽早提交。

### 13.4 未备案的替代（仅技术验证）

- 香港 / 海外节点（免大陆备案，但有延迟与合规差异）
- 仅 IP + 端口访问（不适合海报二维码与正式产品）

**本 MVP 目标市场若在大陆，仍建议走大陆 ECS + 备案路线。**

---

## 14. 小程序上线前还需要准备什么

> 当前 MVP **未做** 微信小程序；以下为未来迁移前置清单。

| 项 | 说明 |
|----|------|
| 备案完成 | 大陆服务器 + 域名 |
| HTTPS 域名 | `api.xxx.com` 配置合法域名 |
| 微信公众平台 | 注册小程序、主体认证 |
| 服务器域名白名单 | request 合法域名、`uploadFile`、`downloadFile` 等 |
| 业务域名 / web-view | 若 H5 嵌套 |
| 用户隐私协议 | 隐私政策弹窗、用户协议 |
| 内容合规 | 与 Web 相同免责声明；玄学类审核风险预案 |
| 登录体系 | 当前为 `session_key` + localStorage；小程序需 `openid` 等 |
| 支付 | 真实支付需商户号；当前 mock 解锁需产品改造 |
| 海报 / 分享 | 小程序分享卡片与保存相册 API |
| 后端 CORS | 小程序不走浏览器 CORS，但需校验来源 / 鉴权 |
| 性能与包体 | 小程序代码包、首屏、接口超时 |

---

## 15. 购买服务器前检查清单

在下单 ECS / RDS / 域名之前，确认：

- [ ] 本地 MVP 全流程验收通过（问事、今日运势、解锁、海报、历史）
- [ ] `go test ./...`、`npm run build` 通过
- [ ] 已阅读 `.env.production.example` 与 `docs/phase8-production-hardening.md`
- [ ] 已选定 **最小** 或 **稍稳** 方案
- [ ] 预估月预算（ECS + 公网 IP + RDS 可选 + 域名）
- [ ] 确认部署地域（大陆需备案）
- [ ] 域名命名与 `www` / `api` 子域规划已定
- [ ] DeepSeek 账号与 API 额度已准备（若内测要用 AI）
- [ ] 运维方式确定（SSH 密钥、非密码登录）
- [ ] 备份与回滚思路已读（第 10、17 节）
- [ ] 明确本阶段 **不做** 真实支付、广告、小程序

---

## 16. 上线前检查清单

资源到位后、切换流量前：

### 16.1 基础设施

- [ ] ECS 安全组仅开放 22（限 IP）、80、443
- [ ] 8080 / 3000 / 3306 不对公网开放
- [ ] 域名 DNS 解析正确（`www`、`api`）
- [ ] SSL 证书有效，HTTPS 强制跳转
- [ ] Nginx 配置 `proxy_set_header X-Forwarded-Proto`

### 16.2 应用配置

- [ ] `APP_ENV=production`
- [ ] `ENABLE_DEBUG_ROUTES=false`
- [ ] `CORS_ALLOWED_ORIGINS` 为生产前端域名
- [ ] `DEEPSEEK_API_KEY` 已配置且未泄露
- [ ] 前端 build 使用生产 `NEXT_PUBLIC_*`
- [ ] `migrate` 已执行到最新（含 `006`）
- [ ] MySQL 字符集 utf8mb4

### 16.3 功能冒烟（生产 URL）

- [ ] `GET https://api.xxx.com/health` 返回 `db: ok`，时间戳 `+08:00`
- [ ] 首页 → 问事 → 起卦动画 → 结果页
- [ ] 今日运势新起卦 / 重复进入
- [ ] Mock 解锁与完整解读
- [ ] 分享海报二维码可打开结果页
- [ ] 历史记录分页
- [ ] 敏感词拦截仍生效
- [ ] 限流在高压下表现符合预期

### 16.4 安全与运维

- [ ] 数据库强密码、独立账号（非 root 跑应用）
- [ ] 备份任务已配置并验证恢复
- [ ] 日志轮转或监控已配置
- [ ] `.env` 权限与 Git 忽略规则正确

---

## 17. 回滚方案

### 17.1 应用回滚

| 场景 | 操作 |
|------|------|
| 新版本 backend 异常 | `docker compose` 切回上一镜像 tag；或保留 `server.bak` 二进制替换重启 |
| 新版本 frontend 异常 | 切回上一 frontend 镜像；**注意** `NEXT_PUBLIC_*` 与 build 绑定 |
| 配置错误 | 恢复上一版 `.env`，重启对应容器 |

建议：每次发布打 tag，如 `yijing-backend:20260623-1`。

### 17.2 数据库回滚

- **优先：** RDS 快照 / 备份恢复到新实例，切换连接串（影响面大，谨慎）
- **迁移脚本错误：** 不要强行删 `schema_migrations`；根据 SQL 写补偿脚本，在测试库验证后再上生产
- **小版本：** 保留发布前 `mysqldump` 文件，可恢复到临时库对比

### 17.3 域名 / 证书回滚

- DNS 切回旧 IP（TTL 内生效）
- Nginx 恢复上一版 `conf.d` 备份

### 17.4 回滚决策建议

- 仅 frontend 展示问题 → 回滚 frontend
- API 5xx 或数据错误 → 回滚 backend + 评估 DB
- 迁移失败 → **停止发布**，不要重复执行错误 migration

---

## 18. 暂时不做的事项

| 事项 | 原因 |
|------|------|
| 真实支付 | MVP 范围外；需商户号与合规 |
| 真实广告 | MVP 范围外；需 SDK 与审核 |
| 微信小程序迁移 | 单独项目；需备案与微信侧配置 |
| 多机部署 / 负载均衡 | 当前流量不需要 |
| Kubernetes | 运维复杂度过高 |
| Redis 分布式限流 | 单实例内存限流够用；多实例时再考虑 |
| OSS 静态托管全站 | Next standalone + Nginx 足够 |
| 自动 CI/CD 上云 | 可后续 Phase 10 再做 |
| 在本 Phase 直接购买与部署 | 本文档仅方案整理 |

---

## 附录 A：推荐内测服务器规格（阿里云 ECS）

| 方案 | 规格建议 | 约用途 |
|------|----------|--------|
| **最小内测** | 2 vCPU，4 GiB，ESSD 40–60 GiB | Compose 三件套 + Nginx，日活 < 几百 |
| **稍稳内测** | 2–4 vCPU，4–8 GiB + RDS 基础版 1C2G | 业务与 DB 分离，小范围邀请测试 |

操作系统建议：**Alibaba Cloud Linux 3** 或 **Ubuntu 22.04 LTS**，便于 Docker 与 Nginx 安装。

---

## 附录 B：本地可提前准备的交付物（无需服务器）

- [ ] 生产用 `.env` 模板填好占位（不含真实 Key）
- [ ] `docker-compose.prod.yml` 草稿（去掉 MySQL 公网端口等）
- [ ] Nginx 配置草稿（本文第 8 节扩展）
- [ ] 备份 cron 脚本
- [ ] 发布 / 回滚 runbook（基于第 16、17 节勾选）
- [ ] 域名与备案材料清单
- [ ] 生产冒烟用例（可复制 `docs/local-acceptance-test.md` 改 URL）

---

## 附录 C：相关文档

- [phase8-production-hardening.md](./phase8-production-hardening.md) — 安全收口
- [local-acceptance-test.md](./local-acceptance-test.md) — 本地验收
- [README.md](../README.md) — 项目总览

---

*文档版本：Phase 9 部署前方案 · 不含实际部署与密钥*
