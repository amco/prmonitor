#!/bin/bash

BRANCH=`git rev-parse --abbrev-ref HEAD`

if [ "$BRANCH" != "master" ]; then
    echo "Error: must be on branch 'master', currently on '$BRANCH'"
    exit 1
fi

RELEASE=`date +%s`
ORIGINAL=`git branch`

godep update ./...
git checkout -b release-$RELEASE

# add vendor dependencies
git add -f vendor
git commit -m "release $RELEASE"
git push -f heroku "release-$RELEASE":master
git checkout master