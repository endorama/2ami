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

export GITHUB_USER=endorama
export GITHUB_REPO=two-factor-authenticator

echo "Verifying release"
if gothub info | grep "name: '$version'" >/dev/null 2>&1; then
	echo "    This version already exists."
	echo "    Please remove it with 'gothub delete -u $GITHUB_USER -r $GITHUB_REPO -t $version'"
	exit 0
fi


echo "Creating release $version"
if grep "-" "$version" >/dev/null 2>&1; then
	gothub release \
		   --tag "$version" \
		   --name "$version" \
		   --pre-release
else
	gothub release \
		   --tag "$version" \
		   --name "$version"
		   # --description \
fi

for file in dist/*; do
	echo "Adding $file"
	gothub upload \
		   --tag "$version" \
		   --name "$(basename "$file")" \
		   --file "$file";
done
