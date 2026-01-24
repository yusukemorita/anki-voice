package noteaudio

import (
	"anki-voice/ankiconnect"
	"anki-voice/audio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Options struct {
	DryRun         bool
	Overwrite      bool
	RemoveOldAudio bool
}

var soundRegex = regexp.MustCompile(`^\[sound:([^\]]+)\]$`)

func AddAudioToNote(noteID int, ankiMediaDir string, fieldMap map[string]string, options Options) error {
	note, err := ankiconnect.GetNote(noteID, fieldMap)
	if err != nil {
		return err
	}

	log.Printf("--- note: %s ---", note.Phrases["base_d"].Value)

	for field, phrase := range note.Phrases {
		text := sanitizePhraseText(phrase.Value)
		if text == "" {
			continue
		}

		if phrase.Audio != "" && !options.Overwrite {
			continue
		}

		log.Printf("generating audio for: '%s'\n", text)
		outputPath := fmt.Sprintf("./output/%s.mp3", text)
		if err := audio.GenerateMP3(text, outputPath); err != nil {
			return err
		}

		filename := fmt.Sprintf("%d-%s.mp3", note.NoteID, field)
		if err := os.Rename(outputPath, filepath.Join(ankiMediaDir, filename)); err != nil {
			return err
		}

		newAudioFieldValue := fmt.Sprintf("[sound:%s]", filename)
		if options.DryRun {
			log.Printf("skipping note update. audio: %s", newAudioFieldValue)
			continue
		}

		log.Printf("updating field in anki: %s\n", text)
		if err := ankiconnect.UpdateNoteField(note.NoteID, fieldMap[field], newAudioFieldValue); err != nil {
			return err
		}

		if options.RemoveOldAudio {
			if err := removeOldAudioFile(phrase.Audio, newAudioFieldValue, ankiMediaDir); err != nil {
				return err
			}
		}
	}

	return nil
}

func sanitizePhraseText(text string) string {
	trimmed := strings.ReplaceAll(text, "&nbsp;", "")
	return strings.TrimSpace(trimmed)
}

func removeOldAudioFile(oldAudioValue, newAudioValue, ankiMediaDir string) error {
	if oldAudioValue == "" || oldAudioValue == newAudioValue {
		return nil
	}

	trimmed := strings.TrimSpace(oldAudioValue)
	matches := soundRegex.FindStringSubmatch(trimmed)
	if len(matches) != 2 {
		log.Printf("unexpected audio format: %s", oldAudioValue)
		return nil
	}

	oldFilename := matches[1]
	oldFilePath := filepath.Join(ankiMediaDir, oldFilename)
	if err := os.Remove(oldFilePath); err != nil {
		log.Printf("failed to remove old audio %s: %v", oldFilePath, err)
	}

	return nil
}
