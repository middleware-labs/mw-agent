#!/usr/bin/env bash
# Merges new "HASH  filename" lines into a release's checksums.txt. Several
# workflows publish to one release and the OpAMP server reads this file, so a
# job must keep the lines of the others. A missing file is normal for the first
# job; any other failure stops here, before an incomplete file can be uploaded.
#
# Usage: merge-release-checksums.sh <version> <owner/repo> <file with new lines>
set -euo pipefail

version=$1
repo=$2
new_entries=$3

work_dir=$(mktemp -d)
trap 'rm -rf "$work_dir"' EXIT
current="$work_dir/current.txt"
merged="$work_dir/checksums.txt"
touch "$current"

assets=$(gh release view "$version" --repo "$repo" --json assets --jq '.assets[].name')
if grep -qx 'checksums.txt' <<<"$assets"; then
  gh release download "$version" --repo "$repo" --pattern checksums.txt --output "$current" --clobber
fi

# Keep every current line whose file isn't in the new entries, then add those.
awk 'NR == FNR { replaced[$2] = 1; next } !($2 in replaced)' "$new_entries" "$current" > "$merged"
cat "$new_entries" >> "$merged"

gh release upload "$version" "$merged" --clobber --repo "$repo"
