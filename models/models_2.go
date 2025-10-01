package models

// This also exists

type SubjectResponse struct {
	Subject   string `json:"subject"`
	SubjectID int    `json:"subject_id"`
	Types     []int  `json:"types"`
}
