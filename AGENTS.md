# AGENTS.md

本文件是 Codex / Cursor / 维护者在本仓库工作的固定上下文。除非用户明确要求更新规则，否则执行任务前应先遵守本文件，再结合具体任务 Prompt。

## 1. 项目简介

「文易传统文化」是一个微信小程序 + Go 后端项目，以传统文化符号为载体，提供学习参考、趣味解读、自我观察与行动节奏整理。

当前用户侧核心模块：

- 问事起卦
- 八字简析
- 奇门问事
- 历史记录
- 分享卡片 / 分享长图
- 首页模块引导

当前后端核心能力：

- 六爻问事与今日一卦 API
- 通用 `analysis_records` 分析报告 API
- 八字 / 奇门简析、完整报告生成与解锁
- DeepSeek 优先生成完整报告，失败使用模板 fallback
- 内部算法灰度：`bazi-v2-poc`、`qimen-v2-poc`、`qimen-v2-professional`

产品定位必须保持为传统文化学习参考，不是预测、改命或专业建议服务。

## 2. 技术栈

- 微信小程序：`miniprogram/`
- 后端：Go
- Web/H5：Next.js + TypeScript + Tailwind CSS，位于 `frontend/`
- 数据库：MySQL 8.0
- AI：DeepSeek / mock / template fallback
- 部署：Docker Compose、Nginx、阿里云 ECS
- 数据库脚本：`sql/` 按编号迁移

## 3. 目录结构

- `backend/`：Go API、handler、service、repository、model、migrate/server 命令
- `backend/internal/handler/analysis.go`：八字 / 奇门 analysis API handler
- `backend/internal/model/analysis.go`：analysis 模型、模块类型、算法版本常量
- `backend/internal/service/bazi/`：八字算法、报告、DeepSeek prompt、v2 POC
- `backend/internal/service/qimen/`：奇门算法、报告、professional 预览与灰度
- `miniprogram/pages/`：小程序页面
- `miniprogram/components/`：小程序自定义组件
- `miniprogram/utils/`：API、配置、session、分享长图、模块视图构建
- `frontend/`：Web/H5 端；其中 `frontend/AGENTS.md` 有 Next.js 特别规则
- `sql/`：MySQL schema / seed / migration
- `deploy/`：Nginx 与 ECS 脚本
- `docs/`：阶段设计、开发说明、验收记录与上下文
- `.deploy-patches/`：部署补丁临时目录，不提交

## 4. 固定工程规则

- 先读文档和现有代码，再改动。
- 保持变更范围与任务目标一致，不做顺手重构。
- 不改变默认算法，除非用户明确要求并确认影响面。
- 不引入广告、支付、微信登录、手机号授权或小程序码能力，除非用户明确要求。
- 不把密钥、AppSecret、DeepSeek Key、服务器密码、私钥、真实 adUnitId 写入仓库。
- 不把真实微信 AppID 反复写入公开文档；文档中使用 `<WECHAT_APP_ID>` 或类似占位符。
- API 地址集中维护，不能散落到页面文件。
- 小程序普通入口不能传 `algorithm_version`，普通用户不能看到算法选择 UI。
- 内部算法只能用于内部创建记录后的详情页条件展示。

## 5. 数据库规则

- 数据库为 MySQL 8.0。
- 不使用 `ENUM`。
- 不使用 `TINYINT`。
- 不使用 `CHECK`。
- 状态字段统一使用 `INT` 并写清楚 `COMMENT`。
- 新 SQL 必须放入 `sql/`，按编号追加 migration，不改历史 migration 语义。
- 不执行 SQL，除非用户明确要求。
- 不执行迁移，除非用户明确要求。
- 不用 `docker compose down -v` 清空数据卷。
- 不用 `docker system prune -a`。
- analysis 表约定：八字 / 奇门共用 `analysis_records`，六爻不塞入该表。
- 八字 / 奇门 analysis 记录删除为硬删除 `analysis_records`；问事起卦删除走 `DELETE /divinations/{id}` 软删除。

## 6. 小程序规则

- 小程序源码在 `miniprogram/`。
- 开发 API：`http://123.57.48.214/api/v1`。
- 正式 API 预留：`https://api.wenyiapp.cn/api/v1`，当前不可用，不能切换为默认。
- `develop` / `trial` 当前使用 dev API；`release` 预留 prod API。
- 微信 DevTools 本地调试需要维护者勾选“不校验合法域名、web-view、TLS 版本以及 HTTPS 证书”。
- 不接广告 / 支付 / 微信登录，除非用户明确要求。
- AD0 后八字 / 奇门结果页使用 `free_unlock` + “查看完整报告”，生产 UI 不展示 mock 广告文案。
- 问事起卦结果页使用 `mock_button` 查看完整解析。
- 分享长图不展示完整原问题、出生日期 / 出生时辰原始输入、`session_key`、payload、prompt、小程序码。
- 八字分享给朋友回到 `/pages/bazi/bazi`，不直达私有记录。
- 奇门分享给朋友可带记录 id，但非同 session 打开不得泄露他人隐私。
- 历史页聚合问事 / 八字 / 奇门，列表只展示安全摘要。

## 7. 后端规则

- `analysis` API 使用 `{ code, message, data }` 响应 envelope。
- analysis GET / DELETE / UNLOCK 使用 `X-Session-Key` header，不把 `session_key` 放入 query。
- `POST /analysis/bazi`、`POST /analysis/qimen` 可在 body 和 header 同时传 session；不一致必须拒绝。
- handler 使用 typed request，不得把客户端任意 JSON 直接透传 repository。
- repository 只接收服务端校验、构造后的 payload。
- 日志和错误响应不得输出完整出生信息、完整原问题、`session_key`、payload、prompt、DeepSeek 原始响应或 API key。
- DeepSeek 完整报告失败时使用模板 fallback，不能让用户流程直接崩掉。
- 内容安全边界由 service / prompt / report forbidden 检查共同维护。
- 如果发现疑似 AppSecret、`access_token`、真实 API key、数据库密码或私钥，先停止提交，并提示维护者轮换对应密钥。

默认算法固定：

- 八字默认 `bazi-simple-v1`
- 奇门默认 `qimen-simple-v1`

内部灰度算法：

- `bazi-v2-poc`
- `qimen-v2-poc`
- `qimen-v2-professional`

内部算法规则：

- 可以由内部 API 请求显式传 `algorithm_version` 创建测试记录。
- 普通小程序 / Web 创建流程不传 `algorithm_version`。
- `bazi-v2-poc` 仅在八字详情页按 `result_payload.algorithm_version` 条件展示 v2 区块。
- `qimen-v2-professional` 仅在奇门详情页按 `algorithm_version` 且 `palaces.length === 9` 条件展示九宫区块。
- `qimen-v2-poc` 虽有 9 宫 payload，但小程序 professional 九宫区块不触发。

## 8. 合规文案规则

统一定位：

- 传统文化学习
- 趣味解读
- 自我反思
- 行动建议 / 行动节奏整理

禁止表述：

- 精准预测
- 改命、化灾、转运承诺
- 保证发财、复合、升职、考试结果
- 医疗、法律、投资、赌博建议
- 恐吓式“大凶大吉”断语
- “看广告改运”等商业化与命运改变绑定文案

推荐表述：

- “仅供传统文化学习与自我反思参考”
- “不构成现实决策依据”
- “请结合实际情况判断”
- “更适合作为状态整理与行动提醒”

## 9. Git 规则

- 当前主分支为 `main`。
- 不使用 `git add .`。
- 不提交 `.deploy-patches/`。
- 不使用 `git reset --hard`。
- 不使用 `git clean`。
- 不使用 force push。
- 不覆盖用户已有未提交改动。
- 提交前查看 `git status --short`，只 stage 任务允许文件。
- 不要 stage 未确认文件；按任务明确列出的路径精确 `git add`。
- SECURITY1 / onboarding 类任务可按任务要求精确添加相关文档，例如：

```bash
git add AGENTS.md docs/CODEX_PROJECT_CONTEXT.md docs/miniprogram-dev.md
git commit -m "docs: add Codex context and sanitize app id"
git push origin main
```

如果还改了其他文件，先停止并让用户确认。

## 10. 禁止操作

- 不修改 backend 业务代码，除非任务明确要求。
- 不修改 miniprogram 功能代码，除非任务明确要求。
- 不修改 frontend，除非任务明确要求。
- 不执行 SQL，除非用户明确要求。
- 不修改 `.env` / `.env.*`。
- 不执行部署。
- 不上传体验版。
- 不提审。
- 不改变默认算法。
- 不新增广告 / 支付 / 登录。
- 不提交 `.deploy-patches/`。
- 不运行破坏性命令。
- 不使用 `docker compose down -v`。
- 不使用 `docker system prune -a`。

## 11. 常用检查命令

只读 / 文档任务：

```bash
git diff --check
git diff --stat
git status --short
```

小程序 JS 语法检查示例：

```bash
node --check miniprogram/utils/api.js
node --check miniprogram/utils/bazi.js
node --check miniprogram/utils/qimen.js
node --check miniprogram/pages/history/history.js
node --check miniprogram/pages/analysis-result/analysis-result.js
node --check miniprogram/pages/qimen-result/qimen-result.js
```

后端相关检查示例：

```bash
cd backend && go test ./...
cd backend && go test -count=1 ./internal/service/bazi/... ./internal/service/qimen/... ./internal/handler/...
```

如果只修改 Markdown，不需要运行后端测试。

## 12. 每次任务完成后的输出格式

完成任务后按以下结构汇报：

```text
# 任务完成结果

## 阅读范围
- 已阅读的关键文档：
- 已阅读的关键代码目录：

## 新增 / 修改文件
- 文件：

## 项目理解摘要
- 项目定位：
- 当前模块：
- 当前默认算法：
- 内部灰度算法：
- 当前阻塞项：

## 检查结果
- git diff --check：
- git diff --stat：
- git status：

## Git
- commit hash：
- 是否 push：
- .deploy-patches 是否未提交：

## 下一步建议
- 建议：
```
