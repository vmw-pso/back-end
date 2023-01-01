package data

import (
	"database/sql"
	"time"

	"github.com/vmw-pso/back-end/internal/validator"
)

type ResourceRequest struct {
	ID           int64     `json:"id"`
	ProjectID    int64     `json:"projectId"`
	JobTitle     string    `json:"jobTitle"`
	Skills       []string  `json:"skills"`
	TotalHours   float64   `json:"totalHours"`
	StartDate    time.Time `json:"startDate"`
	HoursPerWeek float64   `json:"hoursPerWeek"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Version      int64     `json:"version"`
}

func ValidateSkills(v *validator.Validator, skills []string) {
	v.Check(len(skills) > 0, "skills", "at least one skill is required")
	v.Check(validator.Unique(skills), "skills", "cannot contain duplicate values")
}

func ValidateStartDate(v *validator.Validator, startDate time.Time) {
	v.Check(startDate.After(time.Now()), "startDate", "cannot be today or in the past")
}

func ValidateResourceRequestStatus(v *validator.Validator, status string) {
	v.Check(status != "", "status", "must be provided")
	statuses := []string{
		"Open",
		"Closed",
	}
	v.Check(validator.PermittedValue(status, statuses...), "status", "is not a recognised status [Open, Closed]")
}

func ValidateResourceRequest(v *validator.Validator, rr ResourceRequest) {
	ValidateSkills(v, rr.Skills)
	ValidateStartDate(v, rr.StartDate)
	ValidateResourceRequestStatus(v, rr.Status)
}

type ResourceRequestModel struct {
	DB *sql.DB
}
