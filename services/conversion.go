package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const convertedDir = "./converted"

func ExtractTextFromFile(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		return ExtractTextFromPDF(filePath)
	case ".docx":
		return extractTextFromDocx(filePath)
	case ".txt", ".md":
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return string(content), nil
	case ".tex":
		return extractPlainTextFromLatex(filePath)
	default:
		return "", fmt.Errorf("unsupported file format: %s", ext)
	}
}

func ExtractTextFromPDF(pdfPath string) (string, error) {
	cmd := exec.Command("pdftotext", pdfPath, "-")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error extracting text from PDF: %v", err)
	}
	return string(output), nil
}

func extractTextFromDocx(docxPath string) (string, error) {
	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "txt", "--outdir", os.TempDir(), docxPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to convert docx to txt: %w", err)
	}

	txtFileName := strings.TrimSuffix(filepath.Base(docxPath), filepath.Ext(docxPath)) + ".txt"
	txtFilePath := filepath.Join(os.TempDir(), txtFileName)

	content, err := os.ReadFile(txtFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read extracted text from docx: %w", err)
	}

	// Clean up the temporary .txt file
	if err := os.Remove(txtFilePath); err != nil {
		fmt.Printf("Error deleting temporary .txt file %s: %v\n", txtFilePath, err)
	}

	return string(content), nil
}

func extractPlainTextFromLatex(texFilePath string) (string, error) {
	data, err := os.ReadFile(texFilePath)
	if err != nil {
		return "", err
	}

	// Basic LaTeX command removal
	re := regexp.MustCompile(`\\[a-zA-Z]+\*?(\[[^\]]*\])?(\{[^}]*\})?`)
	plain := re.ReplaceAllString(string(data), "")

	// Also remove remaining braces
	braces := regexp.MustCompile(`[{}]`)
	plain = braces.ReplaceAllString(plain, "")

	return plain, nil
}
