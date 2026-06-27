#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

AD_PATTERN='观看视频|广告解锁|模拟广告|rewarded_video_mock'
FORBIDDEN_PATTERN='精准预测|必成|必败|大吉|大凶|必发财|必复合|改运|化灾|转运|投资建议|医疗建议|法律建议|赌博建议|军事行动建议'

PASS_COUNT=0
FAIL_COUNT=0

pass() {
  PASS_COUNT=$((PASS_COUNT + 1))
  printf 'PASS %s\n' "$1"
}

fail() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  printf 'FAIL %s\n' "$1" >&2
}

tmp_ad="$(mktemp)"
tmp_forbidden="$(mktemp)"
tmp_forbidden_bad="$(mktemp)"
tmp_appid="$(mktemp)"
tmp_secret="$(mktemp)"
trap 'rm -f "$tmp_ad" "$tmp_forbidden" "$tmp_forbidden_bad" "$tmp_appid" "$tmp_secret"' EXIT

printf '== release privacy checks ==\n'

if grep -RInE "$AD_PATTERN" miniprogram >"$tmp_ad" 2>/dev/null; then
  fail "ad/mock unlock copy found in miniprogram"
  awk -F: '{ printf "%s:%s: <redacted>\n", $1, $2 }' "$tmp_ad" >&2
else
  pass "no ad/mock unlock copy in miniprogram"
fi

grep -RInE "$FORBIDDEN_PATTERN" miniprogram >"$tmp_forbidden" 2>/dev/null || true
if [[ -s "$tmp_forbidden" ]]; then
  awk -F: '
    $1 == "miniprogram/utils/long-poster-canvas.js" { next }
    $1 == "miniprogram/utils/home.js" { next }
    $1 == "miniprogram/pages/about/about.wxml" { next }
    { print }
  ' "$tmp_forbidden" >"$tmp_forbidden_bad"

  if [[ -s "$tmp_forbidden_bad" ]]; then
    fail "forbidden positive prediction copy found outside filters/boundary notes"
    awk -F: '{ printf "%s:%s: <redacted>\n", $1, $2 }' "$tmp_forbidden_bad" >&2
  else
    pass "forbidden terms only in filters/boundary notes"
  fi
else
  pass "no forbidden prediction copy in miniprogram"
fi

git ls-files 'docs/*.md' |
  xargs awk '
    /wx[0-9a-fA-F]{16,}/ {
      printf "%s:%d: <redacted-wechat-appid>\n", FILENAME, FNR
    }
  ' >"$tmp_appid"
if [[ -s "$tmp_appid" ]]; then
  fail "real-looking WeChat AppID found in tracked docs"
  cat "$tmp_appid" >&2
else
  pass "no real-looking WeChat AppID in tracked docs"
fi

git ls-files 'docs/*.md' |
  xargs awk '
    BEGIN { IGNORECASE = 1 }
    /(appsecret|app_secret)[[:space:]]*[:=][[:space:]]*["'\'']?[A-Za-z0-9._-]{12,}/ {
      printf "%s:%d: <redacted-secret>\n", FILENAME, FNR
    }
    /access_token[[:space:]]*[:=][[:space:]]*["'\'']?[A-Za-z0-9._-]{16,}/ {
      printf "%s:%d: <redacted-token>\n", FILENAME, FNR
    }
    /(OPENAI_API_KEY|DEEPSEEK_API_KEY|api[_-]?key)[[:space:]]*[:=][[:space:]]*["'\'']?[A-Za-z0-9._-]{16,}/ {
      printf "%s:%d: <redacted-api-key>\n", FILENAME, FNR
    }
    /sk-[A-Za-z0-9]{20,}/ {
      printf "%s:%d: <redacted-sk-key>\n", FILENAME, FNR
    }
  ' >"$tmp_secret"
if [[ -s "$tmp_secret" ]]; then
  fail "possible real secret/token/API key found in tracked docs"
  cat "$tmp_secret" >&2
else
  pass "no real-looking AppSecret/access_token/API key in tracked docs"
fi

printf '== summary: %d PASS, %d FAIL ==\n' "$PASS_COUNT" "$FAIL_COUNT"
if [[ "$FAIL_COUNT" -ne 0 ]]; then
  exit 1
fi
