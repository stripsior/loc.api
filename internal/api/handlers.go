package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripsior/loc.api/internal/analyzer"
	"github.com/stripsior/loc.api/internal/cache"
	"github.com/stripsior/loc.api/internal/github"
	"github.com/stripsior/loc.api/internal/models"
)

func analyzeRepository(c *gin.Context) {
	var req models.AnalyzeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	hasToken := req.Token != ""
	cacheKey := cache.GenerateCacheKey(req.Repository, req.Branch, hasToken, req.IncludeAuthors, req.Filters)

	cacheInstance, exists := c.Get("cache")
	if exists {
		if resultCache, ok := cacheInstance.(*cache.Cache); ok {
			if cachedResult, found := resultCache.Get(cacheKey); found {
				c.Header("X-Cache", "HIT")
				c.JSON(http.StatusOK, cachedResult)
				return
			}
		}
	}

	c.Header("X-Cache", "MISS")

	owner, repo, err := github.ParseRepository(req.Repository)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_repository",
			Message: err.Error(),
		})
		return
	}

	token := req.Token
	if token == "" {
		token = github.GetToken()
	}
	ghClient := github.NewClient(token)

	ghRepo, err := ghClient.GetRepository(owner, repo)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "repository_not_found",
			Message: fmt.Sprintf("Failed to fetch repository: %v", err),
		})
		return
	}

	if err := ghClient.ValidateRepositorySize(ghRepo); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "repository_too_large",
			Message: err.Error(),
		})
		return
	}

	branches, err := ghClient.GetBranches(owner, repo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "failed_to_fetch_branches",
			Message: fmt.Sprintf("Failed to fetch branches: %v", err),
		})
		return
	}

	isPrivate := ghRepo.Private != nil && *ghRepo.Private

	branch := req.Branch
	if branch == "" {
		branch = github.GetDefaultBranch(ghRepo)
	}

	repoPath, err := github.CloneRepository(owner, repo, branch, token, req.IncludeAuthors)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "clone_failed",
			Message: fmt.Sprintf("Failed to clone repository: %v", err),
		})
		return
	}

	defer github.CleanupRepository(repoPath)

	fileAnalyzer := analyzer.NewAnalyzer(repoPath, req.Filters)

	result, err := fileAnalyzer.Analyze()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "analysis_failed",
			Message: fmt.Sprintf("Failed to analyze repository: %v", err),
		})
		return
	}

	if req.IncludeAuthors {
		authorStats, err := analyzer.AnalyzeAuthors(repoPath, result.Files)
		if err == nil {
			totalLines := result.Summary.TotalLines
			if totalLines > 0 {
				for email, stats := range authorStats {
					stats.Percentage = float64(stats.TotalLines) / float64(totalLines) * 100
					stats.AvatarURL = ghClient.GetUserByEmail(email)
				}
			}
			authorsMap := make(map[string]models.AuthorStats)
			for email, stats := range authorStats {
				authorsMap[email] = *stats
			}
			result.Authors = authorsMap
		}
	}

	result.Repository = models.RepositoryInfo{
		Owner:          owner,
		Name:           repo,
		FullName:       fmt.Sprintf("%s/%s", owner, repo),
		Description:    getStringValue(ghRepo.Description),
		Stars:          getIntValue(ghRepo.StargazersCount),
		Forks:          getIntValue(ghRepo.ForksCount),
		Size:           int64(getIntValue(ghRepo.Size)),
		DefaultBranch:  github.GetDefaultBranch(ghRepo),
		AnalyzedBranch: branch,
		Branches:       branches,
		Private:        isPrivate,
	}

	if exists {
		if resultCache, ok := cacheInstance.(*cache.Cache); ok {
			resultCache.Set(cacheKey, result)
		}
	}

	c.JSON(http.StatusOK, result)
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getIntValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}
