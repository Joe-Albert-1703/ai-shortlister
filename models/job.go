package models

import (
	"ai-shortlister/database"
	"database/sql"
	"log"
	"strings"
)

type JobPosting struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Applicants  []Applicant `json:"applicants"`
}

type Applicant struct {
	ID          int      `json:"id"`
	JobID       string   `json:"job_id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	Grade       float64  `json:"grade"`
	Skills      []string `json:"skills"`
	Description string   `json:"description"`
}

func AddJobPosting(job JobPosting) error {
	_, err := database.DB.Exec("INSERT INTO job_postings (id, title, description) VALUES ($1, $2, $3)",
		job.ID, job.Title, job.Description)
	if err != nil {
		log.Printf("Error adding job posting to DB: %v", err)
		return err
	}
	return nil
}

func GetJobPosting(id string) (JobPosting, bool) {
	var job JobPosting
	err := database.DB.QueryRow("SELECT id, title, description FROM job_postings WHERE id = $1", id).
		Scan(&job.ID, &job.Title, &job.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return JobPosting{}, false
		}
		log.Printf("Error getting job posting from DB: %v", err)
		return JobPosting{}, false
	}

	rows, err := database.DB.Query("SELECT id, job_id, name, grade, skills, description , email, phone FROM applicants WHERE job_id = $1", id)
	if err != nil {
		log.Printf("Error getting applicants for job %s from DB: %v", id, err)
		return job, false // Return job even if applicants can't be fetched
	}
	defer rows.Close()

	for rows.Next() {
		var applicant Applicant
		var skillsStr string // To scan TEXT[] as string
		err := rows.Scan(&applicant.ID, &applicant.JobID, &applicant.Name, &applicant.Grade, &skillsStr, &applicant.Description, &applicant.Email, &applicant.Phone)
		if err != nil {
			log.Printf("Error scanning applicant row: %v", err)
			continue
		}
		// Convert skills string representation to []string
		applicant.Skills = parsePostgresArray(skillsStr)
		job.Applicants = append(job.Applicants, applicant)
	}
	return job, true
}

func ApplicantExists(jobID, email, phone string) (bool, error) {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM applicants WHERE job_id = $1 AND email = $2 AND phone = $3", jobID, email, phone).Scan(&count)
	if err != nil {
		log.Printf("Error checking if applicant exists: %v", err)
		return false, err
	}
	return count > 0, nil
}

func AddApplicantToJob(applicant Applicant) error {
	_, err := database.DB.Exec("INSERT INTO applicants (job_id, name, email, phone, grade, skills, description) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		applicant.JobID, applicant.Name, applicant.Email, applicant.Phone, applicant.Grade, "{"+joinStrings(applicant.Skills)+"}", applicant.Description)
	if err != nil {
		log.Printf("Error adding applicant to DB: %v", err)
		return err
	}
	return nil
}

func GetAllJobPostings() ([]JobPosting, error) {
	rows, err := database.DB.Query("SELECT id, title, description FROM job_postings")
	if err != nil {
		log.Printf("Error getting all job postings from DB: %v", err)
		return nil, err
	}
	defer rows.Close()

	var jobs []JobPosting
	for rows.Next() {
		var job JobPosting
		err := rows.Scan(&job.ID, &job.Title, &job.Description)
		if err != nil {
			log.Printf("Error scanning job posting row: %v", err)
			continue
		}
		// Fetch applicants for each job
		jobWithApplicants, ok := GetJobPosting(job.ID)
		if ok {
			job.Applicants = jobWithApplicants.Applicants
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// Helper to parse PostgreSQL array string representation
func parsePostgresArray(arrayStr string) []string {
	if len(arrayStr) < 2 || arrayStr[0] != '{' || arrayStr[len(arrayStr)-1] != '}' {
		return []string{}
	}
	// Remove curly braces and split by comma
	elements := strings.Split(arrayStr[1:len(arrayStr)-1], ",")
	for i, elem := range elements {
		elements[i] = strings.TrimSpace(elem)
	}
	return elements
}

// Helper to join strings for PostgreSQL array format
func joinStrings(s []string) string {
	var b strings.Builder
	for i, str := range s {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(str)
	}
	return b.String()
}
