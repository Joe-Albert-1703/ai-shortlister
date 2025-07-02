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

func ConvertToPDF(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, ext)
	pdfPath := filepath.Join(convertedDir, name+".pdf")

	switch ext {
	case ".pdf":
		return path, nil
	case ".docx", ".txt", ".md":
		cmd := exec.Command("libreoffice", "--headless", "--convert-to", "pdf", "--outdir", convertedDir, path)
		return pdfPath, cmd.Run()
	case ".tex":
		// Step 1: Extract text
		plainText, err := extractPlainTextFromLatex(path)
		if err != nil {
			return "", fmt.Errorf("failed to extract text: %w", err)
		}

		// Step 2: Overwrite the original .tex file with plain text
		if err := os.WriteFile(path, []byte(plainText), 0644); err != nil {
			return "", fmt.Errorf("failed to overwrite .tex file with plain text: %w", err)
		}

		// Step 3: Convert .txt to .pdf using LibreOffice
		cmd := exec.Command("libreoffice", "--headless", "--convert-to", "pdf", "--outdir", convertedDir, path)
		return pdfPath, cmd.Run()
	default:
		return "", fmt.Errorf("unsupported format: %s", ext)
	}
}

func ExtractTextFromPDF(pdfPath string) (string, error) {
	cmd := exec.Command("pdftotext", pdfPath, "-")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error extracting text: %v", err)
	}
	return string(output), nil
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
