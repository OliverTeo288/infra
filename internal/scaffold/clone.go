// Package scaffold sets up a fresh project directory by cloning a templated
// repo and copying its contents (minus .git) into the current working
// directory.
package scaffold

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"raid/infra/internal/prompt"
)

// ErrAborted is returned when the user declines to continue after a non-empty
// CWD warning.
var ErrAborted = errors.New("aborted by user: target directory is not empty")

// Run executes the full scaffold flow:
//
//  1. Confirm the user has access to the upstream repo host
//  2. Verify CWD is empty (or get explicit user confirmation to overwrite)
//  3. Prompt for SSH vs HTTPS clone method
//  4. Clone the templated repo, drop the .git directory, copy contents into CWD
//
// httpsURL and sshURL come from build-time config; either may be empty if
// the corresponding clone method is not supported in this build.
func Run(httpsURL, sshURL string) error {
	if !prompt.Confirm("Do you have access to SHIPHATS GitLab? (Y/N)") {
		return fmt.Errorf("please ensure you have access to SHIPHATS GitLab before running this command")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	empty, err := isDirEmpty(cwd)
	if err != nil {
		return fmt.Errorf("failed to inspect current directory: %w", err)
	}
	if !empty {
		fmt.Printf("Warning: current directory %s is not empty.\n", cwd)
		fmt.Println("Project files will be copied into this directory and may overwrite existing files of the same name.")
		if !prompt.Confirm("Continue anyway? (Y/N)") {
			return ErrAborted
		}
	}

	options := []string{"Clone with SSH", "Clone with HTTPS"}
	choice, err := prompt.Selection(options, "")
	if err != nil {
		return fmt.Errorf("failed to prompt for cloning method: %w", err)
	}

	var repoURL string
	switch choice {
	case "Clone with SSH":
		if sshURL == "" {
			return fmt.Errorf("GitlabSSHDomain is not configured (missing build-time variable)")
		}
		repoURL = sshURL
	case "Clone with HTTPS":
		if httpsURL == "" {
			return fmt.Errorf("GitlabHTTPSDomain is not configured (missing build-time variable)")
		}
		repoURL = httpsURL
	default:
		return fmt.Errorf("invalid cloning method selected")
	}

	return cloneInto(repoURL, cwd)
}

// isDirEmpty reports whether path contains no entries other than hidden files
// commonly safe to ignore (.git, .DS_Store, .gitkeep).
func isDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	names, err := f.Readdirnames(-1)
	if err != nil {
		return false, err
	}
	for _, n := range names {
		switch n {
		case ".git", ".DS_Store", ".gitkeep":
			continue
		}
		return false, nil
	}
	return true, nil
}

// cloneInto runs `git clone` into a temp subdir of targetDir, removes the
// .git directory (the user is starting fresh), copies contents up to
// targetDir, and removes the temp subdir.
func cloneInto(repoURL, targetDir string) error {
	repoName := repoNameFromURL(repoURL)
	temp := filepath.Join(targetDir, repoName)

	// A previous failed run may have left the temp dir behind, which would
	// cause `git clone` to error with "destination already exists". Ask
	// before deleting so we never silently wipe user data.
	if _, err := os.Stat(temp); err == nil {
		fmt.Printf("Warning: %s already exists (likely from a previous failed run).\n", temp)
		if !prompt.Confirm("Remove it and proceed? (Y/N)") {
			return fmt.Errorf("aborted: cannot clone — destination %s already exists", temp)
		}
		if err := os.RemoveAll(temp); err != nil {
			return fmt.Errorf("failed to remove stale temp directory: %w", err)
		}
	}

	fmt.Printf("Cloning repository %s into %s...\n", repoURL, temp)
	cmd := exec.Command("git", "clone", repoURL, temp)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	if err := os.RemoveAll(filepath.Join(temp, ".git")); err != nil {
		return fmt.Errorf("failed to remove .git directory: %w", err)
	}

	fmt.Printf("Copying contents from %s to %s...\n", temp, targetDir)
	if err := copyDir(temp, targetDir); err != nil {
		return fmt.Errorf("failed to copy contents: %w", err)
	}

	fmt.Printf("Removing temporary directory %s...\n", temp)
	if err := os.RemoveAll(temp); err != nil {
		return fmt.Errorf("failed to remove temporary directory: %w", err)
	}

	fmt.Println("Repository cloned, remote removed, and contents copied successfully.")
	return nil
}

// repoNameFromURL extracts a repo name from either a git-over-SSH URL
// (git@host:owner/repo.git) or HTTPS URL (https://host/owner/repo.git).
// Trailing ".git" and any query string / fragment are stripped.
func repoNameFromURL(repoURL string) string {
	// Strip query and fragment ("?ref=main", "#anchor") which would
	// otherwise end up in the filename.
	if i := strings.IndexAny(repoURL, "?#"); i >= 0 {
		repoURL = repoURL[:i]
	}
	// path.Base understands forward slashes; covers both URL forms.
	name := path.Base(repoURL)
	name = strings.TrimSuffix(name, ".git")
	if name == "" || name == "." || name == "/" {
		// Fallback so we never produce an empty directory name.
		return "template-repo"
	}
	return name
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return copyFile(p, dstPath, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
