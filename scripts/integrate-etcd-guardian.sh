#!/usr/bin/env bash
# 一次性脚本：将同级目录的 etcd-guardian 项目整合到 klaw/modules/etcd-guardian
set -euo pipefail

SRC="/Users/allengaller/Documents/GitHub/kudig-io/etcd-guardian"
DST="/Users/allengaller/Documents/GitHub/kudig-io/klaw/modules/etcd-guardian"

echo "==> 1. 拷贝源码（排除 .git/.qoder/node_modules/重复文件）"
mkdir -p "$DST"
rsync -a --delete \
  --exclude '.git' \
  --exclude '.qoder' \
  --exclude 'README 2.md' \
  --exclude 'web-ui/node_modules' \
  --exclude 'web-ui/dist' \
  "$SRC/" "$DST/"

echo "==> 2. 重写 go.mod module 路径"
cd "$DST"
go mod edit -module github.com/kudig-io/klaw/modules/etcd-guardian
cd "$DST/backend"
go mod edit -module github.com/kudig-io/klaw/modules/etcd-guardian/backend

echo "==> 3. 重写 Go 源码 import 路径"
cd "$DST"
find . -name '*.go' -not -path './web-ui/*' -print0 | xargs -0 sed -i '' \
  -e 's|github.com/etcdguardian/etcdguardian|github.com/kudig-io/klaw/modules/etcd-guardian|g' \
  -e 's|github.com/etcdguardian/backend|github.com/kudig-io/klaw/modules/etcd-guardian/backend|g'

echo "==> 4. 校验残留引用"
if grep -rn 'github.com/etcdguardian' --include='*.go' .; then
  echo "!! 仍有残留引用" && exit 1
fi
echo "OK: 无残留引用"
