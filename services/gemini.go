package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// GeminiResponse represents the structure of the response from the Gemini API
type GeminiOuterResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
		Index        int    `json:"index"`
	} `json:"candidates"`
}

type GeminiResponse struct {
	Name        string   `json:"Name"`
	Grade       float64  `json:"Grade"`
	Skills      []string `json:"Skills"`
	Description string   `json:"Description"`
	Email       string   `json:"Email"`
	Phone       string   `json:"Phone"`
}

func SendToGeminiText(text string, jobDescription string) (*GeminiResponse, error) {
	apiURL := os.Getenv("GEMINI_API_URL")
	if apiURL == "" {
		return nil, fmt.Errorf("GEMINI_API_URL not set")
	}

	// Prompt
	prompt := fmt.Sprintf(`You're an AI recruiter. Analyze the following resume and extract the content and give me
	- their full name
	- a list of their qualifications
	- a rating from 1.000-100.00 based on how they fit the job requirements
	- a very short description of them 1-3 sentences.
	- extract their email address
	- extract their phone number
	
		Job description:
		%s
		`, jobDescription)
	
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
						"Name": {
						"type": "string"
						},
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
						},
					"Email": {
					"type": "string"
					},
					"Phone": {
					"type": "string"
					}
					}
					},
				},
		}`, prompt+text)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var outer GeminiOuterResponse
	if err := json.Unmarshal(resBody, &outer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal outer Gemini response: %w", err)
	}

	// Safety checks
	if len(outer.Candidates) == 0 || len(outer.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("unexpected Gemini response format")
	}

	// Step 2: Extract the inner JSON string and unmarshal
	rawJSON := outer.Candidates[0].Content.Parts[0].Text

	var geminiData GeminiResponse
	if err := json.Unmarshal([]byte(rawJSON), &geminiData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gemini inner JSON: %w", err)
	}

	return &geminiData, nil
}
