package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"ai-shortlister/models"
	"ai-shortlister/services"
	"strings"
)

const uploadDir = "./uploads"

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	jobID := r.URL.Query().Get("jobID")
	if jobID == "" {
		http.Error(w, "Job ID is required for resume upload", http.StatusBadRequest)
		return
	}

	job, ok := models.GetJobPosting(jobID)
	if !ok {
		http.Error(w, "Job posting not found", http.StatusNotFound)
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

	pdfPath, err := services.ConvertToPDF(filename)
	if err != nil {
		http.Error(w, "Conversion failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	text, err := services.ExtractTextFromPDF(pdfPath)
	if err != nil {
		http.Error(w, "Text extraction failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Sending extracted text to Gemini...")
	geminiResponse, err := services.SendToGeminiText(text, job.Description)
	if err != nil {
		fmt.Println("Gemini Error:", err)
		http.Error(w, "Gemini processing failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Recieved grade from gemini")
	applicant := models.Applicant{
		Grade:       geminiResponse.Grade,
		Description: geminiResponse.Description,
		Skills:      geminiResponse.Skills,
	}

	applicant.Name = strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)) // Assuming filename is applicant name
	applicant.JobID = jobID
	err = models.AddApplicantToJob(applicant)
	if err != nil {
		http.Error(w, "Failed to save applicant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Clean up: Delete the original uploaded file and the converted PDF
	if err := os.Remove(filename); err != nil {
		fmt.Printf("Error deleting original file %s: %v\n", filename, err)
	}
	if err := os.Remove(pdfPath); err != nil {
		fmt.Printf("Error deleting converted PDF %s: %v\n", pdfPath, err)
	}

	fmt.Fprintln(w, "Upload and processing complete. Applicant added to job posting.")
}