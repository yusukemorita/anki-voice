package main

import (
	"anki-voice/anki"
	"anki-voice/ankiconnect"
	"anki-voice/noteaudio"
	"flag"
	"log"
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
	limitFlag := flag.Int("limit", 100, "limit the number of cards to update")
	dryRunFlag := flag.Bool("dryrun", false, "set to true to skip update of the note in anki")
	queryFlag := flag.String("query", "", "use an anki query to filter which cards to update")
	overwriteFlag := flag.Bool("overwrite", false, "set to true to overwrite existing audio")
	removeTagFlag := flag.String("removetag", "", "remove the specified tag when update of a note succeeds")
	flag.Parse()

	noteID := *noteIDFlag
	dryRun := *dryRunFlag
	overwrite := *overwriteFlag
	query := *queryFlag
	tagToRemove := *removeTagFlag
	limit := *limitFlag

	ankiMediaDir, err := anki.MediaDir()
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case noteID != 0:
		log.Println("Update one note")
		err = updateOneNote(noteID, ankiMediaDir, tagToRemove, dryRun, overwrite)
		if err != nil {
			log.Fatal(err)
		}
	case query != "":
		log.Println("Update notes that match query")
		ids, err := ankiconnect.QueryNotes(query)
		if err != nil {
			log.Fatal(err)
		}

		for index, id := range ids {
			err = updateOneNote(id, ankiMediaDir, tagToRemove, dryRun, overwrite)
			if err != nil {
				log.Fatal(err)
			}

			if limit != 0 && index+1 >= limit {
				break
			}
		}
	}
}

func updateOneNote(noteID int, ankiMediaDir string, tagToRemove string, dryRun, overwrite bool) error {
	err := noteaudio.AddAudioToNote(noteID, ankiMediaDir, fields, noteaudio.Options{
		DryRun:         dryRun,
		Overwrite:      overwrite,
		RemoveOldAudio: true,
	})
	if err != nil {
		return err
	}

	if !dryRun {
		err = ankiconnect.AddNoteTag(noteID, anki.AudioGeneratedTag)
		if err != nil {
			return err
		}

		if tagToRemove != "" {
			log.Printf("removing tag in anki: %s\n", tagToRemove)
			err = ankiconnect.RemoveNoteTag(noteID, tagToRemove)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
