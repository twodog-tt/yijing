#!/usr/bin/env bash
# 部署或更新易经 MVP（内测）
# 用法：在项目根目录执行 bash deploy/scripts/deploy.sh
# 不包含真实密钥；需事先配置 .env

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${ROOT_DIR}"

COMPOSE_FILE="docker-compose.prod.yml"
ENV_FILE="${ENV_FILE:-.env}"
MAX_WAIT="${MAX_WAIT:-120}"
export COMPOSE_PARALLEL_LIMIT="${COMPOSE_PARALLEL_LIMIT:-1}"

echo "==> 项目目录：${ROOT_DIR}"

# ---------- 更新代码（占位，按实际选择一种）----------
# git pull origin main
# 或 rsync / scp 上传后跳过此步

if [ ! -f "${ENV_FILE}" ]; then
  echo "错误：未找到 ${ENV_FILE}"
  echo "请先创建 .env.internal-test，并让 .env 链接到它"
  echo "然后替换 SERVER_IP、数据库密码、（可选）DEEPSEEK_API_KEY"
  exit 1
fi

# shellcheck disable=SC1090
set -a
source "${ENV_FILE}"
set +a

if grep -Eq '^[A-Z0-9_]+=.*CHANGE_ME' "${ENV_FILE}" 2>/dev/null; then
  echo "错误：${ENV_FILE} 中仍包含 CHANGE_ME 占位密码，请修改后再部署"
  exit 1
fi

PUBLIC_IP="${SERVER_IP:-}"
if [ -z "${PUBLIC_IP}" ] || [ "${PUBLIC_IP}" = "SERVER_IP" ]; then
  echo "错误：请在 ${ENV_FILE} 中设置真实 SERVER_IP"
  exit 1
fi

HEALTH_URL="${HEALTH_URL:-http://127.0.0.1:${BACKEND_PORT:-8080}/api/v1/health}"

echo "==> 串行构建容器（适配 2G 内存 ECS）..."
docker compose -f "${COMPOSE_FILE}" --env-file "${ENV_FILE}" build backend
docker compose -f "${COMPOSE_FILE}" --env-file "${ENV_FILE}" build frontend
echo "==> 启动容器..."
docker compose -f "${COMPOSE_FILE}" --env-file "${ENV_FILE}" up -d

echo "==> 等待服务就绪（migrate 在 backend 启动时自动执行）..."
elapsed=0
until curl -sf "${HEALTH_URL}" >/dev/null 2>&1; do
  sleep 3
  elapsed=$((elapsed + 3))
  if [ "${elapsed}" -ge "${MAX_WAIT}" ]; then
    echo "错误：健康检查超时 ${HEALTH_URL}"
    docker compose -f "${COMPOSE_FILE}" --env-file "${ENV_FILE}" ps
    docker compose -f "${COMPOSE_FILE}" --env-file "${ENV_FILE}" logs --tail=50 backend
    exit 1
  fi
  echo "  等待中... (${elapsed}s)"
done

echo "==> 健康检查通过"
curl -s "${HEALTH_URL}" | head -c 500
echo ""

# backend entrypoint 已自动执行；再次运行用于明确确认所有版本均已记录。
echo "==> 确认 migration 状态..."
docker compose -f "${COMPOSE_FILE}" --env-file "${ENV_FILE}" exec -T backend ./migrate

echo ""
echo "=========================================="
echo "  部署完成"
if [ -n "${PUBLIC_IP}" ] && [ "${PUBLIC_IP}" != "SERVER_IP" ]; then
  echo "  访问地址：http://${PUBLIC_IP}/"
  echo "  API 示例：http://${PUBLIC_IP}/api/v1/health"
else
  echo "  请在浏览器访问：http://<你的ECS公网IP>/"
fi
echo "=========================================="
echo ""
echo "日志：docker compose -f ${COMPOSE_FILE} --env-file ${ENV_FILE} logs -f --tail=100 backend"
echo "备份：bash deploy/scripts/backup-mysql.sh"
