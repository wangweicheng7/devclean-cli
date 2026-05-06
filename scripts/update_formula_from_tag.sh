#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-}"
if [[ -z "${TAG}" ]]; then
  echo "usage: $0 <tag> (example: v0.1.0)" >&2
  exit 2
fi

REPO="wangweicheng7/cleandev-cli"
FORMULA_FILE="homebrew-tap/Formula/cleandev-cli.rb"

URL="https://github.com/${REPO}/archive/refs/tags/${TAG}.tar.gz"
TMP_TAR="$(mktemp -t cleandev-cli.XXXXXX.tar.gz)"

cleanup() {
  rm -f "${TMP_TAR}"
}
trap cleanup EXIT

echo "downloading: ${URL}" >&2
curl -L -o "${TMP_TAR}" "${URL}"

SHA="$(shasum -a 256 "${TMP_TAR}" | awk '{print $1}')"

# Update formula fields:
# - version "..."
# - url "..."
# - sha256 "..."
#
# We use slurp mode to safely replace the url+sha256 pair with an explicit newline.
perl -0777 -pi -e "s/^\\s*version\\s+\"[^\"]*\"\\s*$/  version \"${TAG}\"/m; s/^\\s*url\\s+\"[^\"]*\"\\s*\\n\\s*sha256\\s+\"[^\"]*\"\\s*$/  url \"${URL}\"\\n  sha256 \"${SHA}\"/m" "${FORMULA_FILE}"

echo "updated ${FORMULA_FILE}" >&2
echo "version: ${TAG}" >&2
echo "sha256: ${SHA}" >&2

