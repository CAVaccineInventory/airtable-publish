#!/usr/bin/env bash

set -eu

WHAT=${1:-pipeline}
if [ "$WHAT" = "monitoring" ]; then
	DEPLOY_BRANCH=prod-monitoring
	DEPLOY_LIMIT=origin/prod
elif [ "$WHAT" = "pipeline" ]; then
	DEPLOY_BRANCH=prod
	DEPLOY_LIMIT=origin/main
else
	cat <<EOF

Unknown deploy "$WHAT"


USAGE:
    $(dirname "$0")/deploy.sh [deploy]

This script deploys production by merging and pushing branches:

 - For 'pipeline' deploys (the default), 'main' is merged into 'prod'.

 - For 'monitoring' deploys, 'main' is merged into 'prod-monitoring',
   but no further than 'prod'; this ensures that the monitoring will
   not begin paging on anything that is not yet being published.

EOF
	exit 1
fi

if [ -n "$(git status --untracked-files=no --porcelain)" ]; then
	echo "Untracked changes in local files -- aborting!"
	echo
	echo "For simplicity, this manipulates your working tree to do"
	echo "the merge; this requires a clean working copy.  Stash your"
	echo "changes and try again."
	exit 1
fi

# Make sure we have the most up-to-date main
git fetch --quiet origin

# Figure out which commits we're deploying upto; for pipeline, this is
# just main.  For monitoring, it's also limited by what `prod` includes.
DEPLOY_LIMIT=$(git merge-base origin/main "$DEPLOY_LIMIT")

# First, a couple safety checks:

# First, verify that prod's tree is identical its most recent merge-base with
# main; that is, it has no changes that were not released on `main`.
ORIGIN_DEPLOY="origin/$DEPLOY_BRANCH"
MERGE_BASE=$(git merge-base "$DEPLOY_LIMIT" "$ORIGIN_DEPLOY")
if ! git diff "$MERGE_BASE" "$ORIGIN_DEPLOY" --exit-code; then
	echo "$DEPLOY_BRANCH tree has diverged from what was released on main!"
	echo
	git --no-pager diff --stat "$MERGE_BASE" "$ORIGIN_DEPLOY"
	echo
	git --no-pager log --no-decorate --oneline "$ORIGIN_DEPLOY" "^$DEPLOY_LIMIT" --max-parents=1
	exit 1
fi

# Second, verify that prod has no commits, besides merge commits, that
# are not on `main`.  Given the above check, this is only possible if
# there are changes that are a net no-change result, which would be
# odd.
if [ 0 -ne "$(git rev-list "$ORIGIN_DEPLOY" "^$DEPLOY_LIMIT" --max-parents=1 | wc -l)" ]; then
	echo "$DEPLOY_BRANCH contains non-merge commits that are not in main!"
	echo
	git --no-pager log --no-decorate --oneline "$ORIGIN_DEPLOY" "^$DEPLOY_LIMIT" --max-parents=1
	exit 1
fi

echo
echo "You are about to deploy the following **$WHAT** commits:"
if [ "$WHAT" != "pipeline" ]; then
	echo "  (this is limited to go no further than current 'prod')"
fi
echo '```'
git --no-pager log --no-decorate --oneline "$DEPLOY_LIMIT" "^$ORIGIN_DEPLOY"
echo '```'

echo
echo "Type 'yes' to confirm they look right, and that you have gotten a :thumbsup: from #operations:"
read -r VERIFY
if [ "$VERIFY" != "yes" ]; then
	exit 1
fi

# So we don't create, or rely on the state of a local "prod" branch,
# work on a detached HEAD
CURRENT_BRANCH=$(git branch --show-current)
echo "Checking out detached $DEPLOY_BRANCH..."
git checkout --detach "$ORIGIN_DEPLOY"
echo
echo

# Always create a merge commit, as a marker of what was deployed at
# once.
echo "Merging main (up to $DEPLOY_LIMIT)..."
git merge --no-edit --no-ff "$DEPLOY_LIMIT"
echo
echo

# Push the newly-generated merge commit
echo "Pushing new $DEPLOY_BRANCH to origin..."
git push origin "HEAD:refs/heads/$DEPLOY_BRANCH"
echo
echo

echo "Switching back to $CURRENT_BRANCH..."
git checkout "$CURRENT_BRANCH"
echo
echo

echo "Done!"
