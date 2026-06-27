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
- [x] F1 阶段奇门 unlock forbidden（F5 已开放 mock 解锁）

**明确未做：**

- [x] 小程序奇门页面（Phase F2 已完成，见 §10.14）
- [x] 奇门 DeepSeek 完整报告生成能力（Phase F4，unlock 未开放）
- [x] 奇门 unlock mock 视频解锁（Phase F5）
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
- [x] 本阶段未含 unlock（Phase F5 已单独交付）

**明确未做（F2 当时）：**

- [ ] 完整九宫盘 UI / 专业排盘
- [ ] 长图分享（Phase F6）

---

### 10.15 Phase F4 交付清单（奇门完整报告 + DeepSeek fallback，已完成）

**已实现：**

- [x] `qimen.FullReportGenerator`：DeepSeek 优先生成，失败 fallback 模板
- [x] `ai_provider=deepseek` / `template_fallback` 记录
- [x] Prompt 强制使用安全 `question_summary` 常量；`free_content` 超长截断
- [x] 禁用词校验 + 空内容 fallback
- [x] `analysis.Service.GenerateFullReport` 路由 qimen / bazi 生成器
- [x] 八字 unlock / 六爻业务不受影响

**明确未做：**

- [ ] 奇门长图分享（Phase F6）

---

### 10.16 Phase F5 交付清单（奇门完整报告解锁，已完成）

**后端：**

- [x] `POST /api/v1/analysis/{id}/unlock` 支持 `module_type=qimen`
- [x] 仅允许 `rewarded_video_mock`
- [x] session 所有权 + SQL 条件更新
- [x] 已解锁重复调用不重复生成

**小程序：**

- [x] `qimen-result` mock 视频解锁 + 完整报告展示

**明确未做：**

- [x] 奇门长图分享（Phase F6，见 §10.17）
- [ ] 真实微信激励视频

---

### 10.17 Phase F6 交付清单（奇门结果长图分享，已完成）

**已实现：**

- [x] `qimen-share-card` 长图组件（Canvas 动态高度 + 超长截断）
- [x] 解锁后「生成分享长图」
- [x] `onShareAppMessage` 克制标题 + 仅 id 路径
- [x] 长图不含完整原问题 / session_key / payload / 小程序码

**明确未做（F6 当时）：**

- [ ] Web 端同步（Phase W1，见 §10.18）

---

### 10.18 Phase W1 交付清单（Web 端八字/奇门同步，已完成）

**Web 路由：**

| 路由 | 说明 |
|------|------|
| `/bazi` | 八字表单 + 最近记录 |
| `/qimen` | 奇门表单 + 最近记录 |
| `/analysis/[id]` | 八字/奇门结果页（按 `module_type` 分支） |

**已实现：**

- [x] `frontend/lib/api.ts`：`createBaziAnalysis` / `createQimenAnalysis` / `getAnalysis` / `getAnalysisList` / `deleteAnalysis` / `unlockAnalysis`
- [x] Analysis API 使用 `X-Session-Key` 请求头（非 URL query）
- [x] Web 解锁使用 `rewarded_video_mock`（内测模拟弹窗，无真实广告）
- [x] 八字：表单、免费解读、完整报告、删除记录
- [x] 奇门：表单、免费解读、完整报告、删除记录
- [x] 首页入口：八字简析 / 奇门问事
- [x] 隐私：不展示出生日期/完整原问题/`input_payload`/`result_payload` 原文；列表 subtitle 使用算法版本或安全摘要常量

**W1 自查修复：**

- [x] 结果页加载前先 `createSession`，避免未注册 session 误报「记录不存在」
- [x] 免费解读仅来自 view 层，移除 `record.free_content` 直出 fallback
- [x] 解锁弹窗重开时清空错误；删除/解锁互斥禁用
- [x] 表单页 `pageshow` 刷新最近记录（对齐小程序 `onShow`）

**W1 二次自审修复：**

- [x] `request()` 透传 `AbortError`，解锁超时显示正确文案（非「网络连接失败」）
- [x] 统一 `ensureSession()`：列表/详情/解锁/删除前均注册 session
- [x] `pageshow` 仅在 `event.persisted`（bfcache 恢复）时刷新，避免首屏重复请求
- [x] 列表 subtitle 增加 `safeAnalysisListSubtitle` 防御层，屏蔽意外泄露的原问题
- [x] 解锁进行中通过 `onLoadingChange` 禁用删除按钮
- [x] 已解锁但 `full_content` 为空时提供「重新获取完整报告」（Web + 小程序）

---

### 10.19 跨阶段（F4–F6 + W1）二次自审

**F4 奇门完整报告 + DeepSeek fallback**

| 检查项 | 结论 |
|--------|------|
| DeepSeek → template fallback 链路 | 通过（`go test ./internal/service/qimen/...`） |
| Prompt 强制 `QuestionSummary` 常量 | 通过 |
| `free_content` 320 rune 截断 | 通过 |
| 禁用词 / 空内容 fallback | 通过 |
| 八字 unlock 不受影响 | 通过 |

**F5 奇门完整报告解锁**

| 检查项 | 结论 | 修复 |
|--------|------|------|
| `rewarded_video_mock` only | 通过 | — |
| 重复 unlock 返回已有内容 | 通过 | — |
| **已解锁但 `full_content` 为空** | 原仅 API 返回、未持久化，刷新后仍空 | **已修复**：`UpdateUnlockedFullContent` + 小程序/Web「重新获取」 |

**F6 奇门长图分享**

| 检查项 | 结论 |
|--------|------|
| 长图不含原问题 / session / payload / 小程序码 | 通过 |
| 免费区用 `FREE_CONTENT_POSTER_NOTE` 非 raw `free_content` | 通过 |
| 未解锁不可生成长图 | 通过 |
| `fullContent` 为空时 block 生成 | 通过（F5 修复后更可恢复） |

**W1 Web 八字/奇门**

| 检查项 | 结论 |
|--------|------|
| 路由 `/bazi` `/qimen` `/analysis/[id]` | 通过 |
| `X-Session-Key` header | 通过 |
| 隐私字段不展示 | 通过 |
| `npm run build` + W1 eslint | 通过 |

**验证命令（均已通过）：**

- `go test ./...`
- `node --check`（qimen-result / analysis-result / qimen-share-card）
- `npm run build`（frontend）

---

**明确未做（W1 范围外）：**

- [ ] Web 长图分享（八字/奇门）
- [ ] 历史页合并六爻 + 八字 + 奇门（可后续 W1.1）
- [ ] 真实微信激励视频 / 登录 / 支付

---

### 10.20 Phase UX1 交付清单（小程序 + Web 八字/奇门轻量动效，已完成）

**范围：** 小程序 + Web UI 动效；**不改** backend / sql / deploy。

**小程序新增组件：**

- [x] `miniprogram/components/element-flow/` — 五行标签慢速流转（纯 CSS）
- [x] `miniprogram/components/qimen-grid/` — 九宫格线框装饰 + 光点（纯 CSS）
- [x] `miniprogram/components/result-reveal/` — 结果分段延迟渐入

**Web 新增组件：**

- [x] `frontend/components/motion/ElementFlow.tsx`
- [x] `frontend/components/motion/QimenGrid.tsx`
- [x] `frontend/components/motion/ResultReveal.tsx`
- [x] `frontend/app/globals.css` — 共享 animation 工具类 + `prefers-reduced-motion`

**八字 / 奇门（小程序 + Web 对齐）：**

- [x] 表单页：顶部氛围、`element-flow` / `qimen-grid`、提交呼吸 loading
- [x] 结果页：分段渐入、四柱/段落依次、五行高亮、解锁/完整报告淡入

**技术约束：**

- [x] 不引入 Lottie / 远程资源 / 额外 npm 动画库
- [x] 无频繁 timer + 状态抖动（分类 pulse 单次 timeout，卸载清理）
- [x] 不改变 unlock / 隐私边界 / 文案合规

**明确未做：**

- [ ] Lottie 复杂动效（UX2 评估后仍用手写 CSS，未引入 JSON）

---

### 10.22 Phase UX2 交付清单（可感知动效升级，已完成）

**参考方向（仅思路，未引入代码/资源）：**

- airbnb/lottie-web、wechat-miniprogram/lottie-miniprogram → 未引入（包体/license 风险）
- simeydotme/sparticles 等粒子库 → 未引入；采用 CSS 光点/扫描线
- streamich/awesome-css-animations → 参考 keyframes 思路，项目内手写

**新增组件：**

- [x] `element-orbit` / `ElementOrbit` — 五行环形流转 + 中心「五行流转 / 简化干支」
- [x] `qimen-scan-grid` / `QimenScanGrid` — 九宫扫描线 + 光点路径
- [x] `section-reveal` / `SectionReveal` — 分段上移淡入（80ms 步进）

**覆盖页面：** 小程序 bazi / analysis-result / qimen / qimen-result；Web `/` `/bazi` `/qimen` `/analysis/[id]`

**未改：** backend / sql / deploy / unlock / 隐私边界

---

*Phase UX2 可感知动效已交付。*

*Phase UX1 小程序 + Web 轻量动效已交付。*

### 10.21 Phase UX1 自审摘要

- **修复：** 四柱卡片移除 `result-reveal` 与 `pillar-emerge` 双重动画；小程序 / Web 均补充 `prefers-reduced-motion`；Web 解锁按钮光晕覆盖弹窗打开态。
- **验证：** `node --check`（小程序页面 + 组件）、`npm run build`（frontend）。
- **未改：** backend / sql / deploy / unlock 逻辑。

---

*Phase W1 Web 端同步已交付。*

---

### 10.23 Phase W2 交付清单（Web 首页对齐小程序，进行中）

**范围：** 仅 `frontend/` + 相关 `docs/`；**不改** backend / miniprogram / sql / deploy。

**对齐目标：** Web 首页移动端信息架构与小程序 `pages/index` 一致。

**文案：**

- [x] 顶部 eyebrow：易经问事
- [x] 主标题：从卦象中整理当下思路
- [x] 副标题：以传统文化学习和趣味解读为基础…
- [x] 四标签：传统文化学习 / 趣味卦象解读 / 自我反思整理 / 克制的行动建议
- [x] 底部免责声明（与小程序 index 一致）

**入口顺序（纵向 nav-link 风格）：**

1. 问事起卦（主按钮深色）
2. 八字简析（副标题 + compact `ElementOrbit`）
3. 奇门问事（副标题 + compact `QimenScanGrid`）
4. 今日一卦（淡金边框，非独立大卡片）
5. 历史记录
6. 关于与免责声明

**移除：** Web 首页独立「今日运势」大卡片。

**样式：** `frontend/app/globals.css` 新增 `.home-*` 类；页面背景 `#faf8f3`；桌面端居中窄卡片。

**未改：** API、unlock、DeepSeek、广告、支付、微信登录。

---

### 10.24 Phase UX2.1 交付清单（奇门九宫动效强化，进行中）

**范围：** `miniprogram/components/qimen-scan-grid`、奇门表单/结果页顶栏、Web `QimenScanGrid` + `/qimen` + `/analysis/[id]`；**不改** backend / sql / deploy。

**强化内容：**

- [x] hero 九宫 `168rpx`，居中于标题区
- [x] 横向扫描线 + 斜向光带（5.2s）
- [x] 光点沿九宫路径移动（中心定位 `translate(-50%,-50%)`）
- [x] 9 宫依次轻微高亮
- [x] 淡金渐变背景 + 径向光晕
- [x] Web 同步 + `prefers-reduced-motion` 降级

**未改：** unlock / 删除 / 长图 / API / 隐私边界

---

### 10.25 Phase AD0 交付清单（隐藏 mock 广告，内测 free_unlock）

**范围：** 小程序八字/奇门结果页 + Web analysis 解锁 + 后端 analysis unlock API；**不改** sql / deploy。

**产品变化：**

- [x] 八字/奇门未解锁按钮：「查看完整报告」
- [x] 直接调用 `unlock_type=free_unlock`，不走 `rewarded-ad.js`
- [x] Web analysis 同步 `free_unlock`，移除 mock 广告文案
- [x] 后端新增允许 `free_unlock`；`rewarded_video_mock` 保留 dev 兼容

**仍拒绝：** `rewarded_video` / `paid` / `admin` / analysis 路径的 `mock_button`

**未来：** Phase AD1 接入真实微信激励视频（需流量主 + adUnitId）

---

### 10.26 Phase F7 交付清单（奇门差异化解读优化）

**问题：** 不同问事在 free_content / 完整报告中相似度过高。

**根因：** qimen-simple-v1 主要按 category + 少量 hash 扰动生成；问题文本特征未进入结构化输出与 DeepSeek prompt。

**改动：**

- [x] 新增 `question_profile`（intent_type / time_horizon / decision_pressure / relation_scope / risk_tone）
- [x] 新增 `qimen_lens`（focus / support / caution / pacing theme）
- [x] 新增 `differentiation_seed` + `safe_question_summary`（不含原问题全文）
- [x] `Calculate` / `free_content` / 模板 full_content / DeepSeek prompt 均引用上述字段
- [x] 小程序 / Web 结果页新增「关注主题」卡片展示 lens 差异
- [x] 测试覆盖：同类不同问、不同 category、prompt 隐私、八字 / unlock 不受影响

**仍禁止：** 精准预测、强吉凶、改运、投资/医疗/法律/赌博/军事建议；列表/分享/长图不展示完整原问题。

---

### 10.27 Phase E10 交付清单（八字差异化与报告质量）

**问题：** 不同出生信息的 free_content / 完整报告相似度过高。

**根因：** 反思与行动建议模板固定；DeepSeek prompt 缺少 `bazi_profile` / `interpretation_lens`。

**改动：**

- [x] 新增 `bazi_profile`（日主观察、季节倾向、五行平衡类型、行动风格、反思主题）
- [x] 新增 `interpretation_lens`（可借助/需留意/节奏/五行关系）
- [x] `free_content` / 模板 full_content / DeepSeek prompt 引用上述字段
- [x] 小程序 / Web 结果页新增「解读视角」卡片
- [x] 未知时辰仍不生成时柱；result_payload 不含 birth_date

---

### 10.28 Phase ALG1 交付清单（八字 v2 POC · 后端算法）

**目标：** 在 **不复制 GitHub 源码、不引入 AGPL、不新增外部依赖** 的前提下，自研 Go 实现 `bazi-v2-poc` 最小算法层，与 `bazi-simple-v1` 并存。

**范围（仅后端 POC）：**

- [x] 立春换年（年柱不再按公历 1 月 1 日切换）
- [x] 节气月柱（十二「节」切换月令，年上起月法）
- [x] 日柱沿用 v1 固定基准日算法（已补充 golden 说明）
- [x] 时柱沿用 v1 十二时辰 + 五鼠遁；未知时辰不生成时柱
- [ ] 真太阳时（延后 **ALG1.1**）
- [ ] 大运 / 流年 / 神煞 / 格局 / 旺衰强断（明确不做）

**实现位置：**

- `backend/internal/service/bazi/calendar/` — 节气近似时刻、立春换年、月令
- `backend/internal/service/bazi/calculate_v2.go` — `CalculateV2()` + `ResultPayload()`（`algorithm_version=bazi-v2-poc`）
- `bazi-simple-v1` 的 `Calculate()` **未改默认行为**

**默认策略：** 线上 Create API **默认仍使用** `bazi-simple-v1`；可通过请求参数 `algorithm_version=bazi-v2-poc` 做内部灰度测试。

**ALG1.1（已完成）：** 扩展节气边界 golden tests（立春/惊蛰/清明/小暑/立秋/白露/寒露/立冬/大雪/小寒）；节令时刻仍为世纪万年历公式 + 本地正午近似，非天文台精确时刻。

**ALG1.2（已完成）：** `POST /api/v1/analysis/bazi` 支持可选 `algorithm_version`（`bazi-simple-v1` | `bazi-v2-poc`）；默认 v1；非法值返回参数错误；v2 result_payload 补齐 `pillars` / `pillars_v2` / `calendar_basis` / 差异化字段以兼容现有结果页；free_unlock / DeepSeek / template fallback 均支持 v2。

**参考来源（仅规则/能力矩阵，未复制代码）：** `bazi-skill`（MIT 规则文档）、taibu-core 能力对照（MIT，未引入 npm 包）。

**产品边界：** 仅供传统文化学习参考，不承诺专业排盘准确性；result_payload 不含完整 birth_date / session_key / prompt。

---

### 10.29 Phase REPORT1 交付清单（八字 / 奇门完整报告质量增强）

**目标：** 提升 `full_content` 具体性、差异化与结构化字段引用；**不改** 大运/流年/神煞/奇门完整排盘算法。

**八字完整报告（7 段固定结构）：**

- [x] 一、简要说明 → 七、边界声明
- [x] DeepSeek prompt 与 template fallback 均引用 `bazi_profile` / `interpretation_lens` / `five_elements` / `calendar_basis`（v2）
- [x] 体现 `element_balance_type` / `action_style` / `reflection_theme` 差异
- [x] 未知时辰不生成时柱；v2 说明节气口径为公式近似
- [x] 禁用词检测排除「边界声明」正文；prompt 不含 birth_date / session_key / 原始 payload

**奇门完整报告（7 段固定结构）：**

- [x] 一、问题局势摘要 → 七、边界声明
- [x] DeepSeek prompt 与 template fallback 均引用 `question_profile` / `qimen_lens` / category
- [x] career / relationship / study / decision / general 各有写作重点
- [x] 不输出完整原问题；prompt 使用 `safe_question_summary`

**测试：** 5 组八字样例 + 5 组奇门样例差异化；禁用词 body 检测；`go test` / `go vet` 通过。

**部署：** 仅 backend；无需 frontend / 小程序重编译 / SQL。

---

### 10.30 Phase SHARE1 交付清单（小程序长图分享内容优化）

**目标：** 八字 / 奇门分享长图从「字段罗列 + 完整报告粘贴」优化为「高价值摘要 + 行动要点」。

**八字长图：**

- [x] 展示四柱示意（未知时辰不展示时柱）
- [x] 展示解读视角（五行倾向 / 行动风格 / 反思主题）
- [x] 展示节奏与留意（interpretation_lens）
- [x] 展示 2–3 条行动要点
- [x] 移除完整报告全文粘贴

**奇门长图：**

- [x] 展示分类 + 安全问事摘要（不含完整原问题）
- [x] 展示关注主题 / 可借助 / 需留意 / 行动节奏（qimen_lens）
- [x] 按 category 突出本类关注重点
- [x] 展示 2–3 条反思或行动要点
- [x] 移除完整报告全文粘贴

**隐私与合规：** 不含 birth_date / 完整原问题 / session_key / payload；过滤强预测/改运/投资医疗法律等禁用词。

**部署：** 仅小程序重新编译；无需 backend / frontend / SQL。

---

### 10.30.1 Phase SHARE2 交付清单（问事起卦长图摘要化）

**目标：** 问事起卦 / 卦象分享长图与 SHARE1 对齐，改为摘要 + 行动提醒，不贴完整解析全文。

- [x] 长图结构：卦象概览 / 局势摘要 / 变化观察 / 行动提醒 / 自我反思 / 免责声明
- [x] 展示本卦 / 变卦（无变卦显示「无明显变卦」）与动爻提示
- [x] 不展示完整原问题（使用「用户问题已用于本次卦象梳理」）
- [x] 不贴 `full_content` / 完整解析全文
- [x] 复用 `long-poster-canvas.js` 抽取工具；新增 `divination.js` 构建摘要数据
- [x] 不改 backend / frontend / SQL

**部署：** 仅小程序重新编译；无需 backend / frontend / SQL。

---

### 10.31 Phase H1 交付清单（统一历史记录页）

**目标：** 小程序「历史记录」页升级为问事 / 八字 / 奇门统一入口。

- [x] 顶部筛选：全部 / 问事起卦 / 八字简析 / 奇门问事
- [x] 并发加载三类记录，按 `created_at` 倒序合并
- [x] 点击跳转 `result` / `analysis-result` / `qimen-result`
- [x] 八字 / 奇门支持 DELETE `/api/v1/analysis/{id}` 删除
- [x] 问事记录暂无 DELETE API，历史页不提供删除按钮

**H1.1 更新：**

- [x] 后端新增 `DELETE /api/v1/divinations/{id}`（软删除）
- [x] 历史页问事起卦支持删除

**部署：** 需部署 backend；小程序重新编译；无需 frontend / SQL。

---

### 10.32 Phase BAZI1.2 交付清单（八字 v2 灰度接入）

**目标：** 后端创建八字分析时支持内部选择 `bazi-v2-poc`，默认仍为 `bazi-simple-v1`。

- [x] `POST /api/v1/analysis/bazi` 可选 `algorithm_version`（`bazi-simple-v1` | `bazi-v2-poc`）
- [x] 省略字段 → 默认 `bazi-simple-v1` / `Calculate()`
- [x] `bazi-v2-poc` → `CalculateV2()` + 兼容 `result_payload`（含 `pillars_v2` / `calendar_basis`）
- [x] 非法 `algorithm_version` → 400 参数错误，不 fallback
- [x] `free_unlock` / DeepSeek / template fallback 均支持 v2
- [x] 小程序 / Web **暂不暴露**算法选择 UI
- [x] 不改 SQL / frontend / miniprogram / qimen

**POC 边界：** 节气时刻为公式近似；真太阳时、大运、流年、神煞不在本阶段。

**实现 commit：** `a083882`（`feat(bazi): support v2 algorithm selection for internal rollout`）

**部署：** 随 backend 发布；无需 SQL / frontend / 小程序重编译。

---

### 10.33 Phase ALG2 交付清单（奇门 v2 POC · 后端算法）

**目标：** 在 **不复制 GitHub 源码、不引入 AGPL、不新增外部依赖** 的前提下，自研 Go 实现 `qimen-v2-poc` 最小九宫排盘 POC，与 `qimen-simple-v1` 并存。

**范围（仅后端 POC）：**

- [x] 结构化九宫盘 `palaces`（9 宫：名称 / 天盘地盘干 / 九星 / 八门 / 八神占位）
- [x] `calendar_basis`：节令名称 + 公式近似口径说明
- [x] `dun`：冬至后阳遁 / 夏至后阴遁简化口径 + POC 局数
- [x] `xun` / `chief`：旬首、空亡、值符值使最小占位
- [x] `method_note` / `limits`：明确学习观察边界，不作强预测
- [ ] 专业完整起局 / 转盘算法（延后 **ALG2.3**）
- [ ] API 灰度接入（延后 **ALG2.2**）
- [ ] 大运 / 流年 / 神煞 / 应期 / 强吉凶（明确不做）

**ALG2.1（已完成）：** 扩展 golden fixtures（10 组 + 冬至/夏至边界 + 结构完整性 + 合规）；文档化 POC 与专业口径差异；**仍不接 API**。

**实现位置：**

- `backend/internal/service/qimen/calculate_v2.go` — `CalculateV2()` + `ResultPayload()`（`algorithm_version=qimen-v2-poc`）
- `backend/internal/service/qimen/calendar.go` — 节令近似、阴阳遁、局数 POC
- `backend/internal/service/qimen/palace.go` — 九宫结构 builder
- `backend/internal/service/qimen/algorithm_version.go` — 版本常量与 limits
- `qimen-simple-v1` 的 `Calculate()` **未改默认行为**；Create API **默认仍走 v1**

**默认策略：** 线上 `POST /analysis/qimen` **不接 v2**；仅供后端测试与后续 ALG2.2 灰度评估。

**POC 边界：** 节令时刻复用八字 calendar 公式近似；阴阳遁 / 局数为稳定可测 POC 规则，非专业奇门排盘。

**产品边界：** 仅供传统文化学习与结构化观察；result_payload 不含完整原问题 / session_key / prompt。

---

### 10.34 Phase ALG2.1 交付清单（奇门 v2 POC 口径审计与 golden fixtures）

**目标：** 在 **不改 API / 不改小程序 / 不部署** 前提下，扩展 `qimen-v2-poc` golden fixtures 与边界测试，并文档化当前 POC 口径与专业口径差异。

**范围（仅后端测试与文档）：**

- [x] 10 组 golden fixtures（含立春前、惊蛰、夏至日/次日、小暑、白露、冬至前/冬至日、小寒、次年夏至）
- [x] 冬至 POC 边界：12/21 阴遁 → 12/22 阳遁
- [x] 夏至 POC 边界：6/20 阳遁 → 6/21 阴遁（公历近似，非节气时刻）
- [x] 九宫结构完整性：`index` / `name` / 干 / 星 / 门 / 神 / `summary`；中五宫门为 `—`
- [x] payload 完整性 + 合规禁用词扫描（limits 除外）
- [x] 同一输入多次输出完全一致；category 影响局数 / 旬首 / 宫位轮转
- [ ] API 灰度（延后 **ALG2.2**）
- [ ] 专业转盘起局（延后 **ALG2.3**）

#### 10.34.1 当前 POC 口径审计（与专业口径差异）

| 维度 | 当前 POC 口径 | 专业口径（未实现） | 标注 |
|------|--------------|-------------------|------|
| **节令** | 复用 `bazi/calendar` 十二「节」公式近似时刻；取不晚于问事时刻的最近一节名称 | 精确节气时刻（可含真太阳时、交节秒级） | **POC 近似** |
| **阴阳遁** | 公历简化：12/22 起阳遁至次年 6/20；6/21 起阴遁至 12/21 | 以冬至/夏至**节气交节时刻**切换，并区分超神/接气等 | **POC 近似** |
| **局数** | `hash(RFC3339时间 + category + 阴阳遁) % 9 + 1`，稳定可测 | 拆补 / 置闰 / 茅山等流派按节与符头推算 | **POC 占位规则** |
| **旬首 / 空亡** | 六旬首列表 + 固定空亡表，按日期与 category hash 选取 | 按时家奇门符头、日干支推算旬首与空亡 | **POC 占位** |
| **值符 / 值使** | 默认天禽/开门；优先取「局数对应宫位」的星/门 | 转盘后随旬首、时干、遁局确定 | **POC 占位** |
| **九星 / 八门 / 八神** | 固定九星八门八神名表；按局数 + category + 日期 hash 轮转 | 转盘排布，随遁局、天禽寄宫规则变化 | **POC 占位** |
| **天盘干 / 地盘干** | 十天干表按 `(idx+局数+rotate)` 取模轮转 | 依旬首、时干、转盘规则飞布 | **POC 占位** |
| **中五宫** | 门字段固定 `—`；星神仍参与轮转 | 天禽寄宫（坤二或艮八等流派差异） | **POC 简化** |

**哪些仍是 POC：** 节令时刻、阴阳遁切换日、局数、旬首空亡、值符值使、星门神干排布——均为稳定可测近似/占位，**不是**专业完整奇门排盘。

**哪些后续需专业校准（ALG2.3+）：** 节气交节精确时刻、拆补/置闰局数、符头旬首、转盘九星八门八神飞布、天禽寄宫、真太阳时、时家/日家流派选择。

**默认策略：** ALG2.1 阶段仅后端 POC；**ALG2.2 起** Create API 支持内部可选 `qimen-v2-poc`（默认仍为 v1）。

**ALG2.2（已完成）：** `POST /api/v1/analysis/qimen` 支持可选 `algorithm_version`（`qimen-simple-v1` | `qimen-v2-poc`）；默认 v1；非法值返回 400；v2 result_payload 合并 v1 解读字段 + 九宫结构；free_unlock / DeepSeek / template fallback 均支持 v2；小程序 / Web **暂不暴露**算法选择 UI。

---

### 10.35 Phase ALG2.2 交付清单（奇门 v2 API 灰度接入）

**目标：** 参照 BAZI1.2，后端创建奇门分析时支持内部选择 `qimen-v2-poc`，默认仍为 `qimen-simple-v1`。

- [x] `POST /api/v1/analysis/qimen` 可选 `algorithm_version`（`qimen-simple-v1` | `qimen-v2-poc`）
- [x] 省略字段 → 默认 `qimen-simple-v1` / `Calculate()`
- [x] `qimen-v2-poc` → `CalculateV2()` + `BuildV2APIResultPayload()`（含 v1 解读字段 + 9 宫）
- [x] 非法 `algorithm_version` → 400 参数错误，不 fallback
- [x] `free_unlock` / DeepSeek prompt / template fallback 均支持 v2 payload
- [x] 小程序 / Web **暂不暴露**算法选择 UI
- [x] 不改 SQL / frontend / miniprogram

**请求示例（内部测试）：**

```json
{
  "question": "我最近适合推进这个项目吗？",
  "category": "career",
  "confirm_disclaimer": true,
  "algorithm_version": "qimen-v2-poc"
}
```

**POC 边界：** v2 仍为 POC 近似排盘，非专业完整起局；专业校准延后 **ALG2.3**。

**部署：** 仅 backend；无需 SQL / frontend / 小程序重编译。

---

### 10.36 Phase QIMEN-REPORT2 交付清单（奇门 v2 完整报告增强）

**目标：** 增强 `qimen-v2-poc` 完整报告（DeepSeek prompt + template fallback），更充分使用九宫 payload；**不改** v2 排盘算法、小程序 / Web / SQL。

- [x] v2 fallback 固定 9 段结构（局势摘要 → 边界声明）
- [x] 引用 2–3 重点宫位 + dun/ju/chief + calendar_basis POC 说明
- [x] DeepSeek prompt 使用 `palaces_summary` / `focus_palaces_summary`（非原始 JSON）
- [x] `qimen-simple-v1` 7 段报告 **不受影响**
- [x] 默认 API 仍为 `qimen-simple-v1`；小程序 / Web 暂不展示算法选择
- [ ] 专业排盘口径增强（延后 **ALG2.4** 实现；**ALG2.3-SPEC** 已完成设计）

**实现位置：** `backend/internal/service/qimen/v2_report.go`、`full_content.go`、`prompt_input.go`、`deepseek_full.go`

**部署：** 仅 backend；无需 SQL / frontend / 小程序重编译。

---

### 10.37 Phase ALG2.3-SPEC 交付清单（奇门 v2 专业口径设计与 fixtures 规划）

**目标：** 在 **不接 API、不改小程序 / Web、不部署、不替换 qimen-v2-poc / qimen-simple-v1** 前提下，设计 `qimen-v2-professional` 目标数据结构与 golden fixtures 元数据表，为 **ALG2.4** 第一批专业口径实现做准备。

**本阶段完成：**

- [x] 口径差距审计表（8 维度：节令交节、阴阳遁、局数、旬首/空亡、值符值使、九星八门八神、天盘/地盘干、天禽寄宫）
- [x] `AlgorithmVersionQimenV2Professional = "qimen-v2-professional"` 常量草案
- [x] `CalculationResultV2Professional` + `ResultPayloadDraft()` JSON 结构草案
- [x] `ProfessionalFixturePlans` 10 组 fixtures 元数据（含夏至/冬至/节令边界 + category 差异）
- [x] `ProfessionalModuleRoadmap`（ALG2.4–ALG2.5 模块拆分）
- [x] `TestQimenV2ProfessionalFixturesAreDocumented` 等测试（元数据完整性，非假装已有专业计算）

**实现位置（设计 only，未接 Create / handler）：**

- `backend/internal/service/qimen/professional_spec.go`
- `backend/internal/service/qimen/professional_fixtures_test.go`

**口径差距审计（POC → professional）：**

| 维度 | 当前 `qimen-v2-poc` | 目标 `qimen-v2-professional` | 状态 |
|------|---------------------|--------------------------------|------|
| 节气交节 | bazi/calendar 十二节公式近似 | 精确交节时刻（秒级），可扩展真太阳时 | gap → ALG2.4 |
| 阴阳遁 | 公历 6/21、12/22 简化切换 | 与冬至/夏至节气交节绑定 | gap → ALG2.4 |
| 局数 | hash(RFC3339+category+遁) % 9 + 1 | 拆补 / 置闰 / 三元明确口径 | gap → ALG2.4–ALG2.5 |
| 旬首/空亡 | 六旬首固定表 + hash | 由日时干支推导符头、旬首与空亡 | gap → ALG2.4 |
| 值符/值使 | 按局数宫位占位 | 旬首 + 转盘后星门落宫规则 | gap → ALG2.5 |
| 九星/八门/八神 | 固定名表 + hash 轮转 | 转盘 / 飞布规则 | gap → ALG2.5 |
| 天盘干/地盘干 | 十天干表取模轮转 | 按局数、遁法、旬首排布 | gap → ALG2.5 |
| 天禽寄宫 | 中五宫门为 —，未专业寄宫 | 寄坤二 / 寄艮八等流派口径文档化 | gap → ALG2.5 |

**professional payload 草案（节选）：**

```json
{
  "algorithm_version": "qimen-v2-professional",
  "calendar_basis": {
    "solar_term": "",
    "solar_term_time": "",
    "jieqi_basis": "professional_pending",
    "time_basis": "local_time",
    "note": ""
  },
  "dun": {
    "type": "yang_or_yin",
    "ju": 1,
    "method": "chai_bu_or_zhi_run_pending",
    "yuan": "upper_middle_lower_pending"
  },
  "ganzhi": { "year": "", "month": "", "day": "", "hour": "" },
  "xun": { "xun_shou": "", "empty_branches": [] },
  "chief": {
    "zhi_fu": "",
    "zhi_shi": "",
    "zhi_fu_palace": "",
    "zhi_shi_palace": ""
  },
  "palaces": [],
  "method_note": "",
  "limits": []
}
```

**fixtures 规划（10 组，元数据 only）：**

| 输入时间 | category | 关注边界 |
|----------|----------|----------|
| 2024-02-04 10:30 | general | 小寒区间内 |
| 2024-03-20 09:00 | career | 惊蛰 + category |
| 2024-06-20 23:30 | study | 夏至 POC 边界前（仍阳遁） |
| 2024-06-21 00:30 | study | 夏至 POC 边界后（阴遁） |
| 2024-08-07 15:00 | relationship | 小暑 + category |
| 2024-09-22 18:30 | decision | 白露 + 决策类 |
| 2024-12-21 23:10 | general | 冬至 POC 边界前（阴遁） |
| 2024-12-22 00:30 | general | 冬至 POC 边界后（阳遁） |
| 2025-02-03 11:30 | career | 小寒 + 跨年 |
| 2025-06-21 09:00 | study | 次年夏至重复样例 |

**仍不做（ALG2.3-SPEC）：**

- [ ] API 接入 `qimen-v2-professional`（延后 **ALG2.4+**）
- [ ] 替换 `qimen-simple-v1` 默认策略
- [ ] 删除或替换 `qimen-v2-poc`
- [ ] 专业完整排盘计算实现
- [ ] frontend / miniprogram / SQL / deploy 变更

**默认策略：** 线上仍默认 `qimen-simple-v1`；`qimen-v2-poc` 继续可用于内部灰度与 QIMEN-REPORT2；`qimen-v2-professional` 仅为后续目标版本标识。

**下一步：** **ALG2.4B** 拆补/置闰局数；**ALG2.5** 转盘飞布；**QIMEN-V2-VIEW** 前端九宫展示；**BAZI1.3** 八字后续口径。

---

### 10.38 Phase ALG2.4A 交付清单（奇门 v2 professional 基础层）

**目标：** 实现 `qimen-v2-professional` **基础层**（节气 provider、阴阳遁绑定冬至/夏至、干支/旬首/空亡、preview 计算），**不接 API、不部署、不替换 POC/v1**。

**本阶段完成：**

- [x] `SolarTermProvider` 接口 + `FormulaSolarTermProvider`（十二节 + 冬至/夏至；`formula_approximation`）
- [x] `ResolveProfessionalCalendarBasis` / `ResolveProfessionalDun`（`method=solar_term_boundary`，绑定冬至/夏至交节点）
- [x] `ResolveProfessionalGanZhi` / `ResolveXunFromGanZhi`（时柱优先旬首；不再 hash category）
- [x] `CalculateProfessionalPreview`（`algorithm_version=qimen-v2-professional`；chief/palaces/ju 标记 pending）
- [x] 10 组 fixtures 扩展断言（preview payload 结构 + provider 边界）

**实现位置：**

- `professional_calendar.go` / `professional_ganzhi.go` / `professional_calculate.go`
- `professional_*_test.go`、`professional_calculate_test.go`

**仍为公式近似（非秒级天文台）：** 节令与中气按 `[Y*D+C]-L` 本地正午；后续可替换权威交节表。

**夏至/冬至边界说明：** professional 以 provider 交节时刻为准；与 POC 公历 6/21、12/22 可能在小时级边界上不同（测试与文档已注明）。

**仍不做（ALG2.4A）：**

- [ ] API 接入 `qimen-v2-professional`
- [ ] 拆补 / 置闰 / 三元局数（延后 **ALG2.4B**）
- [ ] 转盘飞布 / 值符落宫 / 天禽寄宫（延后 **ALG2.5**）
- [ ] frontend / miniprogram / SQL / deploy

**默认策略：** 线上仍 `qimen-simple-v1`；`qimen-v2-poc` 内部灰度不变；professional preview 仅供后端测试。

---

### 10.39 Phase ALG2.4B 交付清单（奇门 v2 professional 拆补 / 三元局数第一版）

**目标：** 为 `CalculateProfessionalPreview` 增加第一版拆补局数（`chai_bu`）、三元（`upper`/`middle`/`lower`），**不接 API、不部署**。

**本阶段完成：**

- [x] `ProfessionalJuResult` 局数结果结构
- [x] `ResolveProfessionalYuan` — **方案 A**：节令交节后日序 0–4 upper / 5–9 middle / 10+ lower
- [x] `ResolveChaiBuJu` — 十二节起局映射表 + 三元偏移；阳遁顺行、阴遁逆行；`ju` 1–9
- [x] `ResolveZhiRunJuPending` — 置闰法结构预留（`method=zhi_run_pending`，未接入 preview）
- [x] `CalculateProfessionalPreview` 写入 `dun.ju` / `dun.method=chai_bu` / `dun.yuan`
- [x] fixtures 断言：ju 1–9、yuan 三元、category 不影响局数、夏至/冬至前后 dun.type 与 ju 方向

**十二节起局映射（第一版近似，非完整二十四节气）：**

| 节 | 起局 | 节 | 起局 |
|----|------|-----|------|
| 小寒 | 阳2 | 小暑 | 阴8 |
| 立春 | 阳8 | 立秋 | 阴2 |
| 惊蛰 | 阳1 | 白露 | 阴9 |
| 清明 | 阳3 | 寒露 | 阴7 |
| 立夏 | 阳4 | 立冬 | 阴6 |
| 芒种 | 阳5 | 大雪 | 阴5 |

**仍不做（ALG2.4B）：**

- [ ] API 接入 `qimen-v2-professional`
- [ ] 置闰法完整实现（`zhi_run` 仅 pending）
- [ ] 二十四节气完整映射（延后 **ALG2.4C**）
- [ ] 转盘飞布 / 值符落宫 / 天禽寄宫（**ALG2.5**）
- [ ] frontend / miniprogram / SQL / deploy

**下一步：** ALG2.5 转盘飞布；ALG2.4D 置闰法设计；QIMEN-V2-VIEW。

---

### 10.40 Phase ALG2.4C 交付清单（奇门 v2 professional 二十四节气映射增强）

**目标：** 扩展 professional calendar 至二十四节气数据层，拆补 basis 升级为 `twenty_four_terms_chai_bu_v1`；**不接 API、不部署**。

**本阶段完成：**

- [x] `ProfessionalSolarTerm` + `TwentyFourSolarTermCatalog`（24 名称、index、jie/qi）
- [x] `FormulaSolarTermProvider.TwentyFourTerms` — 十二节公式 + 十二气中点近似 + 夏至/冬至公式
- [x] `ResolveCurrentProfessionalTerm` / `ResolvePreviousProfessionalTerm`（跨年稳定）
- [x] `BaseJuForProfessionalTerm` — 24 节气全覆盖（`pending_verification`）
- [x] `ResolveChaiBuJu` basis → `twenty_four_terms_chai_bu_v1`
- [x] preview `calendar_basis` 使用当前二十四节气结果

**精度说明：**

| 类型 | precision | 说明 |
|------|-----------|------|
| 十二节 | `formula_approximation` | 沿用 bazi/calendar 公式 |
| 夏至/冬至 | `formula_approximation` | 阴阳遁边界公式 |
| 其余十气 | `formula_midpoint_approximation` | 相邻节令中点，pending_verification |

**仍不做（ALG2.4C）：**

- [ ] API 接入 `qimen-v2-professional`
- [ ] 置闰法完整实现（`zhi_run_pending` 仍预留）
- [ ] 转盘飞布 / 值符落宫 / 天禽寄宫（**ALG2.5**）
- [ ] frontend / miniprogram / SQL / deploy

**下一步：** ALG2.5B 寄宫/飞布校准；ALG2.6 professional API 灰度；QIMEN-V2-VIEW。

---

### 10.41 Phase ALG2.5 交付清单（奇门 v2 professional 九宫落盘第一版）

**目标：** 实现 professional preview 九宫落盘、值符值使、地盘/天盘干、九星八门八神第一版；**不接 API、不部署**。

**本阶段完成：**

- [x] `BuildProfessionalEarthPlateStems` — 戊己庚辛壬癸丁丙乙，阳顺阴逆
- [x] `BuildProfessionalStars` / `BuildProfessionalDoors` / `BuildProfessionalDeities`
- [x] `BuildProfessionalHeavenPlateStems` — 相对值符落宫旋转
- [x] `ResolveProfessionalChief` — 旬首+局数映射落宫
- [x] `BuildProfessionalPalaces` — 9 宫完整字段 + `layout_role`
- [x] 天禽暂保留中五宫（`layout_role=center`），中五宫门为 `—`
- [x] `CalculateProfessionalPreview` palaces 不再为空

**第一版口径（pending_verification）：**

| 模块 | 说明 |
|------|------|
| 地盘干 | ju 起宫，阳遁顺布 / 阴遁逆布 |
| 九星/八门/八神 | ju + 阴阳遁转盘 |
| 值符值使 | xun_shou 索引 + ju 映射落宫 |
| 天禽 | 中五宫固定，寄宫流派延后 ALG2.5B |

**仍不做（ALG2.5）：**

- [ ] API 接入 `qimen-v2-professional`
- [ ] 置闰法完整实现
- [ ] 寄宫流派校准（ALG2.5B）
- [ ] frontend / miniprogram / SQL / deploy

**下一步：** ALG2.6 professional API 灰度；QIMEN-V2-VIEW；ALG2.5C 坤二/艮八寄宫流派实现。

---

### 10.42 Phase ALG2.5B 交付清单（奇门 v2 professional 落盘口径校准）

**目标：** 落盘口径版本化、天禽寄宫策略可配置（默认不变）、chief/palace 一致性测试增强；**不接 API、不部署**。

**本阶段完成：**

- [x] `ProfessionalLayoutVersionV1` = `professional_layout_v1_center_tianqin`
- [x] `ProfessionalLayoutConfig` + `TianQinPlacementMode`（`center` / `kun2_pending` / `gen8_pending`）
- [x] preview payload 输出 `layout_version` / `layout_basis`
- [x] `method_note` / `limits` 说明 ALG2.5B 第一版落盘口径
- [x] 默认天禽留中五宫；坤二/艮八寄宫仅结构预留
- [x] `BuildProfessionalLayout` + `ValidateChiefPalaceConsistency` / `ValidateProfessionalPalaceIntegrity`
- [x] 值符落中五时值使 fallback 坤二宫门
- [x] 6 组 fixtures + 稳定性 / category 独立 / POC/v1 回归测试

**仍不做（ALG2.5B）：**

- [ ] API 接入 `qimen-v2-professional`
- [ ] 坤二/艮八寄宫流派正式实现（ALG2.5C）
- [ ] frontend / miniprogram / SQL / deploy
- [ ] 声称最终权威专业排盘

**下一步：** ALG2.6 API 灰度；QIMEN-V2-VIEW；ALG2.5C 寄宫流派实现。

---

### 10.43 Phase ALG2.6 交付清单（奇门 v2 professional API 内部灰度）

**目标：** 创建 API 支持 `algorithm_version=qimen-v2-professional`；默认仍为 `qimen-simple-v1`；`qimen-v2-poc` 保留；**仅 backend 部署**。

**本阶段完成：**

- [x] `ResolveAlgorithmVersion` 接受 `qimen-v2-professional`
- [x] `Create` 走 `CalculateProfessionalPreview` + `BuildProfessionalAPIResultPayload`
- [x] result_payload 含 9 宫、chief、layout_version、ganzhi 及 v1 兼容字段
- [x] `free_unlock` fallback 9 段 professional 报告
- [x] DeepSeek prompt 支持 professional 字段（layout_version / ganzhi / palaces_summary）
- [x] handler / service / model 测试覆盖

**仍不做（ALG2.6）：**

- [ ] 小程序 / Web 暴露算法选择
- [ ] SQL / frontend 部署
- [ ] 声称最终权威专业排盘

**下一步：** QIMEN-V2-VIEW-QA；ALG2.7 professional 报告质量增强；BAZI1.3。

---

### 10.44 Phase QIMEN-V2-VIEW 交付清单（小程序 professional 九宫展示）

**目标：** 小程序奇门结果页在 professional 记录上条件展示九宫与排盘口径；**不改** backend / Web / SQL。

**本阶段完成：**

- [x] `buildQimenView` 识别 `qimen-v2-professional` + 9 宫 `palaces`
- [x] 新增 `qimen-palace-grid` 组件（洛书序 3×3）
- [x] 结果页展示 layout_version / 节令 / 阴阳遁 / 局数 / 三元 / 值符值使 / 九宫
- [x] 边界说明（第一版 / 置闰 pending / 寄宫 pending）
- [x] 长图增加一句 professional 轻量摘要（不含完整九宫）

**仍不做（QIMEN-V2-VIEW）：**

- [ ] 普通用户 algorithm_version 选择 UI
- [ ] 默认创建 professional
- [ ] backend / frontend / SQL 变更

**下一步：** QIMEN-V2-VIEW-QA；ALG2.7；BAZI1.3。

---

### 10.45 Phase QIMEN-V2-VIEW-QA 交付清单（小程序 professional 九宫回归验收）

**目标：** 回归验收 QIMEN-V2-VIEW；确认 v1 / poc / professional 展示边界与隐私合规；**不改** backend / Web / SQL。

**验收通过（自动化 / API / 视图层）：**

- [x] 普通创建默认 `qimen-simple-v1`，结果页不展示 professional 九宫区块
- [x] `qimen-v2-poc` 虽有 `palaces=9`，`buildProfessionalQimenView` 返回 null，不误触发 professional 区块
- [x] `qimen-v2-professional`：`palaces=9`、`layout_version=professional_layout_v1_center_tianqin`、`chief` 非 pending；九宫格 3×3 洛书序；中五宫 `door=—` 有 fallback
- [x] 分享长图仅增加 professional 一句摘要，不绘制完整九宫
- [x] 不展示 session_key / prompt / payload 原始 JSON / 完整原问题
- [x] 无强预测 / 改运化灾文案；禁用词仅在过滤器列表

**仍待本地微信开发者工具确认：**

- [ ] 结果页 UI 渲染、历史页跳转、分享卡片与长图实际导出

**本阶段未做：** backend / frontend / SQL / deploy 变更。

**下一步：** ALG2.7；BAZI1.3；RELEASE-QA。

---

### 10.46 Phase ALG2.7 交付清单（奇门 professional 完整报告质量增强）

**目标：** 提升 `qimen-v2-professional` 完整报告（DeepSeek + fallback）质量；**不改**排盘算法、小程序、Web、SQL。

**本阶段完成：**

- [x] `pickProfessionalFocusPalaces`：首选 `zhi_fu_palace` / `zhi_shi_palace`，再按 category 选辅助宫（career→乾六/离九；relationship→兑七/巽四；study→离九/坤二；decision→艮八/坎一；general→中五/坤二）
- [x] focus 2–3 宫；category 仅影响报告关注角度
- [x] DeepSeek prompt 结构化输入 + 9 段输出要求 + 第一版边界说明
- [x] fallback 9 段：引用 layout_version、dun、chief、focus 宫位（含干/星/门/神）
- [x] 隐私：不输出完整原问题 / session_key / payload / prompt
- [x] 合规：正文无强预测 / 改运化灾

**仍不做：** 排盘算法变更；frontend / miniprogram / SQL。

**部署：** backend-only；curl 验证 `free_unlock` full_content。

**下一步：** ALG2.7-QA；BAZI1.3；RELEASE-QA。

---

### 10.47 Phase ALG2.7-QA 交付清单（奇门 professional 报告质量回归验收）

**目标：** 验收 ALG2.7 professional 报告质量；5 类 category 差异、DeepSeek 9 段、fallback 测试路径、v1/poc/bazi 回归、隐私合规。

**验收通过：**

- [x] ECS @ `8e47cdb`；health ok
- [x] 5 类 professional unlock：9 段、九宫/宫名、值符/值使、第一版说明
- [x] category 差异明显（summary/support/risk/pacing 片段 5 类各不相同）
- [x] fallback：`go test` 覆盖 9 段 template、layout_version、focus、category 差异化
- [x] 回归：v1 / poc / pro / bazi；非法 `qimen-v3` → 400
- [x] 隐私：无完整原问题 / session_key / payload JSON
- [x] 合规：正文无正向强断

**未做：** 排盘算法 / miniprogram / Web / SQL 变更；生产 DeepSeek 配置未改。

**下一步：** BAZI1.3；RELEASE-QA。

---

### 10.48 Phase QIMEN-V2-VIEW-DEVTOOLS-QA 交付清单（微信开发者工具 UI 验收）

**目标：** 验收小程序 professional 九宫 UI、分享/长图、历史跳转；**不改** backend / Web / SQL。

**代码层预检通过：**

- [x] v1 / poc / professional 视图条件正确
- [x] 九宫洛书序、中五宫 `door=—` fallback、值符宫标签
- [x] 分享/长图摘要不含完整原问题；professional 长图不含完整九宫
- [x] 历史页路由正确

**DevTools 待本地确认：**

- [ ] 真实渲染、解锁、分享卡片、长图相册、历史跳转

**测试记录：** professional id=102；v1 id=103；poc id=104（见 `miniprogram-dev.md` §25.21）。

**下一步：** RELEASE-QA；BAZI1.3。

---

### 10.49 Phase MINIAPP-LOCAL-QA 交付清单（DevTools / 真机预览）

**目标：** 真实微信开发者工具与真机预览验收 professional 九宫 UI；**不改** backend / Web / SQL。

**已就绪：**

- [x] 测试记录 id=102/103/104 可加载
- [x] 视图层 / 分享 / 长图逻辑预检通过（§25.21）

**待本地 DevTools / 真机确认：**

- [ ] 九宫真实渲染、中五宫 `—`、值符标签
- [ ] 分享卡片、长图相册、历史跳转

**说明：** 自动化环境无法替代 DevTools；需维护者本地勾选验收清单。

**下一步：** RELEASE-QA。

---

### 10.50 Phase BAZI1.3 交付清单（小程序八字 v2 内部记录展示）

**目标：** 小程序八字结果页在 `bazi-v2-poc` 记录上条件展示节气口径 / 四柱 v2 / 五行结构；**不改** backend / Web / SQL。

**本阶段完成：**

- [x] `buildBaziV2View` 识别 `bazi-v2-poc`
- [x] 新增 `bazi-v2-panel` 组件（四柱 v2、calendar_basis、五行、边界说明）
- [x] 长图增加一句 v2 轻量摘要
- [x] 未知时辰不展示时柱
- [x] v1 布局与创建默认路径不变

**仍不做：** 算法选择 UI；默认创建 v2；backend / SQL 变更。

**下一步：** BAZI1.3-QA；RELEASE-QA。

---

### 10.51 Phase BAZI1.3-QA 验收清单（小程序八字 v2 展示回归）

**Git 基线：** `879784d`（BAZI1.3）；**不改** backend / Web / SQL。

**API 复验（ECS）：**

- [x] id=105 `bazi-v2-poc`：`calendar_basis` / `pillars_v2` / `five_elements` 齐全；时柱 `癸卯`
- [x] id=106 `bazi-simple-v1`：无 v2 字段；默认路径不受影响
- [x] id=107 `bazi-v2-poc` 未知时辰：`pillars_v2` 无 `hour`；响应无 session_key / prompt
- [x] id=107 需 session `bazi-v2-unknown-test`（非 `bazi-v2-view-test`）

**视图层 / 隐私 / 合规：**

- [x] `buildAnalysisView`：105/107 `isBaziV2=true`；106 `false`
- [x] 107 `pillarsV2View.hourUnknown=true`；面板文案「未知时辰，未生成时柱」
- [x] 分享卡片 / 长图不含完整 birth_date / 时辰原始输入 / payload JSON
- [x] 长图 v2 仅一句轻量摘要（`BAZI_V2_POSTER_NOTE`）
- [x] 无 algorithm_version 选择 UI；普通创建仍默认 v1
- [x] 禁用词仅出现在长图过滤器；页面无强预测 / 改运化灾文案

**DevTools UI：** 待本地重新编译后勾选（§25.24）。

**本阶段结论：** 自动化预检通过；无代码小修。

**下一步：** RELEASE-QA；HOME1-QA；BAZI1.4 报告质量增强。

---

### 10.52 Phase HOME1 交付清单（首页与模块引导优化）

**目标：** 优化小程序首页模块入口与使用引导，让用户更清楚问事 / 八字 / 奇门分别适合什么场景；**不改** backend / Web / SQL。

**本阶段完成：**

- [x] 首页四块信息架构（品牌 / 三模块 / 如何选择 / 边界说明）
- [x] `home-module-card` + `home-guide-card` 组件
- [x] `utils/home.js` 集中模块文案
- [x] 关于页补充三模块说明
- [x] 保留历史 / 关于入口；无 algorithm_version / 广告 / 登录

**仍不做：** 默认算法变更；backend / SQL；广告 / 支付 / 微信登录。

**DevTools UI：** 需重新编译后本地确认。

**下一步：** HOME1-QA；RELEASE-QA；BAZI1.4。

---

### 10.53 Phase HOME1-QA 验收清单（首页与模块引导回归）

**Git 基线：** `94c138b`（HOME1）；**不改** backend / Web / SQL。

**代码层预检：**

- [x] `home.js` 三模块文案 / 引导 / 边界与 HOME1 规格一致
- [x] 跳转路径均在 `app.json` 注册
- [x] `home-module-card` / `home-guide-card` 组件注册正常
- [x] 八字 / 奇门创建 API 不传 `algorithm_version`
- [x] 首页无内部算法版本 / 广告 / 强预测文案

**DevTools UI：** 待本地重新编译后勾选（§25.26）。

**本阶段结论：** 自动化预检通过；无代码小修。

**下一步：** RELEASE-QA；BAZI1.4；用户反馈/体验优化。

---

### 10.54 Phase RELEASE-QA-PREP 验收清单（体验版前全模块准备）

**Git 基线：** `17b9371`；**不提审**；**不改** backend / Web / SQL。

**backend health：**

- [x] `/api/v1/health` 与 `/health`：`status=ok`, `db=ok`
- [x] `POST /api/v1/sessions` 可用

**代码层 / API 预检：**

- [x] 首页三模块路由 + 历史 / 关于
- [x] 八字 105/106/107 条件展示与隐私
- [x] 奇门 102/103/104 professional 条件展示
- [x] 普通创建不传 algorithm_version
- [x] 静态合规检查通过

**DevTools 统一清单：** 见 `docs/miniprogram-dev.md` §25.27（待本地勾选）。

**阻塞项：** 仅 DevTools / 真机 UI 未闭环；无 backend 或代码阻塞。

**下一步：** 备案 / 合法域名 → 体验版上传评估。

---

### 10.55 Phase RELEASE-QA 验收清单（体验版前最终验收）

**Git 基线：** `bcfbe59`（RELEASE-QA-PREP）；本阶段 docs @ 新 commit；**不提审**；**不上传体验版**。

**自动化验收（2026-06-27）：**

- [x] backend health：`status=ok`, `db=ok`
- [x] 静态检查 / 合规 grep 通过
- [x] 程序化 24/24：首页路由 + 八字 105/106/107 + 奇门 102/103/104 + 默认创建不传 algorithm_version
- [x] `https://api.wenyiapp.cn` 不可用；HTTP dev ECS 可用

**DevTools / 真机：** 待维护者本地勾选（§25.28）；Cursor 无法替代。

**备案 / 合法域名阻塞：**

- [ ] ICP 备案完成
- [ ] HTTPS API 部署
- [ ] 微信 request 合法域名配置
- **建议上传体验版：否**（先完成上述三项）

**本阶段结论：** 代码与 API 就绪；体验版分发与 UI 闭环为剩余阻塞；无 miniprogram 小修。

**下一步：** DevTools 本地勾选 → 备案 / 合法域名 → 体验版上传。

---

### 10.56 Phase MINIAPP-UX1 交付清单（小程序状态体验优化）

**目标：** 补齐小程序核心路径的加载态、错误态、空状态与防重复提交体验；**不改** backend / Web / SQL / deploy / 默认算法。

**本阶段完成：**

- [x] 问事 / 八字 / 奇门创建中保持按钮 loading / disabled，并在导航失败时提示从历史记录重新进入。
- [x] 网络错误统一为「网络似乎不太稳定，请稍后再试」。
- [x] 结果页记录加载失败统一为可重试错误态，并引导回到历史记录或模块入口。
- [x] 完整报告生成 / 修复失败统一为「报告生成失败，请稍后再试」等明确文案。
- [x] 历史页空状态保留分类入口；删除 loading 按 record type + id 精准锁定。
- [x] 长图相册保存失败提示检查相册权限，保留打开设置授权流程。

**仍不做：** 后端改造；SQL；广告 / 支付 / 微信登录；默认算法切换；普通用户算法选择 UI。

**DevTools / 真机：** 需维护者本地验证弱网、导航失败、相册拒绝授权与 Canvas 长图保存。

**下一步：** RELEASE-QA-RECHECK；DOMAIN1；BAZI1.4。

---
