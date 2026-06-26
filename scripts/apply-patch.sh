#!/bin/bash

set -e

BRANCH_NAME_BUILD="playwright-build"
SCRIPTS_DIR="$(dirname "$0")"

echo "Applying patches..."
echo "==================="

pushd "$SCRIPTS_DIR/.."
PW_VERSION="$(grep -oE "playwrightCliVersion.+\"[0-9\.]+" ./run.go | sed 's/playwrightCliVersion = "/v/g')"
git submodule update --init
pushd playwright

git checkout HEAD --detach

# Always advance the submodule to the pinned version. Previously this only ran when
# a leftover playwright-build branch existed, so the very first roll on a fresh
# submodule patched the OLD version instead of the new tag.
git fetch --tags
git checkout "$PW_VERSION"

if git show-ref -q --heads "$BRANCH_NAME_BUILD"; then
  git branch -D "$BRANCH_NAME_BUILD"
fi

git checkout -b "$BRANCH_NAME_BUILD"
git apply --3way --whitespace=nowarn ../patches/*
git add -A
git commit -m "Applied patches"

popd