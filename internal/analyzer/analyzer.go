package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stripsior/loc.api/internal/models"
)

type Analyzer struct {
	repoPath string
	filters  *models.FilterOptions
}

func NewAnalyzer(repoPath string, filters *models.FilterOptions) *Analyzer {
	return &Analyzer{
		repoPath: repoPath,
		filters:  filters,
	}
}

func (a *Analyzer) Analyze() (*models.AnalyzeResponse, error) {
	startTime := time.Now()

	languageStats := make(map[string]*models.LanguageStats)
	var files []models.FileInfo
	var summary models.SummaryStats

	var mu sync.Mutex
	var wg sync.WaitGroup

	fileChan := make(chan string, 100)
	errorChan := make(chan error, 1)

	numWorkers := 10
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				fileInfo, err := a.analyzeFile(filePath)
				if err != nil {
					select {
					case errorChan <- err:
					default:
					}
					continue
				}

				if fileInfo == nil {
					continue
				}

				mu.Lock()
				files = append(files, *fileInfo)

				if _, exists := languageStats[fileInfo.Language]; !exists {
					languageStats[fileInfo.Language] = &models.LanguageStats{
						Language: fileInfo.Language,
					}
				}
				ls := languageStats[fileInfo.Language]
				ls.FileCount++
				ls.TotalLines += fileInfo.TotalLines
				ls.CodeLines += fileInfo.CodeLines
				ls.CommentLines += fileInfo.CommentLines
				ls.BlankLines += fileInfo.BlankLines
				ls.Bytes += fileInfo.Size

				summary.TotalFiles++
				summary.TotalLines += fileInfo.TotalLines
				summary.TotalCodeLines += fileInfo.CodeLines
				summary.TotalCommentLines += fileInfo.CommentLines
				summary.TotalBlankLines += fileInfo.BlankLines
				summary.TotalBytes += fileInfo.Size
				mu.Unlock()
			}
		}()
	}

	err := filepath.Walk(a.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if a.shouldExcludeDirectory(path) {
				return filepath.SkipDir
			}
			return nil
		}

		fileChan <- path
		return nil
	})

	close(fileChan)
	wg.Wait()
	close(errorChan)

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	select {
	case err := <-errorChan:
		if err != nil {
			return nil, err
		}
	default:
	}

	if summary.TotalLines > 0 {
		for _, ls := range languageStats {
			ls.Percentage = float64(ls.TotalLines) / float64(summary.TotalLines) * 100
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].TotalLines > files[j].TotalLines
	})

	langStatsMap := make(map[string]models.LanguageStats)
	for lang, stats := range languageStats {
		langStatsMap[lang] = *stats
	}

	processingTime := time.Since(startTime)

	return &models.AnalyzeResponse{
		Summary:        summary,
		Languages:      langStatsMap,
		Files:          files,
		ProcessingTime: processingTime.String(),
	}, nil
}

func (a *Analyzer) analyzeFile(filePath string) (*models.FileInfo, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	if !a.shouldAnalyzeFile(filePath, content, fileInfo.Size()) {
		return nil, nil
	}

	language := DetectLanguage(filePath, content)

	stats, err := CountLines(filePath, language)
	if err != nil {
		return nil, fmt.Errorf("failed to count lines in %s: %w", filePath, err)
	}

	relPath, err := filepath.Rel(a.repoPath, filePath)
	if err != nil {
		relPath = filePath
	}

	return &models.FileInfo{
		Path:         relPath,
		Language:     language,
		TotalLines:   stats.Total,
		CodeLines:    stats.Code,
		CommentLines: stats.Comments,
		BlankLines:   stats.Blank,
		Size:         fileInfo.Size(),
	}, nil
}

func (a *Analyzer) shouldAnalyzeFile(filePath string, content []byte, size int64) bool {
	if IsBinary(content) {
		return false
	}

	excludeGenerated := true
	excludeVendored := true

	if a.filters != nil {
		if a.filters.ExcludeGenerated != nil {
			excludeGenerated = *a.filters.ExcludeGenerated
		}
		if a.filters.ExcludeVendored != nil {
			excludeVendored = *a.filters.ExcludeVendored
		}
	}

	if excludeVendored && IsVendored(filePath) {
		return false
	}

	if excludeGenerated && IsGenerated(filePath, content) {
		return false
	}

	if a.filters == nil {
		return true
	}

	if a.filters.MinFileSize != nil && size < *a.filters.MinFileSize {
		return false
	}
	if a.filters.MaxFileSize != nil && size > *a.filters.MaxFileSize {
		return false
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	if len(a.filters.IncludeExtensions) > 0 {
		included := false
		for _, includeExt := range a.filters.IncludeExtensions {
			if strings.ToLower(includeExt) == ext {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	if len(a.filters.ExcludeExtensions) > 0 {
		for _, excludeExt := range a.filters.ExcludeExtensions {
			if strings.ToLower(excludeExt) == ext {
				return false
			}
		}
	}

	return true
}

func (a *Analyzer) shouldExcludeDirectory(dirPath string) bool {
	if strings.Contains(dirPath, ".git") {
		return true
	}

	if a.filters == nil || len(a.filters.ExcludeDirectories) == 0 {
		return false
	}

	dirName := filepath.Base(dirPath)

	for _, excludePattern := range a.filters.ExcludeDirectories {
		if dirName == excludePattern || strings.Contains(dirPath, excludePattern) {
			return true
		}
	}

	return false
}
