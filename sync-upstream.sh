#!/usr/bin/env bash
#
# sync-upstream.sh — 把 gandli/fscanx 的自定义改动重新基于上游 killmonday/fscanx 最新代码
#
# 为什么需要它:
#   fork master 与 upstream master 是"无关历史"(unrelated histories),直接
#   `git merge upstream/master` 会产生全文件级虚假 diff 且可能冲突。
#   本脚本改用"以 upstream 为基底 + cherry-pick 我们的改动"的方式,
#   让冲突在可控的小范围内暴露,且永不 force-push master。
#
# 用法:
#   ./sync-upstream.sh            # 基于 upstream/master 开 sync 分支 + 开 PR
#   ./sync-upstream.sh --dry      # 只打印将 cherry-pick 的 commit,不实际操作
#
set -euo pipefail

DRY=false
[[ "${1:-}" == "--dry" ]] && DRY=true

# 当前分支必须干净
if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "❌ 工作区有未提交改动,先 stash 或 commit 再跑" >&2
  exit 1
fi

echo "→ fetch upstream + origin"
git fetch upstream
git fetch origin

UPSTREAM_HEAD=$(git rev-parse upstream/master)
ORIGIN_HEAD=$(git rev-parse origin/master)
DATE=$(date +%Y%m%d)

echo "  upstream/master = ${UPSTREAM_HEAD:0:7}"
echo "  origin/master   = ${ORIGIN_HEAD:0:7}"

# fork 相对上游的独有 commit(按时间正序,便于顺序 cherry-pick)
MAPS=$(git log --reverse --oneline upstream/master..origin/master | awk '{print $1}')
COUNT=$(echo "$MAPS" | grep -c . || true)
echo "→ 将 cherry-pick ${COUNT} 个本地改动到上游基底:"
echo "$MAPS" | while read -r c; do
  git log -1 --oneline "$c" 2>/dev/null | sed '    '
done

if $DRY; then
  echo "→ --dry 模式,不实际操作"
  exit 0
fi

BRANCH="sync/upstream-${DATE}"
echo "→ 基于 upstream/master 开分支 ${BRANCH}"
git checkout -B "$BRANCH" upstream/master

echo "→ cherry-pick 本地改动"
if ! git cherry-pick $MAPS; then
  echo "❌ cherry-pick 冲突,请手动解决后执行:" >&2
  echo "    git add -A && git cherry-pick --continue" >&2
  echo "    然后: git push origin $BRANCH && gh pr create --base master --head $BRANCH ..." >&2
  exit 1
fi

echo "→ 推分支"
git push origin "$BRANCH" --force-with-lease

echo "→ 开 PR 到 fork master"
gh pr create --repo gandli/fscanx --base master --head "$BRANCH" \
  --title "sync: rebase custom changes onto upstream/${DATE}" \
  --body "$(cat <<EOF
## Purpose
将本 fork 的自定义改动(CI/README)重新基于上游 killmonday/fscanx 最新代码,避免无关历史导致的合并冲突。

## Overview
- 基底:upstream/master (${UPSTREAM_HEAD:0:7})
- cherry-pick 本 fork 独有 commit:${COUNT} 个
- 若上游无新 commit,此 PR 等价于把自定义改动重新叠在原始基点上,diff 应仅含我们的改动

## Verification
- 基于 upstream/master 开分支,cherry-pick 无冲突
- CI (build) 应通过
EOF
)" || echo "⚠️ PR 创建失败(可能已存在),手动检查"

echo "✅ 同步分支已推送,PR 待审。合入即完成上游跟进。"
