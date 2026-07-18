package status

import (
	"bytes"
	"os/exec"
	"strings"
)

// GitInfo holds extracted Git repository information
type GitInfo struct {
	Branch      string
	Commit      string
	WorkingTree string
}

// GetGitInfo executes git commands to retrieve current repo status
func GetGitInfo(rootDir string) (*GitInfo, error) {
	info := &GitInfo{
		Branch:      "Unknown",
		Commit:      "Unknown",
		WorkingTree: "Unknown",
	}

	// Get Branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = rootDir
	if out, err := branchCmd.Output(); err == nil {
		info.Branch = strings.TrimSpace(string(out))
	}

	// Get Commit
	commitCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	commitCmd.Dir = rootDir
	if out, err := commitCmd.Output(); err == nil {
		info.Commit = strings.TrimSpace(string(out))
	}

	// Get Working Tree Status
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = rootDir
	if out, err := statusCmd.Output(); err == nil {
		if len(bytes.TrimSpace(out)) == 0 {
			info.WorkingTree = "Clean"
		} else {
			info.WorkingTree = "Dirty"
		}
	}

	return info, nil
}
