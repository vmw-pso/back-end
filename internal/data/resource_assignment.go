package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/vmw-pso/back-end/internal/validator"
)

type ResourceAssignment struct {
	ID           int64     `json:"id"`
	RequestID    int64     `json:"requestId"`
	Resource     string    `json:"resource"`
	StartDate    time.Time `json:"startDate"`
	EndDate      time.Time `json:"endDate"`
	HoursPerWeek float64   `json:"hours_er_week"`
}

func ValidateHourPerWeek(v *validator.Validator, hoursPerWeek float64) {
	v.Check(hoursPerWeek <= 40, "hoursPerWeek", "must be no more than 40")
}

func ValidateEndData(v *validator.Validator, a ResourceAssignment, budgetHours float64) {
	totalHours := float64(workDays(a.StartDate, a.EndDate)) * a.HoursPerWeek
	v.Check(budgetHours >= totalHours, "endDate", "is beyond the budgeted hours")
	v.Check(a.EndDate.After(a.StartDate), "endDate", "must be after startDate")
}

func ValidateResourceAssignment(v *validator.Validator, a ResourceAssignment, checkValues bool, startDate time.Time, budgetHours float64) {
	ValidateStartDate(v, startDate, a.StartDate)
	if checkValues {
		ValidateHourPerWeek(v, a.HoursPerWeek)
		ValidateEndData(v, a, budgetHours)
	}
}

func workDays(startDate, endDate time.Time) int {
	wd := 0
	for {
		if startDate.Equal(endDate) {
			return wd
		}
		if startDate.Weekday() != time.Saturday && startDate.Weekday() != time.Sunday {
			wd++
		}
		startDate.Add(24 * time.Hour)
	}
}

type ResourceAssignmentModel struct {
	DB *sql.DB
}

func (m *ResourceAssignmentModel) Insert(a *ResourceAssignment) error {
	query := `
		INSERT INTO resource_assignment
		(resource_request_id, employee_id, start_date, end_date, hours_per_week)
		VALUES ($1,
			(SELECT employee_id FROM resource WHERE resource.name=$2),
			$3, $4, $5) RETURNING assignment_id`

	args := []interface{}{
		a.RequestID,
		a.Resource,
		a.StartDate,
		a.EndDate,
		a.HoursPerWeek,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&a.ID)
}

func (m *ResourceAssignmentModel) GetForRequest(reqID int64) ([]*ResourceAssignment, error) {
	query := `
		SELECT assignment_id, r.name, start_date, end_date, hours_per_week
		FROM(resource_assignment a
			INNER JOIN resource r ON a.employee_id=r.employee_id)
		WHERE resource_request_id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, reqID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assignments := []*ResourceAssignment{}

	for rows.Next() {
		var assignment ResourceAssignment
		assignment.RequestID = reqID
		err := rows.Scan(
			&assignment.ID,
			&assignment.Resource,
			&assignment.StartDate,
			&assignment.EndDate,
			&assignment.HoursPerWeek,
		)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, &assignment)
	}
	return assignments, nil
}
