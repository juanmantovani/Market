package environments

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LoadProjectDotEnv reads the repository .env into the process environment.
// The root is the nearest ancestor of the working directory that contains
// pnpm-workspace.yaml, because the server starts from apps/backend. A missing
// root or a missing file is ignored. Variables already present in the
// environment are kept, so a shell override and Docker Compose win.
func LoadProjectDotEnv() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	for {
		marker := filepath.Join(dir, "pnpm-workspace.yaml")
		info, statErr := os.Stat(marker)
		if statErr == nil && !info.IsDir() {
			return loadDotEnv(filepath.Join(dir, ".env"))
		}
		if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return statErr
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	pairs, err := parseDotEnv(file)
	if err != nil {
		return err
	}

	for key, value := range pairs {
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}

// parseDotEnv collects assignments. A later line for the same key replaces
// an earlier one in the file; the caller still refuses to replace a variable
// that the process already has.
func parseDotEnv(reader io.Reader) (map[string]string, error) {
	scanner := bufio.NewScanner(reader)
	pairs := make(map[string]string)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		key, value, ok, err := parseDotEnvLine(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf(".env:%d: %w", lineNumber, err)
		}
		if ok {
			pairs[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return pairs, nil
}

func parseDotEnvLine(line string) (string, string, bool, error) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false, nil
	}

	if rest, found := strings.CutPrefix(line, "export"); found && (rest == "" || rest[0] == ' ' || rest[0] == '\t') {
		line = strings.TrimSpace(rest)
	}

	key, value, found := strings.Cut(line, "=")
	if !found || !validEnvName(strings.TrimSpace(key)) {
		return "", "", false, errors.New("invalid assignment")
	}
	key = strings.TrimSpace(key)

	parsed, err := unquoteEnvValue(value)
	if err != nil {
		return "", "", false, err
	}
	return key, parsed, true, nil
}

func unquoteEnvValue(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	quote := value[0]
	if quote != '"' && quote != '\'' {
		return value, nil
	}
	if len(value) == 1 || value[len(value)-1] != quote {
		return "", errors.New("unclosed quote")
	}
	return value[1 : len(value)-1], nil
}

func validEnvName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		switch {
		case r == '_' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z':
		case i > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}
