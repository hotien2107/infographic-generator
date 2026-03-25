package extraction

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"infographic-generator/backend/internal/modules/documents"
)

type Result struct {
	RawText  string
	Metadata documents.RawContentMetadata
}

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) ExtractFromText(input string) (Result, error) {
	normalized := normalizeText(input)
	if normalized == "" {
		return Result{}, fmt.Errorf("input text is empty after normalization")
	}
	metadata := buildMetadata(documents.SourceTypeText, documents.FileTypeText, normalized, 1)
	return Result{RawText: normalized, Metadata: metadata}, nil
}

func (s *Service) ExtractFromFile(fileType documents.FileType, payload []byte) (Result, error) {
	switch fileType {
	case documents.FileTypeTXT:
		text := normalizeText(string(payload))
		if text == "" {
			return Result{}, fmt.Errorf("txt file is empty after normalization")
		}
		metadata := buildMetadata(documents.SourceTypeFile, documents.FileTypeTXT, text, 1)
		return Result{RawText: text, Metadata: metadata}, nil
	case documents.FileTypePDF:
		text, pageCount, err := extractPDFText(payload)
		if err != nil {
			return Result{}, err
		}
		normalized := normalizeText(text)
		if normalized == "" {
			return Result{}, fmt.Errorf("cannot extract readable text from pdf")
		}
		metadata := buildMetadata(documents.SourceTypeFile, documents.FileTypePDF, normalized, pageCount)
		return Result{RawText: normalized, Metadata: metadata}, nil
	default:
		return Result{}, fmt.Errorf("unsupported file type: %s", fileType)
	}
}

func extractPDFText(payload []byte) (string, int, error) {
	raw := string(payload)
	if !strings.Contains(raw, "%PDF") {
		return "", 0, fmt.Errorf("invalid pdf header")
	}
	pageCount := strings.Count(raw, "/Type /Page")
	if pageCount == 0 {
		pageCount = 1
	}
	// Basic extractor: lấy text nằm trong cặp ngoặc đơn của stream PDF.
	matches := regexp.MustCompile(`\(([^\)\(]{2,})\)`).FindAllStringSubmatch(raw, -1)
	parts := make([]string, 0, len(matches))
	for _, m := range matches {
		candidate := strings.TrimSpace(strings.ReplaceAll(m[1], `\\`, `\`))
		if candidate != "" {
			parts = append(parts, candidate)
		}
	}
	if len(parts) == 0 {
		return "", pageCount, fmt.Errorf("no textual token found in pdf")
	}
	return strings.Join(parts, "\n"), pageCount, nil
}

func buildMetadata(sourceType documents.SourceType, fileType documents.FileType, rawText string, pageCount int) documents.RawContentMetadata {
	return documents.RawContentMetadata{FileType: fileType, SourceType: sourceType, PageCount: max(1, pageCount), SectionHeadings: inferSectionHeadings(rawText), CharacterCount: utf8.RuneCountInString(rawText)}
}

func inferSectionHeadings(rawText string) []string {
	lines := strings.Split(rawText, "\n")
	headings := make([]string, 0, 8)
	seen := map[string]bool{}
	headingPattern := regexp.MustCompile(`^[A-ZÀ-Ỹ0-9][A-ZÀ-Ỹ0-9\s\-:]{3,80}$`)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || len(trimmed) > 80 || !headingPattern.MatchString(trimmed) || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		headings = append(headings, trimmed)
		if len(headings) >= 10 {
			break
		}
	}
	return headings
}

func normalizeText(input string) string {
	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	normalizedLines := make([]string, 0, len(lines))
	lastEmpty := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if lastEmpty {
				continue
			}
			lastEmpty = true
			normalizedLines = append(normalizedLines, "")
			continue
		}
		lastEmpty = false
		normalizedLines = append(normalizedLines, trimmed)
	}
	return strings.TrimSpace(strings.Join(normalizedLines, "\n"))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
