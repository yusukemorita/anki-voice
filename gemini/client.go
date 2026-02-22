package gemini

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type APIError = genai.APIError

const	noteModel = "gemini-2.5-flash"

func NewClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	inner, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, err
	}

	return &GeminiClient{inner: inner}, nil
}

type GeminiClient struct {
	inner *genai.Client
}

func (c *GeminiClient) GenerateVerbJSON(ctx context.Context, word string) (GeminiResponse, error) {
	return c.generateNoteJSON(ctx, word, verbPromptTemplate)
}

func (c *GeminiClient) GenerateNounJSON(ctx context.Context, word string) (GeminiResponse, error) {
	return c.generateNoteJSON(ctx, word, nounPromptTemplate)
}

func (c *GeminiClient) GenerateOtherJSON(ctx context.Context, word string) (GeminiResponse, error) {
	return c.generateNoteJSON(ctx, word, defaultPromptTemplate)
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
