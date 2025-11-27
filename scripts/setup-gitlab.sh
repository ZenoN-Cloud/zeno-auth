#!/bin/bash

set -e

echo "🦊 Setting up GitLab remote..."

# Проверяем, есть ли уже gitlab remote
if git remote | grep -q "^gitlab$"; then
    echo "✅ GitLab remote already exists"
    git remote set-url gitlab git@gitlab.com:zeno-cy/zeno-auth.git
    echo "✅ Updated GitLab remote URL"
else
    git remote add gitlab git@gitlab.com:zeno-cy/zeno-auth.git
    echo "✅ Added GitLab remote"
fi

# Показываем все remotes
echo ""
echo "📋 Current remotes:"
git remote -v

echo ""
echo "🚀 Ready to push to GitLab!"
echo ""
echo "Commands:"
echo "  git push gitlab main          # Push main branch"
echo "  git push gitlab --all         # Push all branches"
echo "  git push gitlab --tags        # Push all tags"
echo "  make gitlab-push              # Push current branch + tags"
