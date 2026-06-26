# 微信小程序开发版使用说明

本文档仅用于微信开发者工具中的开发版调试。当前不提交微信审核，也不配置正式发布环境。

## 1. 当前工程范围

- 小程序源码目录：`miniprogram/`
- 开发 API：`http://123.57.48.214/api/v1`
- 正式 API 预留：`https://api.wenyiapp.cn/api/v1`
- Phase 5 已在问事和今日一卦首次生成流程中接入后端真实六爻逐爻动画

正式 API 必须等待 ICP 备案、HTTPS 证书和微信 request 合法域名配置完成后才能启用。

## 2. 准备项目配置

仓库不会提交真实 AppID。首次使用时，在项目根目录执行：

```bash
cp miniprogram/project.config.json.example miniprogram/project.config.json
```

示例文件默认使用：

```json
{
  "appid": "touristappid"
}
```

可以保留 `touristappid` 做基础界面调试，也可以将复制后的 `project.config.json` 中 `appid` 改为自己的小程序 AppID 或开发测试号。

不要在代码或 Git 中写入 AppSecret。`project.config.json` 和 `project.private.config.json` 已加入 `.gitignore`。

## 3. 导入微信开发者工具

1. 打开微信开发者工具。
2. 选择“导入项目”。
3. 项目目录选择仓库下的 `miniprogram/`。
4. AppID 选择自己的小程序 AppID、开发测试号，或使用示例中的 `touristappid`。
5. 项目名称可填写“易经问事开发版”。
6. 完成导入后点击“编译”。

首页应能打开，并可进入：今日一卦、问事起卦、八字简析、历史记录、关于与免责声明。

## 4. 关闭合法域名校验进行开发调试

当前开发 API 使用 HTTP 公网 IP，不能用于正式发布配置。在微信开发者工具中：

1. 打开右上角“详情”。
2. 进入“本地设置”。
3. 勾选“不校验合法域名、web-view（业务域名）、TLS 版本以及 HTTPS 证书”。
4. 重新编译项目。

`project.config.json.example` 中也预设了 `setting.urlCheck=false`，但仍建议导入后在开发者工具界面确认一次。

这项设置只用于开发调试。正式发布前必须恢复合法域名校验。

## 5. API 环境配置

API 地址统一维护在：

```text
miniprogram/utils/config.js
```

当前环境：

| 环境 | API Base URL | 状态 |
|---|---|---|
| dev | `http://123.57.48.214/api/v1` | 微信开发者工具联调使用 |
| prod | `https://api.wenyiapp.cn/api/v1` | 仅预留，暂不可用 |

不要在页面文件里重复填写 API 地址。环境映射规则为：

- 微信开发版 `develop`：使用 dev
- 微信体验版 `trial`：当前仍使用 dev
- 微信正式版 `release`：使用预留 prod
- 无法识别环境或使用 `touristappid`：默认使用 dev

正式域名当前不可用，不要为了测试而把开发版切换到 prod。备案、HTTPS 和微信 request 合法域名全部完成后，再验证正式环境。

## 6. 开发接口检查

Phase 3 提供仅开发环境可用的页面：

```text
pages/debug/debug
```

该页面不出现在正常首页导航中。微信开发者工具里可以通过以下方式打开：

1. 点击工具栏“普通编译”旁的下拉菜单。
2. 选择“添加编译模式”。
3. 启动页面填写 `pages/debug/debug`。
4. 保存并编译。

正式环境打开该页面只会显示“正式环境不开放此调试页面”。

### 6.1 测试 health

点击 Health 的“检查”，期望显示：

```text
服务：ok；数据库：ok
```

该接口使用原始响应模式，不按普通 `{ code, message, data }` 解析。

### 6.2 测试 session_key

点击 Session 的“检查”：

1. 首次检查会向 `POST /sessions` 传空 `session_key`，由后端生成。
2. 返回值保存到小程序 Storage 的 `yijing_session_key`。
3. 页面只显示掩码，不展示完整值。
4. 再次检查会复用本地值并确保服务端会话存在。

“清除本地调试会话”仅用于调试，并带二次确认。清除后将无法通过新会话继续查看原匿名历史。

### 6.3 测试 categories

点击 Categories 的“检查”，期望显示分类数量和分类名称。该操作不创建业务记录。

### 6.4 测试今日运势

点击 Today Fortune 的“检查”，会使用 UTC+8 日期调用：

```text
POST /daily-fortune/today
```

首次调用会创建当天的真实匿名记录；当天再次调用应显示“已有记录”并返回相同记录 ID。页面不会展示完整 AI 报告。

## 7. 常见错误排查

### 网络连接失败

- 确认 ECS 和后端容器正在运行。
- 确认阿里云安全组已开放 80。
- 确认本机可以访问 `http://123.57.48.214/api/v1/health`。

### Health 返回 404 或 HTML

- 检查宿主机 Nginx 是否保留 `/api/` 前缀并反代到 backend。
- 检查请求基址是否仍为 `http://123.57.48.214/api/v1`。

### `40301` 完整解读未解锁

这是新记录的正常状态。结果页应继续显示卦象和免费解读，并展示“模拟解锁完整解读”按钮；不应显示整页错误。如果已解锁记录仍返回该错误，检查本机匿名 session_key 是否被清除。

### `40002` 问题不支持

说明问题触发了后端内容边界。换成自我反思、状态整理或行动选择类问题，不要在小程序端绕过后端判断。

### `42901` 请求过于频繁

停止连续提交并稍后重试。小程序不会自动重放被限流的起卦或解锁请求。

### 提示“不在以下 request 合法域名列表中”

- 在开发者工具“详情 → 本地设置”中关闭合法域名校验。
- 确认修改后已经重新编译。

### 请求超时

- 检查后端和 MySQL 是否健康。
- 普通请求默认超时 20 秒；后续 DeepSeek 完整解读解锁单独使用 90 秒。

### prod 域名无法访问

当前 `https://api.wenyiapp.cn/api/v1` 只是预留配置。ICP 备案、HTTPS 和微信 request 合法域名未完成前，不要切换到 prod。

### 动画结束后没有跳转

1. 确认 API 返回了有效 `record.id`。
2. 点击“跳过动画”验证 cancel 事件是否可以进入结果页。
3. 到历史记录中检查 record 是否已经保存。
4. 重新编译后再试；不要重复提交同一个问题制造多条记录。

### 历史记录为空

- 首次使用或新匿名 session 本来就没有历史。
- 检查是否在开发检查页清除了 `yijing_session_key`。
- 检查是否清理过开发者工具 Storage。
- 清除 Storage 后形成新匿名历史属于当前设计，不是数据库丢失。

## 8. Phase 4 页面验收流程

开始前确认开发者工具已关闭合法域名校验，并且 Health 检查显示服务和数据库均为 `ok`。

### 8.1 首页入口

1. 使用“普通编译”打开首页。
2. 确认展示传统文化学习、趣味解读、自我反思和行动建议定位。
3. 依次进入今日一卦、问事起卦、历史记录和关于页面。
4. 首页会在后台轻量初始化匿名会话；初始化失败不会阻塞入口展示。

### 8.2 今日一卦

1. 进入“今日一卦”，确认页面日期为 UTC+8 当天日期。
2. 点击“查看今日一卦”。
3. 首次调用应生成记录并进入结果页。
4. 返回后再次调用，应提示“今日一卦已生成”，并打开同一个结果 ID。
5. 页面不会在前端重新随机六爻。

### 8.3 问事起卦

1. 进入“问事起卦”，确认显示事业、关系、学习、选择和近期状态。
2. “今日运势”不应出现在普通事项选择器中。
3. 分别检查未选分类、空问题、少于 5 字和未勾选免责声明的提示。
4. 选择分类，输入 5–200 字的自我反思或行动选择问题，勾选免责声明。
5. 点击“开始起卦”，成功后应直接进入结果页，本阶段没有逐爻动画。

### 8.4 敏感问题

输入涉及医疗、法律、投资、赌博或伤害相关的测试问题。后端返回 `40002` 时，小程序应展示友好提示，不跳转结果页，也不自动重复提交。

### 8.5 结果页与免费解读

1. 确认展示问题、事项类型、起卦时间、本卦、变卦和动爻。
2. 确认六爻从上到下展示，并标记阴阳和动爻。
3. 确认免费解读正常显示。
4. 新记录的完整解读区域应显示“模拟解锁完整解读”。

### 8.6 模拟解锁完整解读

1. 点击“模拟解锁完整解读”。
2. 等待过程中应显示提示；DeepSeek 请求最长允许 90 秒。
3. 成功后应展示：一句话总结、总体判断、当前处境、机会点、风险点、行动建议、情绪提醒、自我反思问题和免责声明。
4. 刷新或重新进入结果页，完整解读仍应可见。
5. 页面和控制台不应输出完整 AI 报告或敏感配置。

### 8.7 历史记录

1. 进入历史记录，确认刚才的问事或今日一卦存在。
2. 每条记录应显示事项类型、问题摘要、本卦、变卦和创建时间。
3. 点击记录应进入对应结果页。
4. 下拉页面应重新加载第一页。
5. 记录超过 20 条时，可通过触底或“加载更多”继续加载。

### 8.8 后端不可用

可在本地临时把 dev 地址改成一个不可达测试地址，或在受控环境暂停后端后编译：

1. 页面应显示“网络连接失败”或超时提示。
2. 页面应提供重试入口。
3. 测试完成后必须恢复 `http://123.57.48.214/api/v1`，不要提交错误地址。

不要为测试关闭生产数据库或删除 Docker 数据卷。

### 8.9 `40301` 未解锁状态

打开一条尚未解锁的新记录。完整解读接口返回 `40301` 时，结果页应正常展示基础信息和免费解读，并显示模拟解锁按钮；不应显示整页错误。

### 8.10 `42901` 限流提示

仅在开发测试会话中短时间重复提交问事，触发限流后应显示“请求过于频繁，请稍后再试”，且不自动重试。不要使用朋友的真实内测会话进行压力测试。

## 9. Phase 5 逐爻动画验收

动画只读取后端返回的 `divination.lines`、本卦、变卦和动爻，不调用随机函数，也不在小程序端计算卦象。

### 9.1 问事起卦动画

1. 进入问事页，完成分类、问题和免责声明校验。
2. 点击“开始起卦”后，按钮应保持禁用，避免重复提交。
3. API 返回前只显示普通提交状态，不播放伪造动画。
4. API 返回真实 record 后，打开全屏动画。
5. 动画依次展示准备、初爻至上爻、本卦、变卦和整理解读阶段。
6. 动画结束后只跳转一次结果页。

### 9.2 今日一卦首次生成动画

1. 在一个当天尚未生成今日记录的匿名 session 中进入今日一卦。
2. 点击按钮并等待 API 返回。
3. `is_existing=false` 时应播放与问事页相同的逐爻组件，模式文案显示“正在整理今日状态”。
4. 动画结束后进入今日结果页。

### 9.3 今日一卦重复进入

当天再次点击今日一卦，若后端返回 `is_existing=true`：

1. 不播放完整逐爻动画。
2. 提示“今日一卦已生成，将为你打开今日结果”。
3. 直接打开原记录。

### 9.4 动画与结果一致性

在动画阶段记录以下内容，并与结果页逐项核对：

- 动画本卦名称 = 结果页本卦名称
- 动画变卦名称 = 结果页变卦名称
- 动画动爻位置 = 结果页动爻位置
- 动画按 position 1→6 生成，结果页按传统排布从上爻到初爻展示
- 每爻的值、阴阳和动静状态完全一致

小程序 API 当前只返回爻值 6/7/8/9，没有返回三枚硬币的具体排列。因此动画中的三枚圆形硬币只作轻量过程视觉，最终信息明确展示真实爻值、阴阳和动静，不伪造正反组合。

### 9.5 API 失败

使用受控的不可达开发地址模拟网络失败：

1. 应关闭提交状态并显示错误。
2. 不应打开 `casting-overlay`。
3. 不应跳转结果页。
4. 测试后立即恢复 dev API 地址。

### 9.6 Timer 清理和重复点击

- 动画播放时返回或关闭页面，组件 `detached` 生命周期应清理全部 timer。
- 点击“跳过动画”会取消 timer，并直接打开后端已经生成的结果。
- 动画期间再次点击页面按钮不应产生第二次请求。
- finish/cancel 使用页面导航锁，不能重复跳转。

### 9.7 动画时长

| 阶段 | 时长 |
|---|---:|
| 准备 | 400ms |
| 每爻 | 400ms × 6 |
| 本卦 | 500ms |
| 变卦 | 500ms |
| 整理解读 | 400ms |
| 合计 | 约 4.2 秒 |

## 10. Phase 6 完整验收清单

### 10.1 代码与配置

- [ ] `app.json`、全部页面 JSON 可以解析
- [ ] 全部 JavaScript 通过语法检查
- [ ] 页面引用的自定义组件文件完整
- [ ] 没有 `Math.random` 或其他前端随机起卦
- [ ] 没有真实 AppID、AppSecret、DeepSeek Key、密码或私钥
- [ ] 没有输出完整 AI 报告或完整 session_key
- [ ] develop/touristappid 默认使用 dev API
- [ ] backend、frontend、sql 和部署配置没有变更

### 10.2 开发者工具基础启动

- [ ] 导入目录选择 `miniprogram/`
- [ ] 本地 `project.config.json` 使用测试 AppID 或 `touristappid`
- [ ] 关闭合法域名、TLS 和 HTTPS 校验
- [ ] 首页编译无 WXML、WXSS、JavaScript 报错
- [ ] Console 没有敏感数据或完整 AI 报告

### 10.3 核心流程

- [ ] 首页四个入口均可访问
- [ ] 今日首次生成播放逐爻动画
- [ ] 今日重复进入提示并跳过完整动画
- [ ] 问事分类不包含“今日运势”
- [ ] 问事字段和免责声明校验正常
- [ ] 40002、42901 提示友好且不自动重试
- [ ] 动画按初爻到上爻展示，总时长约 4.2 秒
- [ ] 动画与结果页本卦、变卦、动爻、六爻一致
- [ ] 结果页免费解读和九字段完整报告正常
- [ ] 40301 只显示锁定态，不影响基础结果
- [ ] 模拟解锁等待和超时提示正常
- [ ] 历史空状态、下拉刷新和加载更多正常
- [ ] 清理 Storage 后形成新匿名历史

### 10.4 机型与视觉

在微信开发者工具中分别选择 iPhone SE、iPhone 14 和一台大屏 Android：

- [ ] 页面没有横向滚动或文本溢出
- [ ] 主要按钮至少约 44px 高
- [ ] 底部内容不被安全区遮挡
- [ ] 长问题、免费解读和完整报告正常换行
- [ ] 动画内容在小屏可纵向滚动，不被裁切
- [ ] 历史列表可正常滚动
- [ ] loading 与 error 状态居中、清晰且可重试

### 10.5 自动化验收边界

可通过静态脚本和模拟 `wx.request` 自动检查：语法、路由、组件依赖、业务错误分类、页面状态、timer 清理、防重复提交和真实 API 数据一致性。

不同模拟机型的最终视觉、开发者工具 Console 和真机表现必须在安装微信开发者工具后人工确认，不能用普通浏览器完全替代。

## 11. 已知限制与正式发布前清单

### 11.1 已知限制

- 当前 HTTP IP 仅用于微信开发者工具，并依赖关闭合法域名/TLS/HTTPS 校验。
- HTTP IP 不作为正式版、审核版或稳定真机调试方案。
- 当前没有微信登录；历史仅通过本机匿名 session_key 关联。
- 清理 Storage 或更换设备后，不会自动恢复原匿名历史。
- 硬币圆形图案是过程视觉；API 未返回三枚硬币具体排列，因此只以爻值、阴阳和动静作为最终信息。
- 动画在切后台时由小程序运行时暂停/恢复，页面导航锁保证不会重复跳转。
- 调试页面仍包含在开发工程中，但正式环境会禁用其操作区域。

### 11.2 开发阶段使用 HTTP IP

1. 环境保持 `develop` 或使用 `touristappid`。
2. API Base URL 保持 `http://123.57.48.214/api/v1`。
3. 开发者工具中关闭合法域名、TLS 和 HTTPS 校验。
4. 仅用于本地开发和页面联调，不提交审核。

### 11.3 正式发布前必须完成

- [ ] `wenyiapp.cn` ICP 备案完成
- [ ] `api.wenyiapp.cn` 部署有效 HTTPS 证书
- [ ] 微信公众平台配置 `request` 合法域名
- [ ] 验证 `https://api.wenyiapp.cn/api/v1` 全部接口
- [ ] release 环境确认切换到 prod Base URL
- [ ] 恢复合法域名、TLS 和 HTTPS 校验
- [ ] 删除或关闭开发调试入口后再评估审核

## 12. Phase 6 分享卡片与分享海报验收

本阶段只使用小程序原生分享能力和本地 Canvas。海报不会请求外部图片，不会上传用户问题到第三方制图服务，也不会生成或伪造二维码。

### 12.1 微信原生分享卡片

- [ ] 首页右上角分享菜单可发送卡片，标题为传统文化学习定位
- [ ] 今日一卦页右上角分享菜单可发送卡片，分享路径回到今日页
- [ ] 结果页右上角分享菜单和“分享给朋友”按钮均可发送卡片
- [ ] 结果页分享路径包含当前记录 ID，格式为 `/pages/result/result?id=xxx`
- [ ] 分享标题不包含精准预测、改命、化灾或保证结果等承诺
- [ ] 分享卡片当前不配置远程 `imageUrl`

### 12.2 结果页分享海报

1. 打开一条已加载完成的结果记录。
2. 点击“生成分享海报”。
3. 确认出现竖版 3:4 海报预览。
4. 海报应包含产品名、合规副标题、事项类型、问题摘要、本卦、变卦、动爻、免费解读摘要、免责声明和搜索引导。
5. 海报不得包含完整 AI 报告、真实或伪造的小程序码。
6. 点击预览外部或“关闭”应关闭预览。

### 12.3 相册权限与保存

- [ ] 首次保存时，小程序请求相册权限
- [ ] 允许权限后，图片成功保存并显示成功提示
- [ ] 拒绝权限后，显示友好说明，并仅在用户再次点击保存时询问是否打开设置
- [ ] 在设置中开启权限后，可以继续保存当前海报
- [ ] 取消打开设置不会关闭海报预览，也不会反复弹窗
- [ ] 保存失败时页面不崩溃，并显示“保存失败，请稍后重试”

开发者工具中的相册保存行为与真机可能存在差异，最终需要在测试手机上至少验证一次首次授权、拒绝授权和设置恢复三条路径。

### 12.4 海报视觉与内容边界

在 iPhone SE、iPhone 14 和大屏 Android 模拟器分别检查：

- [ ] 海报保持约 3:4 竖版比例，预览不横向溢出
- [ ] 长问题最多展示三行并省略
- [ ] 免费解读仅展示摘要，最多五行并省略
- [ ] 本卦、变卦和动爻与结果页一致
- [ ] 六爻图形来自当前结果数据，不进行前端随机
- [ ] 底部显示“微信搜索：文易传统文化”，不显示伪造二维码
- [ ] 免责声明完整可读
- [ ] 海报不出现完整 AI 报告或敏感调试信息

### 12.5 实现说明

- 组件：`components/share-poster/`
- Canvas：微信小程序 Canvas 2D，本地纯色、文字和线条绘制
- 输出：通过 `wx.canvasToTempFilePath` 生成本地临时 PNG
- 保存：通过 `wx.saveImageToPhotosAlbum` 保存到系统相册
- 权限：使用 `scope.writePhotosAlbum`，拒绝后通过用户主动操作引导 `wx.openSetting`
- 当前不使用外部图片 URL、第三方二维码服务或真实小程序码接口

## 13. Phase B 激励视频 mock 适配层

本阶段只实现 **mock 激励视频适配层**，为后续微信流量主激励视频预留结构，不接真实广告。

### 13.1 当前实现范围

- 适配层：`miniprogram/utils/rewarded-ad.js`
- 配置：`miniprogram/utils/config.js` 中 `ad` 字段
- **八字**结果页主流程：确认弹窗 → `controller.show()` → `completed === true` 才调用 unlock（`rewarded_video_mock`）
- **卦象**结果页（Phase E9 起）：不再使用激励视频 mock；改为「查看完整解析」+ `mock_button` unlock
- 后端 unlock_type：`rewarded_video_mock`（Phase B 允许，八字仍在用）

### 13.2 环境默认

| 环境 | 映射 | `ad.enabled` | `ad.mode` | 说明 |
|------|------|--------------|-----------|------|
| develop | dev | `true` | `mock` | 开发版 mock |
| trial | dev | `true` | `mock` | 体验版仍为 mock（内测） |
| release | prod | `false` | `disabled` | 正式版默认关闭 |
| 未知 env + 开发者工具 | dev | `true` | `mock` | 仅工具内允许 mock |
| 未知 env + 真机 | prod | `false` | `disabled` | fail closed |

- 空 `envVersion` **不会** 自动当作 dev
- `rewarded-ad` controller 缺少 `env` 时默认 `disabled`，不默认 dev
- `rewardedVideoAdUnitId` 当前留空
- **未接** 真实微信激励视频、流量主、真实 adUnitId
- **没有** 服务端广告完成强验证
- **没有** Phase B 数据库 migration
- `wechat + 空 adUnitId` 返回 `invalid_config`，**不会** 自动回退 mock
- release / prod **禁止** mock
- 激励视频 mock **仅允许**在标准化后的 `dev` 环境运行（`enabled === true` 且 `mode === "mock"` 且 `env === "dev"`）
- release / prod / 未知真机环境均默认 `disabled`；`rewarded-ad` 适配器自身也会 fail closed，不依赖页面调用方兜底
- 传入 `env: "release"`、`env: "prod"`、空 env 或未知字符串时，即使 `mode: "mock"` 也返回 `disabled`，不会回退 mock

### 13.3 unlock_type 规则（Phase B）

**允许：**

- `mock_button`
- `mock_ad`
- `rewarded_video_mock`

**不允许：**

- `rewarded_video`（真实广告阶段再启用）

结果页主流程使用 `rewarded_video_mock`，不再默认 `mock_button`。

> **Phase E9 更新（卦象）：** `pages/result/result` 已改为「查看完整解析」+ `mock_button` 直接解锁，**不再**走 mock 激励视频。八字结果页 `pages/analysis-result/analysis-result` **仍保留** `rewarded_video_mock` 解锁流程。

`unlockDivination(id, options)` 必须显式传入 `unlockType`；缺少时 API 层直接抛出配置错误（fail closed）。

### 13.4 流程锁与页面卸载防护

- 结果页使用 `unlockFlowRunning` + `unlockFlowToken` 覆盖弹窗、广告、unlock 全链路
- 主按钮与 dev 调试按钮共用同一把锁；进行中再次点击提示「正在处理中，请稍候」
- `onUnload` 设置 `pageUnloaded`、dispose controller、清理 scroll timer
- 异步返回后先校验页面未卸载且 flow token 仍有效，再 `setData` / toast / scroll

### 13.5 开发调试

dev 环境 **卦象**结果页曾提供 mock 激励视频调试按钮；**Phase E9 已移除**。八字结果页仍可通过 `rewarded-ad` mock 流程验收解锁。

也可通过 `config.ad.mockOutcome = "completed" | "cancelled"` 切换（仅影响仍接入 `rewarded-ad` 的页面，当前为八字）。

广告失败提示按 `reason` 区分（如 `disabled`、`invalid_config`、`load_failed` 等），不再一律显示「完整观看后才能解锁」。

### 13.6 Phase B 验收

- [ ] 快速双击 / 主按钮与 dev 按钮交替点击只产生一个流程
- [ ] 取消弹窗后释放锁，可再次点击
- [ ] mock 完整观看：`completed=true` 后调用一次 unlock
- [ ] mock 中途退出：不调用 unlock
- [ ] `disabled` / `invalid_config` / `load_failed`：不调用 unlock，提示对应文案
- [ ] 页面 `onUnload` 后不再 setData / toast / scroll
- [ ] 未知真机环境不会进入 mock
- [ ] 缺少 `unlockType` 时 API 层 fail closed
- [ ] 不打印完整报告、完整 session_key、广告原始错误对象

### 13.7 真实商业化前必须完成

- 关闭 `mock_button`、`mock_ad`、`rewarded_video_mock`
- 配置真实公开 `adUnitId`（不含 AppSecret）
- 启用 `rewarded_video` 与服务端校验策略
- 开通流量主并通过微信审核

## 14. Phase E3：八字简析小程序页面

Phase E3 已在小程序接入 Phase E1/E2 后端 API，形成「八字简析」基础体验闭环（**当时**仅免费简析 + 删除；mock 解锁完整报告见 **§15 Phase E5**）。

### 14.1 新增页面

| 页面 | 路径 | Phase E3 交付范围 |
|---|---|---|
| 八字简析表单 | `pages/bazi/bazi` | 采集出生日期、时辰（或时辰未知）、免责声明；提交后创建记录 |
| 八字简析结果 | `pages/analysis-result/analysis-result?id={id}` | 展示免费简析、删除记录；**不含**完整报告解锁（Phase E5 新增） |

首页新增「八字简析」入口卡片，文案为传统文化学习定位，不使用「精准算命」「命运测算」等表述。

### 14.2 API 封装（Phase E3 范围，`miniprogram/utils/api.js`）

| 方法 | 后端 | Session 传递 |
|---|---|---|
| `createBaziAnalysis(params)` | `POST /analysis/bazi` | body `session_key` + `X-Session-Key` header |
| `getAnalysis(id)` | `GET /analysis/{id}` | **仅** `X-Session-Key` header |
| `getAnalysisList({ page, page_size })` | `GET /analysis?module=bazi` | **仅** `X-Session-Key` header |
| `deleteAnalysis(id)` | `DELETE /analysis/{id}` | **仅** `X-Session-Key` header |

Phase E5 另增 `unlockAnalysis`；见 §15。

**隐私约束：**

- GET / DELETE **不得**把 `session_key` 放入 query
- 小程序日志与错误提示 **不得**打印完整 `session_key`、出生日期、出生时辰、`input_payload`、`result_payload`

### 14.3 历史记录（方案 A）

八字页底部展示「最近记录」列表，调用 `GET /analysis?module=bazi`：

- 展示：创建时间、模块名称（八字简析）、算法版本摘要
- **不展示**出生日期 / 时辰
- 点击跳转结果页

八字记录 **不混入** 六爻 `pages/history/history`，避免分页与排序混乱。

### 14.4 结果页展示范围（Phase E3 交付时）

Phase E3 交付时，结果页 **仅展示免费八字简析**：

- 方法说明（`bazi-simple-v1`）
- 简化干支示意（年 / 月 / 日 / 时柱；时辰未知时不显示时柱）
- 日主、五行倾向、反思焦点、行动建议
- `free_content`
- 免责声明

Phase E3 当时 **未接入**：完整报告解锁、DeepSeek / AI、奇门、分享海报。  
**Phase E5** 已在同一结果页新增 mock 激励视频解锁与模板完整报告；见 §15。

### 14.5 删除能力

结果页底部「危险操作」区提供「删除记录」按钮（弱化样式，不与主操作抢视觉）：

1. 弹窗确认：「删除后不可恢复，是否确认删除？」
2. 调用 `DELETE /analysis/{id}`（`X-Session-Key` header）
3. 成功后返回上一页或跳转八字页
4. 删除后再次打开该 ID 应提示记录不存在

### 14.6 微信开发者工具验收清单（Phase E3 范围）

开始前确认 Health 正常，并已关闭合法域名校验。

**API 层：**

- [ ] `createBaziAnalysis` 可创建记录并返回 `id`
- [ ] `getAnalysis` / `getAnalysisList` / `deleteAnalysis` 使用 `X-Session-Key` header
- [ ] GET / DELETE 请求 URL **不含** `session_key` query

**八字表单页：**

- [ ] 首页「八字简析」入口可进入
- [ ] 未选日期不能提交
- [ ] 未选时辰且未勾选「时辰未知」不能提交
- [ ] 勾选「时辰未知」后可提交（时辰选择隐藏）
- [ ] 未勾选免责声明不能提交
- [ ] 合法输入提交成功并跳转结果页
- [ ] 最近记录列表不展示出生日期 / 时辰
- [ ] 网络错误有友好提示

**结果页：**

- [ ] 展示免费解读与干支示意
- [ ] 时辰未知记录不显示时柱（显示提示文案）
- [ ] 详情页不展示具体出生日期（仅「出生信息已用于本次简析」）
- [ ] 删除记录成功；列表与详情均不再出现
- [ ] 删除后再次打开该 ID 提示不存在

**语法检查（本地）：**

```bash
node --check miniprogram/utils/api.js
node --check miniprogram/utils/bazi.js
node --check miniprogram/pages/bazi/bazi.js
node --check miniprogram/pages/analysis-result/analysis-result.js
```

## 15. Phase E5：八字完整报告 mock 解锁

Phase E5 在八字结果页新增「完整报告」与 `rewarded_video_mock` 解锁能力。

### 15.1 新增 API

| 方法 | 后端 | Session 传递 |
|---|---|---|
| `unlockAnalysis(id, { unlockType })` | `POST /analysis/{id}/unlock` | **仅** `X-Session-Key` header |

请求体：

```json
{
  "unlock_type": "rewarded_video_mock"
}
```

本阶段 **仅支持** `rewarded_video_mock`。完整报告由后端优先生成（Phase E7：DeepSeek，失败 fallback 模板）；小程序调用方式不变。

> **Phase E7 更新：** 当服务器 `AI_PROVIDER=deepseek` 且配置 Key 时，unlock 会尝试 DeepSeek 生成完整报告；失败或未配置时自动 fallback 到模板，用户仍可解锁。

### 15.2 结果页解锁流程

1. 未解锁：展示免费解读 +「观看视频，解锁完整报告」
2. 点击后走 `rewarded-ad.js` mock 流程
3. mock 完整观看 → 调用 `unlockAnalysis`
4. mock 取消 → 不调用 unlock，提示「需要完整观看后才能解锁」
5. 解锁成功 → 展示完整报告并显示「生成结果卡片」；刷新页面仍保持已解锁状态
6. 未解锁不显示「生成结果卡片」
7. 删除功能继续可用；解锁中 / 删除中 / 生成中防重复点击

**文案边界：**

- 使用「观看视频，解锁完整报告」「完整报告仍基于简化干支文化规则…」
- 不使用「看广告改运」「解锁精准财运婚姻」等表述

### 15.3 隐私约束

- GET / DELETE / UNLOCK 均使用 `X-Session-Key` header
- **不得**把 `session_key` 放入 query
- Console 与错误提示 **不得**打印 `session_key`、出生日期、`input_payload`、`result_payload`、`full_content`

### 15.4 微信开发者工具验收清单

- [ ] 未解锁时显示解锁按钮
- [ ] 点击解锁出现 mock 广告流程
- [ ] 取消 mock 广告不调用 unlock
- [ ] 完整 mock 观看后调用 unlock（URL 无 `session_key` query）
- [ ] unlock 成功后展示完整报告
- [ ] 刷新结果页仍显示已解锁完整报告
- [ ] 删除已解锁记录成功；再次打开提示不存在
- [ ] Console 不打印敏感信息

## 16. Phase E6（部署）：ECS 后端部署（Phase E5 unlock）

Phase E6（部署）将 Phase E5 后端发布到内测 ECS（`http://123.57.48.214/api/v1`），**仅 rebuild backend**，不改服务器 `.env`、Nginx、frontend。

### 16.1 部署命令（服务器 `/opt/yijing`）

```bash
git pull origin main
docker compose -f docker-compose.prod.yml --env-file .env build backend
docker compose -f docker-compose.prod.yml --env-file .env up -d backend
docker compose -f docker-compose.prod.yml --env-file .env exec -T backend ./migrate
curl -s http://127.0.0.1:8080/api/v1/health
```

### 16.2 远程 API 验收（已通过）

- [x] `GET /api/v1/health` → db ok
- [x] `POST /analysis/bazi` 创建记录
- [x] `POST /analysis/{id}/unlock`（`X-Session-Key` + `rewarded_video_mock`）→ `full_content`
- [x] 非法 unlock_type / query session_key → 400
- [x] `DELETE /analysis/{id}` → 404 不可再读

### 16.3 微信开发者工具联调（需人工）

关闭合法域名校验后，在八字结果页验证 mock 解锁 → 完整报告展示 → 刷新仍已解锁 → 删除后不可再开。

## 18. Phase E6（小程序）：八字结果分享卡片

在八字结果页新增「生成结果卡片」，使用本地 canvas 生成竖版摘要卡片，支持预览与保存到相册。

### 18.1 功能范围

- **解锁完整报告后**才显示「生成结果卡片」（`unlock_status === 1` 为主判断；`full_content` 用于展示完整报告）
- 未解锁状态：仅显示「观看视频，解锁完整报告」，**不显示**「生成结果卡片」
- 已解锁后点击生成竖版摘要卡片，支持预览与保存到相册
- 本地 canvas 生成，不请求后端、不生成小程序码
- 复用六爻海报的 canvas / 相册权限模式（`components/bazi-share-card/`；canvas 绘制工具与 `utils/poster-canvas.js` 同源，组件内**内联**以避免依赖分析过滤）
- 组件需在结果页 WXML 中**静态挂载**（勿对 `bazi-share-card` 使用 `wx:if`），否则微信「代码依赖分析」可能忽略该 JS，导致 `selectComponent` 失败

### 18.1.1 解锁判定

- 主条件：`unlock_status === 1`
- 辅助：`full_content` 非空时展示完整报告区块
- 解锁成功后立即显示「生成结果卡片」；刷新已解锁记录页仍显示该按钮

### 18.1.2 错误提示

- 卡片数据异常：「卡片数据暂不可用，请刷新后重试」
- 画布初始化失败：「卡片画布初始化失败，请重新进入页面」
- 不再使用笼统的「生成结果卡片不可用」

### 18.1.3 操作区布局

**未解锁：**

- 主按钮「观看视频，解锁完整报告」（88rpx 高、单行、全宽）
- 下方说明：完整报告仍基于简化干支文化规则…；解锁后可生成结果卡片

**已解锁：**

- 完整报告区块 → 主按钮「生成结果卡片」→ 摘要说明

**底部危险操作：**

- 分隔线 + 「删除记录」（76rpx 高、浅红边框/红字/白底，弱化）

解锁中 / 删除中 / 生成中互斥 loading/disabled。

### 18.2 卡片展示内容

- 文易传统文化 / 八字简析
- 简化干支文化规则说明
- 年柱 / 月柱 / 日柱 / 时柱（时辰未知时显示提示文案）
- 五行倾向摘要
- 反思焦点摘要
- 行动建议 1–2 条
- 底部免责声明

**不展示：** 出生日期、出生时辰、session_key、完整报告、`input_payload` / `result_payload`

### 18.3 微信开发者工具验收清单

- [ ] 未解锁主按钮单行显示、高度正常
- [ ] 删除按钮位于底部危险操作区且视觉弱化
- [ ] 未解锁状态不显示「生成结果卡片」，仅引导「观看视频，解锁完整报告」
- [ ] 完整 mock 观看后成功解锁
- [ ] 解锁成功后显示「生成结果卡片」
- [ ] 刷新已解锁结果页仍显示「生成结果卡片」
- [ ] 点击后可正常生成摘要卡片
- [ ] 卡片不展示出生日期 / 出生时辰
- [ ] 卡片不展示完整报告全文
- [ ] 卡片不展示 session_key / input_payload / result_payload
- [ ] 时辰未知时显示「时辰未知，本次不生成时柱」
- [ ] 保存相册成功
- [ ] 拒绝相册权限有友好提示
- [ ] Console 不打印 session_key / 出生信息 / full_content
- [ ] 删除中 / 解锁中 / 生成中不可重复操作
- [ ] 记录不存在时不可生成

### 18.4 语法检查

```bash
node --check miniprogram/pages/analysis-result/analysis-result.js
node --check miniprogram/components/bazi-share-card/bazi-share-card.js
node --check miniprogram/utils/bazi.js
node --check miniprogram/utils/poster-canvas.js
```

## 19. Phase E7：八字完整报告 DeepSeek（后端）

Phase E7 将八字 unlock 完整报告从纯模板升级为 **DeepSeek 优先生成 + 模板 fallback**。

### 19.1 行为

- `POST /api/v1/analysis/{id}/unlock` 调用方式不变（仍仅 `rewarded_video_mock`）
- 服务器 `AI_PROVIDER=deepseek` 且配置 Key 时：尝试 DeepSeek 生成 `full_content`，`ai_provider=deepseek`
- DeepSeek 失败 / 空内容 / 未配置：fallback 到现有模板，`ai_provider=template_fallback`
- 已解锁重复 unlock：返回已有 `full_content`，**不重复调用 DeepSeek**
- 小程序无需改版；GET 详情在已解锁时仍返回 `full_content`（及 `ai_provider` 字段）

### 19.2 合规与隐私

- Prompt 禁止精准算命、婚姻财运预测、改运化解、疾病寿命等表述
- DeepSeek 输入仅发送简化干支结果（pillars、day_master、five_elements 等），不发送完整出生日期
- 日志不打印 session_key、出生信息、payload、full_content、DeepSeek 原始响应、API Key

### 19.3 明确未做

- 真实微信激励视频
- 付费解锁
- 奇门 unlock
- 新 SQL / 小程序大改版

## 21. Phase E9：八字/卦象长图分享 + 卦象去除 mock 视频解锁

Phase E9 在小程序侧升级分享能力：**八字**支持微信分享给朋友 + 解锁后生成完整分享长图；**卦象**去除 mock 观看视频解锁 UI，改为直接「查看完整解析」，并升级完整解析长图。

### 21.1 八字结果页（`pages/analysis-result/analysis-result`）

**分享给朋友：**

- 实现 `onShareAppMessage`
- 分享标题：「一份传统文化视角的八字简析」（通用标题，不含出生日期/时辰）
- **分享路径：`/pages/bazi/bazi`（入口分享，非私有记录直达）**
- 记录基于匿名 `session_key` 绑定；朋友 session 不同，**无法**通过分享打开你的私有结果页
- 若用户手动访问 `/pages/analysis-result/analysis-result?id={id}` 且 session 不匹配，会显示「记录不存在或已被删除」等通用错误，并提供「返回八字简析 →」入口；**不泄露**记录归属或他人隐私
- 记录加载失败时分享同样回退到 `/pages/bazi/bazi`

**完整分享长图（解锁后权益）：**

- 未解锁：仅显示「观看视频，解锁完整报告」（仍走 `rewarded_video_mock`）
- 已解锁：显示「生成分享长图」
- 长图内容：方法说明、简化干支示意、五行倾向、反思焦点、行动建议、免费解读、完整报告、免责声明
- **不展示：** 出生日期、出生时辰、session_key、input_payload / result_payload 原始 JSON、小程序码
- 本地 canvas 动态高度（`utils/long-poster-canvas.js` + `components/bazi-share-card/`）
- 数据构建：`buildBaziLongPosterData(recordId, view, fullContent)`
- **超长内容：** `estimateRawLongPosterHeight` 返回原始高度；超过 12000px 时在 canvas 内限制 `fullContent` 绘制行数，并在长图底部写入「内容较长，长图仅展示前半部分内容…」；预览弹窗同步提示，**不**称「摘要」
- **真机导出：** 长图较高时通过 `resolveExportPixelRatio` 降低 pixelRatio（最高 12000px 时用 1x），避免内存/导出失败

### 21.2 卦象结果页（`pages/result/result`）

**去除 mock 观看视频：**

- 不再显示「观看视频」「mock 广告」「看广告解锁」相关 UI
- 不再触发 `rewarded-ad.js` / dev mock 广告工具
- 未解锁时显示「查看完整解析」，点击直接调用 `unlockDivination`（`unlock_type: mock_button`）
- 已解锁时直接展示完整解析

**完整解析长图：**

- 按钮文案：「生成解析长图」（仅完整解析已加载后显示）
- 长图内容：问事分类（不展示完整原问题）、本卦/变卦/动爻、免费解读、完整解析、免责声明
- **不展示：** session_key、原始 payload JSON、小程序码
- 本地 canvas 动态高度（`components/share-poster/` 复用 `long-poster-canvas.js`）
- 超长完整解析同样限制绘制行数 + 底部截断说明；高长图降低 pixelRatio

> **与 Phase B / §13 的关系：** 卦象结果页 Phase E9 起使用 `mock_button`；§13 Phase B 原文中「结果页主流程使用 rewarded_video_mock」**仅适用于八字**，卦象已调整，见 §13.3 注记。

### 21.3 Canvas 工具

- 新增 `miniprogram/utils/long-poster-canvas.js`：动态高度、全文换行、`computePosterDimensions`（原始高度 vs canvas 高度）、`resolveExportPixelRatio`、`exportCanvasToTempFile`、`getAlbumPermissionHelpers`
- 保留 `poster-canvas.js` 供旧逻辑兼容

### 21.4 隐私与合规

- Console 不打印 `full_content`、出生信息、session_key、完整 payload
- 不接真实广告、不接 AppSecret、不请求后端生成图片

### 21.5 语法检查

```bash
node --check miniprogram/pages/analysis-result/analysis-result.js
node --check miniprogram/pages/result/result.js
node --check miniprogram/utils/bazi.js
node --check miniprogram/utils/poster-canvas.js
node --check miniprogram/utils/long-poster-canvas.js
node --check miniprogram/components/bazi-share-card/bazi-share-card.js
node --check miniprogram/components/share-poster/share-poster.js
```

### 21.6 微信开发者工具验收清单

**八字：**

- [ ] 结果页支持右上角分享给朋友
- [ ] 分享标题不包含出生日期/时辰
- [ ] 分享路径为 `/pages/bazi/bazi`（入口分享，朋友不会直达私有记录）
- [ ] 未解锁时不显示「生成分享长图」
- [ ] 解锁后显示「生成分享长图」
- [ ] 长图包含免费解读和完整报告
- [ ] 长图不展示出生日期/时辰 / session_key / payload
- [ ] 保存相册成功；拒绝权限有友好提示
- [ ] Console 不打印 full_content / 出生信息 / session_key

**卦象：**

- [ ] 结果页不再显示 mock 观看视频
- [ ] 不再触发 rewarded video mock
- [ ] 可「查看完整解析」
- [ ] 可「生成解析长图」
- [ ] 长图包含卦象解析主要内容
- [ ] 长图不展示 session_key / 原始 payload / 完整原问题
- [ ] 保存相册成功
- [ ] Console 不打印完整 payload / session_key

## 20. 当前明确不做

- 不提交微信审核或正式发布
- 不配置正式 request 合法域名
- 不接微信登录或手机号授权
- 不接微信支付
- 不接真实广告（Phase B 仅 mock 适配层，非真实流量主）
- 不接订阅消息
- 不接真实小程序码接口或伪造二维码
- 不为分享海报请求外部图片或第三方制图服务
- 不写入 AppSecret、DeepSeek API Key、服务器密码或私钥

所有卦象内容保持传统文化学习、趣味解读、自我反思和行动整理定位，不作为预测或医疗、法律、投资等现实决策建议。

## 22. Phase F2：小程序奇门问事页面

Phase F2 接入 Phase F1 奇门后端 API，提供表单、结果展示、最近记录与删除能力。

### 22.1 新增页面

| 页面 | 路径 | 说明 |
|---|---|---|
| 奇门问事表单 | `pages/qimen/qimen` | 问题 + 分类 + 免责声明；最近记录列表 |
| 奇门问事结果 | `pages/qimen-result/qimen-result` | 展示结构化简析 + 免费解读 + 删除 |

首页新增「奇门问事」入口卡片。

### 22.2 API 封装（`miniprogram/utils/api.js`）

| 方法 | 后端 |
|---|---|
| `createQimenAnalysis(params)` | `POST /analysis/qimen` |
| `getQimenAnalysisList({ page, page_size })` | `GET /analysis?module=qimen` |
| `getAnalysis(id)` | `GET /analysis/{id}`（复用） |
| `deleteAnalysis(id)` | `DELETE /analysis/{id}`（复用） |

Session：`POST` body + `X-Session-Key` header；`GET`/`DELETE` 仅 header，**不放 query**。

### 22.3 表单与校验

- `question` 必填，4–120 字
- `category`：`career` / `relationship` / `study` / `decision` / `general`
- `confirm_disclaimer` 必须勾选
- 不采集姓名、手机号、身份证、地址、性别、地理位置
- 提交中防重复点击；`40002` 敏感问题友好提示

### 22.4 结果展示

展示：方法说明、局势梳理、风险观察、行动节奏、自我反思、行动建议、免费解读、简化说明、免责声明。

**不展示：** 完整原问题、session_key、原始 payload JSON。

**本阶段不显示：** 解锁完整报告、观看视频、生成长图、DeepSeek。

### 22.5 最近记录

- 调用 `getQimenAnalysisList`
- 列表仅展示「奇门问事」+ 安全摘要 + 创建时间
- 点击进入 `qimen-result`

### 22.6 语法检查

```bash
node --check miniprogram/utils/api.js
node --check miniprogram/utils/qimen.js
node --check miniprogram/pages/qimen/qimen.js
node --check miniprogram/pages/qimen-result/qimen-result.js
```

### 22.7 微信开发者工具验收清单

- [ ] 首页可见「奇门问事」入口
- [ ] 表单校验：空问题 / 过短 / 过长 / 未勾选免责声明
- [ ] 提交成功跳转结果页
- [ ] 结果页展示局势梳理、风险观察、行动节奏等字段
- [ ] 结果页不展示完整原问题
- [ ] 最近记录不展示完整原问题
- [ ] 删除记录确认后返回奇门页
- [ ] 敏感问题（如投资类）有友好提示
- [ ] Console 不打印 question / session_key / payload

## 23. Phase F4：奇门完整报告（后端）

Phase F4 在后端实现奇门完整报告生成能力，DeepSeek 优先、模板 fallback；**不开放小程序 unlock 入口**（待 Phase F5）。

### 23.1 服务位置

| 组件 | 路径 |
|---|---|
| 完整报告生成器 | `backend/internal/service/qimen/full_report.go` |
| 模板 fallback | `backend/internal/service/qimen/full_content.go` |
| DeepSeek Prompt | `backend/internal/service/qimen/deepseek_full.go` |
| 路由入口 | `analysis.Service.GenerateFullReport` |

### 23.2 行为

- DeepSeek 成功 → `ai_provider=deepseek`
- DeepSeek 失败 / 空内容 / 禁用词 → `template_fallback`
- 免费解读不受影响
- 奇门 `POST /analysis/{id}/unlock` 仍返回 forbidden

### 23.3 验收命令

```bash
cd backend && go test -count=1 ./internal/service/qimen/... ./internal/service/analysis/...
```

## 24. Phase F5：奇门完整报告解锁

### 24.1 后端

- `POST /analysis/{id}/unlock` 支持奇门（`module_type=2`）
- 仅 `unlock_type=rewarded_video_mock`
- Session：`X-Session-Key` header，不放 query

### 24.2 小程序

- `pages/qimen-result`：mock 激励视频 → unlock → 展示完整报告
- 复用 `unlockAnalysis(id, { unlockType: 'rewarded_video_mock' })`

### 24.3 语法检查

```bash
node --check miniprogram/pages/qimen-result/qimen-result.js
node --check miniprogram/utils/api.js
node --check miniprogram/utils/qimen.js
node --check miniprogram/components/qimen-share-card/qimen-share-card.js
```

## 25. Phase F6：奇门结果长图分享

解锁完整报告后，结果页支持「生成分享长图」与「分享给朋友」。

### 25.1 组件

| 组件 | 路径 |
|---|---|
| 奇门长图 | `components/qimen-share-card/qimen-share-card` |
| 数据构建 | `utils/qimen.js` → `buildQimenLongPosterData` |

### 25.2 长图内容

包含：文易传统文化、奇门问事、qimen-simple-v1 说明、局势梳理、风险观察、行动节奏、反思问题、行动建议、免费解读、完整报告、免责声明。

**不包含：** 完整原问题、session_key、payload、小程序码。

### 25.3 分享给朋友

- 标题：「一份传统文化视角的奇门问事简析」
- 路径：`/pages/qimen-result/qimen-result?id={id}`
- 加载失败回退奇门页入口

## 25.1 Phase F7：奇门差异化解读（结果页）

**背景：** 不同问事 free_content 相似度过高。

**后端：** `result_payload` 新增 `question_profile`、`qimen_lens`、`safe_question_summary`；免费/完整解读按问事特征组合生成。

**小程序结果页：** 新增「关注主题」卡片（focus / pacing / caution / support），仍不展示完整原问题。

**验收：**

- [ ] 同类不同问（如两个 career 问法）局势/节奏/建议有差异
- [ ] 沟通类 vs 推进类 intent_type 不同
- [ ] 长图/分享/列表仍不含完整原问题

- [ ] 长图/分享/列表仍不含出生日期、时辰

## 25.2 Phase E10：八字差异化解读（结果页）

**后端：** `result_payload` 新增 `bazi_profile`、`interpretation_lens`；免费/完整解读按五行与日主特征组合生成。

**小程序结果页：** 新增「解读视角」卡片（五行倾向 / 行动风格 / 反思主题 / 节奏建议）。

## 25.3 Phase ALG1：八字 v2 POC（后端，未改小程序）

**说明：** ALG1 仅在 `backend/internal/service/bazi/` 增加 `bazi-v2-poc` 计算函数与测试；**小程序 / Web 仍展示 simple-v1 结果**，待 ALG1.1 灰度后再评估前端字段切换。

**v2 增强点：** 立春换年、节气月柱；不做大运/流年/神煞/真太阳时（ALG1.1+）。

**验收（后端）：**

- [ ] `go test ./internal/service/bazi/...` 通过
- [ ] `CalculateV2` golden tests 覆盖立春/惊蛰/清明/小寒/大雪边界
- [ ] `bazi-simple-v1` golden 未被破坏

## 26. Phase UX1：八字 / 奇门轻量动效

Phase UX1 在小程序与 Web 八字、奇门页面增加贴合传统文化场景的轻量 UI 动效，提升氛围与完成感。**仅改 UI 动效，不改后端、数据库、部署。**

### 26.1 技术原则

- 优先使用 WXSS `animation` / `transition`
- **不引入** Lottie / `lottie-miniprogram`
- **不新增**远程动效资源、npm 包、视频/音频/震动
- 动效不遮挡正文、不高频闪烁、不影响按钮点击
- 不使用 `setInterval` 频繁 `setData`；页面卸载时清理 timer

### 26.2 新增组件

| 组件 | 路径 | 用途 |
|------|------|------|
| `element-flow` | `components/element-flow/` | 木火土金水标签慢速流转高亮（八字表单/结果页） |
| `qimen-grid` | `components/qimen-grid/` | 九宫格线框装饰 + 光点慢速变化（奇门表单/结果页） |
| `result-reveal` | `components/result-reveal/` | 结果分段延迟渐入（`active` + `delay` props） |

### 26.3 八字动效

**表单页（`pages/bazi/bazi`）：**

- 顶部柔和光晕浮动
- `element-flow` 五行标签依次轻微高亮
- 提交 loading 文案「生成中…」+ 按钮呼吸光晕
- 免责声明保持静态、清晰

**结果页（`pages/analysis-result/analysis-result`）：**

- `result-reveal` 分段渐入：方法说明 → 四柱（依次）→ 日主 → 五行 → 建议 → 免费解读
- 五行标签行慢速高亮
- 解锁 loading 按钮光晕；解锁成功后完整报告淡入
- 已解锁时长图按钮一次性浮现（非持续闪烁）

### 26.4 奇门动效

**表单页（`pages/qimen/qimen`）：**

- 顶部 `qimen-grid` 九宫格装饰
- 分类切换轻微选中反馈
- 提交 loading 文案「整理局势中…」+ 按钮呼吸

**结果页（`pages/qimen-result/qimen-result`）：**

- 顶部九宫格装饰
- 局势梳理 / 风险观察 / 行动节奏 / 自我反思 / 行动建议 分段渐入
- 解锁后完整报告淡入

### 26.5 不变项

- **不改变** unlock 行为、`rewarded_video_mock` 流程
- **不接** 真实广告 / 支付 / 微信登录
- **不展示** 出生日期/时辰、完整原问题、`session_key`、`payload`
- **不打印** `full_content` / `session_key` / `payload` 到 Console

### 26.7 Web 端同步（Phase UX1-W）

Web 端与小程序 UX1 对齐，使用 `frontend/app/globals.css` 中的 CSS animation（无 Lottie / 无远程资源）：

| Web 路由 / 组件 | 动效 |
|-----------------|------|
| `/bazi` | 顶部光晕、`ElementFlow`、提交「生成中…」+ 呼吸 |
| `/qimen` | `QimenGrid`、分类选中 pulse、「整理局势中…」 |
| `/analysis/[id]` | `ResultReveal` 分段渐入、完整报告淡入、解锁光晕 |
| `components/motion/*` | `ElementFlow` / `QimenGrid` / `ResultReveal` |

支持 `prefers-reduced-motion: reduce` 自动关闭动画。

**八字：**

- [ ] 表单页有轻量五行氛围动效
- [ ] 提交按钮 loading 清晰（「生成中…」）
- [ ] 结果页卡片分段渐入
- [ ] 四柱 / 五行 / 建议展示更有层次
- [ ] 解锁完整报告后淡入正常
- [ ] 未改变 `rewarded_video_mock` 解锁逻辑
- [ ] 不展示出生日期 / 出生时辰
- [ ] Console 不打印 full_content / session_key / payload
- [ ] 低端机无明显卡顿

**奇门：**

- [ ] 表单页有轻量九宫格氛围
- [ ] 分类选中态有轻微反馈
- [ ] 提交 loading 文案克制（「整理局势中…」）
- [ ] 结果页分段渐入
- [ ] 不展示完整原问题
- [ ] 不展示 session_key / payload
- [ ] 无强预测 / 改运 / 大凶大吉文案
- [ ] 低端机无明显卡顿

### 26.8 微信开发者工具 / 浏览器验收清单

见上文 **八字** / **奇门** 复选框列表；Web 另可在 Chrome DevTools 开启「 prefers-reduced-motion 」模拟验证动画关闭。

### 26.9 Phase UX1 自审（2026-06-26）

| 检查项 | 小程序 | Web | 结论 |
|--------|--------|-----|------|
| 纯 CSS 动效，无 Lottie / 远程资源 | ✓ | ✓ | 通过 |
| 无 `setInterval` / 频繁 `setData` | ✓ | ✓ | 通过 |
| unlock / 隐私 / 文案边界未改 | ✓ | ✓ | 通过 |
| 无 `console.log` 敏感字段 | ✓ | ✓ | 通过 |
| 结果分段 delay 与段落顺序一致 | ✓ | ✓ | 通过 |
| `prefers-reduced-motion` 降级 | ✓ | ✓ | 自审已补齐 |
| 四柱双重渐入 | 已修复 | 已修复 | 移除冗余动画 |
| Web 解锁光晕与弹窗态一致 | — | 已修复 | 弹窗打开时亦显示光晕 |

**残留低风险：** 已解锁时分享长图按钮有约 0.5s 浮现延迟（符合设计）；Web 修复区块未包渐入动画（不影响功能）。

## 27. Phase UX2：可感知动效升级（小程序 + Web）

在 UX1 基础上升级为**更明显**的 CSS/WXSS 动效（参考 lottie-web / sparticles / awesome-css-animations 等开源思路，**未引入第三方库或 Lottie JSON**）。

### 27.1 核心效果

| 效果 | 组件 | 场景 |
|------|------|------|
| A 五行环形流转 | `element-orbit` | 八字表单/结果顶、Web `/bazi`、首页入口 |
| B 四柱翻入 | `section-reveal` + `pillar-card--flip` | 八字结果四柱 |
| C 九宫扫描光点 | `qimen-scan-grid` | 奇门表单/结果顶、Web `/qimen`、首页入口 |
| D 结果分段展开 | `section-reveal` | 八字/奇门结果各 section（80ms 步进） |

### 27.2 技术约束

- 无 Lottie / 无远程动画 / 无 npm 动画库
- 无 setInterval / 无每帧 setData
- 删除区、免责声明使用 `static` 不做夸张动效
- `prefers-reduced-motion` 降级（`app.wxss` / `globals.css`）

### 27.3 验收清单

见 Phase UX2 任务文档；微信开发者工具 + 浏览器实机确认动效**明显可见**。

## 28. Phase W2：Web 首页对齐小程序

**范围：** 仅 Web `frontend/`；小程序首页 `pages/index` 为设计基准，**不改** miniprogram。

### 28.1 对齐项

| 项 | 小程序 | Web W2 |
|----|--------|--------|
| 主标题 | 从卦象中整理当下思路 | 同左 |
| 四标签 positioning-card | ✓ | ✓ |
| 入口顺序 | 问事 → 八字 → 奇门 → 今日 → 历史 → 关于 | 同左 |
| 今日一卦 | 列表入口（淡金边框） | 移除独立「今日运势」卡片 |
| 免责声明 | 页底浅色块 | 同文案 + 顶部 banner 同步 |

### 28.2 保留

- 首页八字 / 奇门入口 compact UX2 动效（低 opacity，不抢主按钮）
- 现有路由：`/ask` `/bazi` `/qimen` `/today` `/history` `/about`

### 28.3 验收

浏览器移动端宽度检查文案、顺序、样式；`npm run build` 通过。

## 30. Phase AD0：流量主未开通前隐藏 mock 广告

**背景：** 微信小程序尚未开通流量主，生产 UI 不展示 mock 广告与「观看视频」文案。

### 30.1 当前产品规则

| 模块 | 未解锁按钮 | unlock_type |
|------|-----------|-------------|
| 八字结果 | 查看完整报告 | `free_unlock` |
| 奇门结果 | 查看完整报告 | `free_unlock` |
| 卦象结果 | 查看完整解析 | `mock_button`（六爻 unlock 服务，不变） |

- `rewarded_video_mock`：**仅** dev/测试兼容，生产 UI 不触发
- `rewarded_video` / `paid` / `admin`：仍拒绝
- 不接真实 `adUnitId`、不接 AppSecret

### 30.2 未来 Phase AD1

开通流量主 + 正式 adUnitId + 微信审核通过后，再接真实激励视频广告。

### 30.3 验收

- [ ] 八字/奇门结果页无「观看视频」「广告解锁」文案
- [ ] 点击「查看完整报告」直接生成完整报告
- [ ] Web 同步无 mock 广告文案
- [ ] `rewarded-ad.js` 保留但生产页面不引用

## 29. Phase UX2.1：奇门九宫动效强化

**范围：** `qimen-scan-grid` 组件 + 奇门表单/结果页顶栏布局 + Web 同步；**不改** backend / unlock / API。

### 29.1 优化项

| 项 | 说明 |
|----|------|
| 尺寸 | hero 模式 `168rpx` / `5.25rem`，compact 首页入口 `112rpx` |
| 布局 | 九宫居中于标题上方，与 eyebrow / title 形成整体 |
| 扫描线 | 横向光带自上而下 + 斜向光带（5.2s 周期） |
| 光点 | `translate(-50%,-50%)` 沿九宫路径移动（≥9 点位） |
| 宫位高亮 | 9 格依次轻微金色高亮（`--cell-i` 错开 delay） |
| 背景 | 淡金渐变 + 径向光晕 |

### 29.2 不变项

- 表单 / 解锁 / 删除 / 长图 / 分享逻辑
- 无 setData / 无 JS 帧动画 / 无新依赖
- `prefers-reduced-motion` 降级
