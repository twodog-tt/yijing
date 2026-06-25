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

### 1.2 奇门问事（简化学习版）（Phase F1 后端 ✅ · Phase F2 小程序 ✅）

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
| `POST` | `/api/v1/analysis/qimen` | 创建奇门分析记录并返回免费解读 | **Phase F1** ✅ |
| `GET` | `/api/v1/analysis/{id}` | 获取单条分析详情（含结构化 result） | **Phase E** 起（公开 handler） |
| `GET` | `/api/v1/analysis` | 按 session 分页列表，支持 `module=bazi|qimen` 筛选 | **Phase E1 / F1** ✅ |
| `DELETE` | `/api/v1/analysis/{id}` | 硬删除当前 session 拥有的分析记录 | **Phase E2** ✅ |
| `POST` | `/api/v1/analysis/{id}/unlock` | 解锁八字完整报告（模板 `full_content`；仅 `rewarded_video_mock`） | **Phase E5** ✅ |
| `GET` | `/api/v1/analysis/{id}/interpretation/free` | 免费解读（独立接口；当前免费内容已在 create/get 返回） | **未实现** |
| `GET` | `/api/v1/analysis/{id}/interpretation/full` | 完整解读独立 GET（未解锁 40301） | **未实现**；Phase E5 通过 unlock 直接返回 `full_content` |

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

**请求头：**

```text
X-Session-Key: …        # 匿名 session，不放 GET query、不放 URL
```

**请求体：**

```json
{
  "session_key": "...",
  "birth_date": "1990-05-20",
  "birth_hour_branch": "wu",
  "birth_hour_unknown": false,
  "confirm_disclaimer": true
}
```

| 字段 | 说明 |
|------|------|
| `birth_date` | 公历 `YYYY-MM-DD` |
| `birth_hour_branch` | 十二地支时辰：`zi`…`hai`；`birth_hour_unknown=true` 时可省略 |
| `birth_hour_unknown` | 为 `true` 时不生成时柱 |
| `confirm_disclaimer` | **必填**，必须为 `true`；不写入 `input_payload` |

**响应（示意）：**

```json
{
  "code": 0,
  "data": {
    "id": 1001,
    "module_type": 1,
    "algorithm_version": "bazi-simple-v1",
    "input_payload": { "birth_date": "1990-05-20", "birth_hour_branch": "wu", "calendar": "gregorian", "timezone": "Asia/Shanghai" },
    "result_payload": {
      "algorithm_version": "bazi-simple-v1",
      "method_note": "本功能采用简化干支文化规则，不等同于专业八字排盘。",
      "pillars": { "year": "庚午", "month": "辛巳", "day": "甲子", "hour": "甲子" },
      "day_master": "甲",
      "five_elements": { "wood": 2, "fire": 2, "earth": 1, "metal": 2, "water": 1 },
      "reflection_focus": "…",
      "action_suggestions": ["…"]
    },
    "free_content": "…",
    "unlock_status": 0,
    "generation_status": 1
  }
}
```

### 2.5 `POST /api/v1/analysis/qimen`（Phase F1 ✅）

**请求体：**

```json
{
  "session_key": "...",
  "question": "我最近适合推进这个计划吗？",
  "category": "career",
  "confirm_disclaimer": true
}
```

| 字段 | 说明 |
|------|------|
| `question` | 必填，4–120 字（Unicode 字符数） |
| `category` | 可选：`career` / `relationship` / `study` / `decision` / `general`；默认 `general` |
| `confirm_disclaimer` | **必填**，必须为 `true` |
| `session_key` | body 或 `X-Session-Key` header；二者同时存在且不一致 → 400 |

**Phase F1 明确未做：**

- 不接 DeepSeek / AI
- 不接 unlock / `full_content`
- 不生成完整九宫盘
- 不采集姓名、手机号、身份证、地址、性别、精确地理位置
- 不做军事、赌博、投资、医疗、法律具体建议

**`result_payload`（`qimen-simple-v1`）：** 含 `algorithm_version`、`method_note`、`question_summary`（非原问题全文）、`category`、`time_context`、`situation_overview`、`risk_observations`、`action_pacing`、`reflection_questions`、`action_suggestions`、`calculation_meta.limits`。

**列表隐私：** `GET /analysis?module=qimen` 不返回 payload / free_content；`question` 字段替换为安全摘要「用户问题已用于本次局势梳理」。

**unlock：** 奇门记录调用 `POST /analysis/{id}/unlock` → **403**（`analysis unlock not supported for this module`）。

起局时间默认服务端当前时间（Asia/Shanghai），不暴露复杂盘式参数。

### 2.6 `GET /api/v1/analysis` 查询参数（Phase E 起）

```text
X-Session-Key: …        # 请求头，不放 query
module=bazi|qimen       # 可选；Phase E1 支持 bazi；Phase F1 起支持 qimen
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

> **Phase E0 已确认**完整策略见 [§10.2 Phase E0：出生信息隐私与删除策略](#102-phase-e0出生信息隐私与删除策略)。

| 项 | 说明 |
|----|------|
| 保存目的 | 仅为生成该次八字简析结果与匿名 session 历史 |
| 采集范围 | 出生日期、出生时辰；**不**采集手机号、微信登录、真实姓名、身份证、精确地址、性别 |
| 绑定方式 | 数据绑定匿名 `session_key` 对应的 `session_id` |
| 保存期限 | **默认保存至用户主动删除**；Phase E 不做自动过期清理 |
| 日志 | **禁止**在日志中打印出生日期、出生时辰、`input_payload`、`result_payload` |
| 列表 API | **不返回**完整出生信息、完整 `input_payload`、`full_content` |
| 详情 API | 仅允许当前 session 查看自己的记录 |
| 鉴权传递 | `session_key` 经请求头传递，**不放 GET query** |
| 删除能力 | Phase E 提供用户删除；**硬删除**整条 `analysis_records` 记录 |

### 3.5 `input_payload` / `result_payload` 示例

**八字 `input_payload`（出生信息仅存于此，详情接口在 session 校验后返回）：**

```json
{
  "birth_date": "1990-05-20",
  "birth_hour_branch": "wu",
  "birth_hour_unknown": false,
  "calendar": "gregorian",
  "timezone": "Asia/Shanghai"
}
```

当 `birth_hour_unknown=true` 时，`input_payload` 不含 `birth_hour_branch`。

**八字 `result_payload`（Phase E1 实际结构；`bazi-simple-v1` ≠ 专业排盘）：**

```json
{
  "algorithm_version": "bazi-simple-v1",
  "method_note": "本功能采用简化干支文化规则，不等同于专业八字排盘。",
  "calculation_meta": {
    "limits": [
      "年柱按公历年份简化推算，未按立春换年",
      "月柱按公历月份固定映射月支，非节气月柱",
      "日柱按公历日期推算，未做真太阳时校正"
    ]
  },
  "pillars": {
    "year": "庚午",
    "month": "辛巳",
    "day": "甲子",
    "hour": "甲子"
  },
  "day_master": "甲",
  "five_elements": {
    "wood": 2,
    "fire": 2,
    "earth": 1,
    "metal": 2,
    "water": 1
  },
  "reflection_focus": "基于简化干支文化规则的学习参考：…",
  "action_suggestions": ["记录近期一件让你有感受的小事…"]
}
```

**E1 契约说明：**

- `result_payload` **不包含** `birth` 对象；出生日期/时辰仅在 `input_payload`。
- `birth_hour_unknown=true` 时，`pillars` **完全省略** `hour` key（不输出空字符串）。
- 列表接口 **不返回** `input_payload` / `result_payload` / `free_content`。

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
| `pages/bazi/bazi` | **Phase E3** ✅ | 八字表单：出生日期、时辰、免责声明 |
| `pages/analysis-result/analysis-result` | **Phase E3/E5/E6/E9（小程序）** ✅ | 八字结果页：免费简析、mock 解锁、微信分享、解锁后完整分享长图、删除 |
| `pages/qimen/qimen` | **Phase F2** ✅ | 奇门问事表单 + 最近记录 |
| `pages/qimen-result/qimen-result` | **Phase F2** ✅ | 奇门结果页（仅免费解读 + 删除） |
| 首页入口「八字简析」 | **Phase E3** ✅ | 稳文案，见 §1.1 |

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
- 八字/奇门：`GET /analysis?module=bazi`（Phase E1 已实现 `module=bazi`）
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
  service/bazi/                  # Phase E1 ✅ 计算 + Create + free_content
  service/qimen/                 # Phase F1 ✅
  handler/analysis.go            # Phase E1 ✅ bazi create / get / list
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
Phase E0：出生信息隐私与删除策略确认 ✅
         · 仅更新本文档，不写业务代码
         · 明确保存期限、删除方案、Phase E 开发约束
    ↓
Phase E1：八字简析 MVP 后端基础 API ✅
         · POST/GET /analysis/bazi 基础 API
         · bazi-simple-v1 计算 + 模板 free_content
         · 不接 AI / unlock / 小程序 / 奇门
    ↓
Phase E2：八字记录删除 API + 隐私删除闭环 ✅
         · DELETE /analysis/{id} 硬删除
         · 仅当前 session 可删自己的记录
    ↓
Phase E：八字简析 MVP（完整）
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

### 10.2 Phase E0：出生信息隐私与删除策略

**状态：** ✅ 已确认（文档-only，Phase E 开发前置条件已满足）

八字模块会采集出生日期、出生时辰等较敏感个人信息。Phase E 开始前，必须先明确保存期限与删除方案。本节为 **Phase E0 正式结论**，Phase E 实现须遵守。

#### 10.2.1 出生信息保存策略

当前 MVP 采用：

| 原则 | 说明 |
|------|------|
| 使用目的 | 出生日期、出生时辰 **仅用于** 生成八字简析报告 |
| 身份绑定 | 数据绑定匿名 `session_key` 对应的 `session_id` |
| 不采集手机号 | ✓ |
| 不接微信登录 | ✓ |
| 不采集真实姓名 | ✓ |
| 不采集身份证 | ✓ |
| 不采集精确地址 | ✓ |
| 不采集性别 | ✓ 八字 MVP 已移除 `gender` |
| 日志最小化 | **不在** 应用日志 / console 中打印出生日期、出生时辰、`input_payload`、`result_payload` |
| 列表接口 | **不返回** 完整出生信息 |
| 详情接口 | **仅允许** 当前 session 查看自己的记录 |
| 鉴权传递 | `session_key` **不放 GET query**；后续 analysis API 统一经 **请求头** 传递（如 `X-Session-Key`） |

#### 10.2.2 保存期限

MVP 采用：

```text
默认保存到用户主动删除，便于用户查看历史记录。
```

补充说明：

- 当前为 **匿名 session 历史**；同一设备/小程序内可回看本 session 生成的记录。
- 用户 **清理小程序缓存** 或丢失本地 `session_key` 后，**可能无法再访问** 旧匿名记录（服务端记录仍在，但客户端无法证明归属）。
- 后续可增加「自动清理超过 N 天记录」的配置项或定时任务；**Phase E 不做** 自动过期清理。

#### 10.2.3 删除方案

Phase E 目标删除策略：

```text
用户删除八字记录时，直接硬删除该 analysis_records 记录。
```

原因：

- 八字记录包含出生日期、出生时辰等敏感输入（存于 `input_payload`）。
- 仅软删除（`status=0`）仍会 **保留** `input_payload` 与解读正文，不符合隐私最小化。
- MVP 阶段 **硬删除** 实现更简单、删除效果更彻底。

**若后续因审计需要改为软删除**，必须在标记删除的同时 **清空或置 NULL** 以下字段，不得保留可还原的出生信息：

- `input_payload`
- `result_payload`
- `free_content`
- `full_content`
- `unlock_type`
- `ai_provider`

Phase E 须实现用户可触达的删除入口（如结果页 / 历史列表删除），后端执行物理删除。**Phase E2 已实现 `DELETE /api/v1/analysis/{id}`。**

#### 10.2.4 Phase E 开发约束

除 Phase D 已有 repository 收口外，Phase E **必须** 遵守：

| 约束 | 说明 |
|------|------|
| Handler 不得透传任意 JSON | handler **不得** 把客户端任意 JSON 原样传给 repository |
| 构造 `input_payload` | bazi service 必须从 **typed request** 构造 `input_payload` |
| 构造 `result_payload` | bazi service 必须从 **计算结果** 构造 `result_payload` |
| Repository 只收服务端数据 | repository 只接收服务端已校验、已构造的数据 |
| 日志 | **不允许** 在 console / log 中输出完整出生信息 |
| 错误信息 | **不允许** 在 API 错误响应中返回完整出生信息 |
| URL | **不允许** 把出生信息放到 URL query |
| 分享海报 | **不允许** 在分享海报中展示完整出生日期 / 时辰 |

qimen 模块（Phase G）沿用同一 handler → typed service → repository 模式。

#### 10.2.5 文案边界

八字简析对外表述须保持 **学习 / 倾向 / 反思** 定位，**非** 命运预测或改运服务：

| 允许 | 禁止 |
|------|------|
| 「基于传统干支文化的性格与五行倾向学习」 | 「精准算命」「一生命运」「婚姻财运预测」等 |
| `bazi-simple-v1` 简化示意排盘 | 宣称等于专业八字排盘 |
| 保守的性格倾向与行动提醒 | 大运、流年、合婚、疾病、寿命、精准财运、改运化解 |
| 时辰为 `unknown` 时省略时柱 | 伪造或猜测时柱 |

与 §1.1、§8、§11 一致：表单页与结果页须展示免责声明；AI 输出 schema 含 `disclaimer` 字段。

### 10.3 Phase E1 交付清单（已完成）

**已实现：**

- [x] `POST /api/v1/analysis/bazi` — 创建八字记录，返回 `result_payload` + `free_content`
- [x] `GET /api/v1/analysis/{id}` — 当前 session 可读详情（`X-Session-Key` 请求头）
- [x] `GET /api/v1/analysis?module=bazi` — 列表摘要（不含完整出生信息 / payload 大字段）
- [x] `service/bazi` — `bazi-simple-v1` 简化计算 + 模板免费解读
- [x] `AnalysisRepository.CreateWithFreeContent` — **一次 INSERT 原子写入** `input_payload` / `result_payload` / `free_content`，`generation_status=1`
- [x] `internal/pkg/sessionkey` — analysis GET 经 header 传 session，拒绝 query `session_key`
- [x] 单元测试：bazi service / repository / handler / golden fixtures

**Phase E1 Codex review 修复（v1.1）：**

- 创建流程改为：先计算 → 先生成 `free_content` → **一次 INSERT**；不再 Create 后 UpdateFreeContent
- `result_payload` **不重复**保存 `birth` 对象；出生输入仅在 `input_payload`
- `birth_hour_unknown=true` 时 `result_payload.pillars` **完全省略** `hour` key（不输出空字符串）
- 出生日期 trim 后标准化为 `YYYY-MM-DD`；拒绝晚于 Asia/Shanghai 当天；保留 1900–2100 范围
- POST 若 header/body 同时传不同 `session_key` → 参数错误；`session_key` 最长 64
- POST JSON 启用 `DisallowUnknownFields()`，拒绝 `gender` / `name` / `phone` 等未知字段
- POST 正式接收 `confirm_disclaimer`，必须为 `true`（不写入 `input_payload`）
- 未知 session 列表仍走统一分页校验（page/page_size 上限与默认值）
- 免费文案使用「简化干支示意」「自我反思与行动建议」；unknown 时辰说明「不生成时柱」

**Phase E1 明确未做（后续 Phase 补齐情况见 §10.7）：**

- [ ] `GET /analysis/{id}/interpretation/full` — 完整解读独立 GET（**仍未实现**；Phase E5 unlock 直接返回模板 `full_content`）
- [x] `POST /analysis/{id}/unlock` — mock 激励视频解锁（**Phase E5** ✅）
- [x] `DELETE /api/v1/analysis/{id}` — 用户硬删除（**Phase E2** ✅）
- [x] 小程序 `pages/bazi/bazi`、`pages/analysis-result/analysis-result`（**Phase E3** ✅）
- [x] 首页「八字简析」入口（**Phase E3** ✅）
- [ ] 奇门 `POST /analysis/qimen`
- [ ] DeepSeek / 完整解读 AI

**请求体（POST /analysis/bazi）：**

```json
{
  "session_key": "...",
  "birth_date": "1995-01-01",
  "birth_hour_branch": "zi",
  "birth_hour_unknown": false,
  "confirm_disclaimer": true
}
```

`session_key` 也可经 `X-Session-Key` 请求头传递；**header 与 body 同时存在时必须一致**，否则拒绝。GET / DELETE 类接口 **仅** header。`confirm_disclaimer` 必填且必须为 `true`。未知 JSON 字段（如 `gender`）直接参数错误。

### 10.4 Phase E2 交付清单（已完成）

**已实现：**

- [x] `DELETE /api/v1/analysis/{id}` — 当前匿名 session 硬删除自己的 `analysis_records`
- [x] `AnalysisRepository.DeleteOwnedByID` — SQL 必须包含 `id + session_id + status=1`
- [x] `analysis.Service.Delete` — 所有权删除封装
- [x] 删除成功返回统一 `{ code: 0, message: "ok" }`，**不返回**被删记录或出生信息
- [x] 删除后 `GET /analysis/{id}` → not found；列表不再出现该记录
- [x] 单元测试：repository / service / handler

**删除策略（与 §10.2.3 一致）：**

```sql
DELETE FROM analysis_records
WHERE id = ? AND session_id = ? AND status = 1
```

- **硬删除**，不可恢复
- 不保留 `input_payload` / `result_payload` / `free_content` / `full_content`
- `session_key` 仅经 `X-Session-Key` 请求头传递；query 中出现 `session_key` → 400
- 他人 session 或无权限 → not found（不暴露记录是否存在）

**隐私意义：** 删除 API 完成后，八字模块才具备向真实测试用户开放的**隐私删除闭环**基础条件；但仍不含 AI 完整解读、unlock、小程序页面、奇门。

### 10.5 Phase E3 交付清单（小程序八字页面 + API 联调，已完成）

**已实现：**

- [x] 小程序 `pages/bazi/bazi` — 出生日期、时辰 / 时辰未知、免责声明确认
- [x] 小程序 `pages/analysis-result/analysis-result` — 免费简析展示、删除记录
- [x] 首页「八字简析」入口卡片
- [x] `miniprogram/utils/api.js` — `createBaziAnalysis` / `getAnalysis` / `getAnalysisList` / `deleteAnalysis`
- [x] GET / DELETE 使用 `X-Session-Key` header；**不把** `session_key` 放入 query
- [x] 八字页内「最近记录」列表（方案 A）；列表不展示出生信息
- [x] 结果页删除闭环（确认 → DELETE → 返回）

**Phase E3 明确未做（Phase E5 已补齐 mock 解锁，见 §10.7）：**

- [ ] `GET /analysis/{id}/interpretation/full` — 完整解读独立 GET（**仍未实现**）
- [x] 广告解锁 / `POST /analysis/{id}/unlock`（**Phase E5** ✅）
- [ ] DeepSeek / AI
- [ ] 奇门
- [ ] 分享海报
- [ ] 六爻历史页合并八字记录

**小程序隐私约束（与 §10.2 一致）：**

- 列表 API 与八字页最近记录 **不展示**出生日期 / 时辰
- 结果页优先展示「出生信息已用于本次简析」，不展示具体出生日期
- 错误提示与日志 **不打印** `session_key`、出生日期、时辰、`input_payload`、`result_payload`

### 10.7 Phase E5 交付清单（八字完整报告 + mock 解锁，已完成）

**已实现：**

- [x] `POST /api/v1/analysis/{id}/unlock` — 仅 `unlock_type=rewarded_video_mock`
- [x] 模板生成 `full_content`（基于 `result_payload` + `free_content`，不接 DeepSeek / AI）
- [x] `AnalysisRepository.UnlockWithFullContent` — SQL 含 `id + session_id + status + unlock_status=0`
- [x] 已解锁重复调用返回现有 `full_content`，不重复 UPDATE
- [x] 非 bazi 模块返回 forbidden
- [x] GET 详情在已解锁时返回 `full_content`
- [x] 小程序结果页接入 rewarded video mock 解锁流程
- [x] `unlockAnalysis(id, { unlockType })` — POST + `X-Session-Key` header

**unlock 约束：**

- 仅允许 `rewarded_video_mock`
- 拒绝 `mock_button` / `mock_ad` / `rewarded_video` / `paid` / `admin`
- `session_key` 仅经 `X-Session-Key` header；query 出现 → 400
- 错误响应不含出生日期、时辰、`input_payload`、`result_payload`、`session_key`
- 日志不打印出生信息与 payload

**Phase E5 明确未做：**

- [x] DeepSeek / 真实 AI 完整解读（**Phase E7 已实现 DeepSeek + fallback**）
- [ ] 真实微信激励视频广告
- [ ] 付费解锁
- [ ] 奇门 unlock
- [ ] 独立 `analysis_unlock_record` 审计表
- [ ] `GET /analysis/{id}/interpretation/full` 分离接口（**仍未实现**；当前 unlock / GET 详情直接返回模板 `full_content`）

### 10.11 Phase E7 交付清单（八字 DeepSeek 完整报告，已完成）

**已实现：**

- [x] unlock 时优先生成 DeepSeek 完整报告（复用 DeepSeek Chat Completions 机制）
- [x] DeepSeek 失败 / 空内容 / 未配置 → 模板 fallback（`ai_provider=template_fallback`）
- [x] DeepSeek 成功 → `ai_provider=deepseek`
- [x] `UnlockWithFullContent` 同步写入 `ai_provider`（SQL 仍限定 `unlock_status=0`）
- [x] 已解锁重复 unlock 不重复调用 DeepSeek
- [x] Prompt 合规边界 + 隐私最小化（不发送完整出生日期）
- [x] 错误响应不暴露 DeepSeek 细节

**仍仅支持：**

- `unlock_type=rewarded_video_mock`（mock 视频，非真实广告）

**Phase E7 明确未做：**

- [ ] 真实微信激励视频
- [ ] 付费解锁
- [ ] 奇门 unlock
- [ ] 新 SQL
- [ ] 小程序 / Web 大改版

### 10.8 Phase E 交付清单（摘要）

- [x] `POST /analysis/bazi` + 模板 free_content（**Phase E1**）
- [x] 用户硬删除（**Phase E2**）
- [x] 小程序八字页面（**Phase E3**）
- [x] 模板 full_content + mock unlock（**Phase E5**）
- [ ] DeepSeek / AI 完整解读
- [ ] 真实广告 / 付费

> **Phase E6 命名说明：** **Phase E6（部署）** 指 ECS 发布 E5 unlock API；**Phase E6（小程序）** 指八字结果页本地分享卡片。二者独立交付，编号并列，避免与部署阶段混淆。

### 10.9 Phase E6（部署）交付清单（ECS 后端部署，已完成）

**部署目标：** 将 Phase E5 unlock API 发布到内测 ECS，供小程序 `http://123.57.48.214/api/v1` 联调。

**执行步骤（`/opt/yijing`）：**

```bash
git pull origin main
docker compose -f docker-compose.prod.yml --env-file .env build backend
docker compose -f docker-compose.prod.yml --env-file .env up -d backend
docker compose -f docker-compose.prod.yml --env-file .env exec -T backend ./migrate
```

**未改动：** 服务器 `.env`、Nginx、frontend 镜像、MySQL volume、安全组。

**远程验收（2026-06-25）：**

| 项 | 结果 |
|---|---|
| `GET /api/v1/health` | ok / db ok |
| `POST /analysis/bazi` | 创建成功 |
| `POST /analysis/{id}/unlock`（`rewarded_video_mock` + `X-Session-Key`） | 返回 `full_content` |
| `GET /analysis/{id}` 已解锁 | 含 `full_content` |
| 非法 `unlock_type` | 400 |
| query `session_key` | 400 |
| `DELETE /analysis/{id}` | 成功；再次 GET → 404 |

**小程序侧：** 微信开发者工具关闭合法域名校验后，可验收八字结果页 mock 解锁完整报告（见 `docs/miniprogram-dev.md` §15–§16）。

### 10.10 Phase E6（小程序）交付清单（八字结果分享卡片，已完成）

**已实现：**

- [x] 解锁完整报告后显示「生成结果卡片」（`unlock_status === 1`）
- [x] 未解锁不显示卡片入口，仅引导 mock 视频解锁
- [x] 本地 canvas 竖版摘要卡片（`components/bazi-share-card/`）
- [x] 保存到相册 + 权限引导
- [x] 卡片不含出生日期 / 时辰 / 完整报告全文 / session_key / payload / 小程序码
- [x] 生成中 / 保存中 / 删除中 / 解锁中互斥

**产品规则（Phase E6 修订）：**

- 结果卡片是**解锁完整报告后的权益**，未解锁不允许生成
- 生成卡片仍只展示摘要（干支示意、五行倾向、反思焦点、行动建议 1–2 条、免责声明）
- 不展示出生日期、出生时辰、完整报告全文、`session_key`、`input_payload` / `result_payload`

**操作区 UI（Phase E6 修订）：**

- 未解锁：单行主按钮「观看视频，解锁完整报告」+ 说明文案
- 已解锁：完整报告 → 「生成结果卡片」+ 摘要说明
- 底部弱化「删除记录」，不与主操作抢视觉

**明确未做：**

- [ ] 后端生成图片
- [ ] 小程序码
- [ ] 分享完整报告到卡片

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
| 出生信息保存期限 | 隐私政策文案 | **Phase E0 已确定**：保存至用户主动删除 |
| 删除历史/清除出生信息 | 用户数据删除方案 | **Phase E0 已确定；Phase E2 已实现 DELETE API** |

---

## 13. 文档版本

| 版本 | 说明 |
|------|------|
| Phase C v1 | 初始设计稿 |
| Phase D v1 | 底座-only 实现；Codex review 项回写 |
| Phase E0 v1 | 出生信息隐私与删除策略确认（文档-only） |
| Phase E1 v1 | 八字基础 API + bazi-simple-v1 + 模板 free_content |
| Phase E1 v1.1 | Codex review 修复：原子 INSERT、隐私收缩、session/分页/日期校验 |
| Phase E1 v1.2 | 契约对齐：`confirm_disclaimer`、`module=bazi`、`result_payload` 文档 |
| Phase E2 v1 | 八字记录硬删除 API + 隐私删除闭环 |
| Phase E3 v1 | 小程序八字页面 + API 联调（免费简析 + 删除；不含 unlock / AI / 奇门） |
| Phase E5 v1 | 八字模板完整报告 + `rewarded_video_mock` unlock（不接 DeepSeek / 真实广告） |
| Phase E6（部署）v1 | ECS 后端部署 Phase E5 unlock API（仅 rebuild backend，未改 `.env`） |
| Phase E6（小程序）v1 | 八字结果页本地 canvas 分享卡片（解锁后权益；摘要不含出生信息） |
| Phase E7 v1 | 八字 unlock DeepSeek 完整报告 + 模板 fallback |
| Phase E9 v1 | 八字分享给朋友 + 解锁后完整分享长图；卦象去除 mock 视频解锁 + 完整解析长图 |

---

### 10.12 Phase E9 交付清单（小程序长图分享 + 卦象解锁调整，已完成）

**八字（`pages/analysis-result`）：**

- [x] `onShareAppMessage`：通用标题；分享路径为 `/pages/bazi/bazi`（入口分享，因匿名 session 不同，朋友无法直达私有记录）
- [x] 手动带 `id` 打开失败时展示通用错误 + 返回八字页入口，不泄露记录归属
- [x] 解锁后「生成分享长图」（含免费解读 + `full_content`）
- [x] 超长长图：限制绘制行数 + 底部「仅展示前半部分」说明；高长图降低 pixelRatio
- [x] 长图不展示出生日期/时辰、session_key、payload、小程序码
- [x] 八字仍保留「观看视频，解锁完整报告」（`rewarded_video_mock`）

**卦象（`pages/result`）：**

- [x] 去除 mock 观看视频 UI 与 `rewarded-ad` 流程
- [x] 「查看完整解析」直接 `mock_button` unlock（后端已有支持，未改 backend）
- [x] 「生成解析长图」覆盖完整解析主要内容
- [x] 长图不展示完整原问题（仅分类 +「问事主题已用于本次解析」）

**Canvas：**

- [x] `utils/long-poster-canvas.js`：`computePosterDimensions`、`resolveExportPixelRatio`、动态高度长图工具
- [x] `bazi-share-card` / `share-poster` 升级完整长图；超长内容真实截断而非 silent crop

**明确未做：**

- [ ] 真实微信激励视频（卦象侧已去除 mock；八字侧仍保留 mock 解锁入口）
- [ ] 后端生成图片 / 小程序码 / AppSecret

---

| Phase F1 v1 | 奇门问事后端基础 API（`qimen-simple-v1` 模板免费解读；无 unlock / AI / 小程序） |

---

### 10.13 Phase F1 交付清单（奇门问事后端基础 API，已完成）

**已实现：**

- [x] `POST /api/v1/analysis/qimen` 创建奇门记录 + 模板 `free_content`
- [x] `GET /api/v1/analysis/{id}` 复用现有详情
- [x] `GET /api/v1/analysis?module=qimen` 分页列表（摘要不暴露完整原问题）
- [x] `DELETE /api/v1/analysis/{id}` 复用硬删除
- [x] `qimen-simple-v1` 简化学习版：局势梳理 / 风险观察 / 行动节奏 / 自我反思 / 行动建议
- [x] 风险拦截：静态关键词 + 复用 `sensitive.Service`（DB 敏感词）
- [x] 奇门 unlock forbidden（403）

**明确未做：**

- [x] 小程序奇门页面（Phase F2 已完成，见 §10.14）
- [ ] 奇门 unlock / DeepSeek 完整报告
- [ ] 完整九宫盘 UI / 专业排盘
- [ ] 军事、赌博、投资、医疗、法律具体建议

---

### 10.14 Phase F2 交付清单（小程序奇门问事页面，已完成）

**已实现：**

- [x] `pages/qimen/qimen` 表单页：question / category / confirm_disclaimer
- [x] `pages/qimen-result/qimen-result` 结果页：局势梳理、风险观察、行动节奏、反思、建议、免费解读
- [x] 首页「奇门问事」入口卡片
- [x] `createQimenAnalysis` / `getQimenAnalysisList` API 封装
- [x] 最近记录列表 + 详情跳转
- [x] 详情页删除（确认后返回奇门页）
- [x] 不展示完整原问题、session_key、原始 payload
- [x] 不显示 unlock / 视频 / 长图 / DeepSeek

**明确未做：**

- [ ] 奇门 unlock / DeepSeek 完整报告
- [ ] 完整九宫盘 UI / 专业排盘
- [ ] 长图分享

---

*Phase F1 后端基础 API 已交付。Phase F2 小程序奇门页面已交付。后续可接入 unlock / 长图 / 高级 UI。*
