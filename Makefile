.PHONY: generate-card voice

GENERATE_LIMIT ?= 10
VOICE_LIMIT ?= 10

# Allow: make gen -- 10  (or make voice -- 10)
# Use the first extra goal after the target as the limit.
LIMIT_ARG := $(word 2,$(MAKECMDGOALS))
ifeq ($(LIMIT_ARG),)
LIMIT_ARG := $(GENERATE_LIMIT)
endif

gen:
	go run ./cmd/generate-card -limit $(LIMIT_ARG)

voice:
	go run ./cmd/voice -query "tag:audio" -removetag "audio" -overwrite -limit $(LIMIT_ARG)

# Consume extra goals (e.g. the "10" in "make gen -- 10")
%:
	@:
