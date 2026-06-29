# CODEX_PROJECT_CONTEXT

更新时间：2026-06-29
任务：HOME2：首页信息架构与视觉层级优化

本文件记录 Codex 对当前项目的上下文理解。后续任务开始前应先阅读本文件和根目录 `AGENTS.md`。

## 0. SECURITY1 排查结论

GitHub secret scanning 告警文件为 `docs/miniprogram-dev.md`，类型为 Tencent WeChat API App ID。排查结论：

- 告警内容为微信 AppID，不是 AppSecret。
- 已将公开文档中的真实 AppID 替换为 `<WECHAT_APP_ID>`。
- 未发现真实 AppSecret 或 `access_token` 泄露。
- 未发现真实 DeepSeek / OpenAI API key、数据库密码、服务器密钥或私钥泄露。
- 仓库存在 `session_key` 示例、DeepSeek 变量名和占位符，这些用于文档 / 测试说明，不是生产密钥。
- 本机 `miniprogram/project.config.json` 可能包含真实 AppID，但该文件被 `.gitignore` 忽略，不属于仓库提交内容；不要修改真实运行配置，除非维护者明确要求。

## 1. 项目当前定位

「文易传统文化」是微信小程序 + Go 后端项目，以传统文化模型为载体，提供学习参考、趣味解读、自我观察和行动节奏整理。

当前不是预测、改命、医疗、法律、投资或赌博建议产品。所有页面、报告、分享卡片、长图与 Prompt 都应使用合规边界文案。

当前体验版前代码与 API 自动化预检基本就绪，但体验版分发被备案、HTTPS API 域名、微信 request 合法域名和本地 DevTools / 真机 UI 勾选阻塞。

## 2. 用户侧模块说明

- 问事起卦：用户选择事项类型并输入 5-200 字问题，后端生成六爻、本卦、变卦、动爻、免费解读和完整解析。
- 感情关系观察：作为问事起卦的低优先级场景入口，通过 `/pages/ask/ask?scene=relationship` 切换表单文案和问题模板，复用问事创建 / 结果 / 解锁 / 分享 / 历史链路，不是新的第四个主模块。
- 今日一卦：同一匿名 session 同一本地日期只生成一条记录，复用六爻结果页与解读链路。
- 八字简析：采集公历出生日期、十二时辰或时辰未知，输出简化干支文化示意、五行倾向、反思焦点、行动建议、完整报告。
- 奇门问事：采集问题与分类，按当前时间生成局势梳理、风险观察、行动节奏、反思问题、行动建议、完整报告。
- 历史记录：统一展示问事 / 八字 / 奇门，可筛选、跳转详情和删除。
- 分享卡片 / 长图：小程序本地 Canvas 生成，不请求后端生成图片，不展示隐私字段或完整原始输入。
- 首页模块引导：HOME1 后首页强调三模块入口、场景选择和使用边界。

## 3. 后端 API 结构

基础 API：

- `GET /health`
- `GET /api/v1/health`
- `POST /api/v1/sessions`
- `GET /api/v1/categories`

六爻 / 今日一卦：

- `POST /api/v1/divinations`
- `GET /api/v1/divinations/{id}`
- `GET /api/v1/divinations?session_key=...`
- `DELETE /api/v1/divinations/{id}`，使用 `X-Session-Key`，软删除
- `GET /api/v1/divinations/{id}/interpretation/free`
- `GET /api/v1/divinations/{id}/interpretation/full`
- `POST /api/v1/divinations/{id}/unlock`
- `POST /api/v1/daily-fortune/today`

八字 / 奇门 analysis：

- `POST /api/v1/analysis/bazi`
- `POST /api/v1/analysis/qimen`
- `GET /api/v1/analysis/{id}`
- `GET /api/v1/analysis?module=bazi|qimen&page=1&page_size=20`
- `DELETE /api/v1/analysis/{id}`，硬删除 analysis 记录
- `POST /api/v1/analysis/{id}/unlock`

analysis 约定：

- GET / DELETE / UNLOCK 必须通过 `X-Session-Key` 传 session，不放 query。
- POST 可 body + header 传 session，但不一致会拒绝。
- 响应统一 `{ code, message, data }`。
- `analysis_records.module_type`：`1=八字`，`2=奇门`。
- `analysis_records.algorithm_version` 存算法版本。
- 解锁当前允许 `free_unlock` 和 `rewarded_video_mock`；AD0 后生产 UI 使用 `free_unlock`。

## 4. 小程序页面结构

`miniprogram/app.json` 当前注册页面：

- `pages/index/index`：首页模块引导
- `pages/today/today`：今日一卦
- `pages/ask/ask`：问事起卦表单
- `pages/result/result`：六爻 / 今日结果页
- `pages/history/history`：统一历史记录
- `pages/bazi/bazi`：八字简析表单和最近记录
- `pages/qimen/qimen`：奇门问事表单和最近记录
- `pages/qimen-result/qimen-result`：奇门结果页
- `pages/analysis-result/analysis-result`：八字结果页
- `pages/about/about`：关于与免责声明
- `pages/debug/debug`：开发调试页，正式环境禁用操作区

关键 utils / components：

- `miniprogram/utils/config.js`：API 与广告配置集中处。
- `miniprogram/utils/api.js`：API 封装。
- `miniprogram/utils/bazi.js`：八字 view、v2 条件展示、长图数据。
- `miniprogram/utils/qimen.js`：奇门 view、professional 条件展示、长图数据。
- `miniprogram/utils/home.js`：首页模块配置与合规文案。
- `miniprogram/utils/home.js`：LOVE1 后同时维护首页低优先级场景入口。
- `miniprogram/utils/long-poster-canvas.js`：长图摘要化与 Canvas 工具。
- `components/bazi-v2-panel/`：八字 v2 内部展示。
- `components/qimen-palace-grid/`：奇门 professional 九宫展示。
- `components/bazi-share-card/`、`components/qimen-share-card/`、`components/share-poster/`：分享长图。
- `components/home-module-card/`：首页三大主模块卡片；`components/home-guide-card/` 为 HOME1 引导组件，HOME2 首页当前不再注册。

## 5. 八字模块状态

默认算法：`bazi-simple-v1`。

内部灰度算法：`bazi-v2-poc`。

当前能力：

- `POST /analysis/bazi` 支持出生日期、时辰 / 时辰未知、免责声明。
- 普通小程序创建不传 `algorithm_version`，因此默认 `bazi-simple-v1`。
- 内部可传 `algorithm_version=bazi-v2-poc` 创建灰度记录。
- `bazi-v2-poc` 支持立春换年、节气月柱和 v2 payload 兼容字段，但仍是 POC，不是真太阳时、专业排盘、大运流年或神煞。
- BAZI1.4 后 `bazi-v2-poc` 完整报告使用 v2 专用 DeepSeek prompt 和 fallback 8 段结构，包含排盘口径、四柱观察、五行分布观察、行动节奏与边界声明。
- v2 prompt 使用 `calendar_basis`、`pillars_v2_summary`、`five_elements_summary`、`bazi_profile`、`interpretation_lens`、`method_note`、`limits`，但不输出完整出生日期、`session_key`、payload、raw JSON、内部 prompt 或密钥。
- BAZI1.4-DEPLOY-QA 已将 dev ECS backend 部署到 `29beebf`，公网 dev API 验证 v2 完整报告 8 段结构、节气口径、四柱、五行、未知时辰和 v1 回归均通过。
- TEST1.1 后 `scripts/check-api-smoke.sh` 固化 `bazi-v2-poc` 未知时辰检查：请求必须显式传 `birth_hour_unknown=true`，解锁后完整报告应包含未知时辰说明且不生成干支时柱。
- TEST1.1 后端 API 回归测试补充了八字 v2 未知时辰 create / unlock：create 响应不得回填时柱，unlock 完整报告不得伪造干支时柱。
- DOCS-SYNC1 后文档统一说明当前 smoke 期望为 `15 PASS / 0 FAIL`；旧 `13 PASS / 0 FAIL` 仅作为 TEST1 或 BAZI1.4-DEPLOY-QA 当时历史结果保留。
- 结果页仅在 `result_payload.algorithm_version === "bazi-v2-poc"` 时展示 v2 区块。
- 未知时辰不得生成或伪造时柱。
- 完整报告走 DeepSeek 优先，失败 fallback。
- AD0 后结果页按钮为“查看完整报告”，小程序生产 UI 不触发 mock 广告。
- 分享和长图不展示出生日期 / 出生时辰原始输入、`session_key`、payload。

当前测试记录线索：

- v2 正常：id=105，session=`bazi-v2-view-test`
- v1 默认：id=106，session=`bazi-v1-view-test`
- v2 未知时辰：id=107，session=`bazi-v2-unknown-test`

## 6. 奇门模块状态

默认算法：`qimen-simple-v1`。

内部灰度算法：

- `qimen-v2-poc`
- `qimen-v2-professional`

当前能力：

- `POST /analysis/qimen` 支持问题、分类、免责声明。
- 分类：`career` / `relationship` / `study` / `decision` / `general`。
- 普通小程序创建不传 `algorithm_version`，因此默认 `qimen-simple-v1`。
- 内部可传 `qimen-v2-poc` 或 `qimen-v2-professional` 创建灰度记录。
- `qimen-v2-poc` 有 9 宫 payload，但小程序 professional 九宫区块不触发。
- `qimen-v2-professional` 当前为第一版 professional 预览 / 落盘口径：二十四节气近似、拆补三元、九宫、值符值使、layout_version。
- 详情页仅在 `algorithm_version === "qimen-v2-professional"` 且 `palaces.length === 9` 时展示 professional 九宫。
- 完整报告走 DeepSeek 优先，失败 fallback；ALG2.7 后 professional 报告为 9 段结构并按 category 差异化。
- AD0 后结果页按钮为“查看完整报告”，使用 `free_unlock`。
- 分享和长图不展示完整原问题、`session_key`、payload，professional 长图只加一句摘要，不画完整九宫。

当前测试记录线索：

- professional：id=102，session=`qimen-devtools-prof`
- v1：id=103，session=`qimen-devtools-v1`
- poc：id=104，session=`qimen-devtools-poc`

## 7. 问事模块状态

当前能力：

- 问事起卦是 P0 核心流程。
- 后端按三枚硬币法生成真实 `lines`、本卦、变卦、动爻。
- 小程序起卦动画只读取后端返回结果，不在前端随机或计算卦象。
- 结果页展示基础卦象、免费解读、完整解析。
- AD0 后卦象结果页不再展示 mock 视频，按钮为“查看完整解析”，使用 `mock_button` unlock。
- 长图已摘要化，不贴完整解析全文，不展示完整原问题。
- 问事记录已支持 `DELETE /divinations/{id}`，历史页可删除。
- LOVE1 增加 `scene=relationship` 感情关系观察入口，只改小程序文案、模板和首页入口；提交 payload 仍为既有 `category_id` + `question`，不新增后端字段，不改变问事算法。
- 感情关系观察问题模板只填充输入框，用户可继续编辑，不自动提交。
- 感情关系观察边界文案强调传统文化学习参考与自我观察，不用于判断对方真实想法，也不替代现实沟通与判断。

## 8. 历史记录状态

H1 / H1.1 后历史页已经升级为统一入口：

- 筛选：全部 / 问事起卦 / 八字简析 / 奇门问事
- 并发加载 `GET /divinations`、`GET /analysis?module=bazi`、`GET /analysis?module=qimen`
- 按 `created_at` 倒序合并
- 点击跳转对应详情页
- 八字 / 奇门调用 `DELETE /analysis/{id}` 硬删除
- 问事起卦调用 `DELETE /divinations/{id}` 软删除
- 列表不展示完整原问题、出生信息、payload、session_key

## 9. 分享卡片 / 长图状态

当前状态：

- 首页、结果页支持微信原生分享。
- 六爻结果页分享路径为 `/pages/result/result?id=...`。
- 八字结果页分享路径回到 `/pages/bazi/bazi`，避免朋友直接打开私有记录。
- 奇门结果页分享路径为 `/pages/qimen-result/qimen-result?id=...`，非同 session 不应泄露记录内容。
- 长图均由本地 Canvas 绘制，不请求外部图片，不生成或伪造二维码 / 小程序码。
- SHARE1 / SHARE2 后八字、奇门、卦象长图都改为摘要 + 行动要点，不再粘贴完整报告全文。
- 长图禁止出现出生日期 / 时辰原始输入、完整原问题、`session_key`、payload、prompt、小程序码。
- 长图真实导出、相册权限和真机保存仍需维护者在微信 DevTools / 真机本地勾选。

## 10. 首页引导状态

HOME2 后首页结构：

- 顶部品牌区：文易传统文化，短说明“把问题整理清楚”，提供“开始问事 / 查看历史”入口。
- 三大主模块：问事起卦、八字简析、奇门问事，保持最高视觉层级。
- 常见场景：感情关系观察为唯一已上线场景入口，跳转 `/pages/ask/ask?scene=relationship`；事业选择、学习规划、人际沟通仅展示“规划中”。
- 传统文化小工具：梦境意象解析、姓名笔画观察、起名灵感助手、感情签仅为规划展示，不可点击，不新增页面或 API。
- 使用边界说明：不承诺具体结果、不替代现实决策、不提供投资医疗法律等专业建议。
- 保留历史记录、关于与说明入口。

HOME2 代码层完成后仍需维护者在微信 DevTools / 真机勾选首页渲染、relationship 跳转、规划项不可点击和窄屏不重叠。

## 11. 已完成阶段列表

重点已完成阶段：

- AD0：流量主未开通前隐藏 mock 广告，八字 / 奇门使用 `free_unlock`。
- F7 / E10 / RC1：奇门 / 八字差异化解读与相关回归线索已落文档。
- BAZI1 / BAZI1.3 / BAZI1.3-QA：八字 v2 内部记录展示与回归预检完成。
- ALG2 / ALG2.7 / ALG2.7-QA：奇门 v2 POC、professional 报告质量增强与回归完成。
- QIMEN-V2-VIEW：小程序 professional 九宫条件展示完成。
- MINIAPP-LOCAL-QA 代码层：测试数据与代码预检就绪，真实 UI 待本地勾选。
- HOME1 / HOME1-QA：首页模块引导与代码层 QA 完成。
- RELEASE-QA-PREP：体验版前全模块验收准备完成。
- RELEASE-QA：体验版前自动化最终验收通过，仍不上传体验版、不提审。
- MINIAPP-UX1：小程序加载态 / 错误态 / 空状态、防重复提交与相册权限失败提示优化完成。
- TEST1：release 前静态合规、隐私扫描、API smoke 脚本与 `docs/release-checklist.md` 完成。
- BAZI1.4：八字 `bazi-v2-poc` 报告质量增强完成，v2 专用 prompt、fallback 8 段、排盘口径与五行表达测试覆盖已补齐。
- BAZI1.4-DEPLOY-QA：dev ECS backend-only 部署到 `29beebf`，八字 v2 报告线上验证与 v1 / 奇门 smoke 回归完成。
- TEST1.1：release smoke 增加八字 v2 未知时辰 create / unlock 检查，release checklist 补充 `birth_hour_unknown=true` 口径。
- TEST1.1：后端 handler API 回归测试增加八字 v2 未知时辰 create / unlock 覆盖。
- DOCS-SYNC1：同步 TEST1.1 后测试文档状态，明确当前 `check-api-smoke.sh` 为 15 PASS / 0 FAIL。
- LOVE1：感情关系观察 MVP，作为问事起卦场景入口复用既有 API / 结果 / 历史 / 分享链路，不改后端、不改默认算法。
- LOVE1-QA：代码层验证首页入口、relationship scene、普通问事、八字 / 奇门入口、合规文案、分享 / 长图 / 历史隐私均符合预期；仍需维护者本地 DevTools / 真机勾选。
- UI1：小程序首页、问事结果页、问事分享长图的视觉层级与可读性优化完成；不改后端、Web、SQL、默认算法和隐私边界。
- HOME2：首页首屏、三大主模块、常见场景与传统文化小工具规划区重排完成；未上线工具不可点击，不改后端、Web、SQL、默认算法。

其他重要已完成阶段：

- Phase E1/E2/E3/E5/E7/E9：八字 API、删除、小程序页面、解锁、DeepSeek、长图。
- Phase F1/F2/F4/F5/F6：奇门 API、小程序页面、完整报告、解锁、长图。
- Phase H1/H1.1：统一历史记录与问事删除补齐。
- SHARE1 / SHARE2：八字 / 奇门 / 问事长图摘要化。
- REPORT1：完整报告质量增强。
- UX1 / UX2 / UX2.1：八字 / 奇门动效和九宫动效强化。
- MINIAPP-UX1：核心小程序路径状态体验统一，不改后端、SQL、默认算法。
- TEST1：回归脚本与验收命令固化，不改业务逻辑、不部署、不执行 SQL。
- TEST1.1：未知时辰 smoke 与 release checklist 收口，不改业务逻辑、不部署、不执行 SQL。
- BAZI1.4：八字 v2 完整报告结构化质量增强，不改默认算法、不改小程序、不执行 SQL。
- BAZI1.4-DEPLOY-QA：backend-only 部署并验证 dev API，不部署 frontend、不改小程序、不修改 `.env*`。
- LOVE1：小程序低风险场景入口，三模块结构保持不变，不新增广告 / 支付 / 登录。
- LOVE1-QA：仅记录 QA 结果，不改 backend / Web / SQL / deploy / `.env*`，不上传体验版、不提审。
- UI1：仅小程序 UI / 长图绘制与文档记录；首页主入口层级更清晰，结果页长问题改为独立问题卡，问事长图卡片换行和预览层级增强。
- HOME2：仅小程序首页 / about / 文档；首页新增规划中展示区但不新增真实功能、页面或 API。
- W1 / W2：Web 八字 / 奇门同步与 Web 首页对齐。

## 12. 当前最新 commit 线索

本次 HOME2 开始时：

- 当前分支：`main`
- 当前 HEAD：`c754d4a fix(frontend): match miniapp home layout`
- HOME2 不改 backend / Web / SQL / deploy / `.env*`，不新增真实功能模块，不改变默认算法

HOME2 代码层结论：

- 顶部品牌区改为短文案 + “开始问事 / 查看历史”，第一屏信息更轻。
- 三大主模块仍为问事起卦、八字简析、奇门问事，路径不变，普通创建流程不传内部算法字段。
- 感情关系观察仍位于 `HOME_SCENE_ITEMS`，路径为 `/pages/ask/ask?scene=relationship`，不是第四个主模块。
- 事业选择、学习规划、人际沟通和传统文化小工具仅显示“规划中”，不跳转、不新增页面或 API。
- about 页同步“当前已开放 / 规划中”说明。
- UI1 的问事结果页与长图可读性结论继续保留。

LOVE1-QA 代码层结论保留：

- 感情关系观察入口位于 `HOME_SCENE_ITEMS`，路径为 `/pages/ask/ask?scene=relationship`，不是第四个主模块。
- 问事页 `onLoad(options)` 读取 `scene=relationship` 后切换标题、说明、边界提示、placeholder 和 6 个模板；模板点击只填充 textarea，不自动提交。
- `createDivination` 仍只提交 `session_key`、`category_id`、`question`、`confirm_disclaimer`，不新增 `scene`、`algorithm_version` 或其他后端字段。

历史部署线索：

- ECS 部署前 HEAD：`8e47cdb feat(qimen): improve professional v2 report quality`
- ECS 部署后 HEAD：`29beebf feat(bazi): improve v2 report quality`
- 远端：`origin git@github.com:twodog-tt/yijing.git`
- 工作区清理后无 `.deploy-patches/`，不得重新提交该类临时补丁目录

近期 commit 线索：

- `3b1a3c4 test: add bazi unknown-hour API regression`
- `ad8da95 test: add bazi unknown-hour smoke`
- `5aa5092 docs(bazi): record v2 report deploy QA results`
- `29beebf feat(bazi): improve v2 report quality`
- `e9c048d test: add release regression checks`
- `a290ae4 docs: record TEST1 release checks`
- `0ed4ad0 docs: record miniapp UX improvements`
- `0a2ae16 feat(miniprogram): improve loading error and empty states`
- `d1caf4c docs: add Codex context and sanitize app id`
- `bcfbe59 docs(miniprogram): record release QA prep results`
- `17b9371 docs(miniprogram): record home guidance QA results`
- `94c138b feat(miniprogram): improve home module guidance`
- `d8a7e90 docs(miniprogram): record bazi v2 view QA results`
- `879784d feat(miniprogram): show bazi v2 analysis details`
- `cc883f7 docs(miniprogram): record local DevTools QA results`
- `38bef14 docs(miniprogram): record qimen professional view DevTools QA`

文档内历史线索：

- `a083882`：八字 v2 内部灰度接入。
- `8e47cdb`：奇门 professional 报告质量增强 QA 的 ECS 基线。
- `bcfbe59`：RELEASE-QA-PREP 基线。

BAZI1.4-DEPLOY-QA 验证线索：

- `bazi-v2-poc` 正常时辰：id=123，session=`bazi14-deploy-test`，8 段完整报告通过。
- `bazi-v2-poc` 未知时辰：id=125，session=`bazi14-unknown-test-2`，需传 `birth_hour_unknown=true`，不伪造时柱。
- `bazi-simple-v1` 回归：id=124，session=`bazi14-v1-regression`，不混入 v2 8 段结构。
- `scripts/check-api-smoke.sh`：13 PASS / 0 FAIL（BAZI1.4-DEPLOY-QA 当时结果），奇门 v1 / poc / professional 均通过。

TEST1.1 验证线索：

- `scripts/check-api-smoke.sh` 增加 `bazi-v2-poc unknown-hour` create / free_unlock。
- `backend/internal/handler/analysis_test.go` 增加八字 v2 未知时辰 create / unlock API 回归测试。
- smoke 不打印完整报告正文，只输出 id、algorithm_version、`birth_hour_unknown` 与 PASS / FAIL 摘要。
- 本阶段验证结果：`scripts/check-api-smoke.sh` 15 PASS / 0 FAIL。

DOCS-SYNC1 记录：

- 文档已同步当前 `check-api-smoke.sh` 期望结果：`15 PASS / 0 FAIL`。
- 历史 `13 PASS / 0 FAIL` 仅保留在 TEST1 / BAZI1.4-DEPLOY-QA 当时验收语境中，并注明后续 TEST1.1 已升级为 15 PASS。
- release checklist 明确：未知时辰 API 必须显式传 `birth_hour_unknown=true`，只省略 `birth_hour_branch` 会返回 400。

## 13. 当前阻塞项

当前已知阻塞项：

1. 微信 DevTools / 真机真实 UI 仍需维护者本地勾选。
2. 备案未完成。
3. HTTPS API 域名未完成。
4. 微信 request 合法域名未配置。
5. `https://api.wenyiapp.cn/api/v1/health` 当前不可用。
6. 当前 dev API 为 `http://123.57.48.214/api/v1`。

补充说明：

- Cursor / Codex 当前环境不能替代微信 DevTools 编译、预览、真机相册保存验证。
- DevTools 本地调试依赖关闭合法域名、TLS 和 HTTPS 校验。
- 上传体验版后真机请求可能被 HTTP IP 与未配置合法域名阻塞。
- 当前不建议上传体验版，不提审。

## 14. 体验版前置条件

体验版前必须完成：

- 维护者本地打开微信 DevTools，重新编译并勾选 RELEASE-QA 表格。
- 至少真机验证首页、问事、八字、奇门、历史、分享卡片、长图保存。
- `wenyiapp.cn` ICP 备案完成。
- `api.wenyiapp.cn` 可用并部署 HTTPS 证书。
- 微信公众平台配置 request 合法域名。
- 验证 `https://api.wenyiapp.cn/api/v1/health` 可用。
- 验证正式 API 全部核心接口。
- release 环境确认切换到 prod Base URL。
- 恢复合法域名、TLS 和 HTTPS 校验。
- 确认无广告 / 支付 / 微信登录入口。
- 确认无内部算法选择 UI。
- 确认默认算法仍为 `bazi-simple-v1` 和 `qimen-simple-v1`。

## 15. 后续建议阶段

建议优先级：

1. DOMAIN1：处理备案、HTTPS API 域名、微信 request 合法域名和 prod health。该阶段是体验版分发的主要阻塞。
2. RELEASE-QA-RECHECK：DOMAIN1 和维护者 DevTools / 真机勾选完成后，再做一次体验版前回归，不上传体验版前不要跳过。
3. BAZI1.5 / BAZI-QA：如继续推进八字质量，可基于 dev API 真实报告样本继续优化文案稳定性，但不是体验版分发阻塞。

当前不建议：

- 不进入 AD1，除非流量主、真实 adUnitId、审核与服务端校验策略都已准备。
- 不进入支付 / 会员 / 次数包阶段，除非产品和资质单独确认。
- 不把 `bazi-v2-poc` 或 `qimen-v2-professional` 改成普通用户默认。
- 不为了体验版临时切换到不可用的 `https://api.wenyiapp.cn/api/v1`。
