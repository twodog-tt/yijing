# Phase 10：阿里云 ECS 内测环境部署指南

> **目标：** 单台 ECS + Docker Compose + 宿主机 Nginx，供少量朋友通过 **公网 IP** 远程体验 MVP  
> **不做：** 真实支付、广告、微信小程序、本仓库内自动上云执行  
> **后续：** 备案完成后可平滑切换域名与 HTTPS（见第 14 节）

相关文档：[phase9-aliyun-deployment-plan.md](./phase9-aliyun-deployment-plan.md)

---

## 部署包文件一览

| 文件 | 用途 |
|------|------|
| `docker-compose.prod.yml` | 生产 Compose（端口仅本机） |
| `.env.internal-test.example` | 内测环境变量模板 |
| `deploy/nginx/yijing-ip.conf` | IP 访问 Nginx 反代 |
| `deploy/scripts/server-init.sh` | 新服务器初始化 |
| `deploy/scripts/deploy.sh` | 构建、启动、健康检查 |
| `deploy/scripts/backup-mysql.sh` | MySQL 逻辑备份 |

---

## 1. 阿里云 ECS 购买建议

### 1.1 规格（内测）

| 项 | 建议 |
|----|------|
| 规格 | **2 vCPU / 2 GiB**（配合 2 GiB swap，内测够用） |
| 系统盘 | ESSD **40–60 GiB** |
| 操作系统 | Alibaba Cloud Linux 3 或 Ubuntu 22.04 LTS |
| 地域 | 大陆节点（后续备案）；仅技术验证可考虑香港（路线不同） |
| 公网 | 分配 **弹性公网 IP** |
| 计费 | 包月或按量均可 |

### 1.2 不必此时购买

- RDS（内测 MySQL 跑在 Docker 内即可）
- OSS、SLB、WAF
- 域名与 SSL（IP 内测阶段不需要）

---

## 2. 系统初始化步骤

### 2.1 登录服务器

```bash
ssh root@<ECS公网IP>
# 建议：配置 SSH 密钥登录，禁用密码登录（可选）
```

### 2.2 运行初始化脚本

将项目上传到服务器后（见第 5 节），执行：

```bash
cd /opt/yijing   # 或你的部署目录
sudo bash deploy/scripts/server-init.sh
```

脚本将完成：

- 系统更新
- 创建 2 GiB swap（已存在时不会覆盖）
- 安装 Docker、Docker Compose 插件
- 安装并启用 Nginx
- 创建 `/opt/yijing` 与 `backups` 目录

环境变量 `DEPLOY_DIR` 可覆盖默认部署路径：

```bash
sudo DEPLOY_DIR=/home/deploy/yijing bash deploy/scripts/server-init.sh
```

---

## 3. 安装 Docker / Docker Compose

`server-init.sh` 已通过 [get.docker.com](https://get.docker.com) 安装 Docker。

手动验证：

```bash
docker --version
docker compose version
```

若 Compose 插件缺失（部分镜像）：

```bash
# Ubuntu / Debian
sudo apt-get install -y docker-compose-plugin

# Alibaba Cloud Linux / CentOS 系
sudo dnf install -y docker-compose-plugin
```

---

## 4. 安装 Nginx

`server-init.sh` 已安装。手动验证：

```bash
nginx -v
systemctl status nginx
```

安装内测站点配置：

```bash
sudo cp deploy/nginx/yijing-ip.conf /etc/nginx/conf.d/yijing.conf
# Ubuntu 默认站点会与 default_server 冲突；只移除软链接，保留原配置文件
if [ -L /etc/nginx/sites-enabled/default ]; then
  sudo unlink /etc/nginx/sites-enabled/default
fi
sudo nginx -t
sudo systemctl reload nginx
```

---

## 5. 项目上传方式

任选一种（`deploy.sh` 中 `git pull` 为注释占位）：

### 方式 A：Git（推荐）

```bash
sudo mkdir -p /opt/yijing && sudo chown "$USER":"$USER" /opt/yijing
cd /opt/yijing
git clone <你的仓库地址> .
```

### 方式 B：本机 rsync

```bash
rsync -avz --exclude node_modules --exclude .next --exclude .git \
  ./ user@<ECS_IP>:/opt/yijing/
```

### 方式 C：打包上传

```bash
tar czf yijing.tar.gz --exclude=node_modules --exclude=frontend/.next .
scp yijing.tar.gz user@<ECS_IP>:/opt/yijing/
# 服务器上 tar xzf yijing.tar.gz
```

---

## 6. `.env` / 内测配置说明

```bash
cd /opt/yijing
if [ ! -e .env.internal-test ]; then
  install -m 600 .env.internal-test.example .env.internal-test
else
  echo ".env.internal-test 已存在，未覆盖"
fi
if [ ! -e .env ]; then ln -s .env.internal-test .env; fi
if [ ! -e .env.production ]; then ln -s .env.internal-test .env.production; fi
nano .env.internal-test
```

三个入口使用同一份配置，避免密码重复维护；上述命令不会覆盖已有环境文件。

### 6.1 必须修改的项

| 变量 | 说明 |
|------|------|
| `SERVER_IP` | ECS 公网 IP（全文替换占位） |
| `MYSQL_ROOT_PASSWORD` | MySQL root 强密码 |
| `MYSQL_PASSWORD` | 应用库用户密码；Compose 会注入 backend 的 `DB_PASSWORD` |
| `CORS_ALLOWED_ORIGINS` | `http://<公网IP>` |
| `NEXT_PUBLIC_API_BASE_URL` | `http://<公网IP>/api/v1` |
| `NEXT_PUBLIC_SITE_URL` | `http://<公网IP>` |

一键替换示例（将 `123.45.67.89` 换成真实 IP）：

```bash
sed -i 's/SERVER_IP/123.45.67.89/g' .env.internal-test
# 再手动改 CHANGE_ME 密码
```

### 6.2 生产安全项（内测亦应遵守）

| 变量 | 内测值 |
|------|--------|
| `APP_ENV` | `production` |
| `ENABLE_DEBUG_ROUTES` | **`false`** |
| `ENABLE_RATE_LIMIT` | `true` |
| `AI_PROVIDER` | `mock` 或 `deepseek` |

### 6.3 DeepSeek Key

仅在服务器 `.env.internal-test` 中设置，**不要**写入前端或 Git：

```env
AI_PROVIDER=deepseek
DEEPSEEK_API_KEY=sk-xxxxxxxx
```

修改 `AI_PROVIDER` 或 Key 后只需重启 backend：

```bash
docker compose -f docker-compose.prod.yml up -d backend
```

修改 `NEXT_PUBLIC_*` 后必须 **重新 build frontend**：

```bash
docker compose -f docker-compose.prod.yml build frontend
docker compose -f docker-compose.prod.yml up -d frontend
```

---

## 7. Docker Compose 启动方式

### 7.1 一键部署（推荐）

```bash
cd /opt/yijing
bash deploy/scripts/deploy.sh
```

### 7.2 手动步骤

```bash
docker compose -f docker-compose.prod.yml --env-file .env build
docker compose -f docker-compose.prod.yml --env-file .env up -d
docker compose -f docker-compose.prod.yml ps
```

### 7.3 端口策略（prod compose）

| 服务 | 宿主机暴露 |
|------|------------|
| mysql | **不映射** |
| backend | `127.0.0.1:8080` 仅本机 |
| frontend | `127.0.0.1:3000` 仅本机 |
| 公网入口 | Nginx `80` |

---

## 8. Migration 执行方式

### 8.1 自动（默认）

`backend` 容器启动时 `docker-entrypoint.sh` 会：

1. 等待 MySQL 就绪  
2. 执行 `./migrate`（读取 `schema_migrations`）  
3. 启动 `./server`

首次启动且数据卷为空时，`sql/*.sql` 也会通过 MySQL `docker-entrypoint-initdb.d` 初始化；之后以 **migrate 工具** 为准。

### 8.2 手动补跑

```bash
docker compose -f docker-compose.prod.yml exec backend ./migrate
```

### 8.3 确认版本

```bash
docker compose -f docker-compose.prod.yml exec mysql \
  mysql -u root -p"$MYSQL_ROOT_PASSWORD" yijing \
  -e "SELECT * FROM schema_migrations ORDER BY version;"
```

应包含至 `006_daily_fortune.sql`。

---

## 9. Nginx 反代配置

配置文件：`deploy/nginx/yijing-ip.conf`

| 路径 | 上游 |
|------|------|
| `/` | `http://127.0.0.1:3000`（frontend） |
| `/api/` | `http://127.0.0.1:8080`（backend，保留 `/api/v1` 前缀） |
| `/health` | `http://127.0.0.1:8080/health` |

浏览器访问 API 基址：`http://<IP>/api/v1`（与 `NEXT_PUBLIC_API_BASE_URL` 一致）。

---

## 10. 安全组端口说明

在阿里云控制台 → ECS → 安全组 → **入方向**：

| 端口 | 授权对象 | 说明 |
|------|----------|------|
| **22** | 你的办公 IP /32 | SSH，**不要**对 `0.0.0.0/0` 长期开放 |
| **80** | `0.0.0.0/0` | HTTP，Nginx 入口 |
| **443** | 暂不开放 | 上 HTTPS 后再开 |

**不要开放：**

| 端口 | 原因 |
|------|------|
| 3306 | MySQL 仅 Docker 内网 |
| 8080 | backend 仅 127.0.0.1 |
| 3000 | frontend 仅 127.0.0.1 |

**出方向：** 默认允许即可；backend 需访问 `api.deepseek.com`（若用 DeepSeek）。

---

## 11. DeepSeek Key 配置说明

1. 在 [DeepSeek 开放平台](https://platform.deepseek.com) 创建 API Key  
2. 写入服务器 `/opt/yijing/.env.internal-test`（权限 `600`）
3. 设置 `AI_PROVIDER=deepseek`  
4. 重启 backend 容器  
5. **切勿** 提交到 Git、文档或前端环境变量  

验证：新起卦 → mock 解锁 → 观察完整解读是否生成；失败时 backend 日志会有 fallback 记录。

---

## 12. 日志查看方式

```bash
cd /opt/yijing

# 所有服务
docker compose -f docker-compose.prod.yml logs -f --tail=200

# 单独
docker compose -f docker-compose.prod.yml logs -f backend
docker compose -f docker-compose.prod.yml logs -f frontend
docker compose -f docker-compose.prod.yml logs -f mysql

# Nginx
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

---

## 13. 备份方式

### 13.1 手动备份

```bash
cd /opt/yijing
bash deploy/scripts/backup-mysql.sh
```

输出：`backups/yijing_yijing_YYYYMMDD_HHMMSS.sql.gz`

### 13.2 定时备份（cron 示例）

```bash
crontab -e
# 每天 3:00 备份
0 3 * * * cd /opt/yijing && bash deploy/scripts/backup-mysql.sh >> /var/log/yijing-backup.log 2>&1
```

### 13.3 恢复（谨慎，会覆盖数据）

```bash
set -a
source .env
set +a
gunzip -c backups/yijing_yijing_20260623_030000.sql.gz | \
  docker compose -f docker-compose.prod.yml exec -T mysql \
  mysql -u root -p"$MYSQL_ROOT_PASSWORD" yijing
```

恢复前建议先备份当前库。

---

## 14. 回滚方式

### 14.1 仅应用回滚

```bash
cd /opt/yijing
git switch --detach <上一版本commit>
docker compose -f docker-compose.prod.yml --env-file .env build
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

或保留上一版镜像 tag，修改 compose `image:` 指向旧 tag。

### 14.2 配置回滚

```bash
# 先比较并人工确认，不直接覆盖当前配置
diff -u .env.internal-test .env.internal-test.bak || true
# 确认后再执行：cp .env.internal-test.bak .env.internal-test
docker compose -f docker-compose.prod.yml up -d
```

### 14.3 数据库回滚

从 `backups/*.sql.gz` 恢复到临时库验证后，再决定是否覆盖生产库（见 13.3）。

### 14.4 停止服务

```bash
docker compose -f docker-compose.prod.yml down
# 保留数据卷；加 -v 会删除 MySQL 数据，慎用
```

---

## 15. 内测验收清单

将 `http://<ECS_IP>` 发给朋友前，自行勾选：

| # | 项 | 期望 |
|---|-----|------|
| 1 | `http://<IP>/api/v1/health` | `code` 正常，`db: ok` |
| 2 | 首页 | 免责声明、问事与今日运势入口 |
| 3 | `/ask` 起卦 | 动画 → 结果页 |
| 4 | `/today` 新起卦 | 动画；重复进入打开已有结果 |
| 5 | 免费解读 / mock 解锁 | 正常 |
| 6 | DeepSeek（若启用） | 解锁后完整解读 |
| 7 | 分享海报 | 二维码为 `http://<IP>/divination/{id}` |
| 8 | `/history` | 列表正常 |
| 9 | 手机浏览器 | 布局可读 |
| 10 | `ENABLE_DEBUG_ROUTES=false` | `/api/v1/debug/*` 不可访问 |
| 11 | 安全组 | 仅 22（限 IP）、80 公网 |

---

## 16. 完整内测部署流程（速查）

```text
1. 准备 ECS（2C2G + 2G swap）+ 公网 IP + 安全组（22 限 IP，80 公网）
2. SSH 登录
3. 上传代码到 /opt/yijing
4. sudo bash deploy/scripts/server-init.sh
5. 创建 `.env.internal-test`，让 `.env` / `.env.production` 链接到它，再改 IP、密码
6. sudo cp deploy/nginx/yijing-ip.conf /etc/nginx/conf.d/yijing.conf
7. sudo nginx -t && sudo systemctl reload nginx
8. bash deploy/scripts/deploy.sh
9. 浏览器打开 http://<IP>/ 走验收清单
10. 配置 cron 备份（可选）
```

---

## 17. 有域名和备案后再改什么

| 项 | IP 内测 | 域名 + HTTPS 后 |
|----|---------|-----------------|
| Nginx `server_name` | `_` | `www.xxx.com` |
| 监听 | 仅 80 | 80 跳转 443 + SSL 证书 |
| `NEXT_PUBLIC_SITE_URL` | `http://IP` | `https://www.xxx.com` |
| `NEXT_PUBLIC_API_BASE_URL` | `http://IP/api/v1` | `https://api.xxx.com/api/v1` 或同域 |
| `CORS_ALLOWED_ORIGINS` | `http://IP` | `https://www.xxx.com` |
| 安全组 | 80 | 增加 443 |
| 海报二维码 | `http://IP/...` | `https://www.xxx.com/...` |
| 前端镜像 | — | **必须重新 build** |

API 拆分为 `api.xxx.com` 时，可参考 `deploy/nginx/` 新增独立 server 块（phase9 文档第 8 节）。

---

## 18. 常见问题

### 前端能开页但 API 失败

- 检查 `NEXT_PUBLIC_API_BASE_URL` 是否为 `http://<IP>/api/v1`  
- 是否修改后未 `build frontend`  
- Nginx `location /api/` 是否生效：`curl http://127.0.0.1/api/v1/health`

### CORS 错误

- `CORS_ALLOWED_ORIGINS` 必须与浏览器地址栏一致（含 `http://`，无尾斜杠）  
- 改后重启 backend

### migrate 报错 checksum

- 勿随意改已执行的 `sql/*.sql`  
- 在测试环境解决后再动生产

### 磁盘不足

- `docker system prune`（谨慎）  
- 清理旧 `backups/*.sql.gz`

---

*文档版本：Phase 10 内测部署包 · 不含真实密钥与自动上云执行*
