package anki

import (
	"fmt"
	"os"
	"path/filepath"
)

func MediaDir() (string, error) {
	// anki media directory setup
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	ankiMediaDir := filepath.Join(homeDir, "Library", "Application Support", "Anki2", "User 1", "collection.media")
	info, err := os.Stat(ankiMediaDir)
	if err != nil {
		return "", fmt.Errorf("anki media directory missing: %v", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("anki media path is not a directory: %s", ankiMediaDir)
	}

	return ankiMediaDir, nil
}
