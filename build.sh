#!/usr/bin/env bash

set -eou pipefail

version="$(git tag | sort --version-sort -r | head -1)"

last_tag_ref=$(git log --decorate | grep $version | cut -d' ' -f 2 | head -c 7)

last_commit_ref="$(git log -1 --oneline | cut -d' ' -f 1)"

echo $version $last_tag_ref $last_commit_ref

if [ ! $last_tag_ref == $last_commit_ref ]; then
	echo "last commit is not last tag"
	echo "last tag: $version"
fi

[ -d dist/ ] && mv dist/ dist.prev/

# force using go modules when inside GOPATH
export GO111MODULE=on

xgo \
    --targets="darwin/amd64,linux/amd64" \
    --dest=dist \
    --ldflags "-X main.version=$version" \
    -v -x \
    github.com/endorama/2ami

sudo chown $USER: -R dist

gpg_sign_key="edoardo.tenani@protonmail.com"
checksum_file="dist/2ami-${version}_checksums.txt"

orig_darwin_file="dist/2ami-darwin-10.6-amd64"
orig_linux_file="dist/2ami-linux-amd64"
darwin_file="dist/2ami-${version}-darwin-10.6-amd64"
linux_file="dist/2ami-${version}-linux-amd64"

mv -f "$orig_darwin_file" "$darwin_file"
mv -f "$orig_linux_file" "$linux_file"

gpg2 \
	-u ${gpg_sign_key} \
	--output ${darwin_file}.sig \
	--detach-sign ${darwin_file}

gpg2 \
	-u ${gpg_sign_key} \
	--output ${linux_file}.sig \
	--detach-sign ${linux_file}

sha1sum \
	"$darwin_file" \
	"$linux_file" \
	${checksum_file} \
	> $checksum_file

gpg2 \
	-u ${gpg_sign_key} \
	--output ${checksum_file}.sig \
	--detach-sign ${checksum_file}

rm -r dist.prev/
