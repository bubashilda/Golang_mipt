package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func ListFiles(repositoryRoot, targetRevision string) ([]string, error) {
	gitCmd := exec.Command("git", "ls-tree", "-r", targetRevision)
	gitCmd.Dir = repositoryRoot

	output, cmdErr := gitCmd.Output()
	if cmdErr != nil {
		return nil, fmt.Errorf("failed to list files in repository %q at revision %q: %w", repositoryRoot, targetRevision, cmdErr)
	}

	rawLines := strings.Split(string(output), "\n")
	var filePaths []string
	for _, line := range rawLines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			filePaths = append(filePaths, parts[1])
		}
	}
	return filePaths, nil
}

func Blame(repoDir, targetFile, targetRevision string) ([]byte, error) {
	gitBlame := exec.Command("git", "blame", "--porcelain", targetFile, targetRevision)
	gitBlame.Dir = repoDir

	blameOutput, execErr := gitBlame.Output()
	if execErr != nil {
		return nil, fmt.Errorf("git blame failed for file %q in %q at %q: %w", targetFile, repoDir, targetRevision, execErr)
	}
	return blameOutput, nil
}

func LogLast(repoPath, fileName, revision, logFormat string) ([]byte, error) {
	if logFormat == "" {
		logFormat = "medium"
	}

	logCmd := exec.Command("git", "log", "-1", fmt.Sprintf("--pretty=%s", logFormat), revision, "--", fileName)
	logCmd.Dir = repoPath

	logOutput, err := logCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve git log for %q in %q with revision %q: %w", fileName, repoPath, revision, err)
	}
	return logOutput, nil
}
