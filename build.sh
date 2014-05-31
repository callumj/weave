#!/bin/bash

CURRENT_BRANCH=`git branch | sed -n '/\* /s///p'`
VERSION=`cat VERSION`

if ! [[ "${CURRENT_BRANCH}" == "master" ]]
then
  if [[ "${CURRENT_BRANCH}" == "development" ]]
  then
    VERSION="${VERSION}-dev"
  else
    echo "Builds are only performed on master!"
    exit -1
  fi
fi

# vet the source (capture errors because the current version does not use exit statuses currently)
VET=`go tool vet . 2>&1 >/dev/null`

if ! [ -n "$VET" ]
then
  echo "All good"
  goxc -pv ${VERSION} -d builds go-vet go-test go-install xc codesign copy-resources archive-zip archive-tar-gz pkg-build rmbin
else
  echo "$VET"
  exit -1
fi