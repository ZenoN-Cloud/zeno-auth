#!/bin/bash

set -e

echo "🦊 First push to GitLab setup"
echo ""

# Цвета для вывода
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Проверяем, что мы в git репозитории
if [ ! -d .git ]; then
    echo "❌ Not a git repository"
    exit 1
fi

# Настраиваем GitLab remote
echo -e "${BLUE}📋 Step 1: Setting up GitLab remote${NC}"
if git remote | grep -q "^gitlab$"; then
    echo "✅ GitLab remote already exists"
    git remote set-url gitlab git@gitlab.com:zeno-cy/zeno-auth.git
else
    git remote add gitlab git@gitlab.com:zeno-cy/zeno-auth.git
    echo "✅ Added GitLab remote"
fi

# Показываем текущие remotes
echo ""
echo -e "${BLUE}📋 Current remotes:${NC}"
git remote -v
echo ""

# Проверяем текущую ветку
CURRENT_BRANCH=$(git branch --show-current)
echo -e "${BLUE}📋 Current branch: ${GREEN}${CURRENT_BRANCH}${NC}"
echo ""

# Проверяем статус
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${YELLOW}⚠️  You have uncommitted changes${NC}"
    echo ""
    git status --short
    echo ""
    read -p "Do you want to commit them? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git add .
        echo ""
        echo "Enter commit message:"
        read COMMIT_MSG
        git commit -m "$COMMIT_MSG"
        echo -e "${GREEN}✅ Changes committed${NC}"
    else
        echo -e "${YELLOW}⚠️  Skipping commit${NC}"
    fi
fi

# Пушим в GitLab
echo ""
echo -e "${BLUE}📋 Step 2: Pushing to GitLab${NC}"
echo ""
read -p "Push to GitLab? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo -e "${BLUE}Pushing branch ${GREEN}${CURRENT_BRANCH}${BLUE}...${NC}"
    git push gitlab "$CURRENT_BRANCH"
    
    echo ""
    echo -e "${BLUE}Pushing tags...${NC}"
    git push gitlab --tags || echo "No tags to push"
    
    echo ""
    echo -e "${GREEN}✅ Successfully pushed to GitLab!${NC}"
    echo ""
    echo -e "${BLUE}📋 Next steps:${NC}"
    echo "1. Go to https://gitlab.com/zeno-cy/zeno-auth"
    echo "2. Setup CI/CD variables (see .gitlab/CI_VARIABLES.md)"
    echo "3. Create a merge request if needed"
    echo ""
    echo -e "${BLUE}📋 Useful commands:${NC}"
    echo "  git push gitlab main          # Push main branch"
    echo "  git push gitlab --all         # Push all branches"
    echo "  make gitlab-push              # Push current branch + tags"
else
    echo -e "${YELLOW}⚠️  Push cancelled${NC}"
fi
