package review

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// GitDiffFile represents a single file's diff output.
type GitDiffFile struct {
	Path    string
	Content string // raw diff content for this file
	IsNew   bool
}

// RunGitDiff runs git diff with the given range and returns per-file diffs.
func RunGitDiff(dir, diffRange string) ([]*GitDiffFile, error) {
	args := []string{"diff", "--no-color"}
	if diffRange != "" {
		parts := strings.Fields(diffRange)
		args = append(args, parts...)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(out) > 0 {
			_ = exitErr
		} else {
			return nil, fmt.Errorf("git diff: %w", err)
		}
	}

	return splitDiffByFile(string(out)), nil
}

// RunGitDiffStat returns the list of changed file paths.
func RunGitDiffStat(dir, diffRange string) ([]string, error) {
	args := []string{"diff", "--name-only"}
	if diffRange != "" {
		parts := strings.Fields(diffRange)
		args = append(args, parts...)
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

// splitDiffByFile splits a unified diff into per-file chunks.
func splitDiffByFile(diff string) []*GitDiffFile {
	if diff == "" {
		return nil
	}

	var files []*GitDiffFile
	lines := strings.Split(diff, "\n")
	var current *GitDiffFile
	var buf []string

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			// Flush previous
			if current != nil {
				current.Content = strings.Join(buf, "\n")
				files = append(files, current)
			}
			// Parse file path: "diff --git a/path b/path"
			parts := strings.Fields(line)
			path := ""
			if len(parts) >= 4 {
				path = strings.TrimPrefix(parts[3], "b/")
			}
			current = &GitDiffFile{Path: path}
			buf = []string{line}
		} else if current != nil {
			if strings.HasPrefix(line, "new file mode") {
				current.IsNew = true
			}
			buf = append(buf, line)
		}
	}

	if current != nil {
		current.Content = strings.Join(buf, "\n")
		files = append(files, current)
	}

	return files
}

// extractFuncDiff extracts diff hunks that overlap with a function's line range.
// lineStart/lineEnd refer to the NEW file's line numbers.
func extractFuncDiff(fileDiff string, lineStart, lineEnd int) string {
	lines := strings.Split(fileDiff, "\n")
	var result []string
	var inHunk bool
	var hunkNewStart, hunkNewEnd int
	var hunkBuf []string

	flushHunk := func() {
		if inHunk && hunkNewEnd >= lineStart && hunkNewStart <= lineEnd {
			result = append(result, hunkBuf...)
		}
		hunkBuf = nil
		inHunk = false
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			flushHunk()
			// Parse: @@ -old,count +new,count @@
			hunkNewStart, hunkNewEnd = parseHunkRange(line)
			inHunk = true
			hunkBuf = append(hunkBuf, line)
		} else if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") ||
			strings.HasPrefix(line, "new file") || strings.HasPrefix(line, "old mode") ||
			strings.HasPrefix(line, "new mode") {
			continue // skip file headers
		} else if inHunk {
			hunkBuf = append(hunkBuf, line)
			// Track new-file line advancement to update hunkNewEnd
			if !strings.HasPrefix(line, "-") {
				hunkNewEnd++
			}
		}
	}
	flushHunk()

	if len(result) == 0 {
		return ""
	}
	return strings.Join(result, "\n")
}

// parseHunkRange parses "@@ -10,5 +12,7 @@" and returns (newStart, newEnd).
func parseHunkRange(hunkHeader string) (int, int) {
	// Find the +N,M part
	plusIdx := strings.Index(hunkHeader, "+")
	if plusIdx < 0 {
		return 0, 0
	}
	rest := hunkHeader[plusIdx+1:]
	spaceIdx := strings.Index(rest, " ")
	if spaceIdx > 0 {
		rest = rest[:spaceIdx]
	}
	parts := strings.SplitN(rest, ",", 2)
	start, _ := strconv.Atoi(parts[0])
	count := 1
	if len(parts) > 1 {
		count, _ = strconv.Atoi(parts[1])
	}
	return start, start + count - 1
}
