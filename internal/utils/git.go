package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func CloneRepo(repoURL string) error {
	targetDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get the current working directory: %v", err)
	}

	repoName := filepath.Base(repoURL)
	repoName = repoName[:len(repoName)-len(filepath.Ext(repoName))]
	tempCloneDir := filepath.Join(targetDir, repoName)

	cmd := exec.Command("git", "clone", repoURL, tempCloneDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Cloning repository %s into %s...\n", repoURL, tempCloneDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	gitDir := filepath.Join(tempCloneDir, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return fmt.Errorf("failed to remove .git directory: %v", err)
	}

	fmt.Printf("Copying contents from %s to %s...\n", tempCloneDir, targetDir)
	if err := copyDir(tempCloneDir, targetDir); err != nil {
		return fmt.Errorf("failed to copy contents: %v", err)
	}

	fmt.Printf("Removing temporary directory %s...\n", tempCloneDir)
	if err := os.RemoveAll(tempCloneDir); err != nil {
		return fmt.Errorf("failed to remove temporary directory: %v", err)
	}

	fmt.Println("Repository cloned, remote removed, and contents copied successfully.")
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return copyFile(path, dstPath, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
