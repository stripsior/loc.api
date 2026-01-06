package github

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func CloneRepository(ctx context.Context, owner, repo, branch, token string, includeAuthors bool) (string, error) {
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("loc-api-%s-%s-*", owner, repo))
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	url := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	cloneOptions := &git.CloneOptions{
		URL:           url,
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Progress:      nil,
	}

	if !includeAuthors {
		cloneOptions.Depth = 1
	}

	if token != "" {
		cloneOptions.Auth = &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}
	}

	// Clone with context support for timeout/cancellation
	_, err = git.PlainCloneContext(ctx, tmpDir, false, cloneOptions)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	return tmpDir, nil
}

func CleanupRepository(path string) error {
	if path == "" {
		return nil
	}

	if !filepath.HasPrefix(path, os.TempDir()) {
		return fmt.Errorf("refusing to remove directory outside temp: %s", path)
	}

	return os.RemoveAll(path)
}
