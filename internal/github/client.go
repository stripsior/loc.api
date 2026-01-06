package github

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

const MaxRepoSizeMB = 500
const MaxRepoSizeKB = MaxRepoSizeMB * 1024

type Client struct {
	client *github.Client
	ctx    context.Context
}

func NewClient(token string) *Client {
	ctx := context.Background()
	var client *github.Client

	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	return &Client{
		client: client,
		ctx:    ctx,
	}
}

func ParseRepository(input string) (owner, repo string, err error) {
	input = strings.TrimSpace(input)

	input = strings.TrimPrefix(input, "https://")
	input = strings.TrimPrefix(input, "http://")

	input = strings.TrimPrefix(input, "github.com/")
	input = strings.TrimPrefix(input, "www.github.com/")

	input = strings.TrimSuffix(input, "/")
	input = strings.TrimSuffix(input, ".git")

	parts := strings.Split(input, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repository format: expected 'owner/repo', got '%s'", input)
	}

	owner = parts[0]
	repo = parts[1]

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid repository format: owner and repo cannot be empty")
	}

	return owner, repo, nil
}

func (c *Client) GetRepository(owner, repo string) (*github.Repository, error) {
	repository, _, err := c.client.Repositories.Get(c.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}

	return repository, nil
}

func (c *Client) GetBranches(owner, repo string) ([]string, error) {
	var allBranches []string
	opts := &github.BranchListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		branches, resp, err := c.client.Repositories.ListBranches(c.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch branches: %w", err)
		}

		for _, branch := range branches {
			if branch.Name != nil {
				allBranches = append(allBranches, *branch.Name)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allBranches, nil
}

func (c *Client) GetUserByEmail(email string) string {
	query := fmt.Sprintf("%s in:email", email)
	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 1},
	}

	result, _, err := c.client.Search.Users(c.ctx, query, opts)
	if err == nil && result.GetTotal() > 0 && len(result.Users) > 0 {
		if result.Users[0].AvatarURL != nil {
			return *result.Users[0].AvatarURL
		}
	}

	return getGravatarURL(email)
}

func getGravatarURL(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	hash := md5.Sum([]byte(email))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=identicon&s=200", hash)
}

func (c *Client) ValidateRepositorySize(repository *github.Repository) error {
	if repository.Size == nil {
		return fmt.Errorf("repository size information not available")
	}

	sizeKB := *repository.Size
	if sizeKB > MaxRepoSizeKB {
		return fmt.Errorf("repository size (%d KB) exceeds the maximum allowed size (%d MB)",
			sizeKB, MaxRepoSizeMB)
	}

	return nil
}

func GetDefaultBranch(repository *github.Repository) string {
	if repository.DefaultBranch != nil {
		return *repository.DefaultBranch
	}
	return "main"
}

func GetToken() string {
	return os.Getenv("GITHUB_TOKEN")
}
