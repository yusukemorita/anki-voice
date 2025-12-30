## anki-voice Usage

### prerequisites

1. Run the piper voice generation docker container
  ```console
  docker compose up -d
  ```

2. Have anki (with the ankiconnect addon installed) running

### fill in missing audio in one note 

```sh
go run main.go -note 123456789 # replace 123456789 with the note id
```

### add audio for all notes that match an anki query

```sh
# to only fill in missing audio
go run cmd/voice/main.go -query "tag:audio"
# or, to overwrite all audio
go run cmd/voice/main.go -query "tag:audio" -overwrite

# generate audio for all chatgpt generated tags that don't have generated audio yet
go run cmd/voice/main.go -query "tag:chatgpt-generated -tag:audio-generated"

# generate and overwrite audio for all audio tags, and then remove the tag
go run cmd/voice/main.go -query "tag:audio" -removetag "audio" -overwrite -limit 10
```

## generate-card usage

`generate-card` automatically generates an anki card for a given word, complete with audio.
Prerequisites are the same as `anki-voice`.

### generate a note for a single word

```console
go run cmd/generate-card/main.go benehmen
```

### generate notes for words in german_vocab.txt

```console
go run cmd/generate-card/main.go -limit 10
```

## references

- [piper HTTP API](https://github.com/OHF-Voice/piper1-gpl/blob/main/docs/API_HTTP.md)
- [AnkiConnect](https://github.com/amikey/anki-connect)
- [Gemini AI Studio](https://aistudio.google.com)
