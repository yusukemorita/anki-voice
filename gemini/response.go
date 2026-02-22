package gemini

import (
	"encoding/json"
	"strings"
)

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

func parseResponse(text string) (GeminiResponse, error) {
	jsonText := strings.TrimPrefix(text, "```json")
	jsonText = strings.TrimSuffix(jsonText, "```")

	var response GeminiResponse
	if err := json.Unmarshal([]byte(jsonText), &response); err != nil {
		return GeminiResponse{}, err
	}

	return response, nil
}
