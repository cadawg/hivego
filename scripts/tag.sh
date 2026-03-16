#!/usr/bin/env bash
set -e

usage() {
    echo "Usage: $0 [-minor|-patch]"
    echo "  -minor  increment the minor version (e.g. v0.1.0 → v0.2.0)"
    echo "  -patch  increment the patch version (e.g. v0.1.0 → v0.1.1)"
    exit 1
}

[ $# -eq 1 ] || usage

BUMP=$1
case "$BUMP" in
    -minor|-patch) ;;
    *) usage ;;
esac

# Get the latest semver tag; default to v0.0.0 if none exists
LAST=$(git tag --sort=-version:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -1)
LAST=${LAST:-v0.0.0}

# Strip leading 'v' and split into parts
VERSION=${LAST#v}
MAJOR=$(echo "$VERSION" | cut -d. -f1)
MINOR=$(echo "$VERSION" | cut -d. -f2)
PATCH=$(echo "$VERSION" | cut -d. -f3)

case "$BUMP" in
    -minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    -patch)
        PATCH=$((PATCH + 1))
        ;;
esac

NEXT="v${MAJOR}.${MINOR}.${PATCH}"

echo "Last tag: $LAST"
echo "New tag:  $NEXT"

git tag -a "$NEXT" -m "$NEXT"
echo "Tagged $NEXT (not pushed — run: git push origin $NEXT)"