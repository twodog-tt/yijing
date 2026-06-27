# Release Regression Scripts

本目录存放体验版 / release 前可复用的只读或安全 POST 检查脚本。

## Scripts

- `check-miniprogram-static.sh`
  - 检查小程序 JS 语法。
  - 检查广告 / mock 解锁文案。
  - 检查强预测、改运、投资医疗法律等禁用词是否只出现在过滤器或边界否定说明。
  - 运行 `git diff --check`。

- `check-release-privacy.sh`
  - 检查小程序广告文案与强预测正向文案。
  - 检查跟踪文档中是否出现真实微信 AppID。
  - 检查跟踪文档中是否出现疑似真实 AppSecret、access_token 或 API key。
  - 不扫描 `.env*`，不输出真实密钥值。

- `check-api-smoke.sh`
  - 默认检查 `http://123.57.48.214` 与 `http://123.57.48.214/api/v1`。
  - 可用环境变量覆盖：

```bash
API_BASE=http://127.0.0.1:8080/api/v1 ROOT_BASE=http://127.0.0.1 bash scripts/check-api-smoke.sh
```

  - 动态创建测试记录，不依赖历史 id。
  - 覆盖 health、sessions、八字 v1/v2、奇门 v1/poc/professional、非法 algorithm_version、analysis free_unlock、问事起卦 create/unlock。
  - 只输出 id、algorithm_version、关键字段摘要，不输出完整 payload、完整报告正文或 session_key 响应内容。

## Release 前建议顺序

```bash
bash scripts/check-miniprogram-static.sh
bash scripts/check-release-privacy.sh
bash scripts/check-api-smoke.sh
git diff --check
git status --short
```

`check-api-smoke.sh` 会在 dev API 上创建新的测试记录。正式发布前仍需维护者本地完成微信 DevTools / 真机预览、相册权限与合法域名检查。
