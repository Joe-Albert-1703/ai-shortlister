package handlers

import (
	"encoding/json"
	"net/http"
	"sort"

	"ai-shortlister/models"
	"github.com/google/uuid"
)

func CreateJobPosting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	var job models.JobPosting
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job.ID = uuid.New().String()
	err = models.AddJobPosting(job)
	if err != nil {
		http.Error(w, "Failed to create job posting: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

func GetJobApplicants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	jobID := r.URL.Query().Get("jobID")
	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	job, ok := models.GetJobPosting(jobID)
	if !ok {
		http.Error(w, "Job posting not found", http.StatusNotFound)
		return
	}

	// Sort applicants by grade in decreasing order
	sort.Slice(job.Applicants, func(i, j int) bool {
		return job.Applicants[i].Grade > job.Applicants[j].Grade
	})

	json.NewEncoder(w).Encode(job.Applicants)
}


func GetAllJobPostingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	jobs, err := models.GetAllJobPostings()
	if err != nil {
		http.Error(w, "Failed to retrieve job postings: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(jobs)
}

func GetSingleJobPostingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	jobID := r.URL.Query().Get("jobID")
	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	job, ok := models.GetJobPosting(jobID)
	if !ok {
		http.Error(w, "Job posting not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(job)
}