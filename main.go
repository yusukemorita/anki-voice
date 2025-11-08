package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {
	text := "test 2, hallo"
	outputPathWav := "./output.wav"
	outputPathMp3 := "./output.mp3"

	// TODO: check that piper is running

	log.Println("generating wav...")
	err := generateWav(text, outputPathWav)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(outputPathWav)

	log.Println("converting mp3...")
	// TODO: check that ffmpeg is available
	err = convertWavToMp3(outputPathWav, outputPathMp3)
	if err != nil {
		log.Fatal(err)
	}
}

func generateWav(text, outputPath string ) error {
	req, err := http.NewRequest(http.MethodPost, "http://localhost:9999", strings.NewReader(text))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
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
