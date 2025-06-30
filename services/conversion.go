package services

import (
	"fmt"
	"os/exec"
	"path/filepath"
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
		cmd := exec.Command("pdflatex", "-output-directory", convertedDir, path)
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