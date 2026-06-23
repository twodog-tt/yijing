#!/usr/bin/env bash
# 新 ECS 初始化：Docker、Docker Compose 插件、Nginx、部署目录
# 用法：sudo bash deploy/scripts/server-init.sh
# 不包含任何真实密钥

set -euo pipefail

DEPLOY_DIR="${DEPLOY_DIR:-/opt/yijing}"
BACKUP_DIR="${DEPLOY_DIR}/backups"
SWAP_FILE="${SWAP_FILE:-/swapfile}"
SWAP_SIZE_MB="${SWAP_SIZE_MB:-2048}"

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
  apt-get install -y ca-certificates curl gnupg lsb-release nginx git nano
elif command -v dnf &>/dev/null; then
  dnf install -y ca-certificates curl gnupg git nginx
fi

echo "==> 检查 2G swap..."
if swapon --show=NAME --noheadings | grep -Fxq "${SWAP_FILE}"; then
  echo "swap 已启用：${SWAP_FILE}"
elif [ -e "${SWAP_FILE}" ]; then
  echo "错误：${SWAP_FILE} 已存在但未启用；为避免覆盖，请手动检查后重试"
  exit 1
else
  if ! fallocate -l "${SWAP_SIZE_MB}M" "${SWAP_FILE}"; then
    dd if=/dev/zero of="${SWAP_FILE}" bs=1M count="${SWAP_SIZE_MB}" status=progress
  fi
  chmod 600 "${SWAP_FILE}"
  mkswap "${SWAP_FILE}"
  swapon "${SWAP_FILE}"
  if ! grep -Fq "${SWAP_FILE} none swap sw 0 0" /etc/fstab; then
    printf '%s none swap sw 0 0\n' "${SWAP_FILE}" >> /etc/fstab
  fi
fi
swapon --show

echo "==> 安装 Docker（若未安装）..."
if ! command -v docker &>/dev/null; then
  if ! curl -fsSL https://get.docker.com | sh; then
    if command -v apt-get &>/dev/null; then
      echo "Docker 官方安装地址不可用，改用 Ubuntu 软件源..."
      apt-get install -y docker.io docker-compose-v2
    else
      echo "错误：Docker 安装失败，请检查网络"
      exit 1
    fi
  fi
else
  echo "Docker 已安装：$(docker --version)"
fi
systemctl enable --now docker

if [ -n "${SUDO_USER:-}" ] && [ "${SUDO_USER}" != "root" ]; then
  usermod -aG docker "${SUDO_USER}"
  echo "已将 ${SUDO_USER} 加入 docker 组；首次部署前请退出 SSH 并重新登录"
fi

echo "==> 安装 Docker Compose 插件（若未安装）..."
if ! docker compose version &>/dev/null; then
  if command -v apt-get &>/dev/null; then
    apt-get install -y docker-compose-v2 || apt-get install -y docker-compose-plugin
  elif command -v dnf &>/dev/null; then
    dnf install -y docker-compose-plugin || true
  fi
fi
docker compose version

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
echo "  Swap：${SWAP_FILE} (${SWAP_SIZE_MB}M)"
echo ""
echo "下一步："
echo "  1. 将项目代码上传到 ${DEPLOY_DIR}"
echo "  2. 创建 .env.internal-test，并让 .env / .env.production 链接到它"
echo "  3. sudo cp deploy/nginx/yijing-ip.conf /etc/nginx/conf.d/yijing.conf"
echo "  4. sudo nginx -t && sudo systemctl reload nginx"
echo "  5. bash deploy/scripts/deploy.sh"
