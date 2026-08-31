// Package envfile loads KEY=VALUE pairs from a .env file into the process
// environment, so local development doesn't require manually exporting
// every variable in every new shell session. Values already set in the
// real environment always win - a .env entry never overrides one, matching
// the usual convention (so CI/production env vars are never shadowed by a
// stray .env file that happens to be present).
package envfile

import (
	"bufio"
	"os"
	"strings"
)

// Load reads path (typically ".env") and calls os.Setenv for every
// KEY=VALUE line whose KEY is not already set in the environment. It is
// silent and a no-op if path does not exist - .env is optional; real
// environment variables (e.g. set by a process manager or container) are
// enough on their own.
func Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		if _, alreadySet := os.LookupEnv(key); alreadySet {
			continue
		}
		os.Setenv(key, value)
	}
	return scanner.Err()
}
