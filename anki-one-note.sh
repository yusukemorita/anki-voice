#!/bin/bash

# Usage when the note id is 123456789:
# bash anki-one-note.sh 123456789

set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Error: missing text to synthesize."
  exit 1
fi
NOTE_ID=$1

ANKI_URL="localhost:8765"

NOTE=$(curl --json @- "$ANKI_URL" <<JSON
{
  "action": "notesInfo",
  "version": 5,
  "params": {
    "notes": [${NOTE_ID}]
  }
}
JSON
)

# echo $NOTE | jq '.'

WORD=$(echo $NOTE | jq -c '.result[0].fields.full_d.value')

echo "Updating note: $WORD"

## declare an array variable
declare -a arr=("s1" "s2" "s3" "s4" "s5" "s6" "s7" "s8" "s9")

## loop through each sentence
for sn in "${arr[@]}"; do
  SENTENCE=$(echo $NOTE | jq -r ".result[0].fields.$sn.value")
  [[ -z "$SENTENCE" ]] && continue  

  echo ""
  echo "--- $sn: $SENTENCE START ---"

  # skip if sentence is blank
  # [[ -z "${SENTENCE//[[:space:]]/}" ]] && continue


  # TODO: add output argument to generate.sh
  bash ./generate.sh "$SENTENCE" # outputs to output.mp3

  MP3_NAME="$NOTE_ID-$sn.mp3"

  # TODO: 
  mv ./output.mp3 ~/Library/Application\ Support/Anki2/User\ 1/collection.media/$MP3_NAME

  echo "Updating audio field..."

  FIELD_NAME="${sn}a"

  curl --json @- "$ANKI_URL" <<JSON
{
  "action": "updateNoteFields",
  "version": 5,
  "params": {
    "note": {
      "id": ${NOTE_ID},
      "fields": {
        "$FIELD_NAME": "[sound:${MP3_NAME}]"
      }
    }
  }
}
JSON

  echo "Adding tag..."

  curl --json @- "$ANKI_URL" <<JSON
{
  "action": "addTags",
  "version": 5,
  "params": {
    "notes": [${NOTE_ID}],
    "tags": "audio-generated"
  }
}
JSON


  echo ""
  echo "--- $sn DONE ---"
  echo ""
done

echo "Done!"
