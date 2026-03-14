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

type FileSet struct {
	Key      string
	UpPath   string
	DownPath string
}

func ValidateDirectory(path string) (int, error) {
	files, err := CollectDirectory(path)
	if err != nil {
		return 0, err
	}

	return len(files), nil
}

func CollectDirectory(path string) ([]FileSet, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read migration dir: %w", err)
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
		return nil, fmt.Errorf("no migration pair found in %s", filepath.Clean(path))
	}

	missing := make([]string, 0)
	for key, state := range pairs {
		if !state.up || !state.down {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("migration pair incomplete for: %s", strings.Join(missing, ", "))
	}

	keys := make([]string, 0, len(pairs))
	for key := range pairs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]FileSet, 0, len(keys))
	for _, key := range keys {
		result = append(result, FileSet{
			Key:      key,
			UpPath:   filepath.Join(path, key+".up.sql"),
			DownPath: filepath.Join(path, key+".down.sql"),
		})
	}

	return result, nil
}
