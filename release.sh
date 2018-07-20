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

github_user=endorama
github_repo=two-factor-authenticator

echo "Creating release $version"
gothub release --user "$github_user" --repo "$github_repo" \
			   --tag "$version" \
			   --name "$version" \
			   --pre-release

for file in dist/*; do
	echo "Adding $file"
	gothub upload --user "$github_user" --repo "$github_repo" \
				  --tag "$version" \
				  --file "$file"
done
