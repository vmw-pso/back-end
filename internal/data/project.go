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
	ID             int64  `json:"id"`
	OpportunityID  string `json:"opportunityId,omitempty"`
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
		INSERT INTO projects
		(opportunity_id, changepoint_id, revenue_type, name, customer, end_customer, project_manager_id, status)
		VALUES ($1, $2, $3, $4, $5, $6
			   (SELECT id FROM resources WHERE name=$7),
			   $8)
		RETURNING id`

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

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&p.ID)
}

func (m *ProjectModel) Get(id int64) (*Project, error) {
	if id < 1 {
		return nil, ErrNotFound
	}

	query := `
		SELECT opportunity_id, changepoint_id, revenue_type, name, customer, end_customer, resources.name, status
		FROM(projects
			INNER JOIN resources ON resources.id=projects.project_manager_id)
		WHERE projects.id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p Project
	p.ID = id

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&p.OpportunityID,
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
		UPDATE projects
		SET opportunity_id=$1, changepoint_id=$2, revenue_type=$3, name=$4, customer=$5, end_customer=$6,
		    project_manager_id=(SELECT id FROM resources WHERE resources.name=$7),
			status=$8
		WHERE projects.id=$9`

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
		p.ID,
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

func (m *ProjectModel) GetAll(customer, endCustomer, projectManager, status string, filters Filters) ([]*Project, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), p.id, p.opportunity_id, p.changepoint_id, p.revenue_type, p.name, p.customer, p.end_customer, r.name as projectManager, p.status
		FROM (projects p
			INNER JOIN resources r ON p.project_manager_id=r.id)
		WHERE (p.customer=$1 OR $1='')
		AND (p.end_customer=$2 OR $2='')
		AND (r.name=$3 OR $3='')
		AND (p.status=$4 or $4='')
		ORDER BY %s %s, id ASC
		LIMIT $5 OFFSET $6`, fmt.Sprintf("p.%s", filters.sortColumn()), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []interface{}{customer, endCustomer, projectManager, status, filters.limit(), filters.offset()}

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
			&project.ID,
			&project.OpportunityID,
			&project.ChangepointID,
			&project.RevenueType,
			&project.Name,
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
