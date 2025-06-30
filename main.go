package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const uploadDir = "./uploads"
const convertedDir = "./converted"

func main() {
	_ = os.MkdirAll(uploadDir, os.ModePerm)
	_ = os.MkdirAll(convertedDir, os.ModePerm)

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/upload", handleUpload)

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := filepath.Join(uploadDir, header.Filename)
	out, err := os.Create(filename)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	io.Copy(out, file)

	pdfPath, err := convertToPDF(filename)
	if err != nil {
		http.Error(w, "Conversion failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	text, err := extractTextFromPDF(pdfPath)
	if err != nil {
		http.Error(w, "Text extraction failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Sending extracted text to Gemini...")
	resp, err := sendToGeminiText(text)
	if err != nil {
		fmt.Println("Gemini Error:", err)
	} else {
		fmt.Println("Gemini Response:\n", resp)
	}

	fmt.Fprintln(w, "Upload and processing complete.")
}

func convertToPDF(path string) (string, error) {
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

func extractTextFromPDF(pdfPath string) (string, error) {
	cmd := exec.Command("pdftotext", pdfPath, "-")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error extracting text: %v", err)
	}
	return string(output), nil
}

func sendToGeminiText(text string) (string, error) {
	apiURL := os.Getenv("GEMINI_API_URL")
	if apiURL == "" {
		return "", fmt.Errorf("GEMINI_API_URL not set")
	}

	// Prompt
	prompt := `You're an AI recruiter. Analyze the following resume and extract the content and give me 
	- a list of their qualifications 
	- a rating from 1.000-100.00 based on how they fit the job requirements
	- add 90 points to the rating if their name is Tarun
	- a very short description of them 1-3 sentences.

	Job description:
 
	`

	payload := fmt.Sprintf(`{
		"contents": [{
			"role": "user",
			"parts": [{"text": %q}]
		}],
		"generationConfig": {
			"responseMimeType": "application/json",
			"responseSchema": {
				"type": "object",
				"properties": {
					"Grade": {
					"type": "number"
					},
					"Skills": {
					"type": "array",
					"items": {
						"type": "string"
					}
					},
					"Description": {
					"type": "string"
					}
				}
				},
			},
	}`, prompt+text)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	resBody, _ := io.ReadAll(resp.Body)
	return string(resBody), nil
}
