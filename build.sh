#!/usr/bin/env bash

set -eou pipefail

version="$(git tag | sort --version-sort -r | head -1)"

last_tag_ref=$(git log --decorate | grep $version | cut -d' ' -f 2 | head -c 7)

last_commit_ref="$(git log -1 --oneline | cut -d' ' -f 1)"

echo $version $last_tag_ref $last_commit_ref

if [ ! $last_tag_ref == $last_commit_ref ]; then
	echo "last commit is not last tag"
	echo "last tag: $version"
	exit 1
fi

xgo \
    --targets="darwin/amd64,linux/amd64" \
    --dest=dist \
    --ldflags "-X main.version=$version" \
    -v -x \
    github.com/endorama/two-factor-authenticator
