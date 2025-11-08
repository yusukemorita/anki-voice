package ankiconnect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

const ankiURI = "http://localhost:8765"

type Note struct {
	NoteID  int
	Phrases map[string]Phrase // key: field name
}

type Phrase struct {
	Value string // the actual phrase in the target language
	Audio string // the audio field value. format is typically [sound:filename.mp3], and can also be empty.
}

// GetNote retrieves the fields of the note with the given noteID
func GetNote(noteID int, fields map[string]string) (Note, error) {
	payload := map[string]any{
		"action":  "notesInfo",
		"version": 5,
		"params": map[string]any{
			"notes": []int{noteID},
		},
	}

	responseBody, err := sendRequest(payload)
	if err != nil {
		return Note{}, err
	}

	if errMsg := gjson.GetBytes(responseBody, "error"); errMsg.Exists() && errMsg.String() != "" {
		return Note{}, fmt.Errorf("anki connect error: %s", errMsg.String())
	}

	noteResult := gjson.GetBytes(responseBody, "result.0")
	if !noteResult.Exists() {
		return Note{}, fmt.Errorf("note %d not found", noteID)
	}

	result := Note{
		NoteID:  noteID,
		Phrases: make(map[string]Phrase),
	}

	for field, audioField := range fields {
		fieldValue := noteResult.Get(fmt.Sprintf("fields.%s.value", field)).String()
		audioFieldValue := noteResult.Get(fmt.Sprintf("fields.%s.value", audioField)).String()

		result.Phrases[field] = Phrase{Value: fieldValue, Audio: audioFieldValue}
	}

	return result, nil
}

// QueryNotes retrieves note IDs with the given anki query
func QueryNotes(query string) ([]int, error) {
	payload := map[string]any{
		"action":  "findNotes",
		"version": 5,
		"params": map[string]any{
			"query": query,
		},
	}

	responseBody, err := sendRequest(payload)
	if err != nil {
		return nil, err
	}

	if errMsg := gjson.GetBytes(responseBody, "error"); errMsg.Exists() && errMsg.String() != "" {
		return nil, fmt.Errorf("anki connect error: %s", errMsg.String())
	}

	log.Println(string(responseBody))

	gjsonResult := gjson.GetBytes(responseBody, "result")
	if !gjsonResult.Exists() {
		return nil, errors.New("gjsonResult doesn't exist")
	}

	var ids []int

	for _, result := range gjsonResult.Array() {
		ids = append(ids, int(result.Int()))
	}

	return ids, nil
}


func UpdateNoteField(noteID int, fieldName, fieldValue string) error {
	payload := map[string]any{
		"action":  "updateNoteFields",
		"version": 5,
		"params": map[string]any{
			"note": map[string]any{
				"id": noteID,
				"fields": map[string]any{
					fieldName: fieldValue,
				},
			},
		},
	}

	_, err := sendRequest(payload)
	if err != nil {
		return err
	}

	return nil
}

func AddNoteTag(noteID int, tag string) error {
	payload := map[string]any{
		"action":  "addTags",
		"version": 5,
		"params": map[string]any{
			"notes": []int{noteID},
			"tags":  tag,
		},
	}

	_, err := sendRequest(payload)
	if err != nil {
		return err
	}

	return nil
}

func sendRequest(payload map[string]any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	response, err := http.Post(ankiURI, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("request anki connect: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anki connect returned %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	if errMsg := gjson.GetBytes(responseBody, "error"); errMsg.Exists() && errMsg.String() != "" {
		return nil, fmt.Errorf("anki connect error: %s", errMsg.String())
	}

	return responseBody, nil
}
