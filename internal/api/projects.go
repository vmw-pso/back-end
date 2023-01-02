package api

import (
	"errors"
	"net/http"

	"github.com/vmw-pso/back-end/internal/data"
	"github.com/vmw-pso/back-end/internal/validator"
)

func (api *API) handleCreateProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			OpportunityID  string `json:"opportunityId"`
			ChangepointID  string `json:"changepointId"`
			RevenueType    string `json:"revenueType"`
			Name           string `json:"name"`
			Customer       string `json:"customer"`
			EndCustomer    string `json:"endCustomer"`
			ProjectManager string `json:"projectManager"`
			Status         string `json:"status"`
		}

		err := api.readJSON(w, r, &input)
		if err != nil {
			api.badRequestResponse(w, r, err)
			return
		}

		project := data.Project{
			OpportunityID:  input.OpportunityID,
			ChangepointID:  input.ChangepointID,
			RevenueType:    input.RevenueType,
			Name:           input.Name,
			Customer:       input.Customer,
			EndCustomer:    input.EndCustomer,
			ProjectManager: input.ProjectManager,
			Status:         input.Status,
		}

		v := validator.New()

		if data.ValidateProject(v, project); !v.Valid() {
			api.failedValidationResponse(w, r, v.Errors)
			return
		}

		err = api.models.Projects.Insert(&project)
		if err != nil {
			api.serverErrorResponse(w, r, err)
			return
		}

		err = api.writeJSON(w, http.StatusCreated, envelope{"project": project}, nil)
		if err != nil {
			api.serverErrorResponse(w, r, err)
		}
	}
}

func (api *API) handleUpdateProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := api.readIDStringParam(r)

		project, err := api.models.Projects.Get(id)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrNotFound):
				api.notFoundResponse(w, r)
			default:
				api.serverErrorResponse(w, r, err)
			}
			return
		}

		var input struct {
			ChangepointID  *string `json:"changepointId"`
			RevenueType    *string `json:"revenueType"`
			Name           *string `json:"name"`
			Customer       *string `json:"customer"`
			EndCustomer    *string `json:"endCustomer"`
			ProjectManager *string `json:"projectManager"`
			Status         *string `json:"status"`
		}

		err = api.readJSON(w, r, &input)
		if err != nil {
			api.badRequestResponse(w, r, err)
			return
		}

		if input.ChangepointID != nil {
			project.ChangepointID = *input.ChangepointID
		}

		if input.RevenueType != nil {
			project.RevenueType = *input.RevenueType
		}

		if input.Name != nil {
			project.Name = *input.Name
		}

		if input.Customer != nil {
			project.Customer = *input.Customer
		}

		if input.EndCustomer != nil {
			project.EndCustomer = *input.EndCustomer
		}

		if input.ProjectManager != nil {
			project.ProjectManager = *input.ProjectManager
		}

		if input.Status != nil {
			project.Status = *input.Status
		}

		v := validator.New()

		if data.ValidateProject(v, *project); !v.Valid() {
			api.failedValidationResponse(w, r, v.Errors)
			return
		}

		err = api.models.Projects.Update(project)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrEditConflict):
				api.editConflictResponse(w, r)
			default:
				api.serverErrorResponse(w, r, err)
			}
			return
		}

		err = api.writeJSON(w, http.StatusOK, envelope{"project": project}, nil)
		if err != nil {
			api.serverErrorResponse(w, r, err)
		}
	}
}

func (api *API) handleShowProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := api.readIDStringParam(r)

		project, err := api.models.Projects.Get(id)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrNotFound):
				api.notFoundResponse(w, r)
			default:
				api.serverErrorResponse(w, r, err)
			}
			return
		}

		err = api.writeJSON(w, http.StatusCreated, envelope{"project": project}, nil)
		if err != nil {
			api.serverErrorResponse(w, r, err)
		}
	}
}

func (api *API) handleListProjects() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Customer       string
			EndCustomer    string
			ProjectManager string
			Status         string
			RevenueType    string
			data.Filters
		}

		v := validator.New()

		qs := r.URL.Query()

		input.Customer = api.readString(qs, "customer", "")
		input.EndCustomer = api.readString(qs, "endCustomer", "")
		input.ProjectManager = api.readString(qs, "projectManager", "")
		input.Status = api.readString(qs, "status", "")
		input.RevenueType = api.readString(qs, "revenueType", "")
		input.Filters.Page = api.readInt(qs, "page", 1, v)
		input.Filters.PageSize = api.readInt(qs, "pageSize", 20, v)
		input.Filters.Sort = api.readString(qs, "sort", "opportunity_id")
		input.Filters.SortSafelist = []string{"opportunity_id", "customer", "end_customer", "project_manager", "-opportunity_id", "-customer", "-end_customer", "-project_manager"}

		if data.ValidateFilters(v, input.Filters); !v.Valid() {
			api.failedValidationResponse(w, r, v.Errors)
			return
		}

		projects, metadata, err := api.models.Projects.GetAll(input.Customer, input.EndCustomer,
			input.ProjectManager, input.Status, input.RevenueType, input.Filters)
		if err != nil {
			api.serverErrorResponse(w, r, err)
			return
		}

		err = api.writeJSON(w, http.StatusOK, envelope{"projects": projects, "metadata": metadata}, nil)
		if err != nil {
			api.serverErrorResponse(w, r, err)
		}
	}
}
