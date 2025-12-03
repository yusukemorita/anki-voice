# anki-voice

## Usage

### prerequisites

1. Run the piper voice generation docker container
  ```console
  docker compose up -d
  ```

2. Have anki (with the ankiconnect addon installed) running

### fill in missing audio in one note 

```sh
docker compose up -d
go run main.go -note 123456789 # replace 123456789 with the note id
```

### add audio for all notes that match an anki query

```sh
docker compose up -d
# to only fill in missing audio
go run cmd/voice/main.go -query "tag:audio"
# or, to overwrite all audio
go run cmd/voice/main.go -query "tag:audio" -overwrite

# generate audio for all chatgpt generated tags that don't have generated audio yet
go run cmd/voice/main.go -query "tag:chatgpt-generated -tag:audio-generated"

# generate and overwrite audio for all audio tags, and then remove the tag
go run cmd/voice/main.go -query "tag:audio" -removetag "audio" -overwrite limit 10
```

## references

- [piper HTTP API](https://github.com/OHF-Voice/piper1-gpl/blob/main/docs/API_HTTP.md)
- [AnkiConnect](https://github.com/amikey/anki-connect)
