package main

import (
	"anki-voice/ankiconnect"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	outputPathMp3 = "./output.mp3"
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
	if len(os.Args) != 2 {
		log.Fatal("noteID argument required")
	}
	noteID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("invalid noteID %q: %v", os.Args[1], err)
	}

	note, err := ankiconnect.GetNote(noteID, fields)
	if err != nil {
		log.Fatal(err)
	}

	for field, phrase := range note.Phrases {
		if phrase.Value == "" {
			continue
		}

		if strings.Contains(phrase.Audio, fmt.Sprintf("-%s.mp3", field)) {
			// audio has already been generated
			continue
		}

		log.Printf("generating audio for: %s...\n", phrase.Value)

		outputPathWav := fmt.Sprintf("./output/%s.wav", phrase.Value)

		err = generateWav(phrase.Value, outputPathWav)
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(outputPathWav)

		log.Println("converting mp3...")
		// TODO: check that ffmpeg is available
		err = convertWavToMp3(outputPathWav, fmt.Sprintf("./output/%s.mp3", phrase.Value))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func generateWav(text, outputPath string) error {
	// TODO: try to adjust speed with length_scale? currently sending this just causes the voice
	// to read through the whole payload
	// payload := map[string]any{
	// 	"text":  text,
	// 	"length_scale": 1.2,
	// }

	// body, err := json.Marshal(payload)
	// if err != nil {
	// 	return fmt.Errorf("marshal request failed: %w", err)
	// }

	// response, err := http.Post("http://localhost:9999", "application/json", bytes.NewReader(body))
	// if err != nil {
	// 	return err
	// }
	// defer response.Body.Close()

	response, err := http.Post("http://localhost:9999", "application/json", strings.NewReader(text))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, response.Body); err != nil {
		return err
	}

	return nil
}

func convertWavToMp3(input, output string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("ffmpeg", "-y", "-i", input, "-codec:a", "libmp3lame", "-b:a", "192k", output)
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %v\nDetails:\n%s", err, stderr.String())
	}
	return nil
}

func keys(m map[string]string) []string {
	allKeys := make([]string, 0, len(m))
	for k := range m {
		allKeys = append(allKeys, k)
	}

	return allKeys
}
