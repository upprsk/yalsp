package analysis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

type yalcLogLine struct {
	FileID       int    `json:"file_id"`
	Span         [2]int `json:"span"`
	Prefix       string `json:"prefix"`
	FullPath     string `json:"fullpath"`
	OriginalPath string `json:"original_path"`
	Message      string `json:"message"`
}

func yalcSingleFileCmd(path, filename string) *exec.Cmd {
	return exec.Command(path, filename, "--file", "--error-format", "json")
}

func runYalcSingleFile(path, filename string) ([]yalcLogLine, error) {
	cmd := yalcSingleFileCmd(path, filename)

	var stderrBuffer bytes.Buffer
	cmd.Stderr = &stderrBuffer

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("run: %w", err)
	}

	var lines []yalcLogLine

	dec := json.NewDecoder(&stderrBuffer)
	for {
		var line yalcLogLine
		err := dec.Decode(&line)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("json decode: %w", err)
		}

		lines = append(lines, line)
	}

	return lines, nil
}
