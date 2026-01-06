package analyzer

import (
	"bufio"
	"bytes"
	"os"
	"strings"
)

type CommentSyntax struct {
	SingleLine     []string
	MultiLineStart []string
	MultiLineEnd   []string
}

var commentSyntaxMap = map[string]CommentSyntax{
	"Go":         {SingleLine: []string{"//"}, MultiLineStart: []string{"/*"}, MultiLineEnd: []string{"*/"}},
	"JavaScript": {SingleLine: []string{"//"}, MultiLineStart: []string{"/*"}, MultiLineEnd: []string{"*/"}},
	"TypeScript": {SingleLine: []string{"//"}, MultiLineStart: []string{"/*"}, MultiLineEnd: []string{"*/"}},
	"Java":       {SingleLine: []string{"//"}, MultiLineStart: []string{"/*"}, MultiLineEnd: []string{"*/"}},
	"C":          {SingleLine: []string{"//"}, MultiLineStart: []string{"/*"}, MultiLineEnd: []string{"*/"}},
	"C++":        {SingleLine: []string{"//"}, MultiLineStart: []string{"/*"}, MultiLineEnd: []string{"*/"}},
	"C#":         {SingleLine: []string{"//"}, MultiLineStart: []string{"/*"}, MultiLineEnd: []string{"*/"}},
	"Rust":       {SingleLine: []string{"//"}, MultiLineStart: []string{"/*"}, MultiLineEnd: []string{"*/"}},
	"Swift":      {SingleLine: []string{"//"}, MultiLineStart: []string{"/*"}, MultiLineEnd: []string{"*/"}},
	"Kotlin":     {SingleLine: []string{"//"}, MultiLineStart: []string{"/*"}, MultiLineEnd: []string{"*/"}},
	"Python":     {SingleLine: []string{"#"}, MultiLineStart: []string{`"""`}, MultiLineEnd: []string{`"""`}},
	"Ruby":       {SingleLine: []string{"#"}, MultiLineStart: []string{"=begin"}, MultiLineEnd: []string{"=end"}},
	"PHP":        {SingleLine: []string{"//", "#"}, MultiLineStart: []string{"/*"}, MultiLineEnd: []string{"*/"}},
	"Shell":      {SingleLine: []string{"#"}, MultiLineStart: []string{}, MultiLineEnd: []string{}},
	"Bash":       {SingleLine: []string{"#"}, MultiLineStart: []string{}, MultiLineEnd: []string{}},
	"HTML":       {SingleLine: []string{}, MultiLineStart: []string{"<!--"}, MultiLineEnd: []string{"-->"}},
	"CSS":        {SingleLine: []string{}, MultiLineStart: []string{"/*"}, MultiLineEnd: []string{"*/"}},
	"SQL":        {SingleLine: []string{"--"}, MultiLineStart: []string{"/*"}, MultiLineEnd: []string{"*/"}},
	"Lua":        {SingleLine: []string{"--"}, MultiLineStart: []string{"--[["}, MultiLineEnd: []string{"]]"}},
	"R":          {SingleLine: []string{"#"}, MultiLineStart: []string{}, MultiLineEnd: []string{}},
	"YAML":       {SingleLine: []string{"#"}, MultiLineStart: []string{}, MultiLineEnd: []string{}},
	"TOML":       {SingleLine: []string{"#"}, MultiLineStart: []string{}, MultiLineEnd: []string{}},
	"Dockerfile": {SingleLine: []string{"#"}, MultiLineStart: []string{}, MultiLineEnd: []string{}},
}

type LineStats struct {
	Total    int64
	Code     int64
	Comments int64
	Blank    int64
}

func CountLines(filePath string, language string) (*LineStats, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats := &LineStats{}
	syntax := commentSyntaxMap[language]

	scanner := bufio.NewScanner(file)
	// Increase buffer size to handle files with very long lines (e.g., minified SVG files)
	// Default is 64KB, we set it to 10MB to handle edge cases
	const maxScanTokenSize = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	inMultiLineComment := false
	currentMultiLineEnd := ""

	for scanner.Scan() {
		stats.Total++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			stats.Blank++
			continue
		}

		isComment := false

		if inMultiLineComment {
			isComment = true
			if strings.Contains(line, currentMultiLineEnd) {
				inMultiLineComment = false
				currentMultiLineEnd = ""
			}
		} else {
			for i, start := range syntax.MultiLineStart {
				if strings.Contains(trimmedLine, start) {
					isComment = true
					inMultiLineComment = true
					if i < len(syntax.MultiLineEnd) {
						currentMultiLineEnd = syntax.MultiLineEnd[i]
					}
					if currentMultiLineEnd != "" && strings.Contains(line, currentMultiLineEnd) {
						inMultiLineComment = false
						currentMultiLineEnd = ""
					}
					break
				}
			}

			if !isComment {
				for _, prefix := range syntax.SingleLine {
					if strings.HasPrefix(trimmedLine, prefix) {
						isComment = true
						break
					}
				}
			}
		}

		if isComment {
			stats.Comments++
		} else {
			stats.Code++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

func IsBlankFile(filePath string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	return len(bytes.TrimSpace(content)) == 0, nil
}
