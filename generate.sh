#!/bin/bash

# Usage:
# bash generate.sh "hello hi"

set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Error: missing text to synthesize."
  exit 1
fi
TEXT=$1

echo "generating audio...:  $TEXT"

curl -s -X POST "http://localhost:9999" \
                  -H "Content-Type: application/json" \
                  -d "$TEXT" \
                  --output ./output.wav

echo "converting to mp3..."

ffmpeg \
  -hide_banner -loglevel error -nostdin -y \
  -i ./output.wav -c:a libmp3lame -b:a 192k ./output.mp3

rm ./output.wav

echo "mp3 generated!"

# afplay "./output.mp3"
