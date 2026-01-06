package analyzer

import (
	"bufio"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/stripsior/loc.api/internal/models"
)

type AuthorInfo struct {
	Name  string
	Email string
}

func AnalyzeAuthors(repoPath string, files []models.FileInfo) (map[string]*models.AuthorStats, error) {
	authorStats := make(map[string]*models.AuthorStats)
	filesProcessed := make(map[string]bool)

	for _, file := range files {
		fullPath := filepath.Join(repoPath, file.Path)

		if filesProcessed[fullPath] {
			continue
		}
		filesProcessed[fullPath] = true

		authors, err := getFileAuthors(repoPath, file.Path)
		if err != nil {
			continue
		}

		for email, author := range authors {
			if _, exists := authorStats[email]; !exists {
				authorStats[email] = &models.AuthorStats{
					Name:  author.Name,
					Email: email,
				}
			}

			stats := authorStats[email]
			stats.TotalLines += author.TotalLines
			stats.CodeLines += author.CodeLines
			stats.CommentLines += author.CommentLines
			stats.BlankLines += author.BlankLines
			stats.FileCount++
		}
	}

	return authorStats, nil
}

func getFileAuthors(repoPath, relPath string) (map[string]*models.AuthorStats, error) {
	cmd := exec.Command("git", "blame", "--line-porcelain", relPath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	authors := make(map[string]*models.AuthorStats)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	var currentEmail string
	var currentName string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "author-mail ") {
			email := strings.TrimPrefix(line, "author-mail ")
			email = strings.Trim(email, "<>")
			currentEmail = email
		} else if strings.HasPrefix(line, "author ") {
			name := strings.TrimPrefix(line, "author ")
			currentName = name
		} else if strings.HasPrefix(line, "\t") {
			if currentEmail == "" {
				continue
			}

			if _, exists := authors[currentEmail]; !exists {
				authors[currentEmail] = &models.AuthorStats{
					Name:  currentName,
					Email: currentEmail,
				}
			}

			lineContent := strings.TrimPrefix(line, "\t")
			trimmed := strings.TrimSpace(lineContent)

			author := authors[currentEmail]
			author.TotalLines++

			if trimmed == "" {
				author.BlankLines++
			} else if isCommentLine(trimmed) {
				author.CommentLines++
			} else {
				author.CodeLines++
			}
		}
	}

	return authors, scanner.Err()
}

func isCommentLine(line string) bool {
	commentPrefixes := []string{"//", "#", "/*", "*", "*/", "<!--", "-->"}
	for _, prefix := range commentPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}
