package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vmw-pso/back-end/internal/validator"
)

type Project struct {
	OpportunityID  string `json:"opportunityId"`
	ChangepointID  string `json:"changepointId,omitempty"`
	RevenueType    string `json:"revenueType"`
	Name           string `json:"name"`
	Customer       string `json:"customer"`
	EndCustomer    string `json:"endCustomer,omitempty"`
	ProjectManager string `json:"projectManager"`
	Status         string `json:"status"`
}

func ValidateRevenueType(v *validator.Validator, revenueType string) {
	revenueTypes := []string{
		"Fixed Fee",
		"T&M",
	}
	v.Check(validator.PermittedValue(revenueType, revenueTypes...), "revenueType", "is not a recognised revenue type [T&M, FF]")
}

func ValidateProjectName(v *validator.Validator, name string) {
	v.Check(name != "", "name", "must be provided")
	v.Check(len(name) <= 256, "name", "cannot be more than 256 bytes")
}

func ValidateCustomer(v *validator.Validator, customer string) {
	v.Check(customer != "", "customer", "must be provided")
	v.Check(len(customer) < 256, "customer", "cannot be more than 256 bytes")
}

func ValidateProjectManager(v *validator.Validator, projectManager string) {
	// TODO: Change this to pull PMs from database into cache on start
	projectManagers := []string{
		"Kim Slocum",
		"Nisha Halim",
	}
	v.Check(validator.PermittedValue(projectManager, projectManagers...), "projectManager", "is not a Project Manager")
}

func ValidateStatus(v *validator.Validator, status string) {
	// TODO: Change this to pull statuses from database into cache at start
	statuses := []string{
		"Staged",
		"At Risk",
		"Work in progress",
		"Inactive",
		"Complete",
	}
	v.Check(validator.PermittedValue(status, statuses...), "status", "is not a recognised status")
}

func ValidateProject(v *validator.Validator, project Project) {
	ValidateRevenueType(v, project.RevenueType)
	ValidateName(v, project.Name)
	ValidateCustomer(v, project.Customer)
	ValidateProjectManager(v, project.ProjectManager)
	ValidateStatus(v, project.Status)
}

type ProjectModel struct {
	DB *sql.DB
}

func (m *ProjectModel) Insert(p *Project) error {
	query := `
		INSERT INTO project
		(opportunity_id, changepoint_id, revenue_type, name, customer, end_customer, project_manager_id, status_id)
		VALUES ($1, $2, $3, $4, $5, $6
			   (SELECT employee_id FROM resource WHERE name=$7),
			   (SELECT status_id FROM project_status WHERE status=$8))`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []interface{}{
		p.OpportunityID,
		p.ChangepointID,
		p.RevenueType,
		p.Name,
		p.Customer,
		p.EndCustomer,
		p.ProjectManager,
		p.Status,
	}

	return m.DB.QueryRowContext(ctx, query, args...).Scan()
}

func (m *ProjectModel) Get(id string) (*Project, error) {
	if id == "" {
		return nil, ErrNotFound
	}

	query := `
		SELECT changepoint_id, revenue_type, name, customer, end_customer, resource.name, project_status.status
		FROM((project p
			(INNER JOIN resource ON resource.employee_id=p.project_manager_id)
			(INNER JOIN project_status ON p.status_id=project_status.status_id))
		WHERE opportunity_id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p Project
	p.OpportunityID = id

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&p.ChangepointID,
		&p.RevenueType,
		&p.Name,
		&p.Customer,
		&p.EndCustomer,
		&p.ProjectManager,
		&p.Status,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &p, nil
}

func (m *ProjectModel) Update(p *Project) error {
	query := `
		UPDATE project
		SET changepoint_id=$1, revenue_type=$2, name=$3, customer=$4, end_customer=$5,
		    project_manager_id=(SELECT id FROM resources WHERE resources.name=$6),
			status_id=(SELECT status_id FROM project_status WHERE status=$7)
		WHERE opportunity_id=$8`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []interface{}{
		p.ChangepointID,
		p.RevenueType,
		p.Name,
		p.Customer,
		p.EndCustomer,
		p.ProjectManager,
		p.Status,
		p.OpportunityID,
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

func (m *ProjectModel) GetAll(customer, endCustomer, projectManager, status, revenueType string, filters Filters) ([]*Project, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), p.opportunity_id, p.changepoint_id, p.name, p.revenue_type, p.customer, p.end_customer, r.name, ps.status
		FROM ((project p
			INNER JOIN resource r ON r.employee_id=p.project_manager_id)
			INNER JOIN project_status ps ON ps.status_id=p.status_id)
		WHERE (p.customer=$1 OR $1='')
		AND (p.end_customer=$2 OR $2='')
		AND (r.name=$3 OR $3='')
		AND (ps.status=$4 or $4='')
		AND (p.revenue_type::text=$5 OR $5='')
		ORDER BY %s %s, opportunity_id ASC
		LIMIT $6 OFFSET $7`, fmt.Sprintf("p.%s", filters.sortColumn()), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []interface{}{customer, endCustomer, projectManager, status, revenueType, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	projects := []*Project{}

	for rows.Next() {
		var project Project
		err := rows.Scan(
			&totalRecords,
			&project.OpportunityID,
			&project.ChangepointID,
			&project.Name,
			&project.RevenueType,
			&project.Customer,
			&project.EndCustomer,
			&project.ProjectManager,
			&project.Status,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		projects = append(projects, &project)
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return projects, metadata, nil
}
