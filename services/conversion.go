package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// TextExtractor defines the interface for extracting text from files.
type TextExtractor interface {
	ExtractText(filePath string) (string, error)
}

// Extractor implementations for different formats

type PDFExtractor struct{}

func (e PDFExtractor) ExtractText(filePath string) (string, error) {
	cmd := exec.Command("pdftotext", filePath, "-")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error extracting text from PDF: %v", err)
	}
	return string(output), nil
}

type DocxExtractor struct{}

func (e DocxExtractor) ExtractText(filePath string) (string, error) {
	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "txt", "--outdir", os.TempDir(), filePath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to convert docx to txt: %w", err)
	}

	txtFileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)) + ".txt"
	txtFilePath := filepath.Join(os.TempDir(), txtFileName)

	content, err := os.ReadFile(txtFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read extracted text from docx: %w", err)
	}

	_ = os.Remove(txtFilePath)
	return string(content), nil
}

type PlainTextExtractor struct{}

func (e PlainTextExtractor) ExtractText(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

type LatexExtractor struct{}

func (e LatexExtractor) ExtractText(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Remove basic LaTeX commands
	re := regexp.MustCompile(`\\[a-zA-Z]+\*?(\[[^\]]*\])?(\{[^}]*\})?`)
	plain := re.ReplaceAllString(string(data), "")

	// Remove remaining braces
	braces := regexp.MustCompile(`[{}]`)
	plain = braces.ReplaceAllString(plain, "")

	return plain, nil
}

// Registry of file extension to extractor
var extractors = map[string]TextExtractor{
	".pdf":  PDFExtractor{},
	".docx": DocxExtractor{},
	".txt":  PlainTextExtractor{},
	".md":   PlainTextExtractor{},
	".tex":  LatexExtractor{},
}

// ExtractTextFromFile dispatches to the appropriate extractor based on file extension.
func ExtractTextFromFile(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	extractor, ok := extractors[ext]
	if !ok {
		return "", fmt.Errorf("unsupported file format: %s", ext)
	}
	return extractor.ExtractText(filePath)
}
