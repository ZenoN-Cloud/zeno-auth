#!/bin/bash
# Migration script to GitLab

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}🚀 Migrating zeno-auth to GitLab${NC}"

# Add GitLab remote
git remote add gitlab git@gitlab.com:maxim.viazov/zeno-cy/zeno-auth.git

# Push all branches and tags
git push gitlab --all
git push gitlab --tags

echo -e "${GREEN}✅ Migration complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Run: ./deploy/gitlab-setup.sh"
echo "2. Check pipeline: https://gitlab.com/maxim.viazov/zeno-cy/zeno-auth/-/pipelines"