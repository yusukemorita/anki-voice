package main

import (
	"anki-voice/anki"
	"anki-voice/ankiconnect"
	"anki-voice/gemini"
	"anki-voice/noteaudio"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// key: field in German, value: audio field
var audioFields = map[string]string{
	"base_d": "base_a",
	"s1":     "s1a",
	"s2":     "s2a",
	"s3":     "s3a",
	"s4":     "s4a",
}

type GeminiResponse struct {
	FullDeutsch    string `json:"full_d"`
	BaseDeutsch    string `json:"base_d"`
	BaseEnglish    string `json:"base_e"`
	ArticleDeutsch string `json:"artikel_d"`
	PluralDeutsch  string `json:"plural_d"`
	S1             string
	S1e            string
	S2             string
	S2e            string
	S3             string
	S3e            string
	S4             string
	S4e            string
}

func (response GeminiResponse) toMap() map[string]string {
	return map[string]string{
		"full_d":    response.FullDeutsch,
		"base_d":    response.BaseDeutsch,
		"base_e":    response.BaseEnglish,
		"artikel_d": response.ArticleDeutsch,
		"plural_d":  response.PluralDeutsch,
		"s1":        response.S1,
		"s1e":       response.S1e,
		"s2":        response.S2,
		"s2e":       response.S2e,
		"s3":        response.S3,
		"s3e":       response.S3e,
		"s4":        response.S4,
		"s4e":       response.S4e,
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("no .env file found; using existing environment")
	}

	GEMINI_API_KEY := os.Getenv("GEMINI_API_KEY")
	if GEMINI_API_KEY == "" {
		log.Fatal("GEMINI_API_KEY is not set")
	}

	VOCAB_DIR := os.Getenv("VOCAB_DIR")
	if VOCAB_DIR == "" {
		log.Fatal("VOCAB_DIR is not set")
	}

	// general setup
	ctx := context.Background()
	ankiMediaDir, err := anki.MediaDir()
	if err != nil {
		log.Fatal(err)
	}

	// Gemini setup
	geminiClient, err := gemini.NewClient(ctx, GEMINI_API_KEY)
	if err != nil {
		log.Fatal(err)
	}

	// check that anki is running
	_, err = ankiconnect.QueryNotes("test")
	if err != nil {
		log.Fatalf("error response from anki, is anki running?\n%s", err)
	}

	wordFlag := flag.String("word", "", "word to generate a note for")
	limitFlag := flag.Int("limit", 50, "maximum number of notes to generate")
	flag.Parse()

	word := *wordFlag
	limit := *limitFlag

	if word == "" {
		generateNoteForWordsInVocabDir(geminiClient, ankiMediaDir, VOCAB_DIR, limit)
	} else {
		generateNote(word, geminiClient, ankiMediaDir)
	}
}

func generateNoteForWordsInVocabDir(geminiClient gemini.Client, ankiMediaDir string, vocabDir string, limit int) {
	entries, err := vocabEntriesFromDir(vocabDir)
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	for _, entry := range entries {
		generateErr := generateNote(entry.word, geminiClient, ankiMediaDir)

		var apiErr *gemini.APIError
		if errors.As(generateErr, &apiErr) {
			detailsStr := fmt.Sprintf("%v", apiErr.Details)

			delay, err := extractRetryDelay(detailsStr)
			if err != nil {
				log.Fatalf("failed to extract retry delay. original gemini error:\n%s\n\nextract error:\n%s\n", generateErr, err)
			}

			log.Printf("retry delay: %v", delay)
			time.Sleep(delay)

			// retry after delay, this time fail if error is returned
			generateErr = generateNote(entry.word, geminiClient, ankiMediaDir)
			if generateErr != nil {
				log.Fatal(generateErr)
			}
		}

		if err := os.Remove(entry.path); err != nil {
			log.Fatalf("failed to delete vocab file %s: %v", entry.path, err)
		}
		count++
		if count >= limit {
			log.Printf("reached limit %d\n", limit)
			break
		}
		time.Sleep(time.Second * 2)
	}
}

func generateNote(word string, geminiClient gemini.Client, ankiMediaDir string) error {
	// retrieve result from Gemini
	result, err := geminiClient.GenerateNoteJSON(context.Background(), word)
	if err != nil {
		return err
	}
	log.Printf("Gemini response: \n%s\n", result)

	// remove the code block that Gemini prefers to add to the response
	jsonText := result
	jsonText = strings.TrimPrefix(jsonText, "```json")
	jsonText = strings.TrimSuffix(jsonText, "```")

	// unmarshal the JSON that was in the code block
	var response GeminiResponse
	err = json.Unmarshal([]byte(jsonText), &response)
	if err != nil {
		return err
	}

	// add the note
	log.Println("Adding note...")
	noteID, err := ankiconnect.AddNote(response.toMap())
	if err != nil {
		if strings.Contains(err.Error(), "cannot create note because it is a duplicate") {
			log.Println("skipping duplicate note")
			return nil
		} else {
			return err
		}
	}
	log.Printf("Added note: %d", noteID)

	// add audio to the note
	addAudioToNote(noteID, ankiMediaDir)
	log.Printf("Added audio to note: %d", noteID)

	return nil
}

func addAudioToNote(noteID int, ankiMediaDir string) error {
	log.Printf("adding audio tag to note: %d", noteID)
	err := ankiconnect.AddNoteTag(noteID, anki.AudioTag)
	if err != nil {
		return err
	}

	err = noteaudio.AddAudioToNote(noteID, ankiMediaDir, audioFields, noteaudio.Options{
		Overwrite: true,
	})
	if err != nil {
		return err
	}

	err = ankiconnect.AddNoteTag(noteID, anki.AudioGeneratedTag)
	if err != nil {
		return err
	}

	log.Printf("removing audio tag from note: %d", noteID)
	err = ankiconnect.RemoveNoteTag(noteID, anki.AudioTag)
	if err != nil {
		return err
	}

	return nil
}

type vocabEntry struct {
	word      string
	path      string
	createdAt time.Time
}

func vocabEntriesFromDir(vocabDir string) ([]vocabEntry, error) {
	entries, err := os.ReadDir(vocabDir)
	if err != nil {
		return nil, err
	}

	words := make([]vocabEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			// ignore hidden files
			continue
		}

		base := strings.TrimSuffix(name, filepath.Ext(name))
		base = strings.TrimSpace(base)
		if base == "" {
			// ignore blank file names
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		words = append(words, vocabEntry{
			word:      base,
			path:      filepath.Join(vocabDir, name),
			createdAt: fileCreateTime(info),
		})
	}

	// sort so that files created earlier come first
	sort.SliceStable(words, func(i, j int) bool {
		if words[i].createdAt.Equal(words[j].createdAt) {
			return words[i].word < words[j].word
		}
		return words[i].createdAt.Before(words[j].createdAt)
	})

	return words, nil
}

func extractRetryDelay(errorMessage string) (time.Duration, error) {
	var retryDelayRegex = regexp.MustCompile(`"retryDelay"\s*:\s*"(\d+)s"`)
	m := retryDelayRegex.FindStringSubmatch(errorMessage)
	if m == nil {
		return time.Second * 0, fmt.Errorf("no regex match")
	}

	seconds, err := strconv.Atoi(m[1]) // m[1] is the captured digits
	if err != nil {
		return time.Second * 0, fmt.Errorf("bad number. match: %s, err: %s", m[1], err)
	}

	return time.Duration(seconds) * time.Second, nil
}
