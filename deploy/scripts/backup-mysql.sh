#!/usr/bin/env bash
# MySQL 逻辑备份（Docker 内 yijing-mysql）
# 用法：在项目根目录 bash deploy/scripts/backup-mysql.sh
# 密码从 .env 读取，不写死在脚本中

set -euo pipefail
umask 077

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${ROOT_DIR}"

ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="docker-compose.prod.yml"
CONTAINER="${MYSQL_CONTAINER:-yijing-mysql}"
BACKUP_DIR="${BACKUP_DIR:-${ROOT_DIR}/backups}"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"

if [ ! -f "${ENV_FILE}" ]; then
  echo "错误：未找到 ${ENV_FILE}"
  exit 1
fi

# shellcheck disable=SC1090
set -a
source "${ENV_FILE}"
set +a

DB_NAME="${MYSQL_DATABASE:-${DB_NAME:-yijing}}"
ROOT_PASS="${MYSQL_ROOT_PASSWORD:-}"

if [ -z "${ROOT_PASS}" ]; then
  echo "错误：请在 .env 中设置 MYSQL_ROOT_PASSWORD"
  exit 1
fi

mkdir -p "${BACKUP_DIR}"
chmod 700 "${BACKUP_DIR}"
OUT_FILE="${BACKUP_DIR}/yijing_${DB_NAME}_${TIMESTAMP}.sql.gz"

echo "==> 备份 ${DB_NAME} -> ${OUT_FILE}"

docker compose -f "${COMPOSE_FILE}" exec -T mysql \
  sh -c 'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" exec mysqldump -u root \
    --single-transaction --routines --triggers "$MYSQL_DATABASE"' \
  | gzip > "${OUT_FILE}"

echo "==> 完成，大小：$(du -h "${OUT_FILE}" | cut -f1)"
