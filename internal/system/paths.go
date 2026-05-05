package system

import (
	"os"
	"path/filepath"
)

type Paths struct {
	HomeDir   string `json:"home_dir"`
	ConfigDir string `json:"config_dir"`
	WorkDir   string `json:"work_dir"`
}

func ResolvePaths() (Paths, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return Paths{}, err
	}
	workDir, err := os.Getwd()
	if err != nil {
		return Paths{}, err
	}
	return Paths{
		HomeDir:   homeDir,
		ConfigDir: filepath.Join(configDir, "nea-ai"),
		WorkDir:   workDir,
	}, nil
}
