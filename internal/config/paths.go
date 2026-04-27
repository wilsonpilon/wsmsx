package config

import (
	"os"
	"path/filepath"
)

const (
	AppName = "ws7"
)

func DataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(base, AppName)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", err
	}
	return path, nil
}

func DBPath() (string, error) {
	base, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "ws7.db"), nil
}

