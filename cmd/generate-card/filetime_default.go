//go:build !darwin

package main

import (
	"os"
	"time"
)

func fileCreateTime(info os.FileInfo) time.Time {
	return info.ModTime()
}
