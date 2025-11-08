# anki-voice

## Usage

### generate audio from text

```console
docker compose up -d
bash generate.sh "Der Arzt verschrieb dem Patienten ein neues Arzneimittel."
```

### add audio to existing cards

1. Open the Anki desktop app (with [AnkiConnect](https://github.com/amikey/anki-connect) installed)
2. Run the following

```sh
bash anki-multiple-notes.sh 10 # number of cards to update
```

## references

- piper HTTP API: https://github.com/OHF-Voice/piper1-gpl/blob/main/docs/API_HTTP.md
