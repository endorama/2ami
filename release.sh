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

darwin_file=dist/two-factor-authenticator-darwin-10.6-amd64
linux_file=dist/two-factor-authenticator-linux-amd64

sha1sum \
	"$darwin_file" \
	"$linux_file" \
	> dist/sha1-checksums

github_user=endorama
github_repo=two-factor-authenticator

gothub release --user "$github_user" --repo "$github_repo" \
			   --tag "$version" \
			   --name "$version" \
			   --pre-release

gothub upload --user "$github_user" --repo "$github_repo" \
			  --tag "$version" \
			  --name "two-factor-authenticator-$version-darwin-10.6-amd64" \
			  --file "$darwin_file"

gothub upload --user "$github_user" --repo "$github_repo" \
			  --tag "$version" \
			  --name "two-factor-authenticator-$version-linux-amd64" \
			  --file "$linux_file"

gothub upload --user "$github_user" --repo "$github_repo" \
			  --tag "$version" \
			  --name "two-factor-authenticator-$version-sha1-checksums" \
			  --file dist/sha1-checksums
