package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type APIError = genai.APIError

type GeminiClient struct {
	inner *genai.Client
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

func (response GeminiResponse) ToMap() map[string]string {
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

type Category string

const (
	CategoryNoun  Category = "noun"
	CategoryVerb  Category = "verb"
	CategoryOther Category = "other"

	noteModel = "gemini-2.5-flash"
)

const notePromptTemplate = `
Return the following fields in a JSON structure for the word: %s
The values will be used for creating Anki cards to learn German vocabulary.

* base_d: the base form the German word.
  * When a noun, omit the article. e.g. "Abgas".
	* When a reflexive verb, start with "sich".
* full_d: German word. 
  * When a verb, should be a comma separated list of infinitive, present, simple past, and present perfect. e.g. "analysieren, analysiert, analysierte, hat analysiert"
	* When a reflexive verb, start with "sich". e.g. "sich amüsieren, amüsiert sich, amüsierte sich, hat sich amüsiert"
  * When a noun, should include the article, and the ending in plural. e.g. "das Abgas, -e", "das Alter, -". This is just a combination of the fields artikel_d, base_d, and plural_d.
* base_e: the English translation. e.g. "to analyze"
  * If an English translation is provided in the prompt, make sure base_e covers what is provided
* artikel_d:
  * When a noun, the article. 
  * When not a noun, blank string
* plural_d: 
  * When a noun, the plural ending. "-" if the ending does not change, and e.g. "-e" if an "e" is added.
  * When not a noun, blank string
* s1: first example sentence in German. Create a typical sentence that the word would be used in.
* s1e: english translation of s1
* s2: second example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s2e: english translation of s2
* s3: third example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s3e: english translation of s3
* s4: fourth example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s4e: english translation of s4

Other things to note: 
* If the word is in plural, convert it to singular
* Return ONLY the JSON object wrapped in a json code block, and do not include any other content or text.
`

const detectCategoryPromptTemplate = `Classify the German word into one category: noun, verb, or other.
Word: %s
Return only the category word.`

const nounPromptTemplate = `
Return the following fields in a JSON structure for the noun: %s
This is used for creating Anki cards to learn German vocabulary.

* base_d: the base form the German word. Omit the article.
* full_d: German noun with article and full plural form. e.g. "das Abgas, die Abgase", "das Alter, die Alter"
* base_e: the English translation/meanings. e.g. "exhaust gas"
  * if there area multiple meanings, separate different meanings with a ";", and group similar meanings with "/".
* artikel_d: the article.
* plural_d: the full plural form (include the article and plural word). e.g. "die Abgase", "die Alter"
* s1: 1st example sentence in German
  * Create a typical sentence that starts with the article and noun in nominative.
* s1e: english translation of s1
* s2: 2nd example sentence in German
  * Create a typical sentence that includes the plural of the noun.
	* If the plural of the noun is not used commonly, demonstrate another of the words meanings.
	* If there are no other meanings, just create a different sentence.
* s2e: english translation of s2
* s3: 3rd example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s3e: english translation of s3
* s4: 4th example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s4e: english translation of s4

Other things to note:
* If the word is in plural, convert it to singular
* German and English sentences should end with appropriate punctuation (typically ".")
* Return ONLY the JSON object wrapped in a json code block, and do not include any other content or text.
`

const verbPromptTemplate = `
Return the following fields in a JSON structure for the verb: %s
The values will be used for creating Anki cards to learn German vocabulary.

* base_d: the base form of the German verb. 
  * if the verb is always reflexive, start with "sich".
* full_d: German verb as a comma separated list of infinitive, 3rd person present, 3rd person simple past, and past participle.
  * e.g. "analysieren, analysiert, analysierte, analysiert"
  * if the verb is always reflexive, start with "sich". e.g. "sich amüsieren, amüsiert sich, amüsierte sich, sich amüsiert"
* base_e: the English translation. e.g. "to analyze"
  * always start with "to"
* artikel_d: blank string
* plural_d: blank string
* s1: The first example sentence in German.
  * Create a typical sentence that the word would be used in, with the base form. 
* s1e: english translation of s1
* s2: 2nd example sentence in German
  * Create a typical sentence that the word would be used in, with the present perfect.
* s2e: english translation of s2
* s3: 3rd example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s3e: english translation of s3
* s4: 4th example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s4e: english translation of s4

Other things to note:
* German and English sentences should end with appropriate punctuation (typically ".")
* Return ONLY the JSON object wrapped in a json code block, and do not include any other content or text.
`

func NewClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	inner, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, err
	}

	return &GeminiClient{inner: inner}, nil
}

func (c *GeminiClient) GenerateVerbJSON(ctx context.Context, word string) (GeminiResponse, error) {
	return c.generateNoteJSON(ctx, word, verbPromptTemplate)
}

func (c *GeminiClient) GenerateNounJSON(ctx context.Context, word string) (GeminiResponse, error) {
	return c.generateNoteJSON(ctx, word, nounPromptTemplate)
}

func (c *GeminiClient) GenerateOtherJSON(ctx context.Context, word string) (GeminiResponse, error) {
	return c.generateNoteJSON(ctx, word, notePromptTemplate)
}

func (c *GeminiClient) generateNoteJSON(ctx context.Context, word string, prompt string) (GeminiResponse, error) {
	result, err := c.inner.Models.GenerateContent(
		ctx,
		noteModel,
		genai.Text(fmt.Sprintf(prompt, word)),
		nil,
	)
	if err != nil {
		return GeminiResponse{}, err
	}

	return parseResponse(result.Text())
}

func (c *GeminiClient) DetectCategory(ctx context.Context, word string) (Category, error) {
	prompt := fmt.Sprintf(detectCategoryPromptTemplate, word)

	result, err := c.inner.Models.GenerateContent(
		ctx,
		noteModel,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", err
	}

	category := strings.ToLower(strings.TrimSpace(result.Text()))
	switch category {
	case string(CategoryNoun), string(CategoryVerb), string(CategoryOther):
		return Category(category), nil
	default:
		return "", fmt.Errorf("unexpected category from Gemini: %q", category)
	}
}

func parseResponse(text string) (GeminiResponse, error) {
	jsonText := strings.TrimPrefix(text, "```json")
	jsonText = strings.TrimSuffix(jsonText, "```")

	var response GeminiResponse
	if err := json.Unmarshal([]byte(jsonText), &response); err != nil {
		return GeminiResponse{}, err
	}

	return response, nil
}
