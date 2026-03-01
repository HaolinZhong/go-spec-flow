package review

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// FileDiff represents changes to a single file.
type FileDiff struct {
	Path   string
	IsNew  bool
	Hunks  []*Hunk
}

// Hunk represents a changed region in a file.
type Hunk struct {
	NewStart int
	NewCount int
}

// GetDiff runs git diff and parses the output.
func GetDiff(dir string, base string, commit string) ([]*FileDiff, error) {
	args := []string{"diff", "--unified=0"}
	if commit != "" {
		args = append(args, commit+"~1", commit)
	} else if base != "" {
		args = append(args, base+"...HEAD")
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		// git diff returns exit code 1 when there are changes with some flags
		if exitErr, ok := err.(*exec.ExitError); ok && len(out) > 0 {
			_ = exitErr
		} else {
			return nil, fmt.Errorf("git diff: %w", err)
		}
	}

	return parseDiff(string(out)), nil
}

// GetDiffNames returns just the changed file names.
func GetDiffNames(dir string, base string, commit string) ([]string, error) {
	args := []string{"diff", "--name-only"}
	if commit != "" {
		args = append(args, commit+"~1", commit)
	} else if base != "" {
		args = append(args, base+"...HEAD")
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --name-only: %w", err)
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func parseDiff(diffOutput string) []*FileDiff {
	var diffs []*FileDiff
	var current *FileDiff

	for _, line := range strings.Split(diffOutput, "\n") {
		if strings.HasPrefix(line, "diff --git") {
			if current != nil {
				diffs = append(diffs, current)
			}
			current = &FileDiff{}
			continue
		}

		if current == nil {
			continue
		}

		if strings.HasPrefix(line, "+++ b/") {
			current.Path = strings.TrimPrefix(line, "+++ b/")
			continue
		}

		if strings.HasPrefix(line, "new file mode") {
			current.IsNew = true
			continue
		}

		if strings.HasPrefix(line, "+++ /dev/null") {
			// deleted file, skip
			continue
		}

		if strings.HasPrefix(line, "@@ ") {
			hunk := parseHunkHeader(line)
			if hunk != nil {
				current.Hunks = append(current.Hunks, hunk)
			}
		}
	}

	if current != nil && current.Path != "" {
		diffs = append(diffs, current)
	}

	return diffs
}

// parseHunkHeader parses "@@ -old,count +new,count @@" header.
func parseHunkHeader(line string) *Hunk {
	// Format: @@ -old[,count] +new[,count] @@
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		return nil
	}

	newPart := parts[2] // +new[,count]
	newPart = strings.TrimPrefix(newPart, "+")
	newStart, newCount := parseRange(newPart)

	return &Hunk{
		NewStart: newStart,
		NewCount: newCount,
	}
}

func parseRange(s string) (start, count int) {
	parts := strings.Split(s, ",")
	start, _ = strconv.Atoi(parts[0])
	if len(parts) > 1 {
		count, _ = strconv.Atoi(parts[1])
	} else {
		count = 1
	}
	return
}

// FilterGoFiles returns only .go file diffs.
func FilterGoFiles(diffs []*FileDiff) []*FileDiff {
	var result []*FileDiff
	for _, d := range diffs {
		if filepath.Ext(d.Path) == ".go" {
			result = append(result, d)
		}
	}
	return result
}
