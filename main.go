package main

import (
	"anki-voice/ankiconnect"
	"anki-voice/audio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	outputPathMp3 = "./output.mp3"
)

var (
	// key: field in German, value: audio field
	fields = map[string]string{
		"base_d": "base_a",
		"s1":     "s1a",
		"s2":     "s2a",
		"s3":     "s3a",
		"s4":     "s4a",
		"s5":     "s5a",
		"s6":     "s6a",
		"s7":     "s7a",
		"s8":     "s8a",
		"s9":     "s9a",
	}
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("noteID argument required")
	}
	noteID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("invalid noteID %q: %v", os.Args[1], err)
	}

	note, err := ankiconnect.GetNote(noteID, fields)
	if err != nil {
		log.Fatal(err)
	}

	for field, phrase := range note.Phrases {
		if phrase.Value == "" {
			continue
		}

		if strings.Contains(phrase.Audio, fmt.Sprintf("-%s.mp3", field)) {
			// audio has already been generated
			continue
		}

		log.Printf("generating audio for: %s...\n", phrase.Value)

		outputPath := fmt.Sprintf("./output/%s.mp3", phrase.Value)

		err = audio.GenerateMP3(phrase.Value, outputPath)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func keys(m map[string]string) []string {
	allKeys := make([]string, 0, len(m))
	for k := range m {
		allKeys = append(allKeys, k)
	}

	return allKeys
}
