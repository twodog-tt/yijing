# AI Agent Workflow

本文档说明 **AI 易经 / 今日一卦** 项目中 **ChatGPT、Cursor、Codex** 三个 AI 工具的工作边界与协作流程。

适用技术栈：Go 后端、MySQL、Next.js Web/H5、微信小程序、Docker Compose、Nginx、阿里云 ECS。

> **目标：** 避免 AI 工具越权修改、重复消耗额度、误改生产配置、误提交敏感信息。

---

## 1. 目标

本项目使用多个 AI 工具协作，遵循以下原则：

| 工具 | 角色 |
|------|------|
| **Cursor** | 主力 **实现** |
| **Codex** | 高价值 **审查** 与风险判断 |
| **ChatGPT** | **规划** 与 Prompt 设计 |

**协作原则：**

- 所有 **高风险变更** 必须先 **计划 → 实现 → review**
- **不允许** 任何 AI 工具提交密钥、密码、私钥、AppSecret、DeepSeek Key
- Cursor 额度较充足，承担大部分开发；Codex 额度有限，只用于关键审查
- 生产、安全、数据库、部署相关变更 **必须** 经 Codex（或人工）review 后再合并

---

## 2. 工具角色分工

### ChatGPT

**职责：**

- 产品方向判断
- 功能阶段规划（Phase 划分、MVP 边界）
- 架构取舍建议（单 ECS vs RDS、域名拆分等）
- 生成 **Cursor 执行 Prompt**
- 生成 **Codex Review Prompt**
- 合规、商业化、备案、小程序上线策略建议

**不负责：**

- 直接修改代码仓库
- 直接部署服务器
- 保存或传递真实密钥

---

### Cursor

**职责：**

- 主力开发
- 批量创建 / 修改文件
- 页面与组件实现
- 样式调整
- 文档更新（`docs/`、`README.md`）
- 普通 bug 修复
- Docker / Nginx / 部署脚本落地
- 按 **明确 Prompt** 执行阶段任务

**适合任务：**

- 小程序页面开发
- 分享海报 Canvas
- 激励视频广告 mock 适配层
- Web/H5 页面优化
- Go 后端普通接口开发
- SQL migration 编写
- README 与 docs 更新

**禁止事项：**

- 未经确认 **直接删除数据库** 或数据
- 未经确认 **修改生产环境 `.env`**
- 未经确认 **删除 Docker volume**
- 未经确认 **修改安全组**、对公网开放 3306 / 8080 / 3000
- 把 AppID、AppSecret、DeepSeek Key、服务器密码、SSH 私钥 **写入代码或文档**
- 在没有 review 的情况下 **大改数据库结构**
- 在没有 review 的情况下 **改正式部署配置**

---

### Codex

**职责：**

- **只读审查**（默认）
- `git diff` review
- 安全检查
- 上线前验收
- 部署方案审查
- 数据库 migration 审查
- 小程序审核风险审查
- 广告合规风险审查
- 找隐藏 bug、边界问题和高风险配置

**适合任务：**

- 检查 Cursor 的变更是否越界
- 检查是否泄露敏感信息
- 检查小程序文案是否可能影响审核
- 检查 Nginx / Docker / env 是否有安全风险
- 检查是否误用正式域名
- 检查是否开放内部端口
- 检查是否打印完整 AI 报告或 `session_key`
- 检查是否有重复请求、重复跳转、timer 未清理
- 检查上线前验收清单

**不适合任务：**

- 大量写页面
- 批量补 CSS
- 创建大量组件
- 普通样式调整
- 简单文案改写
- 已经明确的机械性开发任务

**默认要求：**

- Codex **默认先做只读检查**
- **不要直接修改代码**，除非用户明确要求
- 输出 review 结论和 **Cursor 修复指令**
- **重点节省 Codex 额度**

---

## 3. 标准协作流程

```text
1. ChatGPT 制定阶段目标和 Prompt
        ↓
2. Cursor 根据 Prompt 实现
        ↓
3. Cursor 输出：修改文件列表 + 验收结果
        ↓
4. Codex 做只读 review
        ↓
5. Cursor 根据 Codex review 修复（如有）
        ↓
6. 用户本地 / 开发者工具 / 服务器验收
        ↓
7. 通过后再 git commit（必要时 push）
```

**说明：**

- 小改动（纯文案、单文件样式）可省略 Codex，但 **部署 / 安全 / migration** 不可省略
- 用户说「可以修改」后，Cursor 才可按 review 改代码
- commit 前由用户或 Cursor 执行第 8 节检查清单

---

## 4. 每个阶段的文件变更规则

| 阶段 | 允许修改范围 | 说明 |
|------|--------------|------|
| 普通小程序开发 | `miniprogram/`、相关 `docs/` | 不改 backend / deploy |
| Web/H5 开发 | `frontend/`、相关 `docs/` | 不改 miniprogram（除非明确要求） |
| 后端开发 | `backend/`、`sql/`、相关 `docs/` | migration 需 review |
| 部署 | `deploy/`、`docker-compose*.yml`、Nginx、env **example**、相关 `docs/` | 不改业务逻辑 |
| Review | **Codex 默认不修改任何文件** | 只输出结论与修复指令 |

**跨范围修改：** 必须先说明原因，并在 review 中重点检查影响面。

---

## 5. 敏感信息规则

### 明确禁止提交到 Git

| 类型 | 示例 |
|------|------|
| AI / 第三方 Key | DeepSeek API Key、OpenAI Key |
| 微信密钥 | AppSecret、微信支付密钥 |
| 服务器凭证 | root 密码、SSH 私钥 |
| 数据库 | 生产环境密码、真实 DSN |
| 本地私密配置 | `.env`、`project.config.json`、`project.private.config.json` |
| 个人证件 | 任何真实证件信息 |

### 允许提交

| 类型 | 示例 |
|------|------|
| 占位模板 | `.env.example`、`.env.production.example`、`.env.internal-test.example` |
| 小程序示例 | `project.config.json.example`（若存在） |
| 占位符 | `SERVER_IP`、`CHANGE_ME`、空 `DEEPSEEK_API_KEY=` |

**密钥只存在于：** 本机 `.env`、服务器 `/opt/yijing/.env`（权限 `600`），**不进 Git、不进 README、不进前端 `NEXT_PUBLIC_*`**。

---

## 6. Codex Review 固定模板

复制到 Codex 时使用：

```text
请你先做只读检查，不要直接修改代码。

当前阶段：
- [填写阶段名称]

本次变更目标：
- [填写目标]

请重点检查：
1. git diff 是否只改了预期文件
2. 是否有敏感信息
3. 是否有越权修改 backend/frontend/database/deploy
4. 是否有小程序审核风险文案
5. 是否有正式域名/HTTPS/备案未完成却被默认使用
6. 是否有安全风险
7. 是否有重复请求、重复跳转、timer 未清理
8. 是否有错误状态未处理
9. 是否有日志泄露完整 AI 报告或 session_key
10. 是否需要 Cursor 修复

请输出：
- 总体结论
- 高风险问题
- 中风险问题
- 低风险问题
- 建议让 Cursor 修复的明确指令
- 验收清单

不要直接改代码。
```

---

## 7. Cursor 执行固定模板

复制到 Cursor 时使用：

```text
请按以下阶段任务实现。

要求：
- 先确认当前 git status
- 只修改本阶段允许范围内的文件
- 不要提交 Git
- 不要写入任何真实密钥
- 不要修改生产 `.env`
- 不要删除数据库或 Docker volume
- 完成后输出：
  - 修改文件列表
  - 实现内容
  - 验收结果
  - 未完成项
  - 是否修改 backend/frontend/database/deploy
  - 下一阶段建议
```

---

## 8. Git 提交前检查

提交前必须执行：

```bash
git status
git diff --stat
```

并确认：

- [ ] 变更范围符合阶段目标
- [ ] 没有敏感信息（Key、密码、私钥、真实 IP 上的私密配置）
- [ ] 没有误删文件
- [ ] 没有修改不该修改的目录
- [ ] 相关文档已更新
- [ ] 本地构建 / 测试通过（如 `go test ./...`、`npm run build`、小程序开发者工具编译）

---

## 9. 当前项目推荐分工

| 任务 | 实现 | 审查 |
|------|------|------|
| 小程序分享海报 | Cursor | Codex |
| 激励视频广告 mock 适配层 | Cursor | Codex |
| ECS 部署脚本（`deploy/`、`docker-compose.prod.yml`） | Cursor | Codex |
| Nginx / Docker / env 安全 | — | **Codex** |
| 备案通过后的域名 / HTTPS 配置 | Cursor | Codex |
| 小程序正式发布前总验收 | — | **Codex** |
| SQL migration（结构变更） | Cursor | **Codex 必审** |
| 今日运势 / 问事核心流程 | Cursor | 大改时 Codex |
| 文档（phase9/10、验收手册） | Cursor | 部署类文档 Codex |

---

## 10. 总原则

| 原则 | 说明 |
|------|------|
| **Cursor 多做** | 实现、文档、脚本、样式、普通接口 |
| **Codex 少做但做关键判断** | 安全、部署、migration、审核风险、上线验收 |
| **ChatGPT 负责拆解和提示词** | 阶段目标、Prompt、架构与合规建议 |
| **高风险必须 review** | 生产、安全、数据库、部署、跨目录大改 |
| **密钥不进 Git** | 只存在本地或服务器 `.env` |

---

## 相关文档

- [phase8-production-hardening.md](./phase8-production-hardening.md) — 生产安全收口
- [phase9-aliyun-deployment-plan.md](./phase9-aliyun-deployment-plan.md) — 阿里云部署前方案
- [phase10-aliyun-internal-test-deploy.md](./phase10-aliyun-internal-test-deploy.md) — ECS 内测部署
- [local-acceptance-test.md](./local-acceptance-test.md) — 本地验收
- [miniprogram-dev.md](./miniprogram-dev.md) — 小程序开发

---

*文档版本：AI 协作流程 · 不含密钥与生产配置*
