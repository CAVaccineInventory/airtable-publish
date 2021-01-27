#!/usr/bin/env bash

cd "$(dirname "$0")"

set -eu

CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" = "main" ]; then
	FUNCTION_NAME=freshLocationsStaging
	DEPLOY_NAME="staging"
elif [ "$CURRENT_BRANCH" = "prod" ]; then
	FUNCTION_NAME=freshLocations
	DEPLOY_NAME="prod"
else
	echo "Unknown branch '$CURRENT_BRANCH' -- aborting!"
	echo
	echo "Please run this from either the 'prod' or 'main' branches."
	exit 1
fi

if [ -n "$(git status --untracked-files=no --porcelain)" ]; then
	echo "Untracked changes in local files -- aborting!"
	echo
	echo "Because this deploys based on the files in the working copy, please"
	echo "ensure that the working directory has no uncommitted changes."
	exit 1
fi

git pull

echo "Deploying monitoring to $DEPLOY_NAME..."
echo

gcloud functions deploy \
	"$FUNCTION_NAME" \
	--project cavaccineinventory \
	--entry-point CheckFreshness \
	--runtime go113 \
	--set-env-vars "DEPLOY=$DEPLOY_NAME" \
	--trigger-http \
	--allow-unauthenticated \
	--source=.
