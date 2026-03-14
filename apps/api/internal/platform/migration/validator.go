package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type pair struct {
	up   bool
	down bool
}

func ValidateDirectory(path string) (int, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, fmt.Errorf("read migration dir: %w", err)
	}

	pairs := make(map[string]*pair)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		switch {
		case strings.HasSuffix(name, ".up.sql"):
			key := strings.TrimSuffix(name, ".up.sql")
			if _, exists := pairs[key]; !exists {
				pairs[key] = &pair{}
			}
			pairs[key].up = true
		case strings.HasSuffix(name, ".down.sql"):
			key := strings.TrimSuffix(name, ".down.sql")
			if _, exists := pairs[key]; !exists {
				pairs[key] = &pair{}
			}
			pairs[key].down = true
		}
	}

	if len(pairs) == 0 {
		return 0, fmt.Errorf("no migration pair found in %s", filepath.Clean(path))
	}

	missing := make([]string, 0)
	for key, state := range pairs {
		if !state.up || !state.down {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return 0, fmt.Errorf("migration pair incomplete for: %s", strings.Join(missing, ", "))
	}

	return len(pairs), nil
}
