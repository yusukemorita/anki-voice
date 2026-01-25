package main

import (
	"anki-voice/anki"
	"anki-voice/ankiconnect"
	"anki-voice/noteaudio"
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

// key: field in German, value: audio field
var audioFields = map[string]string{
	"base_d": "base_a",
	"s1":     "s1a",
	"s2":     "s2a",
	"s3":     "s3a",
	"s4":     "s4a",
}

const PROMPT = `
Return the following fields in a JSON structure for the word: %s
The values will be used for creating Anki cards to learn German vocabulary.

* base_d: the base form the German word.
  * When a noun, omit the article. e.g. "Abgas".
	* When a reflexive verb, should start with "sich".
* full_d: German word. 
  * When a verb, should be a comma separated list of infinitive, present, simple past, and present perfect. e.g. "analysieren, analysiert, analysierte, hat analysiert"
	* When a reflexive verb, should start with "sich". e.g. "sich amüsieren, amüsiert sich, amüsierte sich, hat sich amüsiert"
  * When a noun, should include the article, and the ending in plural. e.g. "das Abgas, -e", "das Alter, -". This is just a combination of the fields artikel_d, base_d, and plural_d.
* base_e: the English translation. e.g. "to analyze"
  * If an English translation is provided in the prompt, make sure base_e covers what is provided
* artikel_d:
  * When a noun, the article. 
  * When not a noun, blank string
* plural_d: 
  * When a noun, the plural ending. "-" if the ending does not change, and e.g. "-e" if an "e" is added.
  * When not a noun, blank string
* s1: The first example sentence in German. Create a typical sentence that the word would be used in.
* s1e: The English translation of s1.
* s2: The second example sentence in German. If there is more than one meaning of the word, then create a sentence that demonstrates a use of the second meaning.
* s2e: The English translation of s2.
* s3: The third example sentence in German. Only include If there are more than two commonly used meanings of the word. Otherwise, leave blank.
* s3e: The English translation of s3.
* s4: The fourth example sentence in German. Only include If there are more than three commonly used meanings of the word. Otherwise, leave blank.
* s4e: The English translation of s4.

Other things to note: 
* If the word is in plural, convert it to singular
* Return ONLY the JSON object wrapped in a json code block, and do not include any other content or text.
`

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
	geminiClient, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: GEMINI_API_KEY})
	if err != nil {
		log.Fatal(err)
	}

	wordFlag := flag.String("word", "", "word to generate a note for")
	limitFlag := flag.Int("limit", 5, "maximum number of notes to generate")
	flag.Parse()

	word := *wordFlag
	limit := *limitFlag

	if word == "" {
		generateNoteForWordsInVocabFile(geminiClient, ankiMediaDir, VOCAB_DIR, limit)
	} else {
		generateNote(word, geminiClient, ankiMediaDir)
	}
}

func generateNoteForWordsInVocabFile(geminiClient *genai.Client, ankiMediaDir string, vocabDir string, limit int) {
	entries, err := vocabWordsFromFiles(vocabDir)
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	for _, entry := range entries {
		generateNote(entry.word, geminiClient, ankiMediaDir)
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

func generateNote(word string, geminiClient *genai.Client, ankiMediaDir string) {
	// retrieve result from Gemini
	result, err := geminiClient.Models.GenerateContent(
		context.Background(),
		"gemini-2.5-flash",
		genai.Text(fmt.Sprintf(PROMPT, word)),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Gemini response: \n%s\n", result.Text())

	// remove the code block that Gemini prefers to add to the response
	jsonText := result.Text()
	jsonText = strings.TrimPrefix(jsonText, "```json")
	jsonText = strings.TrimSuffix(jsonText, "```")

	// unmarshal the JSON that was in the code block
	var response GeminiResponse
	err = json.Unmarshal([]byte(jsonText), &response)
	if err != nil {
		log.Fatalf("failed to unmarshal response:\n%s\n", result.Text())
	}

	// add the note
	log.Println("Adding note...")
	noteID, err := ankiconnect.AddNote(response.toMap())
	if err != nil {
		if strings.Contains(err.Error(), "cannot create note because it is a duplicate") {
			log.Println("skipping duplicate note")
			return
		} else {
			log.Fatalf("failed to add note: %s", err)
		}
	}
	log.Printf("Added note: %d", noteID)

	// add audio to the note
	addAudioToNote(noteID, ankiMediaDir)
	log.Printf("Added audio to note: %d", noteID)
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
	word string
	path string
}

func vocabWordsFromFiles(vocabDir string) ([]vocabEntry, error) {
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
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		base = strings.TrimSpace(base)
		if base == "" {
			continue
		}
		words = append(words, vocabEntry{
			word: base,
			path: filepath.Join(vocabDir, name),
		})
	}

	return words, nil
}
