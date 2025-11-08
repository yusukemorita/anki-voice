#!/bin/bash

# Usage:
# bash generate.sh "hello hi"

set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Error: missing number of cards to update."
  exit 1
fi
COUNT=$1

ANKI_URL="localhost:8765"

IDS=$(curl -s $ANKI_URL --json @- <<'JSON'
{
  "action": "findNotes",
  "version": 5,
  "params": {"query": "tag:chatgpt-generated -tag:audio-generated"}
}
JSON
)

while IFS= read -r s; do
  bash anki-one-note.sh $s
done < <(echo $IDS | jq ".result[:$COUNT][]")
