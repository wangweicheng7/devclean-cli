#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-}"
if [[ -z "${TAG}" ]]; then
  echo "usage: $0 <tag> (example: v0.2.1)" >&2
  exit 2
fi

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT_DIR}"

if ! command -v gh >/dev/null 2>&1; then
  echo "gh is required (GitHub CLI)" >&2
  exit 1
fi

echo "==> tagging and pushing ${TAG}" >&2
if git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null; then
  echo "tag already exists locally: ${TAG}" >&2
else
  git tag "${TAG}"
fi
git push origin "${TAG}"

echo "==> waiting for Release workflow to finish" >&2
deadline_secs="${DEADLINE_SECS:-900}" # 15 minutes
sleep_secs="${SLEEP_SECS:-10}"
start="$(date +%s)"

while true; do
  # If the GitHub Release already exists with expected artifacts, skip waiting.
  if gh release view "${TAG}" --json assets >/dev/null 2>&1; then
    if gh release view "${TAG}" --json assets --jq '.assets[].name' 2>/dev/null | grep -qx 'checksums.txt'; then
      break
    fi
  fi

  now="$(date +%s)"
  if (( now - start > deadline_secs )); then
    echo "timeout waiting for release workflow for ${TAG}" >&2
    echo "check: gh run list --workflow Release" >&2
    exit 1
  fi

  run_json="$(gh run list --workflow Release -L 20 --json databaseId,headBranch,status,conclusion,createdAt 2>/dev/null || true)"
  if [[ -z "${run_json}" ]]; then
    sleep "${sleep_secs}"
    continue
  fi

  # Find the run where headBranch equals the tag. If JSON parsing fails, return empty.
  run_line="$(printf '%s' "${run_json}" | python3 - "${TAG}" <<'PY'
import json, sys
tag = sys.argv[1]
try:
  data = json.load(sys.stdin)
except Exception:
  sys.exit(0)
for r in data:
  if r.get("headBranch") == tag:
    print(f'{r.get("databaseId")} {r.get("status")} {r.get("conclusion")}')
    break
PY
)"

  if [[ -z "${run_line}" ]]; then
    sleep "${sleep_secs}"
    continue
  fi

  run_id="$(echo "${run_line}" | awk '{print $1}')"
  status="$(echo "${run_line}" | awk '{print $2}')"
  conclusion="$(echo "${run_line}" | awk '{print $3}')"

  if [[ "${status}" != "completed" ]]; then
    sleep "${sleep_secs}"
    continue
  fi
  if [[ "${conclusion}" != "success" ]]; then
    echo "release workflow failed: run_id=${run_id} conclusion=${conclusion}" >&2
    gh run view "${run_id}" --log-failed || true
    exit 1
  fi
  break
done

echo "==> updating formula from release checksums" >&2
make brew-formula-update TAG="${TAG}"

echo "==> publishing formula to tap repo" >&2
make brew-formula-publish

echo "done: ${TAG}" >&2

