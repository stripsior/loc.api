package analyzer

import (
	"path/filepath"

	"github.com/go-enry/go-enry/v2"
)

func DetectLanguage(filePath string, content []byte) string {
	language := enry.GetLanguage(filepath.Base(filePath), content)

	if language == "" {
		return "Unknown"
	}

	return language
}

func IsVendored(filePath string) bool {
	return enry.IsVendor(filePath)
}

func IsGenerated(filePath string, content []byte) bool {
	return enry.IsGenerated(filePath, content)
}

func IsBinary(content []byte) bool {
	return enry.IsBinary(content)
}

func IsDocumentation(filePath string) bool {
	return enry.IsDocumentation(filePath)
}

func ShouldAnalyzeFile(filePath string, content []byte) bool {
	if IsBinary(content) {
		return false
	}

	return true
}
