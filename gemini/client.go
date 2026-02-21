package gemini

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type Client interface {
	GenerateNoteJSON(ctx context.Context, word string) (string, error)
}

type APIError = genai.APIError

type client struct {
	inner *genai.Client
}

const noteModel = "gemini-2.5-flash"

const notePromptTemplate = `
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

func NewClient(ctx context.Context, apiKey string) (Client, error) {
	inner, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, err
	}

	return &client{inner: inner}, nil
}

func (c *client) GenerateNoteJSON(ctx context.Context, word string) (string, error) {
	result, err := c.inner.Models.GenerateContent(
		ctx,
		noteModel,
		genai.Text(fmt.Sprintf(notePromptTemplate, word)),
		nil,
	)
	if err != nil {
		return "", err
	}

	return result.Text(), nil
}
