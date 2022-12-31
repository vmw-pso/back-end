package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/vmw-pso/back-end/internal/validator"
)

type Resource struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	JobTitle       string   `json:"jobTitle"`
	Manager        string   `json:"manager"`
	Workgroup      string   `json:"workgroup"`
	Clearance      string   `json:"clearance"`
	Specialties    []string `json:"specialties"`
	Certifications []string `json:"certifications"`
	Active         bool     `json:"active"`
}

func ValidateID(v *validator.Validator, id int64) {
	v.Check(id != 0, "id", "must be provided")
	v.Check(id > 0, "id", "cannot be a negative number")
}

func ValidateName(v *validator.Validator, name string) {
	v.Check(name != "", "firstName", "must be provided")
	v.Check(len(name) < 256, "firstName", "cannot be more than 256 bytes")
}

func ValidateJobTitle(v *validator.Validator, jobTitle string) {
	jobTitles := []string{
		"Associate Project Manager I",
		"Associate Consultant I",
		"Associate Project Manager II",
		"Associate Consultant II",
		"Project Manager",
		"Consultant",
		"Senior Project Manager",
		"Senior Consultant",
		"Staff Consultant",
		"Consulting Architect",
		"Staff Consulting Architect",
		"Manager - Professional Services - Delivery",
		"Senior Manager - Professional Services - Delivery",
		"Director - Professional Services - Delivery",
	}
	v.Check(validator.PermittedValue(jobTitle, jobTitles...), "jobTitle", "does not exist")
}

func ValidateManager(v *validator.Validator, manager string) {
	managers := []string{
		"Caroline Dimitrovski",
		"Gary Doyle",
		"Lisa Ryan",
		"Peter Stacey",
		"Deborah Brathwaite",
	}
	v.Check(validator.PermittedValue(manager, managers...), "manager", "is not a manager")
}

func ValidateWorkgroup(v *validator.Validator, workgroup string) {
	workgroups := []string{
		"APJ - Managers and Non-Billable",
		"Architects - ANZ",
		"PMO - ANZ",
		"Retainer - ANZ",
		"Server - Australia",
	}
	v.Check(validator.PermittedValue(workgroup, workgroups...), "workgroup", "does not exist")
}

func ValidatorClearance(v *validator.Validator, clearance string) {
	clearances := []string{
		"None",
		"Baseline",
		"NV1",
		"NV2",
		"TSPV",
	}
	v.Check(validator.PermittedValue(clearance, clearances...), "clearance", "must be one of ['None', 'Baseline', 'NV1', 'NV2', 'TSPV']")
}

func ValidateResource(v *validator.Validator, r Resource) {
	ValidateID(v, r.ID)
	ValidateName(v, r.Name)
	ValidateJobTitle(v, r.JobTitle)
	ValidateManager(v, r.Manager)
	ValidateWorkgroup(v, r.Workgroup)
	ValidatorClearance(v, r.Clearance)
	v.Check(validator.Unique(r.Specialties), "specialties", "cannot contain duplicate values")
	v.Check(validator.Unique(r.Certifications), "certifications", "cannot contain duplicate values")
}

type ResourceModel struct {
	DB *sql.DB
}

func (m *ResourceModel) Insert(r *Resource) error {
	query := `
		INSERT INTO resources
		(id, name, email, job_title, manager_id, workgroup_id, clearance, specialties, certifications, active)
		VALUES ($1, $2, $3, 
			   (SELECT title FROM job_titles WHERE title=$4),
			   (SELECT m.id FROM resources m WHERE m.name=$5),
			   (SELECT id FROM workgroups WHERE name=$6),
		       (SELECT level FROM clearances WHERE level=$7), 
			   $8, $9, $10) RETURNING active`

	args := []interface{}{
		r.ID,
		r.Name,
		r.Email,
		r.JobTitle,
		r.Manager,
		r.Workgroup,
		r.Clearance,
		pq.Array(r.Specialties),
		pq.Array(r.Certifications),
		r.Active,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&r.Active)
}

func (m *ResourceModel) Get(id int64) (*Resource, error) {
	if id < 1 {
		return nil, ErrNotFound
	}

	query := `
		SELECT resources.id, resources.name, resources.email, resources.job_title, m.name AS manager, workgroups.name, resources.clearance, resources.specialties, resources.certifications, resources.active
		FROM ((resources
			INNER JOIN resources m ON resources.manager_id=m.id)
			INNER JOIN workgroups ON workgroups.id=resources.workgroup_id)
		WHERE resources.id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var r Resource

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&r.ID,
		&r.Name,
		&r.Email,
		&r.JobTitle,
		&r.Manager,
		&r.Workgroup,
		&r.Clearance,
		pq.Array(&r.Specialties),
		pq.Array(&r.Certifications),
		&r.Active,
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

func (m *ResourceModel) Update(r *Resource) error {
	query := `
		UPDATE resources
		SET name=$1, email=$2, job_title=$3, 
		    manager_id=(SELECT m.id FROM resources m WHERE m.name=$4), 
			workgroup_id=(SELECT id FROM workgroups WHERE name=$5), 
			clearance=$6, specialties=$7, certifications=$8, active=$9
		WHERE id=$10`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []interface{}{
		r.Name,
		r.Email,
		r.JobTitle,
		r.Manager,
		r.Workgroup,
		r.Clearance,
		pq.Array(r.Specialties),
		pq.Array(r.Certifications),
		r.Active,
		r.ID,
	}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan()
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

func (m *ResourceModel) GetAll(name string, workgroups []string, clearance string, specialties []string,
	certifications []string, manager string, active bool, filters Filters) ([]*Resource, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), r.id, r.name, r.email, r.job_title, m.name AS manager, workgroups.name, r.clearance, r.specialties, r.certifications, r.active
		FROM ((resources r
			INNER JOIN resources m ON r.manager_id = m.id)
			INNER JOIN workgroups ON workgroups.id = r.workgroup_id)
		WHERE (workgroups.name = ANY ($1) OR $1 = '{}')
		AND (r.clearance = $2 OR $2 = '')
		AND (r.specialties @> $3 OR $3 = '{}')
		AND (r.certifications @> $4 OR $4 = '{}')
		AND (m.name = $5 OR $5 = '')
		AND (r.active = $6)
		AND (r.name = $7 OR $7 = '')
		ORDER BY %s %s, id ASC
		LIMIT $8 OFFSET $9`, fmt.Sprintf("r.%s", filters.sortColumn()), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []interface{}{pq.Array(workgroups), clearance, pq.Array(specialties), pq.Array(certifications), manager, active, name, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	resources := []*Resource{}

	for rows.Next() {
		var resource Resource
		err := rows.Scan(
			&totalRecords,
			&resource.ID,
			&resource.Name,
			&resource.Email,
			&resource.JobTitle,
			&resource.Manager,
			&resource.Workgroup,
			&resource.Clearance,
			pq.Array(&resource.Specialties),
			pq.Array(&resource.Certifications),
			&resource.Active,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		resources = append(resources, &resource)
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return resources, metadata, nil
}
