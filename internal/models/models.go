package models

type AnalyzeRequest struct {
	Repository     string         `json:"repository" binding:"required"`
	Token          string         `json:"token,omitempty"`
	Branch         string         `json:"branch,omitempty"`
	IncludeAuthors bool           `json:"includeAuthors,omitempty"`
	Filters        *FilterOptions `json:"filters,omitempty"`
}

type FilterOptions struct {
	ExcludeExtensions  []string `json:"excludeExtensions,omitempty"`
	IncludeExtensions  []string `json:"includeExtensions,omitempty"`
	ExcludeDirectories []string `json:"excludeDirectories,omitempty"`
	MinFileSize        *int64   `json:"minFileSize,omitempty"`
	MaxFileSize        *int64   `json:"maxFileSize,omitempty"`
	ExcludeGenerated   *bool    `json:"excludeGenerated,omitempty"`
	ExcludeVendored    *bool    `json:"excludeVendored,omitempty"`
}

type AnalyzeResponse struct {
	Repository     RepositoryInfo           `json:"repository"`
	Summary        SummaryStats             `json:"summary"`
	Languages      map[string]LanguageStats `json:"languages"`
	Authors        map[string]AuthorStats   `json:"authors,omitempty"`
	Files          []FileInfo               `json:"files"`
	ProcessingTime string                   `json:"processingTime"`
}

type RepositoryInfo struct {
	Owner          string   `json:"owner"`
	Name           string   `json:"name"`
	FullName       string   `json:"fullName"`
	Description    string   `json:"description,omitempty"`
	Stars          int      `json:"stars"`
	Forks          int      `json:"forks"`
	Size           int64    `json:"size"`
	DefaultBranch  string   `json:"defaultBranch"`
	AnalyzedBranch string   `json:"analyzedBranch"`
	Branches       []string `json:"branches"`
	Private        bool     `json:"private"`
}

type SummaryStats struct {
	TotalFiles        int   `json:"totalFiles"`
	TotalLines        int64 `json:"totalLines"`
	TotalCodeLines    int64 `json:"totalCodeLines"`
	TotalCommentLines int64 `json:"totalCommentLines"`
	TotalBlankLines   int64 `json:"totalBlankLines"`
	TotalBytes        int64 `json:"totalBytes"`
}

type LanguageStats struct {
	Language     string  `json:"language"`
	FileCount    int     `json:"fileCount"`
	TotalLines   int64   `json:"totalLines"`
	CodeLines    int64   `json:"codeLines"`
	CommentLines int64   `json:"commentLines"`
	BlankLines   int64   `json:"blankLines"`
	Bytes        int64   `json:"bytes"`
	Percentage   float64 `json:"percentage"`
}

type AuthorStats struct {
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	AvatarURL    string  `json:"avatarUrl,omitempty"`
	TotalLines   int64   `json:"totalLines"`
	CodeLines    int64   `json:"codeLines"`
	CommentLines int64   `json:"commentLines"`
	BlankLines   int64   `json:"blankLines"`
	FileCount    int     `json:"fileCount"`
	Percentage   float64 `json:"percentage"`
}

type FileInfo struct {
	Path         string `json:"path"`
	Language     string `json:"language"`
	TotalLines   int64  `json:"totalLines"`
	CodeLines    int64  `json:"codeLines"`
	CommentLines int64  `json:"commentLines"`
	BlankLines   int64  `json:"blankLines"`
	Size         int64  `json:"size"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
