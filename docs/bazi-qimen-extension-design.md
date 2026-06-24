# 八字 / 奇门扩展设计（Phase C）

> **文档性质：** 方案设计 + Phase D 底座实现说明  
> **项目：** 文易传统文化  
> **前置基线：** 六爻问事、今日一卦、起卦动画、免费/完整解读、分享海报、激励视频 mock 解锁、匿名历史  
> **当前背景：** ICP 备案审核中，可并行推进新模块设计与底座规划

---

## 0. 设计原则（全阶段适用）

1. **不把八字、奇门硬塞进 `divination_record`** — 六爻有 `lines` / 本卦 / 变卦，模型已稳定；八字为四柱五行，奇门为局盘信息，字段差异大。
2. **新模块优先使用通用报告表 `analysis_records`** — 解读、解锁、历史在上层抽象，底层按 `module_type` 区分。
3. **八字第一版只做简析** — 不做大运、流年、合婚等深度命理。
4. **奇门第一版先做方案（Phase F），MVP 在 Phase G** — 不做复杂九宫全盘 UI。
5. **文案统一边界** — 传统文化学习、趣味解读、自我反思、行动建议。
6. **禁止表述** — 精准预测、改命、化灾、保证发财/复合、医疗/法律/投资/赌博建议。
7. **不写入任何密钥** — AppSecret、DeepSeek Key、服务器密码、私钥、真实 adUnitId。
8. **不修改生产 `.env`** — 部署与环境变更单独走 review 流程。

与现有文档关系：

| 文档 | 关系 |
|------|------|
| [main-program-module-roadmap.md](./main-program-module-roadmap.md) | 产品优先级与模块矩阵 |
| [ai-agent-workflow.md](./ai-agent-workflow.md) | Cursor 实现 / Codex review 分工 |
| [miniprogram-dev.md](./miniprogram-dev.md) | 小程序开发与 Phase B 解锁 mock |

---

## 1. 模块定位

### 1.1 八字简析 MVP（Phase E 实现）

**定位：**

```text
基于出生时间的传统干支文化学习与性格/五行倾向解读工具
```

**用户价值：** 帮助用户从干支、五行角度整理自我认知，获得可反思的性格倾向与行动提醒，**非**「一生命运预测」。

**第一版包含：**

| 能力 | 说明 |
|------|------|
| 出生日期 | 公历日期（第一版） |
| 出生时辰 | 十二时辰选择；支持 `unknown` |
| 简化干支示意 | 年/月/日/（可选）时柱；**非**专业标准四柱排盘 |
| 日主 | 日柱天干（在可计算范围内） |
| 五行数量简析 | 金木水火土计数与倾向描述，非「旺衰断事」 |
| 性格倾向 | 模板 + AI，保守表述 |
| 行动建议 | 可执行、可反思的短建议 |
| 免费解读 | 短摘要，必得 |
| 完整解读 | 结构化 JSON 报告，锁定 |
| 广告 mock 解锁 | 复用 `rewarded_video_mock`，不接真实广告 |

**第一版明确不做：**

- 大运、流年、流月
- 合婚、子女、六亲断语
- 疾病、寿命、精准财运
- 改运、化解、做法事、佩戴开运物建议
- 农历输入、真太阳时、出生地经度校正（后续单独立项）

**小程序入口文案（稳）：**

```text
八字简析
基于传统干支文化的性格与五行倾向学习
```

**禁止入口/标题：**

```text
精准八字算命
一生命运
婚姻财运预测
```

---

### 1.2 奇门问事（简化学习版）（Phase F 方案 · Phase G 实现）

**定位：**

```text
基于问题与时间的奇门问事（简化学习版）学习工具，输出局势梳理、风险提醒和行动节奏
```

**用户价值：** 在用户提出自我反思类问题后，结合起局时间给出**局势梳理**与**行动节奏**参考，**非**吉凶保证或具体日期承诺。

**第一版包含（Phase G）：**

| 能力 | 说明 |
|------|------|
| 用户问题 | 敏感词拦截，禁止投资/赌博/医疗等 |
| 当前时间起局 | 默认服务器/客户端对齐的「当前时刻」 |
| 局势梳理 | 结构化摘要，保守表述 |
| 风险提醒 | 提醒用户注意盲区，非「必有灾祸」 |
| 行动节奏 | 近期行动整理建议，非具体吉凶日 |
| 免费解读 | 短摘要 |
| 完整解读 | 结构化 JSON |
| 广告 mock 解锁 | 同八字，复用 mock 适配层 |

**第一版明确不做：**

- 复杂九宫全盘交互 UI（神盘拖拽、逐宫点击等）
- 军事、赌博、投资具体建议
- 保证结果、定吉凶日期、择日承诺
- 与六爻混用 `divination_record`
- 高精度流派争议规则（第一版采用「简化起局 + 文档声明边界」）

**入口建议（Phase G 后）：**

```text
奇门问事
基于问题与时间的局势整理与行动节奏学习
```

---

## 2. API 设计草案

### 2.1 路由一览

| 方法 | 路径 | 说明 | 计划阶段 |
|------|------|------|----------|
| `POST` | `/api/v1/analysis/bazi` | 创建八字分析记录并返回免费解读 | **Phase E** |
| `POST` | `/api/v1/analysis/qimen` | 创建奇门分析记录并返回免费解读 | **Phase G** |
| `GET` | `/api/v1/analysis/{id}` | 获取单条分析详情（含结构化 result） | **Phase E** 起（公开 handler） |
| `GET` | `/api/v1/analysis` | 按 session 分页列表，支持 `module_type` 筛选 | **Phase E** 起 |
| `GET` | `/api/v1/analysis/{id}/interpretation/free` | 免费解读 | **Phase E**（八字）· **G**（奇门） |
| `GET` | `/api/v1/analysis/{id}/interpretation/full` | 完整解读（未解锁返回 40301） | **Phase E** · **G** |
| `POST` | `/api/v1/analysis/{id}/unlock` | 解锁完整解读 | **Phase E** 起 |

### 2.2 与现有六爻 API 的关系

**保持不变（不迁移）：**

```text
POST /api/v1/divinations
GET  /api/v1/divinations/{id}
GET  /api/v1/divinations
GET  /api/v1/divinations/{id}/interpretation/free
GET  /api/v1/divinations/{id}/interpretation/full
POST /api/v1/divinations/{id}/unlock
POST /api/v1/daily-fortune/today
```

六爻、今日运势 **继续** 使用 `divination_record` + 现有 handler，**不在 Phase D–E 做破坏性改动**。

### 2.3 通用约定

- **鉴权：** `session_key` 通过 **请求头** 传递（如 `X-Session-Key` 或项目统一 header），**不放 GET query**；POST body 仍可用于创建类接口。
- **响应 envelope：** `{ code, message, data }`，业务码沿用 `40301`（未解锁）等。
- **analysis 解锁（Phase E）：** 新 analysis 主流程 **仅允许** `rewarded_video_mock`；**不让** `mock_button` / `mock_ad` 进入 analysis 主流程。六爻现有 unlock 类型保持不变。
- **拒绝真实广告：** `rewarded_video` 在 analysis 阶段仍拒绝。**Unlock 实现延后到 Phase E**，Phase D 仅预留字段；Create 强制 `unlock_status=0`。
- **限流：** 创建类接口建议复用现有 `rateLimit` 中间件（Phase E/G 注册路由时启用）。
- **AI Provider / AI log：** **延后到 Phase E**；Phase D 仅预留 `ai_provider`、`generation_status` 字段，不新增 `analysis` AI log 表。

### 2.4 `POST /api/v1/analysis/bazi` 请求草案（Phase E）

```json
{
  "session_key": "uuid",
  "birth_date": "1990-05-20",
  "birth_hour_branch": "wu",
  "confirm_disclaimer": true
}
```

| 字段 | 说明 |
|------|------|
| `birth_date` | 公历 `YYYY-MM-DD` |
| `birth_hour_branch` | 十二地支时辰：`zi`…`hai`，或 `unknown` |
| `confirm_disclaimer` | 必须为 `true` |

**响应（示意）：**

```json
{
  "code": 0,
  "data": {
    "id": 1001,
    "module_type": 1,
    "result_payload": { "four_pillars": {}, "day_master": "甲", "five_elements_count": {} },
    "free_content": "…",
    "unlock_status": 0
  }
}
```

### 2.5 `POST /api/v1/analysis/qimen` 请求草案（Phase G）

```json
{
  "session_key": "uuid",
  "question": "近期工作节奏如何整理？",
  "category_id": 1,
  "confirm_disclaimer": true
}
```

可选：`occurred_at`（ISO8601，默认服务端当前时间）。第一版 **不** 暴露复杂盘式参数。

### 2.6 `GET /api/v1/analysis` 查询参数（Phase E 起）

```text
X-Session-Key: …        # 请求头，不放 query
module_type=1|2         # 可选，1=八字 2=奇门
page=1
page_size=20
```

历史页扩展时，六爻记录仍走 `/divinations`；八字/奇门走 `/analysis`，前端合并展示。

---

## 3. 数据表设计（Phase D 已实现）

### 3.1 设计决策

| 决策 | 说明 |
|------|------|
| 单表 `analysis_records` | 八字、奇门共用；六爻不迁入 |
| 解读字段内嵌 | `free_content` / `full_content` 存于本表 |
| 解锁审计 | Phase D **不**新增 `analysis_unlock_record`；Unlock 逻辑 **Phase E** 再实现 |
| AI 日志 | Phase D **不**扩展 `ai_generation_log`；**Phase E** 再定 |
| JSON 字段 | `input_payload` / `result_payload` 使用 MySQL `JSON` 类型 |
| 索引 | 仅复合索引，无单列低基数字段索引 |

### 3.2 表结构（已实现）

**文件：** `sql/007_analysis_records.sql`

```sql
CREATE TABLE IF NOT EXISTS analysis_records (
  id                 BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  session_id         BIGINT      NOT NULL COMMENT '关联 user_sessions.id',
  module_type        INT         NOT NULL COMMENT '模块类型：1=八字 2=奇门',
  algorithm_version  VARCHAR(32) NOT NULL COMMENT '分析规则版本，如 bazi-simple-v1',
  category_id        BIGINT      NULL COMMENT '事项类型ID，部分模块可为空',
  question           VARCHAR(500) NULL COMMENT '用户问题，八字模块为空',
  input_payload      JSON        NOT NULL COMMENT '服务端生成的输入快照JSON',
  result_payload     JSON        NULL COMMENT '非AI计算产生的结构化结果JSON',
  free_content       TEXT        NULL COMMENT '免费解读正文',
  full_content       MEDIUMTEXT  NULL COMMENT '完整解读正文',
  unlock_status      INT         NOT NULL DEFAULT 0 COMMENT '解锁状态：0=未解锁 1=已解锁',
  unlock_type        VARCHAR(32) NULL COMMENT '成功解锁方式',
  ai_provider        VARCHAR(32) NULL COMMENT '完整解读实际使用的AI提供方',
  generation_status  INT         NOT NULL DEFAULT 0 COMMENT '生成状态：0=待生成 1=免费完成 2=完整生成中 3=完整完成 4=免费失败 5=完整失败',
  status             INT         NOT NULL DEFAULT 1 COMMENT '记录状态：1=正常 0=删除',
  created_at         DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at         DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_analysis_session_status_created (session_id, status, created_at, id),
  KEY idx_analysis_session_module_status_created (session_id, module_type, status, created_at, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通用分析报告表';
```

### 3.3 常量约定（Go model，Phase D 已实现）

```text
ModuleTypeBazi   = 1
ModuleTypeQimen  = 2

AnalysisStatusActive  = 1
AnalysisStatusDeleted = 0

AnalysisUnlockStatusLocked   = 0
AnalysisUnlockStatusUnlocked = 1

AnalysisGenerationStatusPending     = 0
AnalysisGenerationStatusFreeDone    = 1
AnalysisGenerationStatusFullGenerating = 2
AnalysisGenerationStatusFullDone    = 3
AnalysisGenerationStatusFreeFailed  = 4
AnalysisGenerationStatusFullFailed  = 5
```

`ValidateModuleType()` 已实现：未知 module_type 拒绝。

### 3.4 出生信息隐私边界

| 项 | 说明 |
|----|------|
| 保存目的 | 仅为生成该次八字简析结果与匿名 session 历史 |
| 保存期限 | **Phase E 开始前必须确定**保存期限与删除历史/清除出生信息方案 |
| 日志 | **禁止**在日志中打印出生日期、出生时辰或完整 `input_payload` |
| 列表 API | **不返回**完整出生信息、完整 `input_payload`、`full_content` |
| 删除能力 | 后续需提供用户删除历史记录能力（Phase E+） |

### 3.5 `input_payload` / `result_payload` 示例

**八字 `input_payload`：**

```json
{
  "birth_date": "1990-05-20",
  "birth_hour_branch": "wu",
  "calendar": "gregorian",
  "timezone": "Asia/Shanghai"
}
```

**八字 `result_payload`（计算输出，非 AI；`bazi-simple-v1` ≠ 专业排盘）：**

```json
{
  "four_pillars": {
    "year": { "stem": "庚", "branch": "午" },
    "month": { "stem": "辛", "branch": "巳" },
    "day": { "stem": "甲", "branch": "子" }
  },
  "day_master": "甲",
  "five_elements_count": { "wood": 2, "fire": 2, "earth": 1, "metal": 2, "water": 1 },
  "calculation_meta": {
    "version": "bazi-simple-v1",
    "hour_unknown": true,
    "limits": ["gregorian_only", "no_true_solar_time", "no_major_luck", "simplified_stem_branch_demo"]
  }
}
```

当 `birth_hour_branch=unknown` 时：**不得生成或伪造时柱**；`four_pillars.hour` 应省略。

**奇门 `input_payload`（Phase G）：**

```json
{
  "question": "…",
  "occurred_at": "2026-06-23T14:30:00+08:00",
  "method": "time_qimen_simple_v1"
}
```

**奇门 `result_payload`（Phase G，示意）：**

```json
{
  "pan_summary": { "ju": "阳遁一局", "zhi_fu": "…", "zhi_shi": "…" },
  "key_palaces": [],
  "calculation_meta": {
    "version": "qimen-simple-v1",
    "limits": ["no_full_grid_ui", "conservative_wording"]
  }
}
```

### 3.6 与 `ai_generation_log` 的关系

**Phase D 不改动。** AI 日志扩展 **延后到 Phase E** 决策。

---

## 4. 八字 report schema（完整解读 JSON，Phase E）

存储于 `full_content`（TEXT/MEDIUMTEXT），结构与六爻 `FullReport` 类似，**独立 schema**。

```json
{
  "summary": "一段整体摘要，强调学习/reflection，不断言具体事件",
  "four_pillars": "简化干支文化示意（年/月/日/可选时），非标准专业排盘",
  "day_master": "日主说明与中性性格倾向",
  "five_elements": "五行数量与倾向，避免「缺什么必补什么」恐吓",
  "personality_tendency": "性格倾向描述，禁止标签化婚恋/疾病断言",
  "reflection_focus": "自我反思切入点（非预测）",
  "action_suggestions": ["建议一", "建议二"],
  "reflection_questions": ["你可以思考…", "…"],
  "disclaimer": "本内容仅供传统文化学习与自我反思参考，不构成预测或专业建议。"
}
```

**免费解读：** Plain text 或短 JSON 摘要，长度控制在现有六爻免费解读同级，**不**泄露完整报告字段。

**AI Prompt 要点（Phase E）：**

- 输入：`result_payload` + 用户可见字段
- 输出：严格 JSON schema
- 禁止：具体年份吉凶、婚恋结果、财运数字、改运方法

---

## 5. 奇门 report schema（完整解读 JSON，Phase G）

```json
{
  "summary": "局势一句话摘要",
  "situation": "当前局势梳理，偏描述性",
  "key_signal": "关键信号（简化），非断语",
  "risk_reminder": "风险提醒，避免恐吓",
  "action_pacing": "行动节奏建议，不承诺具体吉凶日",
  "action_suggestions": ["…", "…"],
  "reflection_questions": ["…", "…"],
  "disclaimer": "本内容仅供传统文化学习与自我反思参考，不构成预测或专业建议。"
}
```

Phase F 将细化：起局简化规则、Prompt 红线、与八字 schema 的 UI 复用点。

---

## 6. 小程序页面规划

### 6.1 页面清单

| 路径 | 阶段 | 说明 |
|------|------|------|
| `pages/bazi/bazi` | **Phase E** | 八字表单：出生日期、时辰、免责声明 |
| `pages/analysis-result/analysis-result` | **Phase E** | 通用结果页：`module_type` 区分展示；免费/完整解读、mock 解锁 |
| `pages/qimen/qimen` | **Phase G** | 奇门问事表单；Phase C–F **仅预留**路由与导航位 |
| 首页入口「八字简析」 | **Phase E** | 稳文案，见 §1.1 |

### 6.2 结果页复用策略

`analysis-result` 参数：`?id=1001&module=bazi|qimen`

| 模块 | 结构化展示区 |
|------|----------------|
| 八字 | 四柱卡片、日主、五行计数条 |
| 奇门 | 局势摘要、关键信号、风险提醒（无九宫格） |

**解锁（Phase E）：** 复用 `rewarded-ad.js`；analysis 主流程 **仅** `rewarded_video_mock`。

**分享海报：** Phase E **可不实现**；后续与六爻 poster 组件抽象 `module_type`。

### 6.3 历史记录扩展（Phase E 后期 / 与 Phase D 协调）

- 六爻：`GET /divinations`
- 八字/奇门：`GET /analysis?module_type=1|2`
- 前端 history 页合并列表，展示模块标签 + 日期 + 摘要

### 6.4 奇门页面（Phase C 仅预留）

- `app.json` 中 **暂不注册** `pages/qimen/qimen`，或在文档/注释中预留
- Phase F 输出 wireframe 与频次限制策略后再建页

---

## 7. 后端包结构（Phase D 已实现 / 后续扩展）

```text
backend/internal/
  model/analysis.go              # Phase D ✅
  repository/analysis.go         # Phase D ✅ Create / FindOwnedByID / ListBySession
  service/analysis/              # Phase D ✅ Get / List only
  service/bazi/                  # Phase E
  service/qimen/                 # Phase G
  handler/                       # Phase E 起注册公开路由
```

**Phase D 明确未做：** 公开 handler、Unlock、AI、八字/奇门算法。

**Unlock（Phase E）：** 新增 analysis unlock 编排，复用 `validateUnlockType` 思路，**不**修改现有 `unlock` 六爻服务。

---

## 8. 八字计算边界（Phase E 必读）

第一版采用 **「bazi-simple-v1 简化干支文化示意」**，**不等于**专业八字排盘。若简化算法不能正确计算标准四柱，**不得**使用「标准四柱排盘」表述；UI 与文档统一使用 **「八字简析」** 或 **「简化干支文化示意」**。

| 项 | Phase E 策略 |
|----|----------------|
| 历法 | **仅公历**输入；不做农历转换 |
| 真太阳时 | **不做** |
| 子时换日 | 采用 **固定规则并文档化**（实现前 Codex review 选定一种） |
| 月柱 | 按 **节气** 或 **简化固定分界** 二选一；若无法保证准确，必须声明误差 |
| 时柱 unknown | **不得生成或伪造时柱**；`result_payload` 省略 hour，`hour_unknown: true` |
| 精度 | **不确定的规则不假装准确**；标注 `calculation_meta.limits` |

**禁止：** 在规则未验证前输出「精准四柱 / 标准排盘」，或在 unknown 时辰下伪造时柱。

---

## 9. 奇门计算边界（Phase F/G 预告）

| 项 | 策略 |
|----|------|
| 起局 | 时间起局简化算法 `qimen-simple-v1` |
| UI | 文本 + 卡片，无交互盘 |
| 问题 | 敏感词 + 类目限制，拒绝投资/赌博/医疗 |
| 频次 | 建议每 session 每日上限（Phase F 定具体值） |

---

## 10. 实施顺序

```text
Phase C：本文档 — 八字/奇门扩展设计 ✅
    ↓
Phase D：analysis_records 通用报告底座（最小版）✅
         · sql/007_analysis_records.sql
         · model / repository / service（Get、List）
         · 不注册公开 API、不 Unlock、不 AI、不算法
         · 不影响现有六爻接口
    ↓  【建议 Codex review Phase D diff】
Phase E：八字简析 MVP
         · service/bazi 排盘（bazi-simple-v1）
         · POST /analysis/bazi
         · interpretation free/full
         · 小程序 bazi + analysis-result + 首页入口
    ↓
Phase F：奇门方案细化
         · 起局规则、Prompt、合规清单、频次策略
         · 仍可不写业务代码
    ↓
Phase G：奇门 MVP
         · POST /analysis/qimen
         · pages/qimen/qimen
         · analysis-result 奇门分支
```

### 10.1 Phase D 交付清单（已完成）

- [x] `sql/007_analysis_records.sql`
- [x] Go：`AnalysisRecord` model、`AnalysisListItem`、`ValidateModuleType`
- [x] Go：`analysis` repository（Create 强制默认 locked 状态、FindOwnedByID、ListBySession）
- [x] Create 仅允许服务端字段；`ValidateAlgorithmVersion` + JSON object 校验
- [x] Go：`analysis` service（Get、List）
- [x] 单元测试：model / repository（sqlmock）/ service
- [x] 文档：Codex review 项已回写本文档

**Phase D 明确未做：**

- 公开 handler / API 路由
- Unlock / AI / 八字·奇门算法
- `analysis_unlock_record` / analysis AI log 表
- 小程序页面 / 历史聚合
- 修改 divination / interpretation / unlock 现有服务

### 10.2 Phase E 前置原则

**Phase E 开始前必须确定：**

- 出生信息保存期限
- 删除历史 / 清除出生信息方案

**Phase E handler / service 约束：**

- Phase E handler **不得**把客户端任意 JSON 原样传给 repository。
- bazi / qimen service 必须从 **typed request** 构造 `input_payload` / `result_payload`，经校验后再调用 `AnalysisRepository.Create`。

### 10.3 Phase E 交付清单（摘要）

- [ ] `POST /analysis/bazi` + 免费/完整解读 + unlock
- [ ] 小程序 `pages/bazi/bazi`、`pages/analysis-result/analysis-result`
- [ ] 首页「八字简析」入口
- [ ] `go test ./...`
- [ ] 规则边界写入用户可见 disclaimer

---

## 11. 合规与文案检查清单

实施各 Phase 前抽检：

- [ ] 无「精准预测」「改命」「化灾」「保证发财/复合」
- [ ] 免责声明在表单与结果页可见
- [ ] 出生信息最小化采集，说明用途
- [ ] 解锁文案为「观看视频解锁完整解读」，非「看视频改运」
- [ ] AI 输出 schema 含 `disclaimer` 字段

---

## 12. 风险与未决项

| 项 | 说明 | 建议决策时机 |
|----|------|--------------|
| 八字月柱算法 | 节气 vs 简化表 | Phase E 前 Codex review |
| 子时换日规则 | 流派差异 | Phase E 实现前固定一种并文档化 |
| AI 日志表 | 是否扩展 `ai_generation_log` | **Phase E** |
| 独立 `analysis_unlock_record` | 是否对齐六爻审计 | **Phase E** |
| 历史页合并 | 前端一次拉两个 API vs 后端聚合 | Phase E UI 阶段 |
| 出生信息保存期限 | 隐私政策文案 | **Phase E 开始前必须确定** |
| 删除历史/清除出生信息 | 用户数据删除方案 | **Phase E 开始前必须确定** |

---

## 13. 文档版本

| 版本 | 说明 |
|------|------|
| Phase C v1 | 初始设计稿 |
| Phase D v1 | 底座-only 实现；Codex review 项回写 |

---

*Phase D 已完成底座最小版。建议将 Phase D diff 交给 Codex review 后再进入 Phase E。*
