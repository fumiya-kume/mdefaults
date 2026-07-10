#!/bin/sh

set -eu

profile=${1:-coverage.out}
minimum=${2:-95.0}

if [ ! -f "$profile" ]; then
  echo "coverage profile not found: $profile" >&2
  exit 1
fi

total=$(go tool cover -func="$profile" | awk '/^total:/ {gsub(/%/, "", $3); print $3}')
if [ -z "$total" ]; then
  echo "could not read total coverage from $profile" >&2
  exit 1
fi

awk -v total="$total" -v minimum="$minimum" 'BEGIN {
  printf "unit coverage: %.1f%% (required: %.1f%%)\n", total, minimum
  if (total + 0 < minimum + 0) {
    exit 1
  }
}'
