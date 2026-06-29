# Release Checklist

本文件记录「文易传统文化」体验版 / release 前可复用的回归检查流程。当前 TEST1 只新增脚本与文档，不修改业务逻辑、不执行 SQL、不部署 backend、不上传体验版。

## 1. 本地静态检查

```bash
bash scripts/check-miniprogram-static.sh
```

覆盖范围：

- 小程序全部 JS 语法：`node --check`
- 广告 / mock 解锁文案：不得出现「观看视频」「广告解锁」「模拟广告」「rewarded_video_mock」
- 强预测 / 改运 / 投资医疗法律等禁用词：只允许出现在过滤器列表或边界否定说明
- `git diff --check`

## 2. API Smoke 检查

默认 dev API：

```bash
bash scripts/check-api-smoke.sh
```

可覆盖 API 地址：

```bash
API_BASE=http://127.0.0.1:8080/api/v1 ROOT_BASE=http://127.0.0.1 bash scripts/check-api-smoke.sh
```

覆盖范围：

- `GET /health`
- `GET /api/v1/health`
- `POST /api/v1/sessions`
- `POST /api/v1/analysis/bazi`：默认 `bazi-simple-v1`
- `POST /api/v1/analysis/bazi`：内部 `bazi-v2-poc`
- `POST /api/v1/analysis/bazi`：内部 `bazi-v2-poc` + `birth_hour_unknown=true`
- `POST /api/v1/analysis/qimen`：默认 `qimen-simple-v1`
- `POST /api/v1/analysis/qimen`：内部 `qimen-v2-poc`
- `POST /api/v1/analysis/qimen`：内部 `qimen-v2-professional`
- 非法 `algorithm_version=qimen-v3` 返回 400 / invalid params
- `POST /api/v1/analysis/{id}/unlock`：`free_unlock`
- `POST /api/v1/analysis/{id}/unlock`：`bazi-v2-poc` 未知时辰报告包含未知时辰说明，且不生成干支时柱
- `POST /api/v1/divinations`：问事起卦
- `POST /api/v1/divinations/{id}/unlock`：`mock_button`

脚本每次动态创建新测试记录，不依赖历史 id；日志只输出 id、algorithm_version、关键字段是否存在，不输出完整报告正文、完整 `result_payload` 或密钥。

TEST1.1 后当前期望结果为 `15 PASS / 0 FAIL`。其中八字 v2 未知时辰用例必须显式传 `birth_hour_unknown=true`；只省略 `birth_hour_branch` 不表示未知时辰，当前 API 会按参数错误返回 400。

## 3. 隐私与合规检查

```bash
bash scripts/check-release-privacy.sh
```

覆盖范围：

- 小程序不出现广告 / mock 解锁文案
- 小程序不出现强预测正向文案
- 跟踪文档中不出现真实微信 AppID；只允许 `<WECHAT_APP_ID>` 或 `wx********`
- 跟踪文档中不出现疑似真实 AppSecret、access_token、API key
- 不扫描 `.env*`
- 不输出真实密钥值

## 4. 小程序 DevTools 检查

维护者本地打开微信开发者工具，使用 dev API 时需勾选「不校验合法域名、web-view、TLS 版本以及 HTTPS 证书」。

HOME2-QA 代码层已确认：首页结构、三大主模块、感情关系观察、规划中场景 / 小工具和 about 页同步均符合预期；以下仍需维护者在 DevTools / 真机本地勾选。

必测路径：

- 首页：三模块入口、历史、关于
  - 第一屏品牌区文案清晰，“开始问事 / 查看历史”入口可用
  - 三大主模块卡片层级清晰，文字 / 标签 / 进入箭头在窄屏不重叠
  - 常见场景中感情关系观察低于三大主模块，事业选择 / 学习规划 / 人际沟通显示为规划中且不跳转
  - 传统文化小工具规划区可见，梦境意象解析、姓名笔画观察、起名灵感助手、感情签均不作为正式入口
  - 首页无内部算法选择、广告、支付、登录或手机号授权入口
- 问事起卦：创建、起卦动画、结果页、完整解析、历史、删除
  - 结果页长问题显示在“本次问题”卡片中，不挤压事项 / 时间元信息
- 感情关系观察：从首页低优先级场景入口进入 `/pages/ask/ask?scene=relationship`，模板只填充问题不自动提交，结果 / 历史 / 分享复用问事链路
  - 首页视觉层级低于三大主模块
  - 问事页标题、说明、边界提示、placeholder、6 个模板显示正确
  - 模板点击后用户可编辑，提交后创建问事记录并进入结果页
  - 分享卡片、长图、历史列表不展示完整原问题
  - 普通问事入口仍是普通文案，八字 / 奇门入口不受影响
- 八字简析：创建、最近记录、结果页、完整报告、历史、删除
- 奇门问事：创建、最近记录、结果页、完整报告、professional 九宫条件展示、历史、删除
- 分享卡片：问事 / 八字 / 奇门分享入口
- 分享长图：生成、预览、保存
  - 问事长图预览标识清楚，长句换行正常，本卦 / 变卦名称不越界
- 相册权限：首次授权、拒绝授权、打开设置后重试
- 弱网 / 失败态：加载失败、记录打不开、报告生成失败

## 5. 八字测试记录

以下 id 只用于当前 dev 环境辅助测试。正式 release check 应优先运行 `scripts/check-api-smoke.sh` 动态创建新记录。

| 类型 | id | session | algorithm_version | 期望 |
|------|----|---------|-------------------|------|
| v2 正常时辰 | 105 | `bazi-v2-view-test` | `bazi-v2-poc` | 展示 v2 节气 / 四柱 / 五行区块 |
| v1 默认路径 | 106 | `bazi-v1-view-test` | `bazi-simple-v1` | 不展示 v2 区块 |
| v2 未知时辰 | 107 | `bazi-v2-unknown-test` | `bazi-v2-poc` | 不生成或伪造时柱 |

未知时辰 API 请求必须显式传 `birth_hour_unknown=true`，不要只省略 `birth_hour_branch`。`scripts/check-api-smoke.sh` 会动态创建 `bazi-v2-poc` 未知时辰记录，并在解锁后检查完整报告包含未知时辰说明、未生成干支时柱。

## 6. 奇门测试记录

以下 id 只用于当前 dev 环境辅助测试。正式 release check 应优先运行 `scripts/check-api-smoke.sh` 动态创建新记录。

| 类型 | id | session | algorithm_version | 期望 |
|------|----|---------|-------------------|------|
| professional | 102 | `qimen-devtools-prof` | `qimen-v2-professional` | 展示九宫 9 宫与 professional 信息 |
| v1 默认路径 | 103 | `qimen-devtools-v1` | `qimen-simple-v1` | 不展示 professional 九宫 |
| v2 POC | 104 | `qimen-devtools-poc` | `qimen-v2-poc` | 不误触发 professional 九宫 |

## 7. 历史页检查

- 全部 / 问事 / 八字 / 奇门筛选可用
- 三类记录详情跳转正确
- 删除态只锁定当前记录
- 删除后列表刷新或本地移除正确
- 空状态显示对应模块入口
- 列表不展示完整原问题、出生信息、`session_key`、payload、prompt

## 8. 长图 / 相册权限检查

- 问事长图不展示完整原问题或完整解析全文
- 八字长图不展示出生日期、出生时辰原始输入或完整报告全文
- 奇门长图不展示完整原问题或完整报告全文
- 长图不展示 `session_key`、payload、prompt、小程序码
- 保存成功提示正常
- 相册权限拒绝后提示检查权限，并可引导打开设置

## 9. 备案 / HTTPS / request 合法域名检查

体验版前必须完成：

- `wenyiapp.cn` ICP 备案
- `https://api.wenyiapp.cn/api/v1/health` 可用
- 微信公众平台配置 request 合法域名
- release 环境确认使用正式 HTTPS API
- 恢复合法域名、TLS 和 HTTPS 校验

当前已知阻塞：

- 备案仍在管局审核中
- HTTPS API 域名未完成
- 微信 request 合法域名未配置
- DevTools / 真机 UI 仍需维护者本地勾选

## 10. 上传体验版前检查

建议顺序：

```bash
bash scripts/check-miniprogram-static.sh
bash scripts/check-release-privacy.sh
bash scripts/check-api-smoke.sh
git diff --check
git status --short
```

全部通过后，维护者再执行微信 DevTools / 真机验收。未完成备案、HTTPS API、request 合法域名之前，不建议上传体验版，不提审。
