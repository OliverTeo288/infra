package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func CloneRepo(repoURL string) error {
	targetDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get the current working directory: %v", err)
	}

	// Extract the repository name from the URL
	repoName := filepath.Base(repoURL)
	repoName = repoName[:len(repoName)-len(filepath.Ext(repoName))]

	// Temporary directory for the cloned repository
	tempCloneDir := filepath.Join(targetDir, repoName)

	// Prepare the git clone command
	args := []string{"clone", repoURL, tempCloneDir}
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute the clone command
	fmt.Printf("Cloning repository %s into %s...\n", repoURL, tempCloneDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	// Remove the .git directory to lose the remote connection
	gitDir := filepath.Join(tempCloneDir, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return fmt.Errorf("failed to remove .git directory: %v", err)
	}

	// Copy all files, including hidden ones (those starting with a '.')
	cpCmd := exec.Command("sh", "-c", fmt.Sprintf("cp -r %s/. %s", tempCloneDir, targetDir))
	cpCmd.Stdout = os.Stdout
	cpCmd.Stderr = os.Stderr
	fmt.Printf("Copying contents from %s to %s...\n", tempCloneDir, targetDir)
	if err := cpCmd.Run(); err != nil {
		return fmt.Errorf("failed to move contents using cp: %v", err)
	}

	// Remove the main cloned folder
	fmt.Printf("Removing temporary directory %s...\n", tempCloneDir)
	if err := os.RemoveAll(tempCloneDir); err != nil {
		return fmt.Errorf("failed to remove temporary directory: %v", err)
	}

	fmt.Println("Repository cloned, remote removed, and contents copied successfully.")
	return nil
}
