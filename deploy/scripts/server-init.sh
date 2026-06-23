#!/usr/bin/env bash
# 新 ECS 初始化：Docker、Docker Compose 插件、Nginx、部署目录
# 用法：sudo bash deploy/scripts/server-init.sh
# 不包含任何真实密钥

set -euo pipefail

DEPLOY_DIR="${DEPLOY_DIR:-/opt/yijing}"
BACKUP_DIR="${DEPLOY_DIR}/backups"

echo "==> 更新系统包..."
if command -v dnf &>/dev/null; then
  dnf makecache -y || true
  dnf update -y || true
elif command -v apt-get &>/dev/null; then
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -y
  apt-get upgrade -y || true
else
  echo "警告：未识别的包管理器，请手动安装 Docker 与 Nginx"
fi

echo "==> 安装基础工具..."
if command -v apt-get &>/dev/null; then
  apt-get install -y ca-certificates curl gnupg lsb-release nginx git
elif command -v dnf &>/dev/null; then
  dnf install -y ca-certificates curl gnupg git nginx
fi

echo "==> 安装 Docker（若未安装）..."
if ! command -v docker &>/dev/null; then
  curl -fsSL https://get.docker.com | sh
  systemctl enable docker
  systemctl start docker
else
  echo "Docker 已安装：$(docker --version)"
fi

echo "==> 安装 Docker Compose 插件（若未安装）..."
if ! docker compose version &>/dev/null; then
  if command -v apt-get &>/dev/null; then
    apt-get install -y docker-compose-plugin || true
  elif command -v dnf &>/dev/null; then
    dnf install -y docker-compose-plugin || true
  fi
fi
docker compose version || echo "警告：请确认 docker compose 插件可用"

echo "==> 启用 Nginx..."
systemctl enable nginx
systemctl start nginx || true

echo "==> 创建部署目录 ${DEPLOY_DIR} ..."
mkdir -p "${DEPLOY_DIR}" "${BACKUP_DIR}"
chmod 755 "${DEPLOY_DIR}"

if [ -n "${SUDO_USER:-}" ] && [ "${SUDO_USER}" != "root" ]; then
  chown -R "${SUDO_USER}:${SUDO_USER}" "${DEPLOY_DIR}"
fi

echo ""
echo "初始化完成。"
echo "  部署目录：${DEPLOY_DIR}"
echo "  备份目录：${BACKUP_DIR}"
echo ""
echo "下一步："
echo "  1. 将项目代码上传到 ${DEPLOY_DIR}"
echo "  2. cp .env.internal-test.example .env 并替换 SERVER_IP / 密码"
echo "  3. sudo cp deploy/nginx/yijing-ip.conf /etc/nginx/conf.d/yijing.conf"
echo "  4. sudo nginx -t && sudo systemctl reload nginx"
echo "  5. bash deploy/scripts/deploy.sh"
