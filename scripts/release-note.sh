#!/bin/bash

# 获取最新两个tag
tags=($(git tag --sort=-creatordate | head -n 2))

if [ ${#tags[@]} -lt 2 ]; then
  echo "❌ 至少需要两个 tag 才能生成提交清单。"
  exit 1
fi

latest_tag=${tags[0]}
previous_tag=${tags[1]}

echo "🔍 正在生成 ${previous_tag}..${latest_tag} 之间的提交清单..."

# 获取提交清单（包含提交ID、提交说明和作者）
commit_list=$(git log --pretty=format:"- %h %s (%an)" "${previous_tag}..${latest_tag}")

# 判断提交清单是否为空
if [ -z "$commit_list" ]; then
  echo "⚠️ 这两个版本之间没有提交。"
  exit 0
fi

# 更新 GitHub Release 描述（使用 gh CLI）
echo "✍️ 正在更新 release ${latest_tag} 的描述..."

gh release edit "${latest_tag}" --notes "${commit_list}"

echo "✅ Release ${latest_tag} 的描述已成功更新为提交清单："
echo
echo "${commit_list}"