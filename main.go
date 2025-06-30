package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"ai-shortlister/database"
	"ai-shortlister/handlers"
)

const uploadDir = "./uploads"
const convertedDir = "./converted"

func main() {
	_ = os.MkdirAll(uploadDir, os.ModePerm)
	_ = os.MkdirAll(convertedDir, os.ModePerm)

	database.InitDB()

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/upload", handlers.HandleUpload)
	http.HandleFunc("/jobs", handlers.CreateJobPosting)
	http.HandleFunc("/jobs/applicants", handlers.GetJobApplicants)
	http.HandleFunc("/jobs/all", handlers.GetAllJobPostingsHandler)
	http.HandleFunc("/jobs/single", handlers.GetSingleJobPostingHandler)
	http.HandleFunc("/applicant/job/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/applicant_job_upload.html")
	})
	http.HandleFunc("/hr/job/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/hr_job_applicants.html")
	})

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
