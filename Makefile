.PHONY: generate-card voice

gen:
	go run ./cmd/generate-card

voice:
	go run ./cmd/voice -query "tag:audio" -removetag "audio" -overwrite
