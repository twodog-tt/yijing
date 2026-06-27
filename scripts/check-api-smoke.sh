#!/usr/bin/env bash
set -euo pipefail

API_BASE="${API_BASE:-http://123.57.48.214/api/v1}"
ROOT_BASE="${ROOT_BASE:-http://123.57.48.214}"
CURL_TIMEOUT="${CURL_TIMEOUT:-120}"

PASS_COUNT=0
FAIL_COUNT=0
LAST_HTTP_CODE=""

pass() {
  PASS_COUNT=$((PASS_COUNT + 1))
  printf 'PASS %s\n' "$1"
}

fail() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  printf 'FAIL %s\n' "$1" >&2
}

have_jq() {
  command -v jq >/dev/null 2>&1
}

json_get() {
  local file="$1"
  local path="$2"
  if have_jq; then
    jq -r ".$path // empty" "$file"
  else
    node - "$path" "$file" <<'NODE'
const fs = require("fs");
const path = process.argv[2].split(".");
const file = process.argv[3];
let value = JSON.parse(fs.readFileSync(file, "utf8"));
for (const key of path) {
  if (value == null) process.exit(0);
  value = /^[0-9]+$/.test(key) ? value[Number(key)] : value[key];
}
if (value == null) process.exit(0);
if (typeof value === "object") console.log(JSON.stringify(value));
else console.log(String(value));
NODE
  fi
}

json_length() {
  local file="$1"
  local path="$2"
  if have_jq; then
    jq -r ".$path | if type == \"array\" then length else 0 end" "$file"
  else
    node - "$path" "$file" <<'NODE'
const fs = require("fs");
const path = process.argv[2].split(".");
const file = process.argv[3];
let value = JSON.parse(fs.readFileSync(file, "utf8"));
for (const key of path) {
  if (value == null) {
    console.log("0");
    process.exit(0);
  }
  value = /^[0-9]+$/.test(key) ? value[Number(key)] : value[key];
}
console.log(Array.isArray(value) ? String(value.length) : "0");
NODE
  fi
}

curl_json() {
  local method="$1"
  local url="$2"
  local body="$3"
  local out="$4"
  shift 4
  local args=(-sS --max-time "$CURL_TIMEOUT" -o "$out" -w "%{http_code}" -X "$method" "$url" -H "Content-Type: application/json")
  if [[ -n "$body" ]]; then
    args+=(-d "$body")
  fi
  while [[ "$#" -gt 0 ]]; do
    args+=("$1")
    shift
  done
  LAST_HTTP_CODE="$(curl "${args[@]}")"
}

expect_code_zero() {
  local file="$1"
  [[ "$(json_get "$file" "code")" == "0" ]]
}

create_session() {
  local session_key="$1"
  local out="$2"
  curl_json POST "$API_BASE/sessions" "{\"session_key\":\"$session_key\"}" "$out"
  if [[ "$LAST_HTTP_CODE" == "200" ]] && expect_code_zero "$out" && [[ -n "$(json_get "$out" "data.session_id")" ]]; then
    pass "session create"
  else
    fail "session create (http=$LAST_HTTP_CODE)"
    return 1
  fi
}

create_bazi() {
  local label="$1"
  local session_key="$2"
  local algorithm_version="$3"
  local out="$4"
  local payload
  if [[ -n "$algorithm_version" ]]; then
    payload="{\"session_key\":\"$session_key\",\"birth_date\":\"1992-03-18\",\"birth_hour_branch\":\"mao\",\"birth_hour_unknown\":false,\"confirm_disclaimer\":true,\"algorithm_version\":\"$algorithm_version\"}"
  else
    payload="{\"session_key\":\"$session_key\",\"birth_date\":\"1992-03-18\",\"birth_hour_branch\":\"mao\",\"birth_hour_unknown\":false,\"confirm_disclaimer\":true}"
  fi
  curl_json POST "$API_BASE/analysis/bazi" "$payload" "$out" -H "X-Session-Key: $session_key"
  local id returned_version
  id="$(json_get "$out" "data.id")"
  returned_version="$(json_get "$out" "data.algorithm_version")"
  if [[ "$LAST_HTTP_CODE" == "200" ]] && expect_code_zero "$out" && [[ -n "$id" ]]; then
    printf '  %s id=%s algorithm_version=%s\n' "$label" "$id" "$returned_version"
    pass "$label create"
  else
    fail "$label create (http=$LAST_HTTP_CODE)"
    return 1
  fi
}

create_bazi_unknown_hour() {
  local label="$1"
  local session_key="$2"
  local out="$3"
  local payload="{\"session_key\":\"$session_key\",\"birth_date\":\"1992-03-18\",\"birth_hour_unknown\":true,\"confirm_disclaimer\":true,\"algorithm_version\":\"bazi-v2-poc\"}"
  curl_json POST "$API_BASE/analysis/bazi" "$payload" "$out" -H "X-Session-Key: $session_key"
  local id returned_version hour_v1 hour_v2
  id="$(json_get "$out" "data.id")"
  returned_version="$(json_get "$out" "data.algorithm_version")"
  hour_v1="$(json_get "$out" "data.result_payload.pillars.hour")"
  hour_v2="$(json_get "$out" "data.result_payload.pillars_v2.hour")"
  if [[ "$LAST_HTTP_CODE" == "200" ]] &&
    expect_code_zero "$out" &&
    [[ -n "$id" ]] &&
    [[ "$returned_version" == "bazi-v2-poc" ]] &&
    [[ -z "$hour_v1" ]] &&
    [[ -z "$hour_v2" ]]; then
    printf '  %s id=%s algorithm_version=%s birth_hour_unknown=true hour_pillar=none\n' "$label" "$id" "$returned_version"
    pass "$label create"
  else
    fail "$label create (http=$LAST_HTTP_CODE)"
    return 1
  fi
}

create_qimen() {
  local label="$1"
  local session_key="$2"
  local algorithm_version="$3"
  local out="$4"
  local payload
  if [[ -n "$algorithm_version" ]]; then
    payload="{\"session_key\":\"$session_key\",\"question\":\"我最近适合如何推进这个计划的节奏？\",\"category\":\"career\",\"confirm_disclaimer\":true,\"algorithm_version\":\"$algorithm_version\"}"
  else
    payload="{\"session_key\":\"$session_key\",\"question\":\"我最近适合如何推进这个计划的节奏？\",\"category\":\"career\",\"confirm_disclaimer\":true}"
  fi
  curl_json POST "$API_BASE/analysis/qimen" "$payload" "$out" -H "X-Session-Key: $session_key"
  local id returned_version palace_count
  id="$(json_get "$out" "data.id")"
  returned_version="$(json_get "$out" "data.algorithm_version")"
  palace_count="$(json_length "$out" "data.result_payload.palaces")"
  if [[ "$LAST_HTTP_CODE" == "200" ]] && expect_code_zero "$out" && [[ -n "$id" ]]; then
    printf '  %s id=%s algorithm_version=%s palaces=%s\n' "$label" "$id" "$returned_version" "$palace_count"
    pass "$label create"
  else
    fail "$label create (http=$LAST_HTTP_CODE)"
    return 1
  fi
}

unlock_analysis() {
  local label="$1"
  local session_key="$2"
  local id="$3"
  local out="$4"
  curl_json POST "$API_BASE/analysis/$id/unlock" '{"unlock_type":"free_unlock"}' "$out" -H "X-Session-Key: $session_key"
  local unlock_status has_full
  unlock_status="$(json_get "$out" "data.unlock_status")"
  has_full="$(json_get "$out" "data.full_content")"
  if [[ "$LAST_HTTP_CODE" == "200" ]] && expect_code_zero "$out" && [[ "$unlock_status" == "1" ]] && [[ -n "$has_full" ]]; then
    pass "$label free_unlock"
  else
    fail "$label free_unlock (http=$LAST_HTTP_CODE)"
    return 1
  fi
}

unlock_bazi_unknown_hour() {
  local label="$1"
  local session_key="$2"
  local id="$3"
  local out="$4"
  curl_json POST "$API_BASE/analysis/$id/unlock" '{"unlock_type":"free_unlock"}' "$out" -H "X-Session-Key: $session_key"
  local unlock_status full_content
  unlock_status="$(json_get "$out" "data.unlock_status")"
  full_content="$(json_get "$out" "data.full_content")"
  if [[ "$LAST_HTTP_CODE" == "200" ]] &&
    expect_code_zero "$out" &&
    [[ "$unlock_status" == "1" ]] &&
    [[ -n "$full_content" ]] &&
    { [[ "$full_content" == *"时辰未知"* ]] || [[ "$full_content" == *"未生成时柱"* ]] || [[ "$full_content" == *"不生成时柱"* ]]; } &&
    [[ ! "$full_content" =~ 时柱[：:][[:space:]]*[甲乙丙丁戊己庚辛壬癸][子丑寅卯辰巳午未申酉戌亥] ]]; then
    pass "$label free_unlock"
  else
    fail "$label free_unlock (http=$LAST_HTTP_CODE)"
    return 1
  fi
}

create_divination() {
  local session_key="$1"
  local out="$2"
  local payload="{\"session_key\":\"$session_key\",\"category_id\":1,\"question\":\"我适合如何安排这周的学习和休息节奏？\",\"confirm_disclaimer\":true}"
  curl_json POST "$API_BASE/divinations" "$payload" "$out"
  local id line_count
  id="$(json_get "$out" "data.id")"
  line_count="$(json_length "$out" "data.lines")"
  if [[ "$LAST_HTTP_CODE" == "200" ]] && expect_code_zero "$out" && [[ -n "$id" ]] && [[ "$line_count" == "6" ]]; then
    printf '  divination id=%s lines=%s\n' "$id" "$line_count"
    pass "divination create"
  else
    fail "divination create (http=$LAST_HTTP_CODE)"
    return 1
  fi
}

unlock_divination() {
  local session_key="$1"
  local id="$2"
  local out="$3"
  local payload="{\"session_key\":\"$session_key\",\"unlock_type\":\"mock_button\"}"
  curl_json POST "$API_BASE/divinations/$id/unlock" "$payload" "$out"
  if [[ "$LAST_HTTP_CODE" == "200" ]] && expect_code_zero "$out" && [[ "$(json_get "$out" "data.unlock_status")" == "1" ]]; then
    pass "divination mock_button unlock"
  else
    fail "divination mock_button unlock (http=$LAST_HTTP_CODE)"
    return 1
  fi
}

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

printf '== API smoke checks ==\n'
printf 'API_BASE=%s\nROOT_BASE=%s\n' "$API_BASE" "$ROOT_BASE"
if ! have_jq; then
  printf 'INFO jq not found; using Node.js JSON parser fallback.\n'
fi

curl_json GET "$ROOT_BASE/health" "" "$tmp_dir/root-health.json"
if [[ "$LAST_HTTP_CODE" == "200" ]] && [[ "$(json_get "$tmp_dir/root-health.json" "status")" == "ok" ]]; then
  pass "GET /health"
else
  fail "GET /health (http=$LAST_HTTP_CODE)"
fi

curl_json GET "$API_BASE/health" "" "$tmp_dir/api-health.json"
if [[ "$LAST_HTTP_CODE" == "200" ]] && [[ "$(json_get "$tmp_dir/api-health.json" "status")" == "ok" ]]; then
  pass "GET /api/v1/health"
else
  fail "GET /api/v1/health (http=$LAST_HTTP_CODE)"
fi

create_session "test-smoke-bazi-v1" "$tmp_dir/session.json"

create_bazi "bazi-simple-v1" "test-smoke-bazi-v1" "" "$tmp_dir/bazi-v1.json"
bazi_v1_id="$(json_get "$tmp_dir/bazi-v1.json" "data.id")"

create_bazi "bazi-v2-poc" "test-smoke-bazi-v2" "bazi-v2-poc" "$tmp_dir/bazi-v2.json"

create_bazi_unknown_hour "bazi-v2-poc unknown-hour" "test-smoke-bazi-v2-unknown" "$tmp_dir/bazi-v2-unknown.json"
bazi_v2_unknown_id="$(json_get "$tmp_dir/bazi-v2-unknown.json" "data.id")"

create_qimen "qimen-simple-v1" "test-smoke-qimen-v1" "" "$tmp_dir/qimen-v1.json"
qimen_v1_id="$(json_get "$tmp_dir/qimen-v1.json" "data.id")"

create_qimen "qimen-v2-poc" "test-smoke-qimen-poc" "qimen-v2-poc" "$tmp_dir/qimen-poc.json"
create_qimen "qimen-v2-professional" "test-smoke-qimen-prof" "qimen-v2-professional" "$tmp_dir/qimen-prof.json"

illegal_payload='{"session_key":"test-smoke-qimen-illegal","question":"我最近适合如何推进这个计划的节奏？","category":"career","confirm_disclaimer":true,"algorithm_version":"qimen-v3"}'
curl_json POST "$API_BASE/analysis/qimen" "$illegal_payload" "$tmp_dir/qimen-illegal.json" -H "X-Session-Key: test-smoke-qimen-illegal"
illegal_message="$(json_get "$tmp_dir/qimen-illegal.json" "message")"
if [[ "$LAST_HTTP_CODE" == "400" ]] && [[ "$illegal_message" == *"algorithm_version"* ]]; then
  pass "illegal qimen algorithm_version rejected"
else
  fail "illegal qimen algorithm_version rejected (http=$LAST_HTTP_CODE)"
fi

unlock_analysis "bazi-simple-v1" "test-smoke-bazi-v1" "$bazi_v1_id" "$tmp_dir/bazi-unlock.json"
unlock_bazi_unknown_hour "bazi-v2-poc unknown-hour" "test-smoke-bazi-v2-unknown" "$bazi_v2_unknown_id" "$tmp_dir/bazi-v2-unknown-unlock.json"
unlock_analysis "qimen-simple-v1" "test-smoke-qimen-v1" "$qimen_v1_id" "$tmp_dir/qimen-unlock.json"

create_divination "test-smoke-divination" "$tmp_dir/divination.json"
divination_id="$(json_get "$tmp_dir/divination.json" "data.id")"
unlock_divination "test-smoke-divination" "$divination_id" "$tmp_dir/divination-unlock.json"

printf '== summary: %d PASS, %d FAIL ==\n' "$PASS_COUNT" "$FAIL_COUNT"
if [[ "$FAIL_COUNT" -ne 0 ]]; then
  exit 1
fi
