package review

import (
	"fmt"
	"os"
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

// DiffMode indicates which type of diff was used.
type DiffMode string

const (
	DiffModeStaged    DiffMode = "staged"
	DiffModeUnstaged  DiffMode = "unstaged"
	DiffModeCommit    DiffMode = "commit"
	DiffModeBase      DiffMode = "base"
)

// DiffOutput holds diff results and metadata about how they were obtained.
type DiffOutput struct {
	Diffs []*FileDiff
	Mode  DiffMode
}

// GetDiff runs git diff and parses the output.
// When no commit/base is specified, tries staged changes first, then falls back to unstaged.
func GetDiff(dir string, base string, commit string) ([]*FileDiff, error) {
	output, err := GetDiffWithMode(dir, base, commit)
	if err != nil {
		return nil, err
	}
	return output.Diffs, nil
}

// GetDiffWithMode is like GetDiff but also returns the diff mode used.
func GetDiffWithMode(dir string, base string, commit string) (*DiffOutput, error) {
	if commit != "" {
		diffs, err := runGitDiff(dir, "diff", "--unified=0", commit+"~1", commit)
		if err != nil {
			return nil, err
		}
		return &DiffOutput{Diffs: diffs, Mode: DiffModeCommit}, nil
	}

	if base != "" {
		diffs, err := runGitDiff(dir, "diff", "--unified=0", base+"...HEAD")
		if err != nil {
			return nil, err
		}
		return &DiffOutput{Diffs: diffs, Mode: DiffModeBase}, nil
	}

	// No commit/base: try staged first, then unstaged
	staged, err := runGitDiff(dir, "diff", "--staged", "--unified=0")
	if err != nil {
		return nil, err
	}
	if len(staged) > 0 {
		return &DiffOutput{Diffs: staged, Mode: DiffModeStaged}, nil
	}

	unstaged, err := runGitDiff(dir, "diff", "--unified=0")
	if err != nil {
		return nil, err
	}
	return &DiffOutput{Diffs: unstaged, Mode: DiffModeUnstaged}, nil
}

func runGitDiff(dir string, args ...string) ([]*FileDiff, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(out) > 0 {
			_ = exitErr
		} else {
			return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
		}
	}
	return parseDiff(string(out)), nil
}

// GetUntrackedGoFiles returns FileDiff entries for untracked .go files.
// Each file is treated as new (IsNew: true) with the entire file as a single hunk.
func GetUntrackedGoFiles(dir string) ([]*FileDiff, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w", err)
	}

	var diffs []*FileDiff
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" || filepath.Ext(line) != ".go" {
			continue
		}
		// Count lines in the file for the hunk
		lineCount, err := countFileLines(filepath.Join(dir, line))
		if err != nil {
			continue
		}
		diffs = append(diffs, &FileDiff{
			Path:  line,
			IsNew: true,
			Hunks: []*Hunk{{NewStart: 1, NewCount: lineCount}},
		})
	}
	return diffs, nil
}

func countFileLines(path string) (int, error) {
	data, err := exec.Command("wc", "-l", path).Output()
	if err != nil {
		// Fallback: read file
		content, err2 := os.ReadFile(path)
		if err2 != nil {
			return 0, err2
		}
		return strings.Count(string(content), "\n") + 1, nil
	}
	parts := strings.Fields(strings.TrimSpace(string(data)))
	if len(parts) == 0 {
		return 0, fmt.Errorf("unexpected wc output")
	}
	n, _ := strconv.Atoi(parts[0])
	return n, nil
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
