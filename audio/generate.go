package audio

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)


func GenerateMP3(text, outputPath string) error {
	wavPath := fmt.Sprintf("%s.wav", outputPath)

	err := generateWav(text, wavPath) 
	if err != nil {
		return err
	}
	defer os.Remove(wavPath)

	err = convertWavToMp3(wavPath, outputPath)
	if err != nil {
		return err
	}

	return nil
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
