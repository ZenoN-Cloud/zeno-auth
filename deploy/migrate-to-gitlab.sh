#!/bin/bash
# Migration script to GitLab

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}🚀 Migrating zeno-auth to GitLab${NC}"

# Add GitLab remote
if ! git remote add gitlab git@gitlab.com:maxim.viazov/zeno-cy/zeno-auth.git; then
    echo "Remote 'gitlab' already exists or failed to add"
fi

# Push all branches and tags
if ! git push gitlab --all; then
    echo "Failed to push branches to GitLab"
    exit 1
fi

if ! git push gitlab --tags; then
    echo "Failed to push tags to GitLab"
    exit 1
fi

echo -e "${GREEN}✅ Migration complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Run: ./deploy/gitlab-setup.sh"
echo "2. Check pipeline: https://gitlab.com/maxim.viazov/zeno-cy/zeno-auth/-/pipelines"