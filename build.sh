#!/bin/bash

if [[ -z "$BUILDBOX_BRANCH" ]]
then
  BUILDBOX_BRANCH=`git branch | sed -n '/\* /s///p'`
fi
VERSION=`cat VERSION`

if ! [[ "${BUILDBOX_BRANCH}" == "master" ]]
then
  if [[ "${BUILDBOX_BRANCH}" == "development" ]]
  then
    VERSION="${VERSION}-dev"
  else
    echo "Builds are only performed on master!"
    exit -1
  fi
fi

rm -r -f tmp/go
rm -r -f builds/${VERSION}

# vet the source (capture errors because the current version does not use exit statuses currently)
echo "Vetting..."
VET=`go tool vet . 2>&1 >/dev/null`

cur=`pwd` 

if ! [ -n "$VET" ]
then
  echo "All good"
  mkdir -p tmp/go
  mkdir -p builds/
  mkdir tmp/go/src tmp/go/bin tmp/go/pkg
  mkdir -p tmp/go/src/github.com/callumj/weave
  cp -R app core remote tools main.go tmp/go/src/github.com/callumj/weave/

  go_src=$'package tools\nconst WeaveVersion string = "'"${VERSION}"$'"'
  echo "$go_src" > tmp/go/src/github.com/callumj/weave/tools/version.go

  mkdir -p builds/${VERSION}/darwin_386 builds/${VERSION}/darwin_amd64 builds/${VERSION}/linux_386 builds/${VERSION}/linux_amd64
  mkdir -p builds/${VERSION}/windows_386 builds/${VERSION}/windows_amd64
  GOPATH="${cur}/tmp/go"
  echo "Getting"
  GOPATH="${cur}/tmp/go" go get -d .
  echo "Starting build"

  GOPATH="${cur}/tmp/go" GOOS=darwin GOARCH=386 go build -o builds/${VERSION}/darwin_386/weave

  GOPATH="${cur}/tmp/go" GOARCH=amd64 GOOS=darwin go build -o builds/${VERSION}/darwin_amd64/weave

  GOPATH="${cur}/tmp/go" GOOS=linux GOARCH=amd64 go build -o builds/${VERSION}/linux_amd64/weave

  GOPATH="${cur}/tmp/go" GOOS=linux GOARCH=386 go build -o builds/${VERSION}/linux_386/weave

  GOPATH="${cur}/tmp/go" GOOS=windows GOARCH=amd64 go build -o builds/${VERSION}/windows_amd64/weave

  GOPATH="${cur}/tmp/go" GOOS=windows GOARCH=386 go build -o builds/${VERSION}/windows_386/weave
else
  echo "$VET"
  exit -1
fi

# rewrite the binaries

FILES=builds/${VERSION}/*/weave
for f in $FILES
do
  str="/weave"
  repl=""
  path=${f/$str/$repl}
  tar  -C ${path} -cvzf "${f}.tgz" weave
  rm ${f}
done