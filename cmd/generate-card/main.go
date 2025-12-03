package main

import (
	"anki-voice/ankiconnect"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"

	"google.golang.org/genai"
)

//go:embed GEMINI_API_KEY
var GEMINI_API_KEY string

const PROMPT = `
Return the following fields in a JSON structure for the word: %s
The values will be used for creating Anki cards to learn German vocabulary.

* full_d: German word. 
  * When a verb, should be a comma separated list of infinitive, present, simple past, and present perfect. e.g. "analysieren, analysiert, analysierte, hat analysiert"
  * When a noun, should include the article, and the ending in plural. e.g. "das Abgas, -e", "das Alter, -". This is just a combination of the fields artikel_d, base_d, and plural_d.
* base_e: the English translation. e.g. "to analyze"
  * If an English translation is provided in the prompt, make sure base_e covers what is provided
* base_d: the base form the German word. 
  * When a noun, omit the article. e.g. "Abgas".
  * When plural, convert to singular.
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
* Return ONLY the JSON object, do not include any other content or text. This means that EVEN BACKSLASHES TO INDICATE CODEBLOCKS SHOULD BE OMITTED!!!
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
	ctx := context.Background()

	fmt.Printf("API KEY: %s\n", GEMINI_API_KEY)

	config := &genai.ClientConfig{APIKey: GEMINI_API_KEY}
	client, err := genai.NewClient(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(fmt.Sprintf(PROMPT, "Sackerl")),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Text())

	var response GeminiResponse
	err = json.Unmarshal([]byte(result.Text()), &response)
	if err != nil {
		log.Fatalf("failed to unmarshal response:\n%s\n", result.Text())
	}

	fmt.Printf("Creating note from gemini response: \n%+v\n", response)

	noteID, err := ankiconnect.AddNote(response.toMap())
	if err != nil {
		log.Fatalf("failed to create note: %s", err)
	}

	fmt.Printf("success! created note: %d", noteID)


}
