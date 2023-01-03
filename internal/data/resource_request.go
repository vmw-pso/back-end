package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/vmw-pso/back-end/internal/validator"
)

type ResourceRequest struct {
	ID            int64                     `json:"id"`
	OpportunityID string                    `json:"opportunityId"`
	JobTitle      string                    `json:"jobTitle"`
	TotalHours    float64                   `json:"totalHours"`
	Skills        []string                  `json:"skills"`
	StartDate     time.Time                 `json:"startDate"`
	HoursPerWeek  float64                   `json:"hoursPerWeek"`
	Status        string                    `json:"status"`
	CreatedAt     time.Time                 `json:"createdAt"`
	UpdatedAt     time.Time                 `json:"updatedAt"`
	Version       int64                     `json:"version"`
	Comments      []*ResourceRequestComment `json:"comments,omitempty"`
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

func (m *ResourceRequestModel) Insert(r *ResourceRequest) error {
	query := `
		INSERT INTO resource_request
		(opportunity_id, job_title_id, total_hours, skills, start_date, hours_per_week, status, created_at, updated_at)
		VALUES ($1,
			   (SELECT title_id FROM job_title WHERE title=$2),
			   $3, $4, $5, $6, $7, $8, $9) RETURNING request_id, version`

	args := []interface{}{
		r.OpportunityID,
		r.JobTitle,
		r.TotalHours,
		pq.Array(r.Skills),
		r.StartDate,
		r.HoursPerWeek,
		r.Status,
		r.CreatedAt,
		r.UpdatedAt,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&r.ID, &r.Version)
}

func (m *ResourceRequestModel) Get(id int64) (*ResourceRequest, error) {
	if id < 1 {
		return nil, ErrNotFound
	}

	query := `
		SELECT r.opportunity_id, j.title, r.total_hours, r.skills, r.start_date, r.hours_per_week, r.status, r.created_at, r.updated_at, r.version
		FROM (resource_request r
			INNER JOIN job_title ON r.job_title_id=job_title.title_id)
		WHERE r.request_id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var r ResourceRequest
	r.ID = id

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&r.OpportunityID,
		&r.JobTitle,
		&r.TotalHours,
		pq.Array(&r.Skills),
		&r.StartDate,
		&r.HoursPerWeek,
		&r.Status,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &r, nil
}

func (m *ResourceRequestModel) Update(r *ResourceRequest) error {
	query := `
		UPDATE resource_request
		SET opportunity_id=$1, job_title_id=(SELECT title_id FROM job_title WHERE title=$2),
		    total_hours=$3, skills=$4, start_date=$5, hours_per_week=$6, status=$7, updated_at=$8, version=$9
		WHERE request_id=$10 
		RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []interface{}{
		r.OpportunityID,
		r.JobTitle,
		r.TotalHours,
		pq.Array(r.Skills),
		r.StartDate,
		r.HoursPerWeek,
		r.Status,
		r.UpdatedAt,
		r.Version + 1,
		r.ID,
	}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&r.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m *ResourceRequestModel) GetForOpportunity(oppId string) ([]*ResourceRequest, error) {
	query := `
		SELECT r.request_id, j.title, r.total_hours, r.skills, r.start_date, r.hours_per_week, r.status, r.created_at, r.updated_at, r.version
		FROM(resource_request r
			INNER JOIN job_title j ON r.job_title_id=j.title_id)
		WHERE r.opportunity_id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, oppId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []*ResourceRequest{}

	for rows.Next() {
		var request ResourceRequest
		request.OpportunityID = oppId
		err := rows.Scan(
			&request.ID,
			&request.JobTitle,
			&request.TotalHours,
			pq.Array(&request.Skills),
			&request.StartDate,
			&request.HoursPerWeek,
			&request.Status,
			&request.CreatedAt,
			&request.UpdatedAt,
			&request.Version,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &request)
	}

	return requests, nil
}
