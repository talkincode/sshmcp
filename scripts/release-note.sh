#!/bin/bash

# Get the latest two tags
tags=($(git tag --sort=-creatordate | head -n 2))

if [ ${#tags[@]} -lt 2 ]; then
    echo "❌ At least two tags are required to generate a commit list."
    exit 1
fi

latest_tag=${tags[0]}
previous_tag=${tags[1]}

echo "🔍 Generating commit list between ${previous_tag}..${latest_tag}..."

# Get commit list (including commit ID, commit message, and author)
commit_list=$(git log --pretty=format:"- %h %s (%an)" "${previous_tag}..${latest_tag}")

# Check if commit list is empty
if [ -z "$commit_list" ]; then
    echo "⚠️ No commits between these two versions."
    exit 0
fi

# Update GitHub Release description (using gh CLI)
echo "✍️ Updating release ${latest_tag} description..."

gh release edit "${latest_tag}" --notes "${commit_list}"

echo "✅ Release ${latest_tag} description has been successfully updated with commit list:"
echo
echo "${commit_list}"
