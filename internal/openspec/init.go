package openspec

import (
	"errors"
	"os"
	"path/filepath"
)

type InitResult struct {
	RootPath   string   `json:"root_path"`
	Created    []string `json:"created"`
	Existing   []string `json:"existing"`
	ConfigPath string   `json:"config_path"`
	StatusPath string   `json:"status_path"`
}

func Init(projectDir string) (InitResult, error) {
	root := filepath.Join(projectDir, "openspec")
	result := InitResult{
		RootPath:   root,
		ConfigPath: filepath.Join(root, "config.yaml"),
		StatusPath: filepath.Join(root, "changes", ".status.yaml"),
	}

	for _, dir := range []string{
		root,
		filepath.Join(root, "specs"),
		filepath.Join(root, "changes"),
		filepath.Join(root, "changes", "archive"),
	} {
		created, err := ensureDir(dir)
		if err != nil {
			return result, err
		}
		if created {
			result.Created = append(result.Created, dir)
		} else {
			result.Existing = append(result.Existing, dir)
		}
	}

	configCreated, err := ensureFile(result.ConfigPath, defaultConfig())
	if err != nil {
		return result, err
	}
	appendPath(&result, result.ConfigPath, configCreated)

	statusCreated, err := ensureFile(result.StatusPath, defaultStatus())
	if err != nil {
		return result, err
	}
	appendPath(&result, result.StatusPath, statusCreated)

	return result, nil
}

func ensureDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return false, errors.New(path + " exists and is not a directory")
		}
		return false, nil
	}
	if !os.IsNotExist(err) {
		return false, err
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return false, err
	}
	return true, nil
}

func ensureFile(path string, content string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return false, nil
	}
	if !os.IsNotExist(err) {
		return false, err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func appendPath(result *InitResult, path string, created bool) {
	if created {
		result.Created = append(result.Created, path)
		return
	}
	result.Existing = append(result.Existing, path)
}

func defaultConfig() string {
	return `project: nea-ai
artifact_store:
  mode: openspec
experimental:
  neabrain: true
rules:
  apply:
    tdd: standard
`
}

func defaultStatus() string {
	return `change: ""
current_phase: INIT
pending_tasks: []
awaiting_approval: false
completed: false
modified_artifacts: []
notes: "initialized by nea-ai"
`
}
