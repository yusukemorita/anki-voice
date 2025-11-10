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
go run main.go -query "tag:audio"
# or, to overwrite all audio
go run main.go -query "tag:audio" -overwrite
```

## references

- [piper HTTP API](https://github.com/OHF-Voice/piper1-gpl/blob/main/docs/API_HTTP.md)
- [AnkiConnect](https://github.com/amikey/anki-connect)
