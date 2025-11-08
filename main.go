package main

import (
	"anki-voice/ankiconnect"
	"anki-voice/audio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	// flags setup
	noteIDFlag := flag.Int("note", 0, "noteID to update audio of")
	dryRunFlag := flag.Bool("dryrun", false, "set to true to skip update of the note in anki")
	overwriteFlag := flag.Bool("overwrite", false, "set to true to overwrite existing audio")
	flag.Parse()

	noteID := *noteIDFlag
	if noteID == 0 {
		log.Fatal("-note argument is required")
	}

	dryRun := *dryRunFlag
	overwrite := *overwriteFlag

	// anki media directory setup
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	ankiMediaDir := filepath.Join(homeDir, "Library", "Application Support", "Anki2", "User 1", "collection.media")
	info, err := os.Stat(ankiMediaDir)
	if err != nil {
		log.Fatalf("anki media directory missing: %v", err)
	}
	if !info.IsDir() {
		log.Fatalf("anki media path is not a directory: %s", ankiMediaDir)
	}

	note, err := ankiconnect.GetNote(noteID, fields)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("--- note: %s ---", note.Phrases["base_d"].Value)

	for field, phrase := range note.Phrases {
		if phrase.Value == "" {
			continue
		}

		if phrase.Audio != "" && !overwrite {
			// audio has already been generated
			continue
		}

		// if strings.Contains(phrase.Audio, fmt.Sprintf("-%s.mp3", field)) {
		// 	// audio has already been generated
		// 	continue
		// }

		log.Printf("generating audio for: %s\n", phrase.Value)
		outputPath := fmt.Sprintf("./output/%s.mp3", phrase.Value)
		err = audio.GenerateMP3(phrase.Value, outputPath)
		if err != nil {
			log.Fatal(err)
		}

		filename := fmt.Sprintf("%d-%s.mp3", noteID, field)
		err = os.Rename(outputPath, filepath.Join(ankiMediaDir, filename))
		if err != nil {
			log.Fatal(err)
		}

		audioFieldValue := fmt.Sprintf("[sound:%s]", filename)
		tag := "audio-generated"
		if dryRun {
			log.Printf("skipping note update. audio: %s, tag: %s", audioFieldValue, tag)
		} else {
			log.Printf("updating field in anki: %s\n", phrase.Value)
			err = ankiconnect.UpdateNoteField(noteID, fields[field], audioFieldValue)
			if err != nil {
				log.Fatal(err)
			}

			err = ankiconnect.AddNoteTag(noteID, tag)
			if err != nil {
				log.Fatal(err)
			}
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
